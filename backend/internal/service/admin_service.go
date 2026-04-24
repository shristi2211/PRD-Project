package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
	"golf-score-lottery/backend/internal/utils"
)

var (
	ErrAdminAlreadyExists = errors.New("admin setup is already complete and locked")
)

// AdminService isolated for administrative/setup workflows.
type AdminService struct {
	adminRepo *repository.AdminRepository
}

func NewAdminService(adminRepo *repository.AdminRepository) *AdminService {
	return &AdminService{adminRepo: adminRepo}
}

// IsSetupComplete returns true if an admin exists, meaning the system is locked down.
func (s *AdminService) IsSetupComplete(ctx context.Context) (bool, error) {
	return s.adminRepo.HasAdmin(ctx)
}

// SetupFirstAdmin registers the first admin user securely. Returns an error if one already exists.
func (s *AdminService) SetupFirstAdmin(ctx context.Context, req *models.RegisterRequest) (string, error) {
	// 1. Lockout Check: Ensure no admin already exists
	exists, err := s.adminRepo.HasAdmin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to check setup status: %w", err)
	}
	if exists {
		return "", ErrAdminAlreadyExists
	}

	// 2. Validate input
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	if err := utils.ValidateRegisterInput(email, req.Password, name); err != nil {
		return "", err
	}

	// 3. Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return "", fmt.Errorf("failed to hash admin password: %w", err)
	}

	// 4. Create admin securely
	adminID, err := s.adminRepo.CreateFirstAdmin(ctx, email, hashedPassword, name)
	if err != nil {
		return "", err
	}

	return adminID, nil
}
