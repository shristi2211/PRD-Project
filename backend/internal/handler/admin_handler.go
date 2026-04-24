package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// HandleGetSetupStatus handles GET /api/auth/setup-status
func (h *AdminHandler) HandleGetSetupStatus(w http.ResponseWriter, r *http.Request) {
	complete, err := h.adminService.IsSetupComplete(r.Context())
	if err != nil {
		log.Printf("ERROR: Failed to check admin setup status: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to check setup status")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]bool{
		"setup_complete": complete,
	})
}

// HandleSetupAdmin handles POST /api/auth/setup
func (h *AdminHandler) HandleSetupAdmin(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	adminID, err := h.adminService.SetupFirstAdmin(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrAdminAlreadyExists) {
			utils.ErrorJSON(w, http.StatusForbidden, "Admin already exists. Setup is locked.")
			return
		}
		// Try to reuse the validation fallback from auth_handler logic
		msg := err.Error()
		if len(msg) > 5 && (msg[:5] == "email" || msg[:4] == "name" || msg[:8] == "password") {
			utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Printf("ERROR: Admin setup failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Admin setup failed")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{
		"message": "SuperAdmin created successfully. Setup is now locked.",
		"user_id": adminID,
	})
}
