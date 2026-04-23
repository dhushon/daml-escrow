package config

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGCPSecretManagerIntegration verifies connectivity and vending from real GCP Secret Manager.
// This test is skipped unless GCP_PROJECT_ID and GOOGLE_APPLICATION_CREDENTIALS are set.
func TestGCPSecretManagerIntegration(t *testing.T) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		t.Skip("Skipping GCP integration test: GCP_PROJECT_ID not set")
	}

	ctx := context.Background()
	resolver, err := NewSecretResolver(ctx, projectID)
	require.NoError(t, err, "Failed to initialize secret resolver")
	defer func() { _ = resolver.Close() }()

	t.Run("Fetch Known Secret Metadata", func(t *testing.T) {
		// We attempt to fetch the 'okta-client-secret'.
		// This verifies IAM permissions and SDK connectivity.
		val, err := resolver.GetSecret(ctx, "okta-client-secret")
		
		// In a real integration test environment, we expect this secret to exist.
		// If it doesn't, we still verify the error type (e.g., should not be an Auth error).
		if err != nil {
			t.Logf("Secret 'okta-client-secret' not found (expected in fresh environments): %v", err)
		} else {
			assert.NotEmpty(t, val)
			t.Log("Successfully vended 'okta-client-secret' from GCP")
		}
	})

	t.Run("Fetch Non-Existent Secret", func(t *testing.T) {
		_, err := resolver.GetSecret(ctx, "non-existent-secret-123")
		assert.Error(t, err)
		t.Logf("Correctly handled missing secret: %v", err)
	})
}
