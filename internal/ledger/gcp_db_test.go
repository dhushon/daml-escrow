package ledger

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGCPDatabaseSchemaIntegration verifies that the target Postgres instance (Cloud SQL)
// has the required tables and reference data for the escrow platform.
func TestGCPDatabaseSchemaIntegration(t *testing.T) {
	dsn := os.Getenv("USER_CONFIG_DSN")
	if dsn == "" {
		t.Skip("Skipping Database integration test: USER_CONFIG_DSN not set")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	t.Run("Verify User Config Table", func(t *testing.T) {
		var tableName string
		err := db.QueryRow(`
			SELECT tablename 
			FROM pg_catalog.pg_tables 
			WHERE schemaname = 'public' AND tablename = 'configs'
		`).Scan(&tableName)
		
		assert.NoError(t, err, "Table 'configs' should exist in the database")
		assert.Equal(t, "configs", tableName)
	})

	t.Run("Verify Schema Structure", func(t *testing.T) {
		rows, err := db.Query(`
			SELECT column_name, data_type 
			FROM information_schema.columns 
			WHERE table_name = 'configs'
		`)
		require.NoError(t, err)
		defer func() { _ = rows.Close() }()

		columns := make(map[string]string)
		for rows.Next() {
			var name, dtype string
			err := rows.Scan(&name, &dtype)
			require.NoError(t, err)
			columns[name] = dtype
		}

		assert.Contains(t, columns, "user_id")
		assert.Contains(t, columns, "config_key")
		assert.Contains(t, columns, "config_value")
		assert.Equal(t, "jsonb", columns["config_value"])
	})
}
