package services

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigService_Unit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := &ConfigService{db: db}

	t.Run("GetConfig - Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_value"}).
			AddRow([]byte(`{"value": 123}`))

		mock.ExpectQuery("SELECT config_value FROM configs").
			WithArgs("user1", "key1").
			WillReturnRows(rows)

		val, err := svc.GetConfig("user1", "key1")
		assert.NoError(t, err)
		assert.JSONEq(t, `{"value": 123}`, string(val))
	})

	t.Run("GetConfig - Not Found", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_value FROM configs").
			WithArgs("user1", "key2").
			WillReturnRows(sqlmock.NewRows([]string{"config_value"})) // Return empty rows set

		val, err := svc.GetConfig("user1", "key2")
		assert.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("SaveConfig - Success", func(t *testing.T) {
		payload := json.RawMessage(`{"theme": "dark"}`)
		mock.ExpectExec("INSERT INTO configs").
			WithArgs("user1", "theme", payload).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := svc.SaveConfig("user1", "theme", payload)
		assert.NoError(t, err)
	})
}
