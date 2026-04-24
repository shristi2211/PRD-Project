package models

import (
	"time"

	"github.com/google/uuid"
)

// Score represents a user's Stableford golf score.
type Score struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Score     int       `json:"score"`
	RoundDate time.Time `json:"round_date"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

// ScoreResponse is the API response for a score.
type ScoreResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Score     int       `json:"score"`
	RoundDate string    `json:"round_date"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a Score to ScoreResponse.
func (s *Score) ToResponse() ScoreResponse {
	return ScoreResponse{
		ID:        s.ID,
		UserID:    s.UserID,
		Score:     s.Score,
		RoundDate: s.RoundDate.Format("2006-01-02"),
		Notes:     s.Notes,
		CreatedAt: s.CreatedAt,
	}
}

// CreateScoreRequest is the input DTO for submitting a new score.
type CreateScoreRequest struct {
	Score     int    `json:"score"`
	RoundDate string `json:"round_date"`
	Notes     string `json:"notes"`
}

// PaginatedScoresResponse is the output DTO for admin score listing.
type PaginatedScoresResponse struct {
	Scores   []ScoreWithUser `json:"scores"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// ScoreWithUser includes user info for admin views.
type ScoreWithUser struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	UserEmail string    `json:"user_email"`
	Score     int       `json:"score"`
	RoundDate string    `json:"round_date"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}
