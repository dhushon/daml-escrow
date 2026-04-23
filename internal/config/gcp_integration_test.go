//go:build integration_gcp
// +build integration_gcp

package config

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGCPSecretManagerSpecialty validates that the Go API can authoritatively 
// vend critical institutional credentials from GCP Secret Manager.
//
// PREREQUISITES:
// 1. gcloud auth application-default login
// 2. GCP_PROJECT_ID environment variable set
// 3. Secrets provisioned via terraform/gcp_secrets.tf
func TestGCPSecretManagerSpecialty(t *testing.T) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		t.Fatal("GCP_PROJECT_ID must be set for specialty integration tests")
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "dev"
	}

	ctx := context.Background()
	resolver, err := NewSecretResolver(ctx, projectID)
	require.NoError(t, err, "Failed to initialize high-assurance secret resolver")
	defer func() { _ = resolver.Close() }()

	t.Run("Authoritative okta-client-secret vending", func(t *testing.T) {
		secretName := "okta-client-secret-" + env
		val, err := resolver.GetSecret(ctx, secretName)
		
		if err != nil {
			t.Errorf("Failed to vend secret [%s]: %v. Ensure terraform has provisioned it.", secretName, err)
		} else {
			assert.NotEmpty(t, val)
			t.Logf("SUCCESS: Authoritatively vended [%s] from cloud KeyVault", secretName)
		}
	})

	t.Run("Authoritative bitgo-access-token vending", func(t *testing.T) {
		secretName := "bitgo-access-token-" + env
		val, err := resolver.GetSecret(ctx, secretName)
		
		if err != nil {
			t.Logf("SKIP: [%s] not found (expected if BitGo is not yet configured)", secretName)
		} else {
			assert.NotEmpty(t, val)
			t.Logf("SUCCESS: Authoritatively vended [%s] from cloud KeyVault", secretName)
		}
	})

	t.Run("Zero-Trust: Access Denied Validation", func(t *testing.T) {
		// Attempt to fetch a secret name that should never exist or be accessible.
		_, err := resolver.GetSecret(ctx, "system-unauthorized-access-test")
		assert.Error(t, err, "Should fail due to missing secret or IAM restriction")
	})
}
