package api

import (
	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestHandler_GetHealth(t *testing.T) {
	// 1. Setup mock dependencies
	logger := zap.NewNop()
	metrics := services.NewMetricsService()
	
	db, mockDB, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
	defer db.Close()
	configSvc := services.NewMockConfigService(db)
	mockDB.ExpectPing()

	mockLedger := new(ledger.MockLedgerClient)
	mockLedger.On("SearchPackageID", mock.Anything, "stablecoin-escrow").Return("pkg-123", nil)

	escrowSvc := services.NewEscrowService(logger, mockLedger, nil, nil, "test-secret")
	
	h := NewHandler(logger, escrowSvc, metrics, configSvc, nil, nil)

	t.Run("Health returns 200 and UP status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		rr := httptest.NewRecorder()
		
		h.GetHealth(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "\"status\":\"UP\"")
	})
}

func (h *Handler) TestRoute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	h.renderJSON(w, map[string]string{"id": id, "status": "mocked"})
}
