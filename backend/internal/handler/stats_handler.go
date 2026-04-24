package handler

import (
	"log"
	"net/http"

	"golf-score-lottery/backend/internal/middleware"
	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

// StatsHandler holds HTTP handlers for dashboard stats endpoints.
type StatsHandler struct {
	statsService *service.StatsService
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// HandleGetUserDashboardStats handles GET /api/stats/dashboard
func (h *StatsHandler) HandleGetUserDashboardStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.statsService.GetUserDashboardStats(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("ERROR: GetUserDashboardStats failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch stats")
		return
	}

	utils.JSON(w, http.StatusOK, stats)
}

// HandleGetAdminDashboardStats handles GET /api/admin/stats/dashboard
func (h *StatsHandler) HandleGetAdminDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsService.GetAdminDashboardStats(r.Context())
	if err != nil {
		log.Printf("ERROR: GetAdminDashboardStats failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch admin stats")
		return
	}

	utils.JSON(w, http.StatusOK, stats)
}

// HandleGetScoreTrend handles GET /api/stats/score-trend
func (h *StatsHandler) HandleGetScoreTrend(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	data, err := h.statsService.GetScoreTrend(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("ERROR: GetScoreTrend failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch score trend")
		return
	}

	utils.JSON(w, http.StatusOK, data)
}

// HandleGetUserGrowth handles GET /api/admin/stats/user-growth
func (h *StatsHandler) HandleGetUserGrowth(w http.ResponseWriter, r *http.Request) {
	data, err := h.statsService.GetUserGrowthData(r.Context())
	if err != nil {
		log.Printf("ERROR: GetUserGrowth failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch user growth data")
		return
	}

	utils.JSON(w, http.StatusOK, data)
}

// HandleGetRevenue handles GET /api/admin/stats/revenue
func (h *StatsHandler) HandleGetRevenue(w http.ResponseWriter, r *http.Request) {
	data, err := h.statsService.GetRevenueData(r.Context())
	if err != nil {
		log.Printf("ERROR: GetRevenue failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch revenue data")
		return
	}

	utils.JSON(w, http.StatusOK, data)
}

// HandleGetCharityDistribution handles GET /api/stats/charity-distribution
func (h *StatsHandler) HandleGetCharityDistribution(w http.ResponseWriter, r *http.Request) {
	data, err := h.statsService.GetCharityDistribution(r.Context())
	if err != nil {
		log.Printf("ERROR: GetCharityDistribution failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch charity distribution")
		return
	}

	utils.JSON(w, http.StatusOK, data)
}
