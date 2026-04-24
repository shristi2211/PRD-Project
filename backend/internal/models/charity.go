package models

import (
	"time"

	"github.com/google/uuid"
)

// Charity represents a charitable organization.
type Charity struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	LogoURL     string    `json:"logo_url"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CharityResponse is the API response for a charity.
type CharityResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	LogoURL     string    `json:"logo_url"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToResponse converts a Charity to CharityResponse.
func (c *Charity) ToResponse() CharityResponse {
	return CharityResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Website:     c.Website,
		LogoURL:     c.LogoURL,
		Active:      c.Active,
		CreatedAt:   c.CreatedAt,
	}
}

// CreateCharityRequest is the input DTO for creating a new charity.
type CreateCharityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Website     string `json:"website"`
	LogoURL     string `json:"logo_url"`
}

// UpdateCharityRequest is the input DTO for updating a charity.
type UpdateCharityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Website     string `json:"website"`
	LogoURL     string `json:"logo_url"`
}

// UserCharitySelection represents a user's selected charity.
type UserCharitySelection struct {
	ID                     uuid.UUID `json:"id"`
	UserID                 uuid.UUID `json:"user_id"`
	CharityID              uuid.UUID `json:"charity_id"`
	ContributionPercentage int       `json:"contribution_percentage"`
	SelectedAt             time.Time `json:"selected_at"`
}

// UserCharitySelectionResponse includes charity details.
type UserCharitySelectionResponse struct {
	ID                     uuid.UUID       `json:"id"`
	CharityID              uuid.UUID       `json:"charity_id"`
	CharityName            string          `json:"charity_name"`
	CharityDescription     string          `json:"charity_description"`
	ContributionPercentage int             `json:"contribution_percentage"`
	SelectedAt             time.Time       `json:"selected_at"`
}

// SelectCharityRequest is the input DTO for selecting multiple charities.
type SelectCharityRequest struct {
	Allocations []struct {
		CharityID              uuid.UUID `json:"charity_id"`
		ContributionPercentage int       `json:"contribution_percentage"`
	} `json:"allocations"`
}

// CharityDistribution represents a charity's share of user selections.
type CharityDistribution struct {
	CharityID   uuid.UUID `json:"charity_id"`
	CharityName string    `json:"charity_name"`
	UserCount   int       `json:"user_count"`
	Percentage  float64   `json:"percentage"`
}
