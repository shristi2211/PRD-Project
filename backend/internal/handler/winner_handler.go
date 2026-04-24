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

// WinnerHandler holds HTTP handlers for winner endpoints.
type WinnerHandler struct {
	winnerService *service.WinnerService
}

func NewWinnerHandler(winnerService *service.WinnerService) *WinnerHandler {
	return &WinnerHandler{winnerService: winnerService}
}

// HandleGetMyWinnings handles GET /api/winners/me
func (h *WinnerHandler) HandleGetMyWinnings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	winnings, err := h.winnerService.GetMyWinnings(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("ERROR: GetMyWinnings failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch winnings")
		return
	}

	if winnings == nil {
		winnings = []models.WinnerResponse{}
	}

	utils.JSON(w, http.StatusOK, winnings)
}

// HandleSubmitProof handles PUT /api/winners/{id}/proof
func (h *WinnerHandler) HandleSubmitProof(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	winnerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid winner ID")
		return
	}

	var req models.SubmitProofRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	if err := h.winnerService.SubmitProof(r.Context(), claims.UserID, winnerID, &req, r.RemoteAddr); err != nil {
		msg := err.Error()
		if isBusinessError(msg) || isValidationError(err) {
			utils.ErrorJSON(w, http.StatusBadRequest, msg)
			return
		}
		if msg == "you can only submit proof for your own winnings" {
			utils.ErrorJSON(w, http.StatusForbidden, msg)
			return
		}
		log.Printf("ERROR: SubmitProof failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to submit proof")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Proof submitted successfully"})
}

// HandleGetPendingVerifications handles GET /api/admin/winners/pending
func (h *WinnerHandler) HandleGetPendingVerifications(w http.ResponseWriter, r *http.Request) {
	winners, err := h.winnerService.GetPendingVerifications(r.Context())
	if err != nil {
		log.Printf("ERROR: GetPendingVerifications failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch pending verifications")
		return
	}

	if winners == nil {
		winners = []models.WinnerResponse{}
	}

	utils.JSON(w, http.StatusOK, winners)
}

// HandleVerifyWinner handles PUT /api/admin/winners/{id}/verify
func (h *WinnerHandler) HandleVerifyWinner(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	winnerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid winner ID")
		return
	}

	var req models.VerifyWinnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	if err := h.winnerService.VerifyWinner(r.Context(), claims.UserID, winnerID, &req, r.RemoteAddr); err != nil {
		msg := err.Error()
		if isBusinessError(msg) {
			utils.ErrorJSON(w, http.StatusBadRequest, msg)
			return
		}
		log.Printf("ERROR: VerifyWinner failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to verify winner")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Winner verification updated"})
}
