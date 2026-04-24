package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/middleware"
	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

// ScoreHandler holds HTTP handlers for score endpoints.
type ScoreHandler struct {
	scoreService *service.ScoreService
}

func NewScoreHandler(scoreService *service.ScoreService) *ScoreHandler {
	return &ScoreHandler{scoreService: scoreService}
}

// HandleCreateScore handles POST /api/scores
func (h *ScoreHandler) HandleCreateScore(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	ipAddress := r.RemoteAddr
	resp, err := h.scoreService.CreateScore(r.Context(), claims.UserID, &req, ipAddress)
	if err != nil {
		msg := err.Error()
		if isValidationError(err) || isBusinessError(msg) {
			utils.ErrorJSON(w, http.StatusBadRequest, msg)
			return
		}
		log.Printf("ERROR: CreateScore failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to create score")
		return
	}

	utils.JSON(w, http.StatusCreated, resp)
}

// HandleUpdateScore handles PUT /api/scores/{id}
func (h *ScoreHandler) HandleUpdateScore(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	scoreID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid score ID")
		return
	}

	var req models.CreateScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	resp, err := h.scoreService.UpdateScore(r.Context(), claims.UserID, scoreID, &req, r.RemoteAddr)
	if err != nil {
		msg := err.Error()
		if isValidationError(err) || isBusinessError(msg) || err.Error() == "you can only edit your own scores" || err.Error() == "you have already logged a score for "+req.RoundDate {
			utils.ErrorJSON(w, http.StatusBadRequest, msg)
			return
		}
		log.Printf("ERROR: UpdateScore failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to update score")
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

// HandleGetMyScores handles GET /api/scores
func (h *ScoreHandler) HandleGetMyScores(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	scores, err := h.scoreService.GetUserScores(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("ERROR: GetMyScores failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch scores")
		return
	}

	if scores == nil {
		scores = []models.ScoreResponse{}
	}

	utils.JSON(w, http.StatusOK, scores)
}

// HandleDeleteScore handles DELETE /api/scores/{id}
func (h *ScoreHandler) HandleDeleteScore(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	scoreID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid score ID")
		return
	}

	if err := h.scoreService.DeleteScore(r.Context(), claims.UserID, scoreID, r.RemoteAddr); err != nil {
		if errors.Is(err, repository.ErrScoreNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, "Score not found")
			return
		}
		msg := err.Error()
		if msg == "you can only delete your own scores" {
			utils.ErrorJSON(w, http.StatusForbidden, msg)
			return
		}
		log.Printf("ERROR: DeleteScore failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to delete score")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Score deleted successfully"})
}

// HandleAdminListScores handles GET /api/admin/scores
func (h *ScoreHandler) HandleAdminListScores(w http.ResponseWriter, r *http.Request) {
	page := parseIntParam(r, "page", 1)
	pageSize := parseIntParam(r, "size", 20)
	search := r.URL.Query().Get("search")

	result, err := h.scoreService.ListAllScores(r.Context(), page, pageSize, search)
	if err != nil {
		log.Printf("ERROR: AdminListScores failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch scores")
		return
	}

	utils.JSON(w, http.StatusOK, result)
}

// HandleAdminDeleteScore handles DELETE /api/admin/scores/{id}
func (h *ScoreHandler) HandleAdminDeleteScore(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	scoreID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid score ID")
		return
	}

	if err := h.scoreService.AdminDeleteScore(r.Context(), claims.UserID, scoreID, r.RemoteAddr); err != nil {
		if errors.Is(err, repository.ErrScoreNotFound) {
			utils.ErrorJSON(w, http.StatusNotFound, "Score not found")
			return
		}
		log.Printf("ERROR: AdminDeleteScore failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to delete score")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Score deleted by admin"})
}

// isBusinessError checks for known business-logic error messages.
func isBusinessError(msg string) bool {
	businessPhrases := []string{"maximum of", "invalid score", "invalid round_date", "pool amount", "no eligible"}
	for _, phrase := range businessPhrases {
		if len(msg) >= len(phrase) && containsStr(msg, phrase) {
			return true
		}
	}
	return false
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
