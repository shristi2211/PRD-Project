package service

import (
	"context"

	"golf-score-lottery/backend/internal/repository"
)

// ReportService handles report generation logic.
type ReportService struct {
	reportRepo *repository.ReportRepository
}

func NewReportService(reportRepo *repository.ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

// GetUserReport returns user summary report data.
func (s *ReportService) GetUserReport(ctx context.Context, from, to string) ([]repository.UserReportRow, error) {
	return s.reportRepo.GetUserReport(ctx, from, to)
}

// GetRevenueReport returns revenue breakdown report data.
func (s *ReportService) GetRevenueReport(ctx context.Context, from, to string) ([]repository.RevenueReportRow, error) {
	return s.reportRepo.GetRevenueReport(ctx, from, to)
}

// GetDrawReport returns draw history report data.
func (s *ReportService) GetDrawReport(ctx context.Context, from, to string) ([]repository.DrawReportRow, error) {
	return s.reportRepo.GetDrawReport(ctx, from, to)
}

// GetCharityReport returns charity allocation report data.
func (s *ReportService) GetCharityReport(ctx context.Context) ([]repository.CharityReportRow, error) {
	return s.reportRepo.GetCharityReport(ctx)
}
