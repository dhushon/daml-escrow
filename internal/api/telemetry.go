package api

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Telemetry provides a high-assurance handle for distributed tracing and metrics.
type Telemetry struct {
	tp *trace.TracerProvider
	mp *metric.MeterProvider
}

// InitTelemetry authoritatively initializes the OTEL SDK.
// It defaults to OTLP HTTP exporters, synchronized with our GKE/Docker collector strategy.
func InitTelemetry(ctx context.Context, serviceName, environment string) (*Telemetry, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel resource: %w", err)
	}

	// 1. Trace Provider
	traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure()) // Internal cluster comms
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// 2. Metric Provider
	metricExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(15*time.Second))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return &Telemetry{tp: tp, mp: mp}, nil
}

// Shutdown authoritatively flushes all telemetry data before service exit.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error
	if err := t.tp.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if err := t.mp.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("telemetry shutdown errors: %v", errs)
	}
	return nil
}
