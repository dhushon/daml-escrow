package services

import (
	"context"
	"github.com/stretchr/testify/mock"
)

// Re-export or use ledger mocks directly in tests.
// This file now only contains mocks specific to the services package if any.

// Analytics Mock
type MockAnalyticsService struct {
	mock.Mock
}

func (m *MockAnalyticsService) ConfirmDeposit(ctx context.Context, tx string, amt float64, cur string) (bool, error) {
	args := m.Called(ctx, tx, amt, cur)
	return args.Bool(0), args.Error(1)
}

func (m *MockAnalyticsService) GetEscrowLifecycle(ctx context.Context, id string, state string) (EscrowLifecycleMetadata, error) {
	args := m.Called(ctx, id, state)
	return args.Get(0).(EscrowLifecycleMetadata), args.Error(1)
}

func (m *MockAnalyticsService) GetWalletHistory(ctx context.Context, addr string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, addr)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// AI Mock
type MockAIProvider struct {
	mock.Mock
}

func (m *MockAIProvider) ClassifyContract(ctx context.Context, allFileData [][]byte, mimeType string) (string, error) {
	args := m.Called(ctx, allFileData, mimeType)
	return args.String(0), args.Error(1)
}

func (m *MockAIProvider) ExtractTerms(ctx context.Context, allFileData [][]byte, mimeType string, contractType string, schema interface{}) (string, error) {
	args := m.Called(ctx, allFileData, mimeType, contractType, schema)
	return args.String(0), args.Error(1)
}
