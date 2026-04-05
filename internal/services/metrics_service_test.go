package services

import (
	"daml-escrow/internal/ledger"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMetricsService(t *testing.T) {
	svc := NewMetricsService()

	t.Run("RecordRequest", func(t *testing.T) {
		svc.RecordRequest(100*time.Millisecond, false)
		svc.RecordRequest(200*time.Millisecond, true)

		avg, count, errRate, _, _, _ := svc.GetSystemPerformance()
		assert.Equal(t, 2, count)
		assert.Equal(t, 150, avg) // (100+200)/2 = 150ms
		assert.Equal(t, 50.0, errRate)
	})

	t.Run("GetHealth", func(t *testing.T) {
		// Mock DB for ConfigService
		db, mockDB, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		defer db.Close()
		configSvc := &ConfigService{db: db}
		mockDB.ExpectPing()

		// Mock Ledger Client
		mockLedger := new(ledger.MockLedgerClient)
		mockLedger.On("SearchPackageID", mock.Anything, "stablecoin-escrow").Return("pkg-123", nil)

		health := svc.GetHealth(configSvc, mockLedger, "test-secret")
		
		assert.Equal(t, "UP", health.Status)
		assert.Equal(t, "1.0.0", health.Version)
		assert.NotEmpty(t, health.Uptime)
		assert.NotEmpty(t, health.StartTime)
		assert.Contains(t, health.Services, "database")
		assert.Contains(t, health.Services, "ledger")
		assert.Contains(t, health.Services, "oracle")
		assert.Equal(t, "UP", health.Services["database"].Status)
		assert.Equal(t, "UP", health.Services["ledger"].Status)
		
		mockDB.ExpectationsWereMet()
		mockLedger.AssertExpectations(t)
	})

	t.Run("GetHealth - Database Down", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
		defer db.Close()
		configSvc := &ConfigService{db: db}
		mockDB.ExpectPing().WillReturnError(fmt.Errorf("db connection refused"))

		mockLedger := new(ledger.MockLedgerClient)
		mockLedger.On("SearchPackageID", mock.Anything, "stablecoin-escrow").Return("pkg-123", nil)

		health := svc.GetHealth(configSvc, mockLedger, "test-secret")
		
		assert.Equal(t, "DEGRADED", health.Status)
		assert.Equal(t, "DOWN", health.Services["database"].Status)
		assert.Contains(t, health.Services["database"].Message, "db connection refused")
		assert.Equal(t, "UP", health.Services["ledger"].Status)
	})

	t.Run("GetHealth - Ledger Down", func(t *testing.T) {
		db, mockDB, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
		defer db.Close()
		configSvc := &ConfigService{db: db}
		mockDB.ExpectPing()

		mockLedger := new(ledger.MockLedgerClient)
		mockLedger.On("SearchPackageID", mock.Anything, "stablecoin-escrow").Return("", fmt.Errorf("ledger timed out"))

		health := svc.GetHealth(configSvc, mockLedger, "test-secret")
		
		assert.Equal(t, "DEGRADED", health.Status)
		assert.Equal(t, "DOWN", health.Services["ledger"].Status)
		assert.Contains(t, health.Services["ledger"].Message, "ledger timed out")
		assert.Equal(t, "UP", health.Services["database"].Status)
	})
}
