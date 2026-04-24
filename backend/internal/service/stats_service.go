package service

import (
	"context"

	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
)

// StatsService handles dashboard statistics.
type StatsService struct {
	statsRepo   *repository.StatsRepository
	charityRepo *repository.CharityRepository
}

func NewStatsService(statsRepo *repository.StatsRepository, charityRepo *repository.CharityRepository) *StatsService {
	return &StatsService{statsRepo: statsRepo, charityRepo: charityRepo}
}

// GetUserDashboardStats returns dashboard statistics for a user.
func (s *StatsService) GetUserDashboardStats(ctx context.Context, userID uuid.UUID) (*models.UserDashboardStats, error) {
	return s.statsRepo.GetUserDashboardStats(ctx, userID)
}

// GetAdminDashboardStats returns dashboard statistics for admin.
func (s *StatsService) GetAdminDashboardStats(ctx context.Context) (*models.AdminDashboardStats, error) {
	return s.statsRepo.GetAdminDashboardStats(ctx)
}

// GetScoreTrend returns score trend chart data for a user.
func (s *StatsService) GetScoreTrend(ctx context.Context, userID uuid.UUID) ([]models.ChartDataPoint, error) {
	return s.statsRepo.GetScoreTrend(ctx, userID)
}

// GetUserGrowthData returns user growth chart data for admin.
func (s *StatsService) GetUserGrowthData(ctx context.Context) ([]models.ChartDataPoint, error) {
	return s.statsRepo.GetUserGrowthData(ctx)
}

// GetRevenueData returns revenue chart data for admin.
func (s *StatsService) GetRevenueData(ctx context.Context) ([]models.ChartDataPoint, error) {
	return s.statsRepo.GetRevenueData(ctx)
}

// GetCharityDistribution returns charity selection distribution.
func (s *StatsService) GetCharityDistribution(ctx context.Context) ([]models.CharityDistribution, error) {
	return s.charityRepo.GetCharityDistribution(ctx)
}
