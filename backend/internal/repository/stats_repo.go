package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"golf-score-lottery/backend/internal/models"
)

// StatsRepository handles all dashboard statistics queries.
type StatsRepository struct {
	pool *pgxpool.Pool
}

func NewStatsRepository(pool *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{pool: pool}
}

// GetUserDashboardStats returns dashboard statistics for a specific user.
func (r *StatsRepository) GetUserDashboardStats(ctx context.Context, userID uuid.UUID) (*models.UserDashboardStats, error) {
	stats := &models.UserDashboardStats{}

	// Rounds played (score count)
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM scores WHERE user_id = $1`, userID,
	).Scan(&stats.RoundsPlayed); err != nil {
		return nil, fmt.Errorf("failed to get rounds: %w", err)
	}

	// Best score
	if err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(score), 0) FROM scores WHERE user_id = $1`, userID,
	).Scan(&stats.BestScore); err != nil {
		return nil, fmt.Errorf("failed to get best score: %w", err)
	}

	// Lottery entries
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM draw_entries WHERE user_id = $1`, userID,
	).Scan(&stats.LotteryEntries); err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// Total winnings
	if err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(prize_amount), 0) FROM winners WHERE user_id = $1 AND verification_status = 'approved'`, userID,
	).Scan(&stats.TotalWinnings); err != nil {
		return nil, fmt.Errorf("failed to get winnings: %w", err)
	}

	return stats, nil
}

// GetAdminDashboardStats returns dashboard statistics for admin.
func (r *StatsRepository) GetAdminDashboardStats(ctx context.Context) (*models.AdminDashboardStats, error) {
	stats := &models.AdminDashboardStats{}

	// Total users
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE role = 'user'`,
	).Scan(&stats.TotalUsers); err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}

	// Active subscriptions
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE subscription_active = true AND role = 'user'`,
	).Scan(&stats.ActiveSubscriptions); err != nil {
		return nil, fmt.Errorf("failed to get active subs: %w", err)
	}

	// Total revenue (sum of all draw pools)
	if err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(platform_fee), 0) FROM draws WHERE status = 'completed'`,
	).Scan(&stats.TotalRevenue); err != nil {
		return nil, fmt.Errorf("failed to get revenue: %w", err)
	}

	// Pending verifications
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM winners WHERE verification_status = 'pending'`,
	).Scan(&stats.PendingVerifications); err != nil {
		return nil, fmt.Errorf("failed to get pending: %w", err)
	}

	// Total draws
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM draws`,
	).Scan(&stats.TotalDraws); err != nil {
		return nil, fmt.Errorf("failed to get draws count: %w", err)
	}

	// Total charities
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM charities WHERE active = true`,
	).Scan(&stats.TotalCharities); err != nil {
		return nil, fmt.Errorf("failed to get charities count: %w", err)
	}

	return stats, nil
}

// GetScoreTrend returns score trend data for a user (last 6 scores).
func (r *StatsRepository) GetScoreTrend(ctx context.Context, userID uuid.UUID) ([]models.ChartDataPoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT TO_CHAR(round_date, 'Mon DD') as label, score
		 FROM scores WHERE user_id = $1
		 ORDER BY round_date ASC
		 LIMIT 10`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get score trend: %w", err)
	}
	defer rows.Close()

	var points []models.ChartDataPoint
	for rows.Next() {
		var p models.ChartDataPoint
		if err := rows.Scan(&p.Label, &p.Value); err != nil {
			return nil, fmt.Errorf("failed to scan trend point: %w", err)
		}
		points = append(points, p)
	}
	return points, nil
}

// GetUserGrowthData returns monthly user registration counts (last 6 months).
func (r *StatsRepository) GetUserGrowthData(ctx context.Context) ([]models.ChartDataPoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT TO_CHAR(date_trunc('month', created_at), 'Mon') as label,
		        COUNT(*) as value
		 FROM users WHERE role = 'user'
		 GROUP BY date_trunc('month', created_at)
		 ORDER BY date_trunc('month', created_at) DESC
		 LIMIT 6`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user growth: %w", err)
	}
	defer rows.Close()

	var points []models.ChartDataPoint
	for rows.Next() {
		var p models.ChartDataPoint
		if err := rows.Scan(&p.Label, &p.Value); err != nil {
			return nil, fmt.Errorf("failed to scan growth point: %w", err)
		}
		points = append(points, p)
	}

	// Reverse to chronological order
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}

// GetRevenueData returns monthly platform fee data (last 6 months).
func (r *StatsRepository) GetRevenueData(ctx context.Context) ([]models.ChartDataPoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT TO_CHAR(draw_date, 'Mon') as label,
		        COALESCE(SUM(platform_fee), 0) as value
		 FROM draws WHERE status = 'completed'
		 GROUP BY date_trunc('month', draw_date), TO_CHAR(draw_date, 'Mon')
		 ORDER BY date_trunc('month', draw_date) DESC
		 LIMIT 6`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue data: %w", err)
	}
	defer rows.Close()

	var points []models.ChartDataPoint
	for rows.Next() {
		var p models.ChartDataPoint
		if err := rows.Scan(&p.Label, &p.Value); err != nil {
			return nil, fmt.Errorf("failed to scan revenue point: %w", err)
		}
		points = append(points, p)
	}

	// Reverse to chronological order
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}
