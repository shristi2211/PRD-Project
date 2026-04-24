package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"golf-score-lottery/backend/internal/config"
	"golf-score-lottery/backend/internal/database"
	"golf-score-lottery/backend/internal/handler"
	"golf-score-lottery/backend/internal/middleware"
	"golf-score-lottery/backend/internal/repository"
	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

func main() {
	// ─── Load Configuration ───────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("✅ Configuration loaded")

	// ─── Initialize Context ───────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ─── Initialize PostgreSQL ────────────────────────────────────
	pgPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgPool.Close()
	log.Println("✅ PostgreSQL connected (pool ready)")

	// ─── Initialize Redis (optional) ──────────────────────────────
	var redisClient *redis.Client
	redisClient, err = database.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Printf("⚠️  Redis unavailable (%v) — running without Redis (using PostgreSQL only)", err)
		redisClient = nil
	} else {
		defer redisClient.Close()
		log.Println("✅ Redis connected")
	}

	// ─── Initialize RSA Key Manager ──────────────────────────────
	keyManager, err := utils.NewKeyManager(cfg.PrivateKeyPath, cfg.PublicKeyPath)
	if err != nil {
		log.Fatalf("Failed to initialize RSA keys: %v", err)
	}
	log.Println("✅ RSA key pair loaded")

	// ─── Wire Dependencies ────────────────────────────────────────
	userRepo := repository.NewUserRepository(pgPool)
	authService := service.NewAuthService(userRepo, keyManager, redisClient, cfg.AccessTokenExpiry, cfg.RefreshTokenExpiry)
	authHandler := handler.NewAuthHandler(authService)

	adminRepo := repository.NewAdminRepository(pgPool)
	adminService := service.NewAdminService(adminRepo)
	adminHandler := handler.NewAdminHandler(adminService)

	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Phase 4 wiring
	activityLogRepo := repository.NewActivityLogRepository(pgPool)
	activityLogService := service.NewActivityLogService(activityLogRepo)
	activityLogHandler := handler.NewActivityLogHandler(activityLogService)

	scoreRepo := repository.NewScoreRepository(pgPool)
	scoreService := service.NewScoreService(scoreRepo, activityLogService)
	scoreHandler := handler.NewScoreHandler(scoreService)

	charityRepo := repository.NewCharityRepository(pgPool)
	charityService := service.NewCharityService(charityRepo, activityLogService)
	charityHandler := handler.NewCharityHandler(charityService)

	winnerRepo := repository.NewWinnerRepository(pgPool)
	drawRepo := repository.NewDrawRepository(pgPool)
	drawService := service.NewDrawService(drawRepo, scoreRepo, winnerRepo, activityLogService)
	drawHandler := handler.NewDrawHandler(drawService)

	winnerService := service.NewWinnerService(winnerRepo, activityLogService)
	winnerHandler := handler.NewWinnerHandler(winnerService)

	statsRepo := repository.NewStatsRepository(pgPool)
	statsService := service.NewStatsService(statsRepo, charityRepo)
	statsHandler := handler.NewStatsHandler(statsService)

	// Phase 5/6 wiring — Reports
	reportRepo := repository.NewReportRepository(pgPool)
	reportService := service.NewReportService(reportRepo)
	reportHandler := handler.NewReportHandler(reportService)

	// ─── Setup Router ─────────────────────────────────────────────
	r := chi.NewRouter()

	// Global middleware (order matters)
	r.Use(middleware.Logger)
	r.Use(middleware.CORSMiddleware(cfg.CORSOrigins))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.JSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})

	// Public auth routes (with optional rate limiting)
	r.Route("/api/auth", func(r chi.Router) {
		if redisClient != nil {
			authRateLimiter := middleware.NewRateLimiter(redisClient, 30, 1*time.Minute)
			r.Use(authRateLimiter.Middleware())
		}
		r.Get("/setup-status", adminHandler.HandleGetSetupStatus)
		r.Post("/setup", adminHandler.HandleSetupAdmin)
		
		r.Post("/register", authHandler.HandleRegister)
		r.Post("/login", authHandler.HandleLogin)
		r.Post("/refresh", authHandler.HandleRefresh)
		r.Post("/logout", authHandler.HandleLogout)
		r.Post("/subscribe", userHandler.HandlePublicSubscribe) // Public: activate subscription before login
	})

	// Truly public, no-auth, IP-based routes
	r.Route("/api/public", func(r chi.Router) {
		r.Post("/ip-subscribe", userHandler.HandleIpSubscribe)
		r.Get("/ip-status", userHandler.HandleIpStatus)
	})

	// Protected routes (require valid JWT)
	r.Route("/api/users", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(keyManager, authService))
		r.Get("/me", authHandler.HandleGetMe)
		r.Put("/me", userHandler.HandleUpdateProfile)
		r.Put("/me/password", userHandler.HandleChangePassword)
		r.Delete("/me", userHandler.HandleDeleteAccount)
	})

	r.Route("/api/subscriptions", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(keyManager, authService))
		r.Post("/start", userHandler.HandleStartSubscription)
		r.Put("/cancel", userHandler.HandleCancelSubscription)
	})

	r.Route("/api/scores", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(keyManager, authService))
		r.Post("/", scoreHandler.HandleCreateScore)
		r.Get("/", scoreHandler.HandleGetMyScores)
		r.Put("/{id}", scoreHandler.HandleUpdateScore)
		r.Delete("/{id}", scoreHandler.HandleDeleteScore)
	})

	r.Route("/api/charities", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(keyManager, authService))
		r.Get("/", charityHandler.HandleListCharities)
	})

	r.Route("/api/charity", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(keyManager, authService))
		r.Post("/select", charityHandler.HandleSelectCharity)
		r.Get("/my-selection", charityHandler.HandleGetMySelection)
	})

	r.Route("/api/winners", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(keyManager, authService))
		r.Get("/me", winnerHandler.HandleGetMyWinnings)
		r.Put("/{id}/proof", winnerHandler.HandleSubmitProof)
	})

	r.Route("/api/stats", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(keyManager, authService))
		r.Get("/dashboard", statsHandler.HandleGetUserDashboardStats)
		r.Get("/score-trend", statsHandler.HandleGetScoreTrend)
		r.Get("/charity-distribution", statsHandler.HandleGetCharityDistribution)
	})

	// Admin-only routes
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(keyManager, authService))
		r.Use(middleware.RequireRole("admin"))
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			utils.JSON(w, http.StatusOK, map[string]string{"message": "Welcome, admin!"})
		})
		
		r.Get("/users", userHandler.HandleListUsers)
		r.Put("/users/{id}/activation", userHandler.HandleToggleActivation)

		r.Get("/subscriptions", userHandler.HandleListUsers) // Reuses user list — includes subscription_type

		r.Get("/scores", scoreHandler.HandleAdminListScores)
		r.Delete("/scores/{id}", scoreHandler.HandleAdminDeleteScore)

		r.Post("/charities", charityHandler.HandleAdminCreateCharity)
		r.Put("/charities/{id}", charityHandler.HandleAdminUpdateCharity)
		r.Put("/charities/{id}/toggle", charityHandler.HandleAdminToggleCharity)

		r.Post("/draws/run", drawHandler.HandleRunDraw)
		r.Post("/draws/simulate", drawHandler.HandleSimulateDraw)
		r.Get("/draws", drawHandler.HandleListDraws)
		r.Get("/draws/{id}", drawHandler.HandleGetDrawDetail)

		r.Get("/winners/pending", winnerHandler.HandleGetPendingVerifications)
		r.Put("/winners/{id}/verify", winnerHandler.HandleVerifyWinner)

		r.Get("/activity-logs", activityLogHandler.HandleGetActivityLogs)
		r.Get("/activity-logs/users", activityLogHandler.HandleGetUsersWithActivity)
		r.Get("/activity-logs/users/{id}/years", activityLogHandler.HandleGetUserActivityYears)
		r.Get("/activity-logs/users/{id}/years/{year}/months", activityLogHandler.HandleGetUserActivityMonths)
		r.Get("/activity-logs/users/{id}/years/{year}/months/{month}", activityLogHandler.HandleGetUserMonthlyActivities)

		r.Get("/reports/users", reportHandler.HandleUserReport)
		r.Get("/reports/revenue", reportHandler.HandleRevenueReport)
		r.Get("/reports/draws", reportHandler.HandleDrawReport)
		r.Get("/reports/charities", reportHandler.HandleCharityReport)

		r.Get("/stats/dashboard", statsHandler.HandleGetAdminDashboardStats)
		r.Get("/stats/user-growth", statsHandler.HandleGetUserGrowth)
		r.Get("/stats/revenue", statsHandler.HandleGetRevenue)
	})

	// ─── Background: Expired Token Cleanup ────────────────────────
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				count, err := userRepo.CleanupExpiredTokens(ctx)
				if err != nil {
					log.Printf("WARNING: Token cleanup failed: %v", err)
				} else if count > 0 {
					log.Printf("🧹 Cleaned up %d expired tokens", count)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// ─── Start Server ─────────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		log.Printf("⚡ Received signal %v, shutting down...", sig)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("ERROR: Server shutdown error: %v", err)
		}
		cancel() // Cancel the main context to stop background goroutines
	}()

	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("📡 CORS origins: %s", cfg.CORSOrigins)
	log.Printf("🔑 Access token expiry: %v", cfg.AccessTokenExpiry)
	log.Printf("🔄 Refresh token expiry: %v", cfg.RefreshTokenExpiry)
	if redisClient == nil {
		log.Println("⚠️  Rate limiting disabled (no Redis)")
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("👋 Server stopped gracefully")
}
