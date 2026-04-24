package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"golf-score-lottery/backend/internal/middleware"
	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/service"
	"golf-score-lottery/backend/internal/utils"
)

// AuthHandler holds HTTP handlers for authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// HandleRegister handles POST /api/auth/register
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	if strings.Contains(ip, ":") && !strings.Contains(ip, "[") { // strip port mapping
		ip = strings.Split(ip, ":")[0]
	}

	userResp, err := h.authService.Register(r.Context(), &req, ip)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			utils.ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}
		// Validation errors are returned as-is (they're user-friendly)
		if isValidationError(err) {
			utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("ERROR: Register failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Registration failed. Please try again.")
		return
	}

	utils.JSON(w, http.StatusCreated, userResp)
}

// HandleLogin handles POST /api/auth/login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	if strings.Contains(ip, ":") && !strings.Contains(ip, "[") { // strip port mapping
		ip = strings.Split(ip, ":")[0]
	}

	loginResp, err := h.authService.Login(r.Context(), &req, ip)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			utils.ErrorJSON(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		if errors.Is(err, service.ErrSubscriptionRequired) {
			utils.ErrorJSON(w, http.StatusForbidden, "Active subscription required. Please subscribe from the landing page.")
			return
		}
		if isValidationError(err) {
			utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("ERROR: Login failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Login failed. Please try again.")
		return
	}

	utils.JSON(w, http.StatusOK, loginResp)
}

// HandleRefresh handles POST /api/auth/refresh
func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	tokenResp, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) || errors.Is(err, service.ErrTokenRevoked) {
			utils.ErrorJSON(w, http.StatusUnauthorized, "Invalid or expired refresh token")
			return
		}
		log.Printf("ERROR: Token refresh failed: %v", err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Token refresh failed. Please login again.")
		return
	}

	utils.JSON(w, http.StatusOK, tokenResp)
}

// HandleLogout handles POST /api/auth/logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var req models.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		log.Printf("ERROR: Logout failed: %v", err)
		// Still return success — logout should be idempotent
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// HandleGetMe handles GET /api/users/me (protected)
func (h *AuthHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userResp, err := h.authService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("ERROR: GetMe failed for user %s: %v", claims.UserID, err)
		utils.ErrorJSON(w, http.StatusInternalServerError, "Failed to fetch user data")
		return
	}

	utils.JSON(w, http.StatusOK, userResp)
}

// isValidationError checks if an error is a user-input validation error
// (safe to expose to the client).
func isValidationError(err error) bool {
	msg := err.Error()
	validationPrefixes := []string{
		"email", "password", "name", "invalid",
	}
	for _, prefix := range validationPrefixes {
		if len(msg) >= len(prefix) && msg[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
