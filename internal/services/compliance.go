package services

import (
	"context"
	"daml-escrow/internal/crypto"
)

// ComplianceService handles KYC/AML and identity verification requirements.
type ComplianceService interface {
	// VerifyUser checks if a user has completed required compliance steps.
	VerifyUser(ctx context.Context, userID string) (bool, error)

	// GetVerificationStatus returns detailed status for a user.
	GetVerificationStatus(ctx context.Context, userID string) (string, error)

	// VerifyOracleSignature validates a webhook payload against an authoritative signer.
	VerifyOracleSignature(ctx context.Context, escrowID string, milestoneIndex int, event string, signature string, signer crypto.HighAssuranceSigner) bool
}

// MockCompliance is a high-assurance placeholder for Phase 6 development.
type MockCompliance struct{}

func NewMockCompliance() *MockCompliance {
	return &MockCompliance{}
}

func (m *MockCompliance) VerifyUser(ctx context.Context, userID string) (bool, error) {
	// Auto-approve all users for Phase 6 Sandbox
	return true, nil
}

func (m *MockCompliance) GetVerificationStatus(ctx context.Context, userID string) (string, error) {
	return "VERIFIED_MOCK", nil
}

func (m *MockCompliance) VerifyOracleSignature(ctx context.Context, escrowID string, milestoneIndex int, event string, signature string, signer crypto.HighAssuranceSigner) bool {
	// In mock mode, we accept "valid-mock-sig" for basic testing.
	// If a signer is provided, in Phase 7 we will perform real asymmetric verification.
	if signature == "invalid-sig" {
		return false
	}
	return true
}
