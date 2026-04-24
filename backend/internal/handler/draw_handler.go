package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/middleware"
	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

// DrawHandler holds HTTP handlers for draw endpoints.
type DrawHandler struct {
	drawService *service.DrawService
}

func NewDrawHandler(drawService *service.DrawService) *DrawHandler {
	return &DrawHandler{drawService: drawService}
}

// HandleRunDraw handles POST /api/admin/draws/run
func (h *DrawHandler) HandleRunDraw(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.RunDrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	result, err := h.drawService.RunDraw(r.Context(), claims.UserID, &req, r.RemoteAddr)
	if err != nil {
		msg := err.Error()
		if isBusinessError(msg) {
			utils.ErrorJSON(w, http.StatusBadRequest, msg)
			return
		}
		log.Printf("ERROR: RunDraw failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to execute draw")
		return
	}

	utils.JSON(w, http.StatusCreated, result)
}

// HandleSimulateDraw handles POST /api/admin/draws/simulate
func (h *DrawHandler) HandleSimulateDraw(w http.ResponseWriter, r *http.Request) {
	var req models.RunDrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	result, err := h.drawService.SimulateDraw(r.Context(), &req)
	if err != nil {
		msg := err.Error()
		if isBusinessError(msg) {
			utils.ErrorJSON(w, http.StatusBadRequest, msg)
			return
		}
		log.Printf("ERROR: SimulateDraw failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to simulate draw")
		return
	}

	utils.JSON(w, http.StatusOK, result)
}

// HandleListDraws handles GET /api/admin/draws
func (h *DrawHandler) HandleListDraws(w http.ResponseWriter, r *http.Request) {
	page := parseIntParam(r, "page", 1)
	pageSize := parseIntParam(r, "size", 10)

	result, err := h.drawService.ListDraws(r.Context(), page, pageSize)
	if err != nil {
		log.Printf("ERROR: ListDraws failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch draws")
		return
	}

	utils.JSON(w, http.StatusOK, result)
}

// HandleGetDrawDetail handles GET /api/admin/draws/{id}
func (h *DrawHandler) HandleGetDrawDetail(w http.ResponseWriter, r *http.Request) {
	drawID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid draw ID")
		return
	}

	result, err := h.drawService.GetDrawDetail(r.Context(), drawID)
	if err != nil {
		log.Printf("ERROR: GetDrawDetail failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch draw detail")
		return
	}

	utils.JSON(w, http.StatusOK, result)
}
