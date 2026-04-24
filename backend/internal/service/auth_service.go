package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
	"golf-score-lottery/backend/internal/utils"
)

// Business-level errors returned to handlers
var (
	ErrUserExists            = errors.New("a user with this email already exists")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrInvalidToken          = errors.New("invalid or expired token")
	ErrTokenRevoked          = errors.New("token has been revoked")
	ErrSubscriptionRequired  = errors.New("active subscription required to login")
	ErrInvalidPlan           = errors.New("invalid subscription plan")
)

// AuthService encapsulates all authentication business logic.
// Redis is optional — if nil, only PostgreSQL is used.
type AuthService struct {
	userRepo    *repository.UserRepository
	keyManager  *utils.KeyManager
	redisClient *redis.Client // may be nil
	accessExp   time.Duration
	refreshExp  time.Duration
}

// NewAuthService creates a new AuthService with all dependencies.
// redisClient can be nil — the service will work without Redis.
func NewAuthService(
	userRepo *repository.UserRepository,
	keyManager *utils.KeyManager,
	redisClient *redis.Client,
	accessExp time.Duration,
	refreshExp time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		keyManager:  keyManager,
		redisClient: redisClient,
		accessExp:   accessExp,
		refreshExp:  refreshExp,
	}
}

// hasRedis checks if Redis is available.
func (s *AuthService) hasRedis() bool {
	return s.redisClient != nil
}

// Register creates a new user account and applies any pending IP subscriptions.
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest, ip string) (*models.UserResponse, error) {
	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	// Validate input
	if err := utils.ValidateRegisterInput(email, req.Password, name); err != nil {
		return nil, err
	}

	// Hash password (bcrypt cost 12)
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Store in database
	user, err := s.userRepo.CreateUser(ctx, email, hashedPassword, name)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Auto-upgrade if there is an IP subscription
	plan, _ := s.userRepo.GetIpSubscription(ctx, ip)
	if plan != "" {
		s.userRepo.SetSubscriptionType(ctx, user.ID, plan)
		user.SubscriptionType = plan
		user.SubscriptionActive = true
	}

	resp := user.ToResponse()
	return &resp, nil
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, ip string) (*models.LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if err := utils.ValidateLoginInput(email, req.Password); err != nil {
		return nil, err
	}

	// Fetch user by email (indexed query)
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials // Don't reveal whether email exists
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify password (constant-time comparison via bcrypt)
	if err := utils.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Subscription gate: block login for non-admin users without active subscription
	if user.Role != "admin" && !user.SubscriptionActive {
		// Auto upgrade if IP has a pending subscription
		plan, _ := s.userRepo.GetIpSubscription(ctx, ip)
		if plan != "" {
			_ = s.userRepo.SetSubscriptionType(ctx, user.ID, plan)
			user.SubscriptionType = plan
			user.SubscriptionActive = true
		} else {
			return nil, ErrSubscriptionRequired
		}
	}

	// Generate access token (RS256)
	accessToken, err := utils.GenerateAccessToken(user, s.keyManager.PrivateKey(), s.accessExp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (RS256) with unique JTI
	refreshToken, jti, err := utils.GenerateRefreshToken(user.ID, s.keyManager.PrivateKey(), s.refreshExp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash the JTI for storage (we never store raw tokens)
	jtiHash := hashToken(jti)
	familyID := uuid.New()

	// Store refresh token hash in PostgreSQL (audit trail)
	if err := s.userRepo.StoreRefreshToken(ctx, user.ID, jtiHash, familyID, time.Now().Add(s.refreshExp)); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Store in Redis for fast O(1) validation (with TTL auto-expiry) — optional
	if s.hasRedis() {
		redisKey := fmt.Sprintf("refresh:%s", jtiHash)
		if err := s.redisClient.Set(ctx, redisKey, user.ID.String(), s.refreshExp).Err(); err != nil {
			log.Printf("WARNING: Failed to cache refresh token in Redis: %v", err)
			// Non-fatal — PostgreSQL is the source of truth
		}
	}

	return &models.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

// RefreshToken validates a refresh token, rotates it, and returns new tokens.
// Implements refresh token rotation for replay attack detection.
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*models.TokenResponse, error) {
	if refreshTokenStr == "" {
		return nil, ErrInvalidToken
	}

	// Validate JWT signature and expiry
	claims, err := utils.ValidateRefreshToken(refreshTokenStr, s.keyManager.PublicKey())
	if err != nil {
		return nil, ErrInvalidToken
	}

	jtiHash := hashToken(claims.ID)

	// Check Redis first if available (fast path)
	if s.hasRedis() {
		redisKey := fmt.Sprintf("refresh:%s", jtiHash)
		_, err = s.redisClient.Get(ctx, redisKey).Result()
		if err == redis.Nil {
			// Token not in Redis — check PostgreSQL (may have been evicted)
			storedToken, dbErr := s.userRepo.GetRefreshTokenByHash(ctx, jtiHash)
			if dbErr != nil {
				if errors.Is(dbErr, repository.ErrTokenNotFound) {
					return nil, ErrInvalidToken
				}
				return nil, fmt.Errorf("failed to validate refresh token: %w", dbErr)
			}

			// If the token was already revoked, this is a reuse attack!
			if storedToken.Revoked {
				log.Printf("SECURITY: Refresh token reuse detected for user %s, family %s", storedToken.UserID, storedToken.FamilyID)
				_ = s.userRepo.RevokeTokenFamily(ctx, storedToken.FamilyID)
				return nil, ErrTokenRevoked
			}
		} else if err != nil {
			log.Printf("WARNING: Redis lookup failed, falling back to PostgreSQL: %v", err)
		}
	}

	// Verify token is not revoked in PostgreSQL (source of truth)
	storedToken, err := s.userRepo.GetRefreshTokenByHash(ctx, jtiHash)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	if storedToken.Revoked {
		return nil, ErrTokenRevoked
	}

	// Revoke the old refresh token (rotation)
	if err := s.userRepo.RevokeRefreshToken(ctx, jtiHash); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}
	if s.hasRedis() {
		redisKey := fmt.Sprintf("refresh:%s", jtiHash)
		s.redisClient.Del(ctx, redisKey) // Remove from Redis (best-effort)
	}

	// Fetch fresh user data
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Generate new access token
	newAccessToken, err := utils.GenerateAccessToken(user, s.keyManager.PrivateKey(), s.accessExp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	// Generate new refresh token (same family for tracking)
	newRefreshToken, newJTI, err := utils.GenerateRefreshToken(user.ID, s.keyManager.PrivateKey(), s.refreshExp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	newJTIHash := hashToken(newJTI)

	// Store new refresh token
	if err := s.userRepo.StoreRefreshToken(ctx, user.ID, newJTIHash, storedToken.FamilyID, time.Now().Add(s.refreshExp)); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// Cache in Redis (optional)
	if s.hasRedis() {
		newRedisKey := fmt.Sprintf("refresh:%s", newJTIHash)
		if err := s.redisClient.Set(ctx, newRedisKey, user.ID.String(), s.refreshExp).Err(); err != nil {
			log.Printf("WARNING: Failed to cache new refresh token in Redis: %v", err)
		}
	}

	return &models.TokenResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout revokes the specified refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshTokenStr string) error {
	if refreshTokenStr == "" {
		return nil // Idempotent — already logged out
	}

	claims, err := utils.ValidateRefreshToken(refreshTokenStr, s.keyManager.PublicKey())
	if err != nil {
		return nil // Token is already invalid, consider logged out
	}

	jtiHash := hashToken(claims.ID)

	// Revoke in PostgreSQL
	if err := s.userRepo.RevokeRefreshToken(ctx, jtiHash); err != nil {
		log.Printf("WARNING: Failed to revoke refresh token in DB: %v", err)
	}

	// Remove from Redis (optional)
	if s.hasRedis() {
		redisKey := fmt.Sprintf("refresh:%s", jtiHash)
		s.redisClient.Del(ctx, redisKey)
	}

	return nil
}

// BlacklistAccessToken adds a revoked access token JTI to Redis with the remaining TTL.
// No-op if Redis is unavailable.
func (s *AuthService) BlacklistAccessToken(ctx context.Context, jti string, expiresAt time.Time) error {
	if !s.hasRedis() {
		return nil // Can't blacklist without Redis
	}

	remaining := time.Until(expiresAt)
	if remaining <= 0 {
		return nil // Already expired, no need to blacklist
	}

	key := fmt.Sprintf("blacklist:%s", jti)
	return s.redisClient.Set(ctx, key, "1", remaining).Err()
}

// IsAccessTokenBlacklisted checks if an access token JTI has been revoked.
// Always returns false (not blacklisted) if Redis is unavailable.
func (s *AuthService) IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	if !s.hasRedis() {
		return false, nil // No Redis = no blacklist = allow
	}

	key := fmt.Sprintf("blacklist:%s", jti)
	_, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Not blacklisted
	}
	if err != nil {
		return false, err
	}
	return true, nil // Blacklisted
}

// GetUserByID retrieves a user by ID (for the /me endpoint).
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

// hashToken creates a SHA-256 hash of a token string.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
