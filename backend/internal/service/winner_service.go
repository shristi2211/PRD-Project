package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"golf-score-lottery/backend/internal/models"
	"golf-score-lottery/backend/internal/repository"
)

// WinnerService handles winner verification business logic.
type WinnerService struct {
	winnerRepo     *repository.WinnerRepository
	activityLogSvc *ActivityLogService
}

func NewWinnerService(winnerRepo *repository.WinnerRepository, activityLogSvc *ActivityLogService) *WinnerService {
	return &WinnerService{winnerRepo: winnerRepo, activityLogSvc: activityLogSvc}
}

// GetMyWinnings returns all wins for a specific user.
func (s *WinnerService) GetMyWinnings(ctx context.Context, userID uuid.UUID) ([]models.WinnerResponse, error) {
	return s.winnerRepo.GetWinnersByUserID(ctx, userID)
}

// SubmitProof allows a winner to submit proof of their score.
func (s *WinnerService) SubmitProof(ctx context.Context, userID uuid.UUID, winnerID uuid.UUID, req *models.SubmitProofRequest, ipAddress string) error {
	if req.ProofURL == "" {
		return fmt.Errorf("proof URL is required")
	}

	// Verify ownership
	winner, err := s.winnerRepo.GetWinnerByID(ctx, winnerID)
	if err != nil {
		if errors.Is(err, repository.ErrWinnerNotFound) {
			return fmt.Errorf("winner record not found")
		}
		return err
	}
	if winner.UserID != userID {
		return fmt.Errorf("you can only submit proof for your own winnings")
	}
	if winner.VerificationStatus != "pending" {
		return fmt.Errorf("proof can only be submitted for pending verifications")
	}

	if err := s.winnerRepo.UpdateWinnerProof(ctx, winnerID, req.ProofURL, req.ProofNotes); err != nil {
		return err
	}

	s.activityLogSvc.LogAction(ctx, &userID, "proof_submitted", "winner", winnerID.String(),
		map[string]interface{}{"proof_url": req.ProofURL}, ipAddress)

	return nil
}

// VerifyWinner allows admin to approve or reject a winner.
func (s *WinnerService) VerifyWinner(ctx context.Context, adminID uuid.UUID, winnerID uuid.UUID, req *models.VerifyWinnerRequest, ipAddress string) error {
	if req.Status != "approved" && req.Status != "rejected" {
		return fmt.Errorf("status must be 'approved' or 'rejected'")
	}
	if req.Status == "rejected" && req.RejectionReason == "" {
		return fmt.Errorf("rejection reason is required when rejecting")
	}

	// Verify winner exists
	winner, err := s.winnerRepo.GetWinnerByID(ctx, winnerID)
	if err != nil {
		if errors.Is(err, repository.ErrWinnerNotFound) {
			return fmt.Errorf("winner record not found")
		}
		return err
	}
	if winner.VerificationStatus != "pending" {
		return fmt.Errorf("this winner has already been %s", winner.VerificationStatus)
	}

	if err := s.winnerRepo.UpdateWinnerVerification(ctx, winnerID, req.Status, req.RejectionReason, adminID); err != nil {
		return err
	}

	s.activityLogSvc.LogAction(ctx, &adminID, "winner_"+req.Status, "winner", winnerID.String(),
		map[string]interface{}{
			"user_id":          winner.UserID.String(),
			"prize_amount":     winner.PrizeAmount,
			"rejection_reason": req.RejectionReason,
		}, ipAddress)

	return nil
}

// GetPendingVerifications returns all pending winner verifications.
func (s *WinnerService) GetPendingVerifications(ctx context.Context) ([]models.WinnerResponse, error) {
	return s.winnerRepo.GetPendingVerifications(ctx)
}
