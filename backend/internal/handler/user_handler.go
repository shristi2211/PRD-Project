package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/middleware"
	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

// UserHandler holds HTTP handlers for profile and admin user management.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// HandleUpdateProfile handles PUT /api/users/me
func (h *UserHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	userResp, err := h.userService.UpdateProfile(r.Context(), claims.UserID, &req)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			utils.ErrorJSON(w, http.StatusConflict, "A user with this email already exists")
			return
		}
		if isValidationError(err) {
			utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("ERROR: UpdateProfile failed for user %s: %v", claims.UserID, err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	utils.JSON(w, http.StatusOK, userResp)
}

// HandleChangePassword handles PUT /api/users/me/password
func (h *UserHandler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	err := h.userService.ChangePassword(r.Context(), claims.UserID, &req)
	if err != nil {
		if errors.Is(err, service.ErrPasswordMismatch) {
			utils.ErrorJSON(w, http.StatusUnauthorized, "Current password is incorrect")
			return
		}
		if errors.Is(err, service.ErrSamePassword) {
			utils.ErrorJSON(w, http.StatusBadRequest, "New password must be different from current password")
			return
		}
		if isValidationError(err) {
			utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("ERROR: ChangePassword failed for user %s: %v", claims.UserID, err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to change password")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// HandleDeleteAccount handles DELETE /api/users/me
func (h *UserHandler) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.userService.DeleteAccount(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("ERROR: DeleteAccount failed for user %s: %v", claims.UserID, err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Account deleted successfully"})
}

// HandleStartSubscription handles POST /api/subscriptions/start
func (h *UserHandler) HandleStartSubscription(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	resp, err := h.userService.StartSubscription(r.Context(), claims.UserID, req.Plan)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPlan) {
			utils.ErrorJSON(w, http.StatusBadRequest, "Invalid plan. Choose 'monthly' or 'yearly'.")
			return
		}
		log.Printf("ERROR: StartSubscription failed for user %s: %v", claims.UserID, err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to start subscription")
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

// HandleCancelSubscription handles PUT /api/subscriptions/cancel
func (h *UserHandler) HandleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, err := h.userService.CancelSubscription(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("ERROR: CancelSubscription failed for user %s: %v", claims.UserID, err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to cancel subscription")
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

// HandlePublicSubscribe handles POST /api/auth/subscribe (public, no auth required)
func (h *UserHandler) HandlePublicSubscribe(w http.ResponseWriter, r *http.Request) {
	var req models.PublicSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	err := h.userService.PublicSubscribe(r.Context(), req.Email, req.Plan)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPlan) {
			utils.ErrorJSON(w, http.StatusBadRequest, "Invalid plan. Choose 'monthly' or 'yearly'.")
			return
		}
		log.Printf("ERROR: PublicSubscribe failed for email %s: %v", req.Email, err)
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Subscription activated successfully"})
}

// HandleListUsers handles GET /api/admin/users?page=1&size=10&search=&status=
func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if pageSize < 1 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	result, err := h.userService.ListUsers(r.Context(), page, pageSize, search, status)
	if err != nil {
		log.Printf("ERROR: ListUsers failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	utils.JSON(w, http.StatusOK, result)
}

// HandleToggleActivation handles PUT /api/admin/users/{id}/activation
func (h *UserHandler) HandleToggleActivation(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.ToggleActivationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	if err := h.userService.ToggleUserActivation(r.Context(), userID, req.Active); err != nil {
		log.Printf("ERROR: ToggleActivation failed for user %s: %v", userID, err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to update user activation")
		return
	}

	status := "activated"
	if !req.Active {
		status = "deactivated"
	}
	utils.JSON(w, http.StatusOK, map[string]string{"message": "User " + status + " successfully"})
}

// HandleIpSubscribe allows public users to lock an IP with a subscription.
func (h *UserHandler) HandleIpSubscribe(w http.ResponseWriter, r *http.Request) {
	var req models.IpSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	if strings.Contains(ip, ":") && !strings.Contains(ip, "[") { // strip port mapping
		ip = strings.Split(ip, ":")[0]
	}

	if err := h.userService.SaveIpSubscription(r.Context(), ip, req.Plan); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "IP subscription saved successfully"})
}

// HandleIpStatus checks if the calling IP is subscribed.
func (h *UserHandler) HandleIpStatus(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	if strings.Contains(ip, ":") && !strings.Contains(ip, "[") { // strip port mapping
		ip = strings.Split(ip, ":")[0]
	}

	status, err := h.userService.CheckIpSubscription(r.Context(), ip)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to check IP status")
		return
	}

	utils.JSON(w, http.StatusOK, status)
}
