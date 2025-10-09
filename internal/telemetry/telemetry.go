package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/savisec/hello-go/internal/config"
)

type Provider struct {
	tracerProvider *trace.TracerProvider
}

func Setup(ctx context.Context, cfg config.TelemetryConfig) (*Provider, error) {
	if !cfg.Enabled {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return &Provider{}, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)

	return &Provider{
		tracerProvider: tracerProvider,
	}, nil
}

func (p *Provider) Shutdown(ctx context.Context) error {
	if p.tracerProvider == nil {
		return nil
	}

	if err := p.tracerProvider.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown tracer provider", "error", err)
		return err
	}

	return nil
}
