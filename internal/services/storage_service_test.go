package services

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageService_Initialization(t *testing.T) {
	ctx := context.Background()

	t.Run("Local Initialization (MinIO)", func(t *testing.T) {
		_ = os.Setenv("STORAGE_ENDPOINT", "http://localhost:9000")
		_ = os.Setenv("STORAGE_ACCESS_KEY", "test")
		_ = os.Setenv("STORAGE_SECRET_KEY", "test-secret")
		defer func() { _ = os.Unsetenv("STORAGE_ENDPOINT") }()

		svc, err := NewStorageService(ctx, "test-bucket")
		assert.NoError(t, err)
		assert.NotNil(t, svc)
		assert.Equal(t, "test-bucket", svc.bucket)
	})

	t.Run("Production-Style Initialization", func(t *testing.T) {
		_ = os.Unsetenv("STORAGE_ENDPOINT")
		
		svc, err := NewStorageService(ctx, "prod-bucket")
		assert.NoError(t, err)
		assert.NotNil(t, svc)
		assert.Equal(t, "prod-bucket", svc.bucket)
	})
}
