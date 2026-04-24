package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"golf-score-lottery/backend/internal/models"
)

var (
	ErrScoreNotFound = errors.New("score not found")
)

// ScoreRepository handles all database operations for scores.
type ScoreRepository struct {
	pool *pgxpool.Pool
}

func NewScoreRepository(pool *pgxpool.Pool) *ScoreRepository {
	return &ScoreRepository{pool: pool}
}

// CreateScore inserts a new score and returns it.
func (r *ScoreRepository) CreateScore(ctx context.Context, userID uuid.UUID, score int, roundDate string, notes string) (*models.Score, error) {
	s := &models.Score{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO scores (user_id, score, round_date, notes, created_at)
		 VALUES ($1, $2, $3::date, $4, NOW())
		 RETURNING id, user_id, score, round_date, notes, created_at`,
		userID, score, roundDate, notes,
	).Scan(&s.ID, &s.UserID, &s.Score, &s.RoundDate, &s.Notes, &s.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create score: %w", err)
	}
	return s, nil
}

// GetScoresByUserID returns all scores for a user ordered by date descending.
func (r *ScoreRepository) GetScoresByUserID(ctx context.Context, userID uuid.UUID) ([]models.Score, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, score, round_date, notes, created_at
		 FROM scores WHERE user_id = $1
		 ORDER BY round_date DESC, created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get scores: %w", err)
	}
	defer rows.Close()

	var scores []models.Score
	for rows.Next() {
		var s models.Score
		if err := rows.Scan(&s.ID, &s.UserID, &s.Score, &s.RoundDate, &s.Notes, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan score: %w", err)
		}
		scores = append(scores, s)
	}
	return scores, nil
}

// GetScoreByID retrieves a single score by ID.
func (r *ScoreRepository) GetScoreByID(ctx context.Context, id uuid.UUID) (*models.Score, error) {
	s := &models.Score{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, score, round_date, notes, created_at
		 FROM scores WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.UserID, &s.Score, &s.RoundDate, &s.Notes, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrScoreNotFound
		}
		return nil, fmt.Errorf("failed to get score: %w", err)
	}
	return s, nil
}

// DeleteScore deletes a score by ID.
func (r *ScoreRepository) DeleteScore(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM scores WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete score: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrScoreNotFound
	}
	return nil
}

// UpdateScore modifies an existing score.
func (r *ScoreRepository) UpdateScore(ctx context.Context, id uuid.UUID, score int, roundDate string, notes string) (*models.Score, error) {
	s := &models.Score{}
	err := r.pool.QueryRow(ctx,
		`UPDATE scores SET score = $1, round_date = $2::date, notes = $3
		 WHERE id = $4
		 RETURNING id, user_id, score, round_date, notes, created_at`,
		score, roundDate, notes, id,
	).Scan(&s.ID, &s.UserID, &s.Score, &s.RoundDate, &s.Notes, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrScoreNotFound
		}
		return nil, fmt.Errorf("failed to update score: %w", err)
	}
	return s, nil
}

// CheckScoreExistsForDate returns true if a user already has a score logged on the provided date.
func (r *ScoreRepository) CheckScoreExistsForDate(ctx context.Context, userID uuid.UUID, roundDate string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM scores WHERE user_id = $1 AND round_date = $2::date`,
		userID, roundDate,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check date constraint: %w", err)
	}
	return count > 0, nil
}

// DeleteOldestScoreByUserID deletes the oldest score for a specific user to make room for new ones.
func (r *ScoreRepository) DeleteOldestScoreByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM scores WHERE id IN (
			SELECT id FROM scores WHERE user_id = $1 ORDER BY round_date ASC, created_at ASC LIMIT 1
		)`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete oldest score: %w", err)
	}
	return nil
}

// CountUserScores returns the number of scores a user has.
func (r *ScoreRepository) CountUserScores(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM scores WHERE user_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count scores: %w", err)
	}
	return count, nil
}

// GetBestScoreByUserID returns the highest score for a user.
func (r *ScoreRepository) GetBestScoreByUserID(ctx context.Context, userID uuid.UUID) (*models.Score, error) {
	s := &models.Score{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, score, round_date, notes, created_at
		 FROM scores WHERE user_id = $1
		 ORDER BY score DESC LIMIT 1`,
		userID,
	).Scan(&s.ID, &s.UserID, &s.Score, &s.RoundDate, &s.Notes, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrScoreNotFound
		}
		return nil, fmt.Errorf("failed to get best score: %w", err)
	}
	return s, nil
}

// ListAllScores returns a paginated list of all scores with user info (admin).
func (r *ScoreRepository) ListAllScores(ctx context.Context, page, pageSize int, search string) ([]models.ScoreWithUser, int, error) {
	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(u.name ILIKE $%d OR u.email ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
	}

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM scores s JOIN users u ON s.user_id = u.id %s", whereClause)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count scores: %w", err)
	}

	// Fetch
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(
		`SELECT s.id, s.user_id, u.name, u.email, s.score, s.round_date, s.notes, s.created_at
		 FROM scores s JOIN users u ON s.user_id = u.id
		 %s ORDER BY s.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1,
	)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list scores: %w", err)
	}
	defer rows.Close()

	var scores []models.ScoreWithUser
	for rows.Next() {
		var s models.ScoreWithUser
		var rd interface{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.UserName, &s.UserEmail, &s.Score, &rd, &s.Notes, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan score: %w", err)
		}
		if t, ok := rd.(interface{ Format(string) string }); ok {
			s.RoundDate = t.Format("2006-01-02")
		}
		scores = append(scores, s)
	}
	return scores, total, nil
}

// GetEligibleUsersForDraw returns users with active subscription and at least one score.
func (r *ScoreRepository) GetEligibleUsersForDraw(ctx context.Context) ([]struct {
	UserID    uuid.UUID
	UserName  string
	UserEmail string
	BestScore int
	ScoreID   uuid.UUID
}, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT ON (u.id) u.id, u.name, u.email, s.score, s.id
		 FROM users u
		 JOIN scores s ON s.user_id = u.id
		 WHERE u.subscription_active = true AND u.role = 'user'
		 ORDER BY u.id, s.score DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get eligible users: %w", err)
	}
	defer rows.Close()

	type eligibleUser struct {
		UserID    uuid.UUID
		UserName  string
		UserEmail string
		BestScore int
		ScoreID   uuid.UUID
	}
	var users []struct {
		UserID    uuid.UUID
		UserName  string
		UserEmail string
		BestScore int
		ScoreID   uuid.UUID
	}
	for rows.Next() {
		var u eligibleUser
		if err := rows.Scan(&u.UserID, &u.UserName, &u.UserEmail, &u.BestScore, &u.ScoreID); err != nil {
			return nil, fmt.Errorf("failed to scan eligible user: %w", err)
		}
		users = append(users, struct {
			UserID    uuid.UUID
			UserName  string
			UserEmail string
			BestScore int
			ScoreID   uuid.UUID
		}(u))
	}
	return users, nil
}
