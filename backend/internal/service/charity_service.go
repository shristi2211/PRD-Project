package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
)

// CharityService handles charity business logic.
type CharityService struct {
	charityRepo    *repository.CharityRepository
	activityLogSvc *ActivityLogService
}

func NewCharityService(charityRepo *repository.CharityRepository, activityLogSvc *ActivityLogService) *CharityService {
	return &CharityService{charityRepo: charityRepo, activityLogSvc: activityLogSvc}
}

// CreateCharity creates a new charity (admin only).
func (s *CharityService) CreateCharity(ctx context.Context, adminID uuid.UUID, req *models.CreateCharityRequest, ipAddress string) (*models.CharityResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("name must not exceed 255 characters")
	}

	charity, err := s.charityRepo.CreateCharity(ctx, name, strings.TrimSpace(req.Description), strings.TrimSpace(req.Website), strings.TrimSpace(req.LogoURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create charity: %w", err)
	}

	s.activityLogSvc.LogAction(ctx, &adminID, "charity_created", "charity", charity.ID.String(),
		map[string]interface{}{"name": name}, ipAddress)

	resp := charity.ToResponse()
	return &resp, nil
}

// UpdateCharity updates a charity (admin only).
func (s *CharityService) UpdateCharity(ctx context.Context, adminID uuid.UUID, charityID uuid.UUID, req *models.UpdateCharityRequest, ipAddress string) (*models.CharityResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	charity, err := s.charityRepo.UpdateCharity(ctx, charityID, name, strings.TrimSpace(req.Description), strings.TrimSpace(req.Website), strings.TrimSpace(req.LogoURL))
	if err != nil {
		return nil, err
	}

	s.activityLogSvc.LogAction(ctx, &adminID, "charity_updated", "charity", charityID.String(),
		map[string]interface{}{"name": name}, ipAddress)

	resp := charity.ToResponse()
	return &resp, nil
}

// ToggleCharityActive toggles a charity's active status (admin only).
func (s *CharityService) ToggleCharityActive(ctx context.Context, adminID uuid.UUID, charityID uuid.UUID, active bool, ipAddress string) error {
	if err := s.charityRepo.ToggleCharityActive(ctx, charityID, active); err != nil {
		return err
	}

	action := "charity_activated"
	if !active {
		action = "charity_deactivated"
	}
	s.activityLogSvc.LogAction(ctx, &adminID, action, "charity", charityID.String(), nil, ipAddress)

	return nil
}

// ListCharities returns charities. Users see only active, admin sees all.
func (s *CharityService) ListCharities(ctx context.Context, activeOnly bool) ([]models.CharityResponse, error) {
	charities, err := s.charityRepo.ListCharities(ctx, activeOnly)
	if err != nil {
		return nil, err
	}

	responses := make([]models.CharityResponse, len(charities))
	for i, c := range charities {
		responses[i] = c.ToResponse()
	}
	return responses, nil
}

// SelectCharity sets a user's charity selections.
func (s *CharityService) SelectCharity(ctx context.Context, userID uuid.UUID, req *models.SelectCharityRequest, ipAddress string) ([]models.UserCharitySelectionResponse, error) {
	if len(req.Allocations) == 0 {
		return nil, fmt.Errorf("you must select at least one charity")
	}

	totalPercentage := 0
	allocs := make([]struct{ CharityID uuid.UUID; Percentage int }, 0)

	for _, alloc := range req.Allocations {
		if alloc.ContributionPercentage <= 0 {
			return nil, fmt.Errorf("percentage must be greater than 0")
		}
		totalPercentage += alloc.ContributionPercentage
		
		// Verify charity exists and is active
		charity, err := s.charityRepo.GetCharityByID(ctx, alloc.CharityID)
		if err != nil {
			return nil, fmt.Errorf("charity not found")
		}
		if !charity.Active {
			return nil, fmt.Errorf("charity '%s' is not currently accepting contributions", charity.Name)
		}

		allocs = append(allocs, struct{ CharityID uuid.UUID; Percentage int }{alloc.CharityID, alloc.ContributionPercentage})
	}

	if totalPercentage != 30 {
		return nil, fmt.Errorf("total contribution percentage must equal exactly 30%% (currently %d%%)", totalPercentage)
	}

	sels, err := s.charityRepo.SetUserCharityAllocations(ctx, userID, allocs)
	if err != nil {
		return nil, fmt.Errorf("failed to select charities: %w", err)
	}

	s.activityLogSvc.LogAction(ctx, &userID, "charities_selected", "charity", "bulk",
		map[string]interface{}{"count": len(sels), "total_percentage": totalPercentage}, ipAddress)

	return s.charityRepo.GetUserCharitySelection(ctx, userID)
}

// GetUserCharitySelection returns the user's current charity selection.
func (s *CharityService) GetUserCharitySelection(ctx context.Context, userID uuid.UUID) ([]models.UserCharitySelectionResponse, error) {
	return s.charityRepo.GetUserCharitySelection(ctx, userID)
}

// GetCharityDistribution returns charity selection distribution stats.
func (s *CharityService) GetCharityDistribution(ctx context.Context) ([]models.CharityDistribution, error) {
	return s.charityRepo.GetCharityDistribution(ctx)
}
