package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReportRepository handles all report-related database queries.
type ReportRepository struct {
	pool *pgxpool.Pool
}

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

// UserReportRow represents a single row in the user report.
type UserReportRow struct {
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	SubscriptionActive bool      `json:"subscription_active"`
	RoundsPlayed       int       `json:"rounds_played"`
	BestScore          int       `json:"best_score"`
	TotalWinnings      float64   `json:"total_winnings"`
	CreatedAt          time.Time `json:"created_at"`
}

// GetUserReport returns all users with their performance summary.
func (r *ReportRepository) GetUserReport(ctx context.Context, from, to string) ([]UserReportRow, error) {
	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if from != "" && to != "" {
		whereClause = fmt.Sprintf("WHERE u.created_at >= $%d AND u.created_at <= $%d", argIdx, argIdx+1)
		args = append(args, from, to)
	}

	query := fmt.Sprintf(`
		SELECT u.name, u.email, u.role, u.subscription_active,
		       COALESCE(s.rounds, 0), COALESCE(s.best, 0),
		       COALESCE(w.total, 0), u.created_at
		FROM users u
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as rounds, COALESCE(MAX(score), 0) as best
			FROM scores WHERE user_id = u.id
		) s ON true
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(prize_amount), 0) as total
			FROM winners WHERE user_id = u.id AND verification_status = 'approved'
		) w ON true
		%s
		ORDER BY u.created_at DESC`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user report: %w", err)
	}
	defer rows.Close()

	var result []UserReportRow
	for rows.Next() {
		var row UserReportRow
		if err := rows.Scan(&row.Name, &row.Email, &row.Role, &row.SubscriptionActive,
			&row.RoundsPlayed, &row.BestScore, &row.TotalWinnings, &row.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user report row: %w", err)
		}
		result = append(result, row)
	}
	return result, nil
}

// RevenueReportRow represents a row in the revenue report.
type RevenueReportRow struct {
	Month         string  `json:"month"`
	Year          int     `json:"year"`
	TotalPool     float64 `json:"total_pool"`
	WinnerPrize   float64 `json:"winner_prize"`
	CharityAmount float64 `json:"charity_amount"`
	PlatformFee   float64 `json:"platform_fee"`
	TotalEntries  int     `json:"total_entries"`
}

// GetRevenueReport returns monthly revenue breakdown.
func (r *ReportRepository) GetRevenueReport(ctx context.Context, from, to string) ([]RevenueReportRow, error) {
	whereClause := "WHERE d.status = 'completed'"
	args := []interface{}{}
	argIdx := 1

	if from != "" && to != "" {
		whereClause += fmt.Sprintf(" AND d.draw_date >= $%d AND d.draw_date <= $%d", argIdx, argIdx+1)
		args = append(args, from, to)
	}

	query := fmt.Sprintf(`
		SELECT TO_CHAR(d.draw_date, 'Month'), d.year,
		       d.total_pool, d.winner_prize, d.charity_amount, d.platform_fee, d.total_entries
		FROM draws d
		%s
		ORDER BY d.year DESC, d.month DESC`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue report: %w", err)
	}
	defer rows.Close()

	var result []RevenueReportRow
	for rows.Next() {
		var row RevenueReportRow
		if err := rows.Scan(&row.Month, &row.Year, &row.TotalPool, &row.WinnerPrize,
			&row.CharityAmount, &row.PlatformFee, &row.TotalEntries); err != nil {
			return nil, fmt.Errorf("failed to scan revenue row: %w", err)
		}
		result = append(result, row)
	}
	return result, nil
}

// DrawReportRow represents a row in the draw report.
type DrawReportRow struct {
	DrawDate     time.Time `json:"draw_date"`
	Month        int       `json:"month"`
	Year         int       `json:"year"`
	Status       string    `json:"status"`
	TotalPool    float64   `json:"total_pool"`
	WinnerPrize  float64   `json:"winner_prize"`
	WinnerName   string    `json:"winner_name"`
	WinnerEmail  string    `json:"winner_email"`
	TotalEntries int       `json:"total_entries"`
}

// GetDrawReport returns detailed draw history.
func (r *ReportRepository) GetDrawReport(ctx context.Context, from, to string) ([]DrawReportRow, error) {
	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if from != "" && to != "" {
		whereClause = fmt.Sprintf("WHERE d.draw_date >= $%d AND d.draw_date <= $%d", argIdx, argIdx+1)
		args = append(args, from, to)
	}

	query := fmt.Sprintf(`
		SELECT d.draw_date, d.month, d.year, d.status, d.total_pool, d.winner_prize,
		       COALESCE(u.name, '-'), COALESCE(u.email, '-'), d.total_entries
		FROM draws d
		LEFT JOIN winners w ON w.draw_id = d.id
		LEFT JOIN users u ON u.id = w.user_id
		%s
		ORDER BY d.draw_date DESC`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get draw report: %w", err)
	}
	defer rows.Close()

	var result []DrawReportRow
	for rows.Next() {
		var row DrawReportRow
		if err := rows.Scan(&row.DrawDate, &row.Month, &row.Year, &row.Status,
			&row.TotalPool, &row.WinnerPrize, &row.WinnerName, &row.WinnerEmail, &row.TotalEntries); err != nil {
			return nil, fmt.Errorf("failed to scan draw row: %w", err)
		}
		result = append(result, row)
	}
	return result, nil
}

// CharityReportRow represents a row in the charity report.
type CharityReportRow struct {
	CharityName   string  `json:"charity_name"`
	Website       string  `json:"website"`
	Active        bool    `json:"active"`
	TotalUsers    int     `json:"total_users"`
	AvgPercentage float64 `json:"avg_percentage"`
	TotalReceived float64 `json:"total_received"`
}

// GetCharityReport returns charity allocation breakdown.
func (r *ReportRepository) GetCharityReport(ctx context.Context) ([]CharityReportRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.name, c.website, c.active,
		       COALESCE(sel.user_count, 0),
		       COALESCE(sel.avg_pct, 0),
		       COALESCE(alloc.total_received, 0)
		FROM charities c
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as user_count, AVG(contribution_percentage)::numeric(5,1) as avg_pct
			FROM user_charity_selections WHERE charity_id = c.id
		) sel ON true
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(d.charity_amount), 0) as total_received
			FROM draws d WHERE d.status = 'completed'
		) alloc ON true
		ORDER BY sel.user_count DESC NULLS LAST`)
	if err != nil {
		return nil, fmt.Errorf("failed to get charity report: %w", err)
	}
	defer rows.Close()

	var result []CharityReportRow
	for rows.Next() {
		var row CharityReportRow
		if err := rows.Scan(&row.CharityName, &row.Website, &row.Active,
			&row.TotalUsers, &row.AvgPercentage, &row.TotalReceived); err != nil {
			return nil, fmt.Errorf("failed to scan charity row: %w", err)
		}
		result = append(result, row)
	}
	return result, nil
}
