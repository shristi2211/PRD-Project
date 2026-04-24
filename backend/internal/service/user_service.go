package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
	"golf-score-lottery/backend/internal/utils"
)

// User-service-level errors
var (
	ErrPasswordMismatch = errors.New("current password is incorrect")
	ErrSamePassword     = errors.New("new password must be different from current password")
)

// UserService handles profile and admin user management business logic.
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// UpdateProfile updates a user's name and email after validation.
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateProfileRequest) (*models.UserResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	// Validate name
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(name) < 2 {
		return nil, fmt.Errorf("name must be at least 2 characters")
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("name must not exceed 255 characters")
	}

	// Validate email
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	user, err := s.userRepo.UpdateUser(ctx, userID, name, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return nil, ErrUserExists
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	resp := user.ToResponse()
	return &resp, nil
}

// ChangePassword verifies the current password, validates the new one, and updates it.
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, req *models.ChangePasswordRequest) error {
	if req.CurrentPassword == "" {
		return fmt.Errorf("current password is required")
	}
	if req.NewPassword == "" {
		return fmt.Errorf("new password is required")
	}

	// Fetch user to verify current password
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	// Verify current password
	if err := utils.CheckPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		return ErrPasswordMismatch
	}

	// Prevent setting the same password
	if err := utils.CheckPassword(req.NewPassword, user.PasswordHash); err == nil {
		return ErrSamePassword
	}

	// Validate new password strength
	if len(req.NewPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(req.NewPassword) > 72 {
		return fmt.Errorf("password must not exceed 72 characters")
	}

	// Hash new password
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.userRepo.UpdateUserPassword(ctx, userID, newHash)
}

// DeleteAccount deletes a user account permanently.
func (s *UserService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	// Refresh tokens are cascade-deleted by the FK constraint in PostgreSQL
	return s.userRepo.DeleteUser(ctx, userID)
}

// ToggleSubscriptionStatus updates the active flag of a user's subscription.
func (s *UserService) ToggleSubscriptionStatus(ctx context.Context, userID uuid.UUID, active bool) (*models.UserResponse, error) {
	if err := s.userRepo.SetSubscriptionStatus(ctx, userID, active); err != nil {
		return nil, fmt.Errorf("failed to update subscription status: %w", err)
	}
	
	// Fetch updated user to return response
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	resp := user.ToResponse()
	return &resp, nil
}

// ListUsers returns a paginated list of users for admin.
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int, search, statusFilter string) (*models.PaginatedUsersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := s.userRepo.ListUsers(ctx, page, pageSize, search, statusFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	userResponses := make([]models.UserResponse, len(users))
	for i, u := range users {
		userResponses[i] = u.ToResponse()
	}

	return &models.PaginatedUsersResponse{
		Users:    userResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ToggleUserActivation sets a user's subscription_active status.
func (s *UserService) ToggleUserActivation(ctx context.Context, userID uuid.UUID, active bool) error {
	return s.userRepo.ToggleUserActivation(ctx, userID, active)
}

// StartSubscription activates a plan for an authenticated user.
func (s *UserService) StartSubscription(ctx context.Context, userID uuid.UUID, plan string) (*models.UserResponse, error) {
	plan = strings.ToLower(strings.TrimSpace(plan))
	if plan != "monthly" && plan != "yearly" {
		return nil, ErrInvalidPlan
	}

	if err := s.userRepo.SetSubscriptionType(ctx, userID, plan); err != nil {
		return nil, fmt.Errorf("failed to start subscription: %w", err)
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	resp := user.ToResponse()
	return &resp, nil
}

// CancelSubscription deactivates a user's subscription (sets to free).
func (s *UserService) CancelSubscription(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error) {
	if err := s.userRepo.SetSubscriptionType(ctx, userID, "free"); err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	resp := user.ToResponse()
	return &resp, nil
}

// PublicSubscribe activates a subscription for an existing user by email (pre-login).
func (s *UserService) PublicSubscribe(ctx context.Context, email, plan string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	plan = strings.ToLower(strings.TrimSpace(plan))

	if plan != "monthly" && plan != "yearly" {
		return ErrInvalidPlan
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found with this email")
	}

	return s.userRepo.SetSubscriptionType(ctx, user.ID, plan)
}

// SaveIpSubscription locks the IP with a plan.
func (s *UserService) SaveIpSubscription(ctx context.Context, ip, plan string) error {
	plan = strings.ToLower(strings.TrimSpace(plan))
	if plan != "monthly" && plan != "yearly" {
		return ErrInvalidPlan
	}
	return s.userRepo.SaveIpSubscription(ctx, ip, plan)
}

// CheckIpSubscription checks if an IP is subscribed.
func (s *UserService) CheckIpSubscription(ctx context.Context, ip string) (*models.IpSubscriptionStatus, error) {
	plan, err := s.userRepo.GetIpSubscription(ctx, ip)
	if err != nil {
		return nil, err
	}
	
	if plan == "" {
		return &models.IpSubscriptionStatus{Active: false}, nil
	}
	return &models.IpSubscriptionStatus{Active: true, PlanType: plan}, nil
}

