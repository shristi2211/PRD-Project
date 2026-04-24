package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
)

// Prize split percentages
const (
	winnerSharePercent  = 60
	charitySharePercent = 30
	platformSharePercent = 10
)

// DrawService handles draw business logic.
type DrawService struct {
	drawRepo       *repository.DrawRepository
	scoreRepo      *repository.ScoreRepository
	winnerRepo     *repository.WinnerRepository
	activityLogSvc *ActivityLogService
}

func NewDrawService(
	drawRepo *repository.DrawRepository,
	scoreRepo *repository.ScoreRepository,
	winnerRepo *repository.WinnerRepository,
	activityLogSvc *ActivityLogService,
) *DrawService {
	return &DrawService{
		drawRepo:       drawRepo,
		scoreRepo:      scoreRepo,
		winnerRepo:     winnerRepo,
		activityLogSvc: activityLogSvc,
	}
}

// RunDraw executes a monthly lottery draw with crypto-random winner selection.
func (s *DrawService) RunDraw(ctx context.Context, adminID uuid.UUID, req *models.RunDrawRequest, ipAddress string) (*models.DrawDetailResponse, error) {
	// Validate month/year
	if req.Month < 1 || req.Month > 12 {
		return nil, fmt.Errorf("invalid month: must be 1-12")
	}
	if req.Year < 2020 {
		return nil, fmt.Errorf("invalid year: must be 2020 or later")
	}
	if req.PoolAmount <= 0 {
		return nil, fmt.Errorf("pool amount must be greater than 0")
	}

	// Prevent duplicate draws
	existing, err := s.drawRepo.GetDrawByMonthYear(ctx, req.Month, req.Year)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing draw: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("a draw has already been completed for %d/%d", req.Month, req.Year)
	}

	// Get eligible users
	eligibleUsers, err := s.scoreRepo.GetEligibleUsersForDraw(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get eligible users: %w", err)
	}
	if len(eligibleUsers) == 0 {
		return nil, fmt.Errorf("no eligible users found (users need active subscription + at least 1 score)")
	}

	// Calculate prize split
	totalPool := req.PoolAmount
	winnerPrize := totalPool * float64(winnerSharePercent) / 100
	charityAmount := totalPool * float64(charitySharePercent) / 100
	platformFee := totalPool * float64(platformSharePercent) / 100

	// Crypto-random winner selection
	winnerIdx, err := cryptoRandomInt(len(eligibleUsers))
	if err != nil {
		return nil, fmt.Errorf("failed to select random winner: %w", err)
	}
	winner := eligibleUsers[winnerIdx]

	// Create draw record
	draw, err := s.drawRepo.CreateDraw(ctx, req.Month, req.Year, totalPool, winnerPrize, charityAmount, platformFee, len(eligibleUsers))
	if err != nil {
		return nil, fmt.Errorf("failed to create draw: %w", err)
	}

	// Create draw entries for all eligible users
	var entries []models.DrawEntryResponse
	for i, u := range eligibleUsers {
		if err := s.drawRepo.CreateDrawEntry(ctx, draw.ID, u.UserID, u.ScoreID, u.BestScore); err != nil {
			log.Printf("WARNING: Failed to create draw entry for user %s: %v", u.UserID, err)
			continue
		}
		entries = append(entries, models.DrawEntryResponse{
			UserID:     u.UserID,
			UserName:   u.UserName,
			UserEmail:  u.UserEmail,
			EntryScore: u.BestScore,
			IsWinner:   i == winnerIdx,
		})
	}

	// Create winner record
	winnerRecord, err := s.winnerRepo.CreateWinner(ctx, draw.ID, winner.UserID, winnerPrize)
	if err != nil {
		return nil, fmt.Errorf("failed to create winner record: %w", err)
	}

	// Log activity
	s.activityLogSvc.LogAction(ctx, &adminID, "draw_executed", "draw", draw.ID.String(),
		map[string]interface{}{
			"month":         req.Month,
			"year":          req.Year,
			"total_pool":    totalPool,
			"winner_prize":  winnerPrize,
			"winner_user":   winner.UserName,
			"total_entries": len(eligibleUsers),
		}, ipAddress)

	winnerResp := &models.WinnerResponse{
		ID:                 winnerRecord.ID,
		DrawID:             winnerRecord.DrawID,
		UserID:             winnerRecord.UserID,
		UserName:           winner.UserName,
		UserEmail:          winner.UserEmail,
		PrizeAmount:        winnerRecord.PrizeAmount,
		VerificationStatus: winnerRecord.VerificationStatus,
		CreatedAt:          winnerRecord.CreatedAt,
	}

	return &models.DrawDetailResponse{
		Draw:    draw.ToResponse(),
		Entries: entries,
		Winner:  winnerResp,
	}, nil
}

// SimulateDraw does a dry-run without persisting any data.
func (s *DrawService) SimulateDraw(ctx context.Context, req *models.RunDrawRequest) (*models.DrawSimulationResult, error) {
	if req.Month < 1 || req.Month > 12 {
		return nil, fmt.Errorf("invalid month: must be 1-12")
	}
	if req.Year < 2020 {
		return nil, fmt.Errorf("invalid year: must be 2020 or later")
	}
	if req.PoolAmount <= 0 {
		return nil, fmt.Errorf("pool amount must be greater than 0")
	}

	eligibleUsers, err := s.scoreRepo.GetEligibleUsersForDraw(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get eligible users: %w", err)
	}

	totalPool := req.PoolAmount
	winnerPrize := totalPool * float64(winnerSharePercent) / 100
	charityAmount := totalPool * float64(charitySharePercent) / 100
	platformFee := totalPool * float64(platformSharePercent) / 100

	result := &models.DrawSimulationResult{
		Month:         req.Month,
		Year:          req.Year,
		EligibleUsers: len(eligibleUsers),
		TotalPool:     totalPool,
		WinnerPrize:   winnerPrize,
		CharityAmount: charityAmount,
		PlatformFee:   platformFee,
	}

	// Build entries
	for _, u := range eligibleUsers {
		result.Entries = append(result.Entries, models.DrawEntryResponse{
			UserID:     u.UserID,
			UserName:   u.UserName,
			UserEmail:  u.UserEmail,
			EntryScore: u.BestScore,
		})
	}

	// Sample winner (crypto-random)
	if len(eligibleUsers) > 0 {
		idx, err := cryptoRandomInt(len(eligibleUsers))
		if err == nil {
			w := eligibleUsers[idx]
			result.SampleWinner = &models.DrawEntryResponse{
				UserID:     w.UserID,
				UserName:   w.UserName,
				UserEmail:  w.UserEmail,
				EntryScore: w.BestScore,
				IsWinner:   true,
			}
		}
	}

	return result, nil
}

// ListDraws returns paginated draws for admin.
func (s *DrawService) ListDraws(ctx context.Context, page, pageSize int) (*models.PaginatedDrawsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	draws, total, err := s.drawRepo.ListDraws(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]models.DrawResponse, len(draws))
	for i, d := range draws {
		responses[i] = d.ToResponse()
	}

	return &models.PaginatedDrawsResponse{
		Draws:    responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetDrawDetail returns full draw details including entries and winner.
func (s *DrawService) GetDrawDetail(ctx context.Context, drawID uuid.UUID) (*models.DrawDetailResponse, error) {
	draw, err := s.drawRepo.GetDrawByID(ctx, drawID)
	if err != nil {
		return nil, err
	}

	entries, err := s.drawRepo.GetDrawEntries(ctx, drawID)
	if err != nil {
		return nil, err
	}

	winners, err := s.winnerRepo.GetWinnersByDrawID(ctx, drawID)
	if err != nil {
		return nil, err
	}

	// Mark the winner in entries
	var winnerResp *models.WinnerResponse
	if len(winners) > 0 {
		winnerResp = &winners[0]
		for i := range entries {
			if entries[i].UserID == winnerResp.UserID {
				entries[i].IsWinner = true
			}
		}
	}

	return &models.DrawDetailResponse{
		Draw:    draw.ToResponse(),
		Entries: entries,
		Winner:  winnerResp,
	}, nil
}

// cryptoRandomInt returns a cryptographically secure random integer in [0, max).
func cryptoRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}
