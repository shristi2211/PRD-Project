package service

import (
	"context"
	"log"

	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
)

// ActivityLogService handles activity logging.
type ActivityLogService struct {
	activityLogRepo *repository.ActivityLogRepository
}

func NewActivityLogService(activityLogRepo *repository.ActivityLogRepository) *ActivityLogService {
	return &ActivityLogService{activityLogRepo: activityLogRepo}
}

// LogAction logs an activity. This is fire-and-forget — errors are only logged, not returned.
func (s *ActivityLogService) LogAction(ctx context.Context, userID *uuid.UUID, action, entityType, entityID string, metadata map[string]interface{}, ipAddress string) {
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	if err := s.activityLogRepo.LogActivity(ctx, userID, action, entityType, entityID, metadata, ipAddress); err != nil {
		log.Printf("WARNING: Failed to log activity [%s]: %v", action, err)
	}
}

// GetActivityLogs returns paginated activity logs for admin.
func (s *ActivityLogService) GetActivityLogs(ctx context.Context, page, pageSize int, userIDFilter, actionFilter string) (*models.PaginatedActivityLogsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := s.activityLogRepo.GetActivityLogs(ctx, page, pageSize, userIDFilter, actionFilter)
	if err != nil {
		return nil, err
	}

	return &models.PaginatedActivityLogsResponse{
		Logs:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetUsersWithActivity returns all users who have activity logs (drill-down level 1).
func (s *ActivityLogService) GetUsersWithActivity(ctx context.Context) ([]repository.ActivityUserSummary, error) {
	return s.activityLogRepo.GetUsersWithActivity(ctx)
}

// GetUserActivityYears returns years with activity for a user (drill-down level 2).
func (s *ActivityLogService) GetUserActivityYears(ctx context.Context, userID uuid.UUID) ([]repository.ActivityYearSummary, error) {
	return s.activityLogRepo.GetUserActivityYears(ctx, userID)
}

// GetUserActivityMonths returns months with activity for a user/year (drill-down level 3).
func (s *ActivityLogService) GetUserActivityMonths(ctx context.Context, userID uuid.UUID, year int) ([]repository.ActivityMonthSummary, error) {
	return s.activityLogRepo.GetUserActivityMonths(ctx, userID, year)
}

// GetUserMonthlyActivities returns detailed logs for a user in a specific month/year (drill-down level 4).
func (s *ActivityLogService) GetUserMonthlyActivities(ctx context.Context, userID uuid.UUID, year, month int) ([]models.ActivityLogResponse, error) {
	return s.activityLogRepo.GetUserMonthlyActivities(ctx, userID, year, month)
}
