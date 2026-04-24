package models

import (
	"time"

	"github.com/google/uuid"
)

// Draw represents a monthly lottery draw.
type Draw struct {
	ID            uuid.UUID `json:"id"`
	DrawDate      time.Time `json:"draw_date"`
	Month         int       `json:"month"`
	Year          int       `json:"year"`
	Status        string    `json:"status"`
	TotalPool     float64   `json:"total_pool"`
	WinnerPrize   float64   `json:"winner_prize"`
	CharityAmount float64   `json:"charity_amount"`
	PlatformFee   float64   `json:"platform_fee"`
	TotalEntries  int       `json:"total_entries"`
	CreatedAt     time.Time `json:"created_at"`
}

// DrawResponse is the API response for a draw.
type DrawResponse struct {
	ID            uuid.UUID `json:"id"`
	DrawDate      time.Time `json:"draw_date"`
	Month         int       `json:"month"`
	Year          int       `json:"year"`
	Status        string    `json:"status"`
	TotalPool     float64   `json:"total_pool"`
	WinnerPrize   float64   `json:"winner_prize"`
	CharityAmount float64   `json:"charity_amount"`
	PlatformFee   float64   `json:"platform_fee"`
	TotalEntries  int       `json:"total_entries"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToResponse converts a Draw to DrawResponse.
func (d *Draw) ToResponse() DrawResponse {
	return DrawResponse{
		ID:            d.ID,
		DrawDate:      d.DrawDate,
		Month:         d.Month,
		Year:          d.Year,
		Status:        d.Status,
		TotalPool:     d.TotalPool,
		WinnerPrize:   d.WinnerPrize,
		CharityAmount: d.CharityAmount,
		PlatformFee:   d.PlatformFee,
		TotalEntries:  d.TotalEntries,
		CreatedAt:     d.CreatedAt,
	}
}

// DrawEntry represents a user's entry in a draw.
type DrawEntry struct {
	ID         uuid.UUID `json:"id"`
	DrawID     uuid.UUID `json:"draw_id"`
	UserID     uuid.UUID `json:"user_id"`
	ScoreID    uuid.UUID `json:"score_id"`
	EntryScore int       `json:"entry_score"`
	CreatedAt  time.Time `json:"created_at"`
}

// DrawEntryResponse includes user info for draw detail views.
type DrawEntryResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	UserName   string    `json:"user_name"`
	UserEmail  string    `json:"user_email"`
	EntryScore int       `json:"entry_score"`
	IsWinner   bool      `json:"is_winner"`
}

// DrawDetailResponse is the detailed view of a draw with entries and winner.
type DrawDetailResponse struct {
	Draw    DrawResponse        `json:"draw"`
	Entries []DrawEntryResponse `json:"entries"`
	Winner  *WinnerResponse     `json:"winner,omitempty"`
}

// RunDrawRequest is the input DTO for running a draw.
type RunDrawRequest struct {
	Month     int     `json:"month"`
	Year      int     `json:"year"`
	PoolAmount float64 `json:"pool_amount"`
}

// DrawSimulationResult is the output of a dry-run draw.
type DrawSimulationResult struct {
	Month         int                 `json:"month"`
	Year          int                 `json:"year"`
	EligibleUsers int                 `json:"eligible_users"`
	TotalPool     float64             `json:"total_pool"`
	WinnerPrize   float64             `json:"winner_prize"`
	CharityAmount float64             `json:"charity_amount"`
	PlatformFee   float64             `json:"platform_fee"`
	SampleWinner  *DrawEntryResponse  `json:"sample_winner,omitempty"`
	Entries       []DrawEntryResponse `json:"entries"`
}

// PaginatedDrawsResponse for listing draws.
type PaginatedDrawsResponse struct {
	Draws    []DrawResponse `json:"draws"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
