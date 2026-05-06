package services

import (
	"context"
	"daml-escrow/internal/ledger"
	"runtime"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type APIStats struct {
	TotalRequests   uint64
	TotalErrors     uint64
	TotalDurationNS uint64
}

type MetricsService struct {
	startTime  time.Time
	apiStats   APIStats
	meter      metric.Meter
	requestCounter  metric.Int64Counter
	errorCounter    metric.Int64Counter
	latencyRecorder metric.Float64Histogram
}

func NewMetricsService() *MetricsService {
	m := otel.GetMeterProvider().Meter("escrow-api")
	
	req, _ := m.Int64Counter("api.requests.total", metric.WithDescription("Total institutional requests"))
	err, _ := m.Int64Counter("api.errors.total", metric.WithDescription("Total request failures"))
	lat, _ := m.Float64Histogram("api.latency", metric.WithDescription("Request latency in milliseconds"), metric.WithUnit("ms"))

	return &MetricsService{
		startTime: time.Now(),
		meter:     m,
		requestCounter:  req,
		errorCounter:    err,
		latencyRecorder: lat,
	}
}

func (s *MetricsService) RecordRequest(ctx context.Context, duration time.Duration, isError bool, accountID, contractID string) {
	// 1. In-Memory Legacy Stats
	atomic.AddUint64(&s.apiStats.TotalRequests, 1)
	atomic.AddUint64(&s.apiStats.TotalDurationNS, uint64(duration.Nanoseconds()))
	if isError {
		atomic.AddUint64(&s.apiStats.TotalErrors, 1)
	}

	// 2. High-Assurance OTEL Metrics
	attrs := []attribute.KeyValue{
		attribute.String("account.id", accountID),
	}
	if contractID != "" {
		attrs = append(attrs, attribute.String("contract.id", contractID))
	}

	s.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	if isError {
		s.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
	s.latencyRecorder.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))
}

func (s *MetricsService) GetHealth(configSvc *ConfigService, ledgerClient ledger.Client, oracleSecret string) ledger.HealthResponse {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	services := make(map[string]ledger.ServiceHealth)
	overallStatus := "UP"

	// 1. Check Database (Postgres) with timeout
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dbCancel()
	
	dbStart := time.Now()
	dbStatus := ledger.ServiceHealth{Status: "UP"}
	if err := configSvc.db.PingContext(dbCtx); err != nil {
		dbStatus.Status = "DOWN"
		dbStatus.Message = "database unreachable: " + err.Error()
		overallStatus = "DEGRADED"
	}
	dbStatus.LatencyMs = time.Since(dbStart).Milliseconds()
	services["database"] = dbStatus

	// 2. Check Ledger (Canton HTTP API) with timeout
	ledgerCtx, ledgerCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer ledgerCancel()

	ledgerStart := time.Now()
	ledgerStatus := ledger.ServiceHealth{Status: "UP"}
	if _, err := ledgerClient.SearchPackageID(ledgerCtx, "stablecoin-escrow"); err != nil {
		ledgerStatus.Status = "DOWN"
		ledgerStatus.Message = "ledger api unreachable: " + err.Error()
		overallStatus = "DEGRADED"
	}
	ledgerStatus.LatencyMs = time.Since(ledgerStart).Milliseconds()
	services["ledger"] = ledgerStatus

	// 3. Check Oracle (Configuration Check)
	oracleStatus := ledger.ServiceHealth{Status: "UP"}
	if oracleSecret == "" || oracleSecret == "development-secret-key" {
		oracleStatus.Status = "DEGRADED"
		oracleStatus.Message = "Using default development secret"
	}
	services["oracle"] = oracleStatus

	return ledger.HealthResponse{
		Status:      overallStatus,
		Version:     "1.0.0",
		Uptime:      time.Since(s.startTime).Round(time.Second).String(),
		StartTime:   s.startTime.Format(time.RFC3339),
		CPUUsage:    float64(runtime.NumGoroutine()) * 0.5,
		MemoryUsage: float64(m.Alloc) / 1024 / 1024,
		Goroutines:  runtime.NumGoroutine(),
		Services:    services,
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
