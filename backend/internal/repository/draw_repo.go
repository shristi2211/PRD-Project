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
	ErrDrawNotFound      = errors.New("draw not found")
	ErrDrawAlreadyExists = errors.New("draw already exists for this month/year")
)

// DrawRepository handles all database operations for draws.
type DrawRepository struct {
	pool *pgxpool.Pool
}

func NewDrawRepository(pool *pgxpool.Pool) *DrawRepository {
	return &DrawRepository{pool: pool}
}

// CreateDraw creates a new draw record.
func (r *DrawRepository) CreateDraw(ctx context.Context, month, year int, totalPool, winnerPrize, charityAmount, platformFee float64, totalEntries int) (*models.Draw, error) {
	d := &models.Draw{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO draws (draw_date, month, year, status, total_pool, winner_prize, charity_amount, platform_fee, total_entries, created_at)
		 VALUES (NOW(), $1, $2, 'completed', $3, $4, $5, $6, $7, NOW())
		 RETURNING id, draw_date, month, year, status, total_pool, winner_prize, charity_amount, platform_fee, total_entries, created_at`,
		month, year, totalPool, winnerPrize, charityAmount, platformFee, totalEntries,
	).Scan(&d.ID, &d.DrawDate, &d.Month, &d.Year, &d.Status, &d.TotalPool, &d.WinnerPrize,
		&d.CharityAmount, &d.PlatformFee, &d.TotalEntries, &d.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create draw: %w", err)
	}
	return d, nil
}

// GetDrawByID retrieves a draw by ID.
func (r *DrawRepository) GetDrawByID(ctx context.Context, id uuid.UUID) (*models.Draw, error) {
	d := &models.Draw{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, draw_date, month, year, status, total_pool, winner_prize, charity_amount, platform_fee, total_entries, created_at
		 FROM draws WHERE id = $1`,
		id,
	).Scan(&d.ID, &d.DrawDate, &d.Month, &d.Year, &d.Status, &d.TotalPool, &d.WinnerPrize,
		&d.CharityAmount, &d.PlatformFee, &d.TotalEntries, &d.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDrawNotFound
		}
		return nil, fmt.Errorf("failed to get draw: %w", err)
	}
	return d, nil
}

// GetDrawByMonthYear checks if a draw already exists for a month/year.
func (r *DrawRepository) GetDrawByMonthYear(ctx context.Context, month, year int) (*models.Draw, error) {
	d := &models.Draw{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, draw_date, month, year, status, total_pool, winner_prize, charity_amount, platform_fee, total_entries, created_at
		 FROM draws WHERE month = $1 AND year = $2`,
		month, year,
	).Scan(&d.ID, &d.DrawDate, &d.Month, &d.Year, &d.Status, &d.TotalPool, &d.WinnerPrize,
		&d.CharityAmount, &d.PlatformFee, &d.TotalEntries, &d.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No draw exists
		}
		return nil, fmt.Errorf("failed to check draw: %w", err)
	}
	return d, nil
}

// ListDraws returns a paginated list of draws.
func (r *DrawRepository) ListDraws(ctx context.Context, page, pageSize int) ([]models.Draw, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM draws`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count draws: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx,
		`SELECT id, draw_date, month, year, status, total_pool, winner_prize, charity_amount, platform_fee, total_entries, created_at
		 FROM draws ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list draws: %w", err)
	}
	defer rows.Close()

	var draws []models.Draw
	for rows.Next() {
		var d models.Draw
		if err := rows.Scan(&d.ID, &d.DrawDate, &d.Month, &d.Year, &d.Status, &d.TotalPool, &d.WinnerPrize,
			&d.CharityAmount, &d.PlatformFee, &d.TotalEntries, &d.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan draw: %w", err)
		}
		draws = append(draws, d)
	}
	return draws, total, nil
}

// CreateDrawEntry inserts a single draw entry.
func (r *DrawRepository) CreateDrawEntry(ctx context.Context, drawID, userID, scoreID uuid.UUID, entryScore int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO draw_entries (draw_id, user_id, score_id, entry_score, created_at)
		 VALUES ($1, $2, $3, $4, NOW())`,
		drawID, userID, scoreID, entryScore,
	)
	if err != nil {
		return fmt.Errorf("failed to create draw entry: %w", err)
	}
	return nil
}

// GetDrawEntries returns all entries for a draw with user info.
func (r *DrawRepository) GetDrawEntries(ctx context.Context, drawID uuid.UUID) ([]models.DrawEntryResponse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT de.id, de.user_id, u.name, u.email, de.entry_score
		 FROM draw_entries de
		 JOIN users u ON u.id = de.user_id
		 WHERE de.draw_id = $1
		 ORDER BY de.entry_score DESC`,
		drawID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get draw entries: %w", err)
	}
	defer rows.Close()

	var entries []models.DrawEntryResponse
	for rows.Next() {
		var e models.DrawEntryResponse
		if err := rows.Scan(&e.ID, &e.UserID, &e.UserName, &e.UserEmail, &e.EntryScore); err != nil {
			return nil, fmt.Errorf("failed to scan draw entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
