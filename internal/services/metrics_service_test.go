package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		health := svc.GetHealth()
		assert.Equal(t, "UP", health.Status)
		assert.Equal(t, "1.0.0", health.Version)
		assert.NotEmpty(t, health.Uptime)
		assert.NotEmpty(t, health.StartTime)
	})
}
