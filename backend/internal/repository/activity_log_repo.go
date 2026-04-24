package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"golf-score-lottery/backend/internal/models"
)

// ActivityLogRepository handles all database operations for activity logs.
type ActivityLogRepository struct {
	pool *pgxpool.Pool
}

func NewActivityLogRepository(pool *pgxpool.Pool) *ActivityLogRepository {
	return &ActivityLogRepository{pool: pool}
}

// LogActivity inserts a new activity log entry.
func (r *ActivityLogRepository) LogActivity(ctx context.Context, userID *uuid.UUID, action, entityType, entityID string, metadata map[string]interface{}, ipAddress string) error {
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO activity_logs (user_id, action, entity_type, entity_id, metadata, ip_address, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		userID, action, entityType, entityID, metaJSON, ipAddress,
	)
	if err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}
	return nil
}

// GetActivityLogs returns paginated activity logs with optional filters.
func (r *ActivityLogRepository) GetActivityLogs(ctx context.Context, page, pageSize int, userIDFilter string, actionFilter string) ([]models.ActivityLogResponse, int, error) {
	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	if userIDFilter != "" {
		conditions = append(conditions, fmt.Sprintf("al.user_id = $%d", argIdx))
		uid, err := uuid.Parse(userIDFilter)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid user ID filter: %w", err)
		}
		args = append(args, uid)
		argIdx++
	}

	if actionFilter != "" {
		conditions = append(conditions, fmt.Sprintf("al.action ILIKE $%d", argIdx))
		args = append(args, "%"+actionFilter+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM activity_logs al %s", whereClause)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count logs: %w", err)
	}

	// Fetch
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(
		`SELECT al.id, al.user_id, COALESCE(u.name, 'System'), COALESCE(u.email, ''),
		        al.action, al.entity_type, al.entity_id, al.metadata, al.ip_address, al.created_at
		 FROM activity_logs al
		 LEFT JOIN users u ON u.id = al.user_id
		 %s ORDER BY al.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1,
	)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get logs: %w", err)
	}
	defer rows.Close()

	var logs []models.ActivityLogResponse
	for rows.Next() {
		var l models.ActivityLogResponse
		var metaBytes []byte
		if err := rows.Scan(&l.ID, &l.UserID, &l.UserName, &l.UserEmail,
			&l.Action, &l.EntityType, &l.EntityID, &metaBytes, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan log: %w", err)
		}
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &l.Metadata)
		}
		if l.Metadata == nil {
			l.Metadata = map[string]interface{}{}
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}

// ActivityUserSummary represents a user with their activity count.
type ActivityUserSummary struct {
	UserID       uuid.UUID `json:"user_id"`
	UserName     string    `json:"user_name"`
	UserEmail    string    `json:"user_email"`
	ActivityCount int      `json:"activity_count"`
	LastActivity string    `json:"last_activity"`
}

// GetUsersWithActivity returns all users who have activity logs.
func (r *ActivityLogRepository) GetUsersWithActivity(ctx context.Context) ([]ActivityUserSummary, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT al.user_id, COALESCE(u.name, 'System'), COALESCE(u.email, ''),
		        COUNT(*) as activity_count,
		        MAX(al.created_at)::text as last_activity
		 FROM activity_logs al
		 LEFT JOIN users u ON u.id = al.user_id
		 WHERE al.user_id IS NOT NULL
		 GROUP BY al.user_id, u.name, u.email
		 ORDER BY MAX(al.created_at) DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to get users with activity: %w", err)
	}
	defer rows.Close()

	var users []ActivityUserSummary
	for rows.Next() {
		var u ActivityUserSummary
		if err := rows.Scan(&u.UserID, &u.UserName, &u.UserEmail, &u.ActivityCount, &u.LastActivity); err != nil {
			return nil, fmt.Errorf("failed to scan user summary: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

// ActivityYearSummary represents a year with activity count for a user.
type ActivityYearSummary struct {
	Year          int `json:"year"`
	ActivityCount int `json:"activity_count"`
}

// GetUserActivityYears returns years with activity for a specific user.
func (r *ActivityLogRepository) GetUserActivityYears(ctx context.Context, userID uuid.UUID) ([]ActivityYearSummary, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT EXTRACT(YEAR FROM created_at)::int as year, COUNT(*) as activity_count
		 FROM activity_logs
		 WHERE user_id = $1
		 GROUP BY EXTRACT(YEAR FROM created_at)
		 ORDER BY year DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity years: %w", err)
	}
	defer rows.Close()

	var years []ActivityYearSummary
	for rows.Next() {
		var y ActivityYearSummary
		if err := rows.Scan(&y.Year, &y.ActivityCount); err != nil {
			return nil, fmt.Errorf("failed to scan year summary: %w", err)
		}
		years = append(years, y)
	}
	return years, nil
}

// ActivityMonthSummary represents a month with activity count.
type ActivityMonthSummary struct {
	Month         int    `json:"month"`
	MonthName     string `json:"month_name"`
	ActivityCount int    `json:"activity_count"`
}

// GetUserActivityMonths returns months with activity for a specific user and year.
func (r *ActivityLogRepository) GetUserActivityMonths(ctx context.Context, userID uuid.UUID, year int) ([]ActivityMonthSummary, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT EXTRACT(MONTH FROM created_at)::int as month,
		        TO_CHAR(created_at, 'Month') as month_name,
		        COUNT(*) as activity_count
		 FROM activity_logs
		 WHERE user_id = $1 AND EXTRACT(YEAR FROM created_at) = $2
		 GROUP BY EXTRACT(MONTH FROM created_at), TO_CHAR(created_at, 'Month')
		 ORDER BY month DESC`, userID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity months: %w", err)
	}
	defer rows.Close()

	var months []ActivityMonthSummary
	for rows.Next() {
		var m ActivityMonthSummary
		if err := rows.Scan(&m.Month, &m.MonthName, &m.ActivityCount); err != nil {
			return nil, fmt.Errorf("failed to scan month summary: %w", err)
		}
		months = append(months, m)
	}
	return months, nil
}

// GetUserMonthlyActivities returns detailed activity logs for a user in a specific month/year.
func (r *ActivityLogRepository) GetUserMonthlyActivities(ctx context.Context, userID uuid.UUID, year, month int) ([]models.ActivityLogResponse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT al.id, al.user_id, COALESCE(u.name, 'System'), COALESCE(u.email, ''),
		        al.action, al.entity_type, al.entity_id, al.metadata, al.ip_address, al.created_at
		 FROM activity_logs al
		 LEFT JOIN users u ON u.id = al.user_id
		 WHERE al.user_id = $1
		   AND EXTRACT(YEAR FROM al.created_at) = $2
		   AND EXTRACT(MONTH FROM al.created_at) = $3
		 ORDER BY al.created_at DESC`, userID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly activities: %w", err)
	}
	defer rows.Close()

	var logs []models.ActivityLogResponse
	for rows.Next() {
		var l models.ActivityLogResponse
		var metaBytes []byte
		if err := rows.Scan(&l.ID, &l.UserID, &l.UserName, &l.UserEmail,
			&l.Action, &l.EntityType, &l.EntityID, &metaBytes, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &l.Metadata)
		}
		if l.Metadata == nil {
			l.Metadata = map[string]interface{}{}
		}
		logs = append(logs, l)
	}
	return logs, nil
}
