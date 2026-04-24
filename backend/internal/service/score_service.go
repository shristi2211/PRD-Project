package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
)

const maxScoresPerUser = 5

// ScoreService handles score business logic.
type ScoreService struct {
	scoreRepo      *repository.ScoreRepository
	activityLogSvc *ActivityLogService
}

func NewScoreService(scoreRepo *repository.ScoreRepository, activityLogSvc *ActivityLogService) *ScoreService {
	return &ScoreService{scoreRepo: scoreRepo, activityLogSvc: activityLogSvc}
}

// CreateScore validates and creates a new score.
func (s *ScoreService) CreateScore(ctx context.Context, userID uuid.UUID, req *models.CreateScoreRequest, ipAddress string) (*models.ScoreResponse, error) {
	// Validate score range
	if req.Score < 1 || req.Score > 45 {
		return nil, fmt.Errorf("invalid score: must be between 1 and 45 (Stableford)")
	}

	// Validate round date
	if req.RoundDate == "" {
		req.RoundDate = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", req.RoundDate); err != nil {
		return nil, fmt.Errorf("invalid round_date format: use YYYY-MM-DD")
	}

	// Check 1 score per date limit
	exists, err := s.scoreRepo.CheckScoreExistsForDate(ctx, userID, req.RoundDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check date limit: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("you have already logged a score for %s", req.RoundDate)
	}

	// Check 5-score limit
	count, err := s.scoreRepo.CountUserScores(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check score count: %w", err)
	}
	if count >= maxScoresPerUser {
		if err := s.scoreRepo.DeleteOldestScoreByUserID(ctx, userID); err != nil {
			return nil, fmt.Errorf("failed to roll over oldest score: %w", err)
		}
	}

	// Create score
	score, err := s.scoreRepo.CreateScore(ctx, userID, req.Score, req.RoundDate, req.Notes)
	if err != nil {
		return nil, fmt.Errorf("failed to create score: %w", err)
	}

	// Log activity
	s.activityLogSvc.LogAction(ctx, &userID, "score_created", "score", score.ID.String(),
		map[string]interface{}{"score": req.Score, "round_date": req.RoundDate}, ipAddress)

	resp := score.ToResponse()
	return &resp, nil
}

// UpdateScore allows modifying an existing score.
func (s *ScoreService) UpdateScore(ctx context.Context, userID uuid.UUID, scoreID uuid.UUID, req *models.CreateScoreRequest, ipAddress string) (*models.ScoreResponse, error) {
	if req.Score < 1 || req.Score > 45 {
		return nil, fmt.Errorf("invalid score: must be between 1 and 45 (Stableford)")
	}

	if req.RoundDate == "" {
		req.RoundDate = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", req.RoundDate); err != nil {
		return nil, fmt.Errorf("invalid round_date format: use YYYY-MM-DD")
	}

	// Verify ownership
	score, err := s.scoreRepo.GetScoreByID(ctx, scoreID)
	if err != nil {
		return nil, err
	}
	if score.UserID != userID {
		return nil, fmt.Errorf("you can only edit your own scores")
	}

	// Check date limit if date changed
	if score.RoundDate.Format("2006-01-02") != req.RoundDate {
		exists, err := s.scoreRepo.CheckScoreExistsForDate(ctx, userID, req.RoundDate)
		if err != nil {
			return nil, fmt.Errorf("failed to check date limit: %w", err)
		}
		if exists && score.RoundDate.Format("2006-01-02") != req.RoundDate {
			return nil, fmt.Errorf("you have already logged a score for %s", req.RoundDate)
		}
	}

	updatedScore, err := s.scoreRepo.UpdateScore(ctx, scoreID, req.Score, req.RoundDate, req.Notes)
	if err != nil {
		return nil, fmt.Errorf("failed to update score: %w", err)
	}

	s.activityLogSvc.LogAction(ctx, &userID, "score_updated", "score", scoreID.String(),
		map[string]interface{}{"old_score": score.Score, "new_score": req.Score}, ipAddress)

	resp := updatedScore.ToResponse()
	return &resp, nil
}

// GetUserScores returns all scores for a user.
func (s *ScoreService) GetUserScores(ctx context.Context, userID uuid.UUID) ([]models.ScoreResponse, error) {
	scores, err := s.scoreRepo.GetScoresByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scores: %w", err)
	}

	responses := make([]models.ScoreResponse, len(scores))
	for i, sc := range scores {
		responses[i] = sc.ToResponse()
	}
	return responses, nil
}

// DeleteScore deletes a user's own score.
func (s *ScoreService) DeleteScore(ctx context.Context, userID uuid.UUID, scoreID uuid.UUID, ipAddress string) error {
	// Verify ownership
	score, err := s.scoreRepo.GetScoreByID(ctx, scoreID)
	if err != nil {
		return err
	}
	if score.UserID != userID {
		return fmt.Errorf("you can only delete your own scores")
	}

	if err := s.scoreRepo.DeleteScore(ctx, scoreID); err != nil {
		return err
	}

	s.activityLogSvc.LogAction(ctx, &userID, "score_deleted", "score", scoreID.String(),
		map[string]interface{}{"score": score.Score}, ipAddress)

	return nil
}

// AdminDeleteScore allows admin to delete any score.
func (s *ScoreService) AdminDeleteScore(ctx context.Context, adminID uuid.UUID, scoreID uuid.UUID, ipAddress string) error {
	score, err := s.scoreRepo.GetScoreByID(ctx, scoreID)
	if err != nil {
		return err
	}

	if err := s.scoreRepo.DeleteScore(ctx, scoreID); err != nil {
		return err
	}

	s.activityLogSvc.LogAction(ctx, &adminID, "score_deleted_by_admin", "score", scoreID.String(),
		map[string]interface{}{"score": score.Score, "user_id": score.UserID.String()}, ipAddress)

	return nil
}

// ListAllScores returns paginated scores for admin.
func (s *ScoreService) ListAllScores(ctx context.Context, page, pageSize int, search string) (*models.PaginatedScoresResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	scores, total, err := s.scoreRepo.ListAllScores(ctx, page, pageSize, search)
	if err != nil {
		return nil, err
	}

	return &models.PaginatedScoresResponse{
		Scores:   scores,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
