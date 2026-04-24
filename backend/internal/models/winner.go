package models

import (
	"time"

	"github.com/google/uuid"
)

// Winner represents a draw winner record.
type Winner struct {
	ID                 uuid.UUID  `json:"id"`
	DrawID             uuid.UUID  `json:"draw_id"`
	UserID             uuid.UUID  `json:"user_id"`
	PrizeAmount        float64    `json:"prize_amount"`
	ProofURL           string     `json:"proof_url"`
	ProofNotes         string     `json:"proof_notes"`
	VerificationStatus string     `json:"verification_status"`
	RejectionReason    string     `json:"rejection_reason"`
	VerifiedBy         *uuid.UUID `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// WinnerResponse is the API response for a winner.
type WinnerResponse struct {
	ID                 uuid.UUID  `json:"id"`
	DrawID             uuid.UUID  `json:"draw_id"`
	UserID             uuid.UUID  `json:"user_id"`
	UserName           string     `json:"user_name,omitempty"`
	UserEmail          string     `json:"user_email,omitempty"`
	DrawMonth          int        `json:"draw_month,omitempty"`
	DrawYear           int        `json:"draw_year,omitempty"`
	PrizeAmount        float64    `json:"prize_amount"`
	ProofURL           string     `json:"proof_url"`
	ProofNotes         string     `json:"proof_notes"`
	VerificationStatus string     `json:"verification_status"`
	RejectionReason    string     `json:"rejection_reason"`
	VerifiedBy         *uuid.UUID `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// SubmitProofRequest is the input DTO for submitting winner proof.
type SubmitProofRequest struct {
	ProofURL   string `json:"proof_url"`
	ProofNotes string `json:"proof_notes"`
}

// VerifyWinnerRequest is the input DTO for admin verification.
type VerifyWinnerRequest struct {
	Status          string `json:"status"` // "approved" or "rejected"
	RejectionReason string `json:"rejection_reason"`
}
