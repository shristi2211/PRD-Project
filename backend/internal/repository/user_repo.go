package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"golf-score-lottery/backend/internal/models"
)

// Common repository errors
var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUserExists     = errors.New("user with this email already exists")
	ErrTokenNotFound  = errors.New("refresh token not found")
)

// UserRepository handles all database operations for users and refresh tokens.
// All methods use single parameterized queries — no N+1 patterns.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// CreateUser inserts a new user and returns the created user.
// Returns ErrUserExists if the email is already taken (unique constraint violation).
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash, name string) (*models.User, error) {
	user := &models.User{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name, role, subscription_active, subscription_type, created_at, updated_at)
		 VALUES ($1, $2, $3, 'user', false, 'free', NOW(), NOW())
		 RETURNING id, email, password_hash, name, role, subscription_active, subscription_type, created_at, updated_at`,
		email, passwordHash, name,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.Role, &user.SubscriptionActive, &user.SubscriptionType, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation (duplicate email)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by primary key.
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, subscription_active, subscription_type, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.Role, &user.SubscriptionActive, &user.SubscriptionType, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// SetSubscriptionStatus updates the user's active subscription status flag.
func (r *UserRepository) SetSubscriptionStatus(ctx context.Context, id uuid.UUID, status bool) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET subscription_active = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetSubscriptionType updates the user's subscription plan and active status.
func (r *UserRepository) SetSubscriptionType(ctx context.Context, id uuid.UUID, planType string) error {
	active := planType == "monthly" || planType == "yearly"
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET subscription_type = $1, subscription_active = $2, updated_at = NOW() WHERE id = $3`,
		planType, active, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription type: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// GetUserByEmail retrieves a user by email. Uses the idx_users_email index.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, subscription_active, subscription_type, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.Role, &user.SubscriptionActive, &user.SubscriptionType, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}


// StoreRefreshToken persists a hashed refresh token to the database.
func (r *UserRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID uuid.UUID, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, family_id, expires_at, revoked, created_at)
		 VALUES ($1, $2, $3, $4, false, NOW())`,
		userID, tokenHash, familyID, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}
	return nil
}

// GetRefreshTokenByHash retrieves a refresh token record by its hash.
func (r *UserRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	rt := &models.RefreshToken{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, family_id, expires_at, revoked, created_at
		 FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.FamilyID,
		&rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return rt, nil
}

// RevokeRefreshToken marks a single refresh token as revoked.
func (r *UserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}

// RevokeTokenFamily revokes all tokens in a family (token reuse detection).
func (r *UserRepository) RevokeTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE family_id = $1`,
		familyID,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke token family: %w", err)
	}
	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user (logout from all devices).
func (r *UserRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}
	return nil
}

// CleanupExpiredTokens removes expired tokens older than 30 days.
// Intended to be called periodically (e.g., background goroutine).
func (r *UserRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < NOW() - INTERVAL '30 days'`,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return result.RowsAffected(), nil
}

// UpdateUser updates a user's name and/or email. Returns the updated user.
func (r *UserRepository) UpdateUser(ctx context.Context, id uuid.UUID, name, email string) (*models.User, error) {
	user := &models.User{}
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET name = $1, email = $2, updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, email, password_hash, name, role, subscription_active, subscription_type, created_at, updated_at`,
		name, email, id,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.Role, &user.SubscriptionActive, &user.SubscriptionType, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		// Check for unique constraint violation (duplicate email)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// UpdateUserPassword updates only the password hash for a user.
func (r *UserRepository) UpdateUserPassword(ctx context.Context, id uuid.UUID, newPasswordHash string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		newPasswordHash, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUser hard-deletes a user. Refresh tokens are cascade-deleted by FK constraint.
func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM users WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ListUsers returns a paginated, searchable, filterable list of users for admin.
// search: ILIKE on name or email. statusFilter: "active", "inactive", or "" for all.
func (r *UserRepository) ListUsers(ctx context.Context, page, pageSize int, search, statusFilter string) ([]models.User, int, error) {
	// Build dynamic WHERE clause
	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	if statusFilter == "active" {
		conditions = append(conditions, fmt.Sprintf("subscription_active = $%d", argIdx))
		args = append(args, true)
		argIdx++
	} else if statusFilter == "inactive" {
		conditions = append(conditions, fmt.Sprintf("subscription_active = $%d", argIdx))
		args = append(args, false)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	// Count total matching rows
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Fetch page
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(
		`SELECT id, email, password_hash, name, role, subscription_active, subscription_type, created_at, updated_at
		 FROM users %s
		 ORDER BY created_at DESC
		 LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1,
	)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.Name,
			&u.Role, &u.SubscriptionActive, &u.SubscriptionType, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, total, nil
}

// ToggleUserActivation sets the subscription_active flag for a user.
func (r *UserRepository) ToggleUserActivation(ctx context.Context, id uuid.UUID, active bool) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET subscription_active = $1, updated_at = NOW() WHERE id = $2`,
		active, id,
	)
	if err != nil {
		return fmt.Errorf("failed to toggle user activation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SaveIpSubscription registers an IP address with a selected plan.
func (r *UserRepository) SaveIpSubscription(ctx context.Context, ip, plan string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ip_subscriptions (ip_address, plan_type, created_at, updated_at) 
		 VALUES ($1, $2, NOW(), NOW())
		 ON CONFLICT (ip_address) DO UPDATE SET plan_type = EXCLUDED.plan_type, updated_at = NOW()`,
		ip, plan,
	)
	if err != nil {
		return fmt.Errorf("failed to save IP subscription: %w", err)
	}
	return nil
}

// GetIpSubscription returns the plan type for an IP if it exists.
func (r *UserRepository) GetIpSubscription(ctx context.Context, ip string) (string, error) {
	var plan string
	err := r.pool.QueryRow(ctx,
		`SELECT plan_type FROM ip_subscriptions WHERE ip_address = $1`,
		ip,
	).Scan(&plan)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil // Not subscribed
		}
		return "", fmt.Errorf("failed to check IP subscription: %w", err)
	}
	return plan, nil
}
