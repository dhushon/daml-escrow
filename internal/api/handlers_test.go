package api

import (
	"daml-escrow/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// We use anonymous interface mocks to keep API tests isolated from service-layer test files
type mockAnalytics struct {
	getLifecycleFn func(id, state string) (interface{}, error)
}

func TestHandler_GetHealth(t *testing.T) {
	// 1. Setup minimal handler
	logger := zap.NewNop()
	metrics := services.NewMetricsService()
	h := NewHandler(logger, nil, metrics, nil, nil)

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
