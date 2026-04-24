package models

import (
	"time"

	"github.com/google/uuid"
)

// ActivityLog represents an audit trail entry.
type ActivityLog struct {
	ID         uuid.UUID              `json:"id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Metadata   map[string]interface{} `json:"metadata"`
	IPAddress  string                 `json:"ip_address"`
	CreatedAt  time.Time              `json:"created_at"`
}

// ActivityLogResponse is the API response for an activity log.
type ActivityLogResponse struct {
	ID         uuid.UUID              `json:"id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	UserName   string                 `json:"user_name,omitempty"`
	UserEmail  string                 `json:"user_email,omitempty"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Metadata   map[string]interface{} `json:"metadata"`
	IPAddress  string                 `json:"ip_address"`
	CreatedAt  time.Time              `json:"created_at"`
}

// PaginatedActivityLogsResponse for admin listing.
type PaginatedActivityLogsResponse struct {
	Logs     []ActivityLogResponse `json:"logs"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// DashboardStats for user dashboard.
type UserDashboardStats struct {
	RoundsPlayed   int     `json:"rounds_played"`
	BestScore      int     `json:"best_score"`
	LotteryEntries int     `json:"lottery_entries"`
	TotalWinnings  float64 `json:"total_winnings"`
}

// AdminDashboardStats for admin dashboard.
type AdminDashboardStats struct {
	TotalUsers           int     `json:"total_users"`
	ActiveSubscriptions  int     `json:"active_subscriptions"`
	TotalRevenue         float64 `json:"total_revenue"`
	PendingVerifications int     `json:"pending_verifications"`
	TotalDraws           int     `json:"total_draws"`
	TotalCharities       int     `json:"total_charities"`
}

// ChartDataPoint for chart data.
type ChartDataPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}
