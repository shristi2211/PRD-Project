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
	ErrCharityNotFound = errors.New("charity not found")
)

// CharityRepository handles all database operations for charities.
type CharityRepository struct {
	pool *pgxpool.Pool
}

func NewCharityRepository(pool *pgxpool.Pool) *CharityRepository {
	return &CharityRepository{pool: pool}
}

// CreateCharity inserts a new charity.
func (r *CharityRepository) CreateCharity(ctx context.Context, name, description, website, logoURL string) (*models.Charity, error) {
	c := &models.Charity{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO charities (name, description, website, logo_url, active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		 RETURNING id, name, description, website, logo_url, active, created_at, updated_at`,
		name, description, website, logoURL,
	).Scan(&c.ID, &c.Name, &c.Description, &c.Website, &c.LogoURL, &c.Active, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create charity: %w", err)
	}
	return c, nil
}

// UpdateCharity updates a charity's details.
func (r *CharityRepository) UpdateCharity(ctx context.Context, id uuid.UUID, name, description, website, logoURL string) (*models.Charity, error) {
	c := &models.Charity{}
	err := r.pool.QueryRow(ctx,
		`UPDATE charities SET name = $1, description = $2, website = $3, logo_url = $4, updated_at = NOW()
		 WHERE id = $5
		 RETURNING id, name, description, website, logo_url, active, created_at, updated_at`,
		name, description, website, logoURL, id,
	).Scan(&c.ID, &c.Name, &c.Description, &c.Website, &c.LogoURL, &c.Active, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCharityNotFound
		}
		return nil, fmt.Errorf("failed to update charity: %w", err)
	}
	return c, nil
}

// ToggleCharityActive toggles a charity's active status.
func (r *CharityRepository) ToggleCharityActive(ctx context.Context, id uuid.UUID, active bool) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE charities SET active = $1, updated_at = NOW() WHERE id = $2`,
		active, id,
	)
	if err != nil {
		return fmt.Errorf("failed to toggle charity: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrCharityNotFound
	}
	return nil
}

// GetCharityByID retrieves a single charity.
func (r *CharityRepository) GetCharityByID(ctx context.Context, id uuid.UUID) (*models.Charity, error) {
	c := &models.Charity{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, website, logo_url, active, created_at, updated_at
		 FROM charities WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.Name, &c.Description, &c.Website, &c.LogoURL, &c.Active, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCharityNotFound
		}
		return nil, fmt.Errorf("failed to get charity: %w", err)
	}
	return c, nil
}

// ListCharities returns all charities, optionally filtered by active status.
func (r *CharityRepository) ListCharities(ctx context.Context, activeOnly bool) ([]models.Charity, error) {
	query := `SELECT id, name, description, website, logo_url, active, created_at, updated_at FROM charities`
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list charities: %w", err)
	}
	defer rows.Close()

	var charities []models.Charity
	for rows.Next() {
		var c models.Charity
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Website, &c.LogoURL, &c.Active, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan charity: %w", err)
		}
		charities = append(charities, c)
	}
	return charities, nil
}

// SetUserCharityAllocations sets a user's charity selections inside a transaction.
func (r *CharityRepository) SetUserCharityAllocations(ctx context.Context, userID uuid.UUID, allocations []struct {
	CharityID uuid.UUID
	Percentage int
}) ([]models.UserCharitySelection, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing allocations
	_, err = tx.Exec(ctx, `DELETE FROM user_charity_selections WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete prior allocations: %w", err)
	}

	var results []models.UserCharitySelection
	for _, alloc := range allocations {
		sel := models.UserCharitySelection{}
		err := tx.QueryRow(ctx,
			`INSERT INTO user_charity_selections (user_id, charity_id, contribution_percentage, selected_at)
			 VALUES ($1, $2, $3, NOW())
			 RETURNING id, user_id, charity_id, contribution_percentage, selected_at`,
			userID, alloc.CharityID, alloc.Percentage,
		).Scan(&sel.ID, &sel.UserID, &sel.CharityID, &sel.ContributionPercentage, &sel.SelectedAt)
		
		if err != nil {
			return nil, fmt.Errorf("failed to insert allocation for charity %s: %w", alloc.CharityID, err)
		}
		results = append(results, sel)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit allocations: %w", err)
	}
	return results, nil
}

// GetUserCharitySelection returns the user's current charity selection with charity details.
func (r *CharityRepository) GetUserCharitySelection(ctx context.Context, userID uuid.UUID) ([]models.UserCharitySelectionResponse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ucs.id, ucs.charity_id, c.name, c.description, ucs.contribution_percentage, ucs.selected_at
		 FROM user_charity_selections ucs
		 JOIN charities c ON c.id = ucs.charity_id
		 WHERE ucs.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get charity selection: %w", err)
	}
	defer rows.Close()

	var selections []models.UserCharitySelectionResponse
	for rows.Next() {
		sel := models.UserCharitySelectionResponse{}
		if err := rows.Scan(&sel.ID, &sel.CharityID, &sel.CharityName, &sel.CharityDescription, &sel.ContributionPercentage, &sel.SelectedAt); err != nil {
			return nil, fmt.Errorf("failed to scan selection: %w", err)
		}
		selections = append(selections, sel)
	}
	
	return selections, nil
}

// GetCharityDistribution returns aggregate charity selection stats.
func (r *CharityRepository) GetCharityDistribution(ctx context.Context) ([]models.CharityDistribution, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT c.id, c.name, COUNT(ucs.id) as user_count
		 FROM charities c
		 LEFT JOIN user_charity_selections ucs ON ucs.charity_id = c.id
		 WHERE c.active = true
		 GROUP BY c.id, c.name
		 ORDER BY user_count DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get charity distribution: %w", err)
	}
	defer rows.Close()

	var totalUsers int
	var distributions []models.CharityDistribution
	for rows.Next() {
		var d models.CharityDistribution
		if err := rows.Scan(&d.CharityID, &d.CharityName, &d.UserCount); err != nil {
			return nil, fmt.Errorf("failed to scan distribution: %w", err)
		}
		totalUsers += d.UserCount
		distributions = append(distributions, d)
	}

	// Calculate percentages
	if totalUsers > 0 {
		for i := range distributions {
			distributions[i].Percentage = float64(distributions[i].UserCount) / float64(totalUsers) * 100
		}
	}

	return distributions, nil
}
