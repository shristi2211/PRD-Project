package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"golf-score-lottery/backend/internal/models"
)

var (
	ErrWinnerNotFound = errors.New("winner not found")
)

// WinnerRepository handles all database operations for winners.
type WinnerRepository struct {
	pool *pgxpool.Pool
}

func NewWinnerRepository(pool *pgxpool.Pool) *WinnerRepository {
	return &WinnerRepository{pool: pool}
}

// CreateWinner inserts a new winner record.
func (r *WinnerRepository) CreateWinner(ctx context.Context, drawID, userID uuid.UUID, prizeAmount float64) (*models.Winner, error) {
	w := &models.Winner{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO winners (draw_id, user_id, prize_amount, verification_status, created_at)
		 VALUES ($1, $2, $3, 'pending', NOW())
		 RETURNING id, draw_id, user_id, prize_amount, proof_url, proof_notes, verification_status, rejection_reason, verified_by, verified_at, created_at`,
		drawID, userID, prizeAmount,
	).Scan(&w.ID, &w.DrawID, &w.UserID, &w.PrizeAmount, &w.ProofURL, &w.ProofNotes,
		&w.VerificationStatus, &w.RejectionReason, &w.VerifiedBy, &w.VerifiedAt, &w.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create winner: %w", err)
	}
	return w, nil
}

// GetWinnerByID retrieves a winner by ID with user info.
func (r *WinnerRepository) GetWinnerByID(ctx context.Context, id uuid.UUID) (*models.WinnerResponse, error) {
	w := &models.WinnerResponse{}
	err := r.pool.QueryRow(ctx,
		`SELECT w.id, w.draw_id, w.user_id, u.name, u.email, d.month, d.year,
		        w.prize_amount, w.proof_url, w.proof_notes, w.verification_status,
		        w.rejection_reason, w.verified_by, w.verified_at, w.created_at
		 FROM winners w
		 JOIN users u ON u.id = w.user_id
		 JOIN draws d ON d.id = w.draw_id
		 WHERE w.id = $1`,
		id,
	).Scan(&w.ID, &w.DrawID, &w.UserID, &w.UserName, &w.UserEmail, &w.DrawMonth, &w.DrawYear,
		&w.PrizeAmount, &w.ProofURL, &w.ProofNotes, &w.VerificationStatus,
		&w.RejectionReason, &w.VerifiedBy, &w.VerifiedAt, &w.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWinnerNotFound
		}
		return nil, fmt.Errorf("failed to get winner: %w", err)
	}
	return w, nil
}

// GetWinnersByDrawID returns all winners for a draw.
func (r *WinnerRepository) GetWinnersByDrawID(ctx context.Context, drawID uuid.UUID) ([]models.WinnerResponse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT w.id, w.draw_id, w.user_id, u.name, u.email, d.month, d.year,
		        w.prize_amount, w.proof_url, w.proof_notes, w.verification_status,
		        w.rejection_reason, w.verified_by, w.verified_at, w.created_at
		 FROM winners w
		 JOIN users u ON u.id = w.user_id
		 JOIN draws d ON d.id = w.draw_id
		 WHERE w.draw_id = $1
		 ORDER BY w.created_at DESC`,
		drawID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get winners: %w", err)
	}
	defer rows.Close()

	var winners []models.WinnerResponse
	for rows.Next() {
		var w models.WinnerResponse
		if err := rows.Scan(&w.ID, &w.DrawID, &w.UserID, &w.UserName, &w.UserEmail, &w.DrawMonth, &w.DrawYear,
			&w.PrizeAmount, &w.ProofURL, &w.ProofNotes, &w.VerificationStatus,
			&w.RejectionReason, &w.VerifiedBy, &w.VerifiedAt, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan winner: %w", err)
		}
		winners = append(winners, w)
	}
	return winners, nil
}

// GetWinnersByUserID returns all wins for a specific user.
func (r *WinnerRepository) GetWinnersByUserID(ctx context.Context, userID uuid.UUID) ([]models.WinnerResponse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT w.id, w.draw_id, w.user_id, u.name, u.email, d.month, d.year,
		        w.prize_amount, w.proof_url, w.proof_notes, w.verification_status,
		        w.rejection_reason, w.verified_by, w.verified_at, w.created_at
		 FROM winners w
		 JOIN users u ON u.id = w.user_id
		 JOIN draws d ON d.id = w.draw_id
		 WHERE w.user_id = $1
		 ORDER BY w.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user winners: %w", err)
	}
	defer rows.Close()

	var winners []models.WinnerResponse
	for rows.Next() {
		var w models.WinnerResponse
		if err := rows.Scan(&w.ID, &w.DrawID, &w.UserID, &w.UserName, &w.UserEmail, &w.DrawMonth, &w.DrawYear,
			&w.PrizeAmount, &w.ProofURL, &w.ProofNotes, &w.VerificationStatus,
			&w.RejectionReason, &w.VerifiedBy, &w.VerifiedAt, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan winner: %w", err)
		}
		winners = append(winners, w)
	}
	return winners, nil
}

// UpdateWinnerProof updates the proof URL and notes for a winner.
func (r *WinnerRepository) UpdateWinnerProof(ctx context.Context, id uuid.UUID, proofURL, proofNotes string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE winners SET proof_url = $1, proof_notes = $2 WHERE id = $3`,
		proofURL, proofNotes, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update proof: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrWinnerNotFound
	}
	return nil
}

// UpdateWinnerVerification updates the verification status of a winner.
func (r *WinnerRepository) UpdateWinnerVerification(ctx context.Context, id uuid.UUID, status string, rejectionReason string, verifiedBy uuid.UUID) error {
	now := time.Now()
	result, err := r.pool.Exec(ctx,
		`UPDATE winners SET verification_status = $1, rejection_reason = $2, verified_by = $3, verified_at = $4
		 WHERE id = $5`,
		status, rejectionReason, verifiedBy, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update verification: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrWinnerNotFound
	}
	return nil
}

// GetPendingVerifications returns all winners with pending verification status.
func (r *WinnerRepository) GetPendingVerifications(ctx context.Context) ([]models.WinnerResponse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT w.id, w.draw_id, w.user_id, u.name, u.email, d.month, d.year,
		        w.prize_amount, w.proof_url, w.proof_notes, w.verification_status,
		        w.rejection_reason, w.verified_by, w.verified_at, w.created_at
		 FROM winners w
		 JOIN users u ON u.id = w.user_id
		 JOIN draws d ON d.id = w.draw_id
		 WHERE w.verification_status = 'pending'
		 ORDER BY w.created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending verifications: %w", err)
	}
	defer rows.Close()

	var winners []models.WinnerResponse
	for rows.Next() {
		var w models.WinnerResponse
		if err := rows.Scan(&w.ID, &w.DrawID, &w.UserID, &w.UserName, &w.UserEmail, &w.DrawMonth, &w.DrawYear,
			&w.PrizeAmount, &w.ProofURL, &w.ProofNotes, &w.VerificationStatus,
			&w.RejectionReason, &w.VerifiedBy, &w.VerifiedAt, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan pending winner: %w", err)
		}
		winners = append(winners, w)
	}
	return winners, nil
}

// CountPendingVerifications returns the count of pending winner verifications.
func (r *WinnerRepository) CountPendingVerifications(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM winners WHERE verification_status = 'pending'`,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending verifications: %w", err)
	}
	return count, nil
}
