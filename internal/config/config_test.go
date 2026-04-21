package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Environment(t *testing.T) {
	// Create a temporary config file
	configContent := `
server:
  port: 8081
auth:
  issuer: "https://test.com"
`
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.Write([]byte(configContent))
	assert.NoError(t, err)
	_ = tmpFile.Close()

	t.Run("Default Environment", func(t *testing.T) {
		_ = os.Unsetenv("ENVIRONMENT")
		_ = os.Unsetenv("AUTH_BYPASS")
		
		cfg, err := LoadConfig(tmpFile.Name())
		assert.NoError(t, err)
		assert.Equal(t, "production", cfg.Auth.Environment)
		assert.False(t, cfg.Auth.AuthBypass)
	})

	t.Run("Custom Environment", func(t *testing.T) {
		_ = os.Setenv("ENVIRONMENT", "dev")
		_ = os.Setenv("AUTH_BYPASS", "true")
		defer func() { _ = os.Unsetenv("ENVIRONMENT") }()
		defer func() { _ = os.Unsetenv("AUTH_BYPASS") }()

		cfg, err := LoadConfig(tmpFile.Name())
		assert.NoError(t, err)
		assert.Equal(t, "dev", cfg.Auth.Environment)
		assert.True(t, cfg.Auth.AuthBypass)
	})
}
