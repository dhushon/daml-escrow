package services

import (
	"runtime"
	"sync/atomic"
	"time"
)

type APIStats struct {
	TotalRequests   uint64
	TotalErrors     uint64
	TotalDurationNS uint64
}

type MetricsService struct {
	startTime    time.Time
	apiStats     APIStats
}

func NewMetricsService() *MetricsService {
	return &MetricsService{
		startTime: time.Now(),
	}
}

func (s *MetricsService) RecordRequest(duration time.Duration, isError bool) {
	atomic.AddUint64(&s.apiStats.TotalRequests, 1)
	atomic.AddUint64(&s.apiStats.TotalDurationNS, uint64(duration.Nanoseconds()))
	if isError {
		atomic.AddUint64(&s.apiStats.TotalErrors, 1)
	}
}

func (s *MetricsService) GetSystemPerformance() (int, int, float64, float64, float64, string) {
	reqs := atomic.LoadUint64(&s.apiStats.TotalRequests)
	errs := atomic.LoadUint64(&s.apiStats.TotalErrors)
	dur := atomic.LoadUint64(&s.apiStats.TotalDurationNS)

	avgLatency := 0
	if reqs > 0 {
		avgLatency = int(dur / reqs / 1000000) // Convert to ms
	}

	errorRate := 0.0
	if reqs > 0 {
		errorRate = (float64(errs) / float64(reqs)) * 100
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsage := float64(m.Alloc) / 1024 / 1024 // MB

	uptime := time.Since(s.startTime).Round(time.Second).String()

	// Mock CPU based on goroutines
	cpuUsage := float64(runtime.NumGoroutine()) * 0.5

	return avgLatency, int(reqs), errorRate, memUsage, cpuUsage, uptime
}
