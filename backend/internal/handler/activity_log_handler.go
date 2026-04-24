package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

// ActivityLogHandler holds HTTP handlers for activity log endpoints.
type ActivityLogHandler struct {
	activityLogService *service.ActivityLogService
}

func NewActivityLogHandler(activityLogService *service.ActivityLogService) *ActivityLogHandler {
	return &ActivityLogHandler{activityLogService: activityLogService}
}

// HandleGetActivityLogs handles GET /api/admin/activity-logs
func (h *ActivityLogHandler) HandleGetActivityLogs(w http.ResponseWriter, r *http.Request) {
	page := parseIntParam(r, "page", 1)
	pageSize := parseIntParam(r, "size", 20)
	userIDFilter := r.URL.Query().Get("user_id")
	actionFilter := r.URL.Query().Get("action")

	result, err := h.activityLogService.GetActivityLogs(r.Context(), page, pageSize, userIDFilter, actionFilter)
	if err != nil {
		log.Printf("ERROR: GetActivityLogs failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch activity logs")
		return
	}

	utils.JSON(w, http.StatusOK, result)
}

// HandleGetUsersWithActivity handles GET /api/admin/activity-logs/users
func (h *ActivityLogHandler) HandleGetUsersWithActivity(w http.ResponseWriter, r *http.Request) {
	users, err := h.activityLogService.GetUsersWithActivity(r.Context())
	if err != nil {
		log.Printf("ERROR: GetUsersWithActivity failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch users with activity")
		return
	}
	utils.JSON(w, http.StatusOK, users)
}

// HandleGetUserActivityYears handles GET /api/admin/activity-logs/users/{id}/years
func (h *ActivityLogHandler) HandleGetUserActivityYears(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	years, err := h.activityLogService.GetUserActivityYears(r.Context(), userID)
	if err != nil {
		log.Printf("ERROR: GetUserActivityYears failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch activity years")
		return
	}

	utils.JSON(w, http.StatusOK, years)
}

// HandleGetUserActivityMonths handles GET /api/admin/activity-logs/users/{id}/years/{year}/months
func (h *ActivityLogHandler) HandleGetUserActivityMonths(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid year")
		return
	}

	months, err := h.activityLogService.GetUserActivityMonths(r.Context(), userID, year)
	if err != nil {
		log.Printf("ERROR: GetUserActivityMonths failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch activity months")
		return
	}

	utils.JSON(w, http.StatusOK, months)
}

// HandleGetUserMonthlyActivities handles GET /api/admin/activity-logs/users/{id}/years/{year}/months/{month}
func (h *ActivityLogHandler) HandleGetUserMonthlyActivities(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid year")
		return
	}

	month, err := strconv.Atoi(chi.URLParam(r, "month"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid month")
		return
	}

	logs, err := h.activityLogService.GetUserMonthlyActivities(r.Context(), userID, year, month)
	if err != nil {
		log.Printf("ERROR: GetUserMonthlyActivities failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch monthly activities")
		return
	}

	utils.JSON(w, http.StatusOK, logs)
}
