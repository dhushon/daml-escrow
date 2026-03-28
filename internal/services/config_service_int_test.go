//go:build integration
// +build integration

package services

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigService_Integration(t *testing.T) {
	dsn := os.Getenv("USER_CONFIG_DSN")
	if dsn == "" {
		t.Skip("Skipping integration test: USER_CONFIG_DSN not set")
	}

	svc, err := NewConfigService(dsn)
	require.NoError(t, err)
	defer svc.Close()

	userID := "test-user-int"
	key := "test-key"
	val := json.RawMessage(`{"active": true}`)

	// 1. Save
	err = svc.SaveConfig(userID, key, val)
	assert.NoError(t, err)

	// 2. Get
	retrieved, err := svc.GetConfig(userID, key)
	assert.NoError(t, err)
	assert.JSONEq(t, string(val), string(retrieved))

	// 3. Update
	newVal := json.RawMessage(`{"active": false}`)
	err = svc.SaveConfig(userID, key, newVal)
	assert.NoError(t, err)

	retrieved2, err := svc.GetConfig(userID, key)
	assert.NoError(t, err)
	assert.JSONEq(t, string(newVal), string(retrieved2))
}
