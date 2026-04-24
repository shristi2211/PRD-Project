package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminRepository isolated for administrative setup tasks.
type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

// HasAdmin checks if any user with the 'admin' role already exists in the database.
func (r *AdminRepository) HasAdmin(ctx context.Context) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE role = 'admin')`,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check admin existence: %w", err)
	}

	return exists, nil
}

// CreateFirstAdmin creates the initial system administrator, bypassing normal registration rules.
func (r *AdminRepository) CreateFirstAdmin(ctx context.Context, email, passwordHash, name string) (string, error) {
	var adminID uuid.UUID
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name, role, subscription_active, created_at, updated_at)
		 VALUES ($1, $2, $3, 'admin', true, NOW(), NOW())
		 RETURNING id`,
		email, passwordHash, name,
	).Scan(&adminID)

	if err != nil {
		return "", fmt.Errorf("failed to create first admin: %w", err)
	}

	return adminID.String(), nil
}
