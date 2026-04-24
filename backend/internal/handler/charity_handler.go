package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/middleware"
	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

// CharityHandler holds HTTP handlers for charity endpoints.
type CharityHandler struct {
	charityService *service.CharityService
}

func NewCharityHandler(charityService *service.CharityService) *CharityHandler {
	return &CharityHandler{charityService: charityService}
}

// HandleListCharities handles GET /api/charities
func (h *CharityHandler) HandleListCharities(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Admin sees all, users see only active
	activeOnly := claims.Role != "admin"

	charities, err := h.charityService.ListCharities(r.Context(), activeOnly)
	if err != nil {
		log.Printf("ERROR: ListCharities failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch charities")
		return
	}

	if charities == nil {
		charities = []models.CharityResponse{}
	}

	utils.JSON(w, http.StatusOK, charities)
}

// HandleSelectCharity handles POST /api/charity/select
func (h *CharityHandler) HandleSelectCharity(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.SelectCharityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	resp, err := h.charityService.SelectCharity(r.Context(), claims.UserID, &req, r.RemoteAddr)
	if err != nil {
		msg := err.Error()
		if isValidationError(err) || isBusinessError(msg) {
			utils.ErrorJSON(w, http.StatusBadRequest, msg)
			return
		}
		log.Printf("ERROR: SelectCharity failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to select charity")
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

// HandleGetMySelection handles GET /api/charity/my-selection
func (h *CharityHandler) HandleGetMySelection(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sel, err := h.charityService.GetUserCharitySelection(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("ERROR: GetMySelection failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch selection")
		return
	}

	if sel == nil {
		utils.JSON(w, http.StatusOK, nil)
		return
	}

	utils.JSON(w, http.StatusOK, sel)
}

// HandleAdminCreateCharity handles POST /api/admin/charities
func (h *CharityHandler) HandleAdminCreateCharity(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateCharityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	resp, err := h.charityService.CreateCharity(r.Context(), claims.UserID, &req, r.RemoteAddr)
	if err != nil {
		if isValidationError(err) {
			utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("ERROR: CreateCharity failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to create charity")
		return
	}

	utils.JSON(w, http.StatusCreated, resp)
}

// HandleAdminUpdateCharity handles PUT /api/admin/charities/{id}
func (h *CharityHandler) HandleAdminUpdateCharity(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	charityID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid charity ID")
		return
	}

	var req models.UpdateCharityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	resp, err := h.charityService.UpdateCharity(r.Context(), claims.UserID, charityID, &req, r.RemoteAddr)
	if err != nil {
		if isValidationError(err) {
			utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("ERROR: UpdateCharity failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to update charity")
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

// HandleAdminToggleCharity handles PUT /api/admin/charities/{id}/toggle
func (h *CharityHandler) HandleAdminToggleCharity(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	charityID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid charity ID")
		return
	}

	var body struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	if err := h.charityService.ToggleCharityActive(r.Context(), claims.UserID, charityID, body.Active, r.RemoteAddr); err != nil {
		log.Printf("ERROR: ToggleCharity failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to toggle charity")
		return
	}

	status := "activated"
	if !body.Active {
		status = "deactivated"
	}
	utils.JSON(w, http.StatusOK, map[string]string{"message": "Charity " + status + " successfully"})
}

// parseIntParam parses an integer query parameter with a default value.
func parseIntParam(r *http.Request, key string, defaultVal int) int {
	val, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || val < 1 {
		return defaultVal
	}
	return val
}
