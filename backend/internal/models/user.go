package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"email"`
	PasswordHash       string    `json:"-"` // Never serialize to JSON
	Name               string    `json:"name"`
	Role               string    `json:"role"`
	SubscriptionActive bool      `json:"subscription_active"`
	SubscriptionType   string    `json:"subscription_type"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// UserResponse is a safe representation of User for API responses.
type UserResponse struct {
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"email"`
	Name               string    `json:"name"`
	Role               string    `json:"role"`
	SubscriptionActive bool      `json:"subscription_active"`
	SubscriptionType   string    `json:"subscription_type"`
	CreatedAt          time.Time `json:"created_at"`
}

// ToResponse converts a User to a safe UserResponse.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:                 u.ID,
		Email:              u.Email,
		Name:               u.Name,
		Role:               u.Role,
		SubscriptionActive: u.SubscriptionActive,
		SubscriptionType:   u.SubscriptionType,
		CreatedAt:          u.CreatedAt,
	}
}

// SubscriptionRequest is the input DTO for subscription management.
type SubscriptionRequest struct {
	Plan string `json:"plan"` // "free", "monthly", "yearly"
}

// PublicSubscribeRequest is for pre-login subscription activation.
type PublicSubscribeRequest struct {
	Email string `json:"email"`
	Plan  string `json:"plan"`
}

// IpSubscribeRequest is for IP-based pre-registration lock.
type IpSubscribeRequest struct {
	Plan string `json:"plan"` // "monthly", "yearly"
}

// IpSubscriptionStatus represents whether an IP is unlocked.
type IpSubscriptionStatus struct {
	Active   bool   `json:"active"`
	PlanType string `json:"plan_type,omitempty"`
}

// RefreshToken represents a stored refresh token.
type RefreshToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	FamilyID  uuid.UUID `json:"family_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterRequest is the input DTO for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest is the input DTO for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the output DTO after successful login.
type LoginResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// RefreshRequest is the input DTO for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse is the output DTO after token refresh.
type TokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest is the input DTO for logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// UpdateProfileRequest is the input DTO for updating user profile.
type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ChangePasswordRequest is the input DTO for changing password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ToggleActivationRequest is the input DTO for admin toggling user activation.
type ToggleActivationRequest struct {
	Active bool `json:"active"`
}

// PaginatedUsersResponse is the output DTO for admin user listing.
type PaginatedUsersResponse struct {
	Users    []UserResponse `json:"users"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
