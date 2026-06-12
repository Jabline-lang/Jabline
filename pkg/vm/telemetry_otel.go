//go:build otel

package vm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace/noop"
)

// OTelConfig configures OpenTelemetry export.
type OTelConfig struct {
	Enabled      bool
	ServiceName  string
	OTLPEndpoint string // gRPC endpoint (e.g., "localhost:4317")
	Insecure     bool
	TraceSampler float64 // 0.0-1.0
}

// DefaultOTelConfig returns sensible defaults.
func DefaultOTelConfig() OTelConfig {
	return OTelConfig{
		Enabled:      false,
		ServiceName:  "jabline",
		OTLPEndpoint: "",
		Insecure:     true,
		TraceSampler: 0.1,
	}
}

// OTelConfigFromEnv reads configuration from environment variables.
func OTelConfigFromEnv() OTelConfig {
	cfg := DefaultOTelConfig()
	if v := os.Getenv("JABLINE_OTEL_ENABLED"); v == "1" || v == "true" {
		cfg.Enabled = true
	}
	if v := os.Getenv("JABLINE_OTEL_ENDPOINT"); v != "" {
		cfg.OTLPEndpoint = v
	}
	if v := os.Getenv("JABLINE_OTEL_INSECURE"); v == "1" || v == "true" {
		cfg.Insecure = true
	}
	if v := os.Getenv("JABLINE_OTEL_SERVICE_NAME"); v != "" {
		cfg.ServiceName = v
	}
	if v := os.Getenv("JABLINE_OTEL_SAMPLE_RATE"); v != "" {
		var rate float64
		if _, err := fmt.Sscanf(v, "%f", &rate); err == nil {
			cfg.TraceSampler = rate
		}
	}
	return cfg
}

// InitOTel initializes the OpenTelemetry SDK.
// Returns a cleanup function that should be called at shutdown.
func InitOTel(ctx context.Context, cfg OTelConfig) (func(), error) {
	if !cfg.Enabled {
		// Use noop tracer
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func() {}, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			attribute.String("jabline.version", "0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTel resource: %w", err)
	}

	// Set up trace exporter
	var traceExporter trace.SpanExporter
	if cfg.OTLPEndpoint != "" {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		traceExporter, err = otlptrace.New(ctx, otlptracegrpc.NewClient(opts...))
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}
	} else {
		traceExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
		}
	}

	// Set up metric exporter (stdout)
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Create tracer provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(5*time.Second),
		),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(cfg.TraceSampler)),
	)
	otel.SetTracerProvider(tp)

	// Create meter provider
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(10*time.Second),
		)),
	)

	cleanup := func() {
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
	}

	return cleanup, nil
}

// StartOTelTrace starts a new trace span, returning the span and updated context.
func StartOTelTrace(ctx context.Context, name string, attrs ...interface{}) (context.Context, func()) {
	tracer := otel.Tracer("jabline")
	ctx, span := tracer.Start(ctx, name)
	for _, attr := range attrs {
		if kv, ok := attr.(attribute.KeyValue); ok {
			span.SetAttributes(kv)
		}
	}
	return ctx, func() { span.End() }
}

// OTelRecordMetric records a metric value.
func OTelRecordMetric(meterName, metricName string, value int64, attrs ...interface{}) {
	meter := otel.Meter(meterName)
	counter, _ := meter.Int64Counter(metricName)
	var kvAttrs []attribute.KeyValue
	for _, attr := range attrs {
		if kv, ok := attr.(attribute.KeyValue); ok {
			kvAttrs = append(kvAttrs, kv)
		}
	}
	counter.Add(context.Background(), value, kvAttrs...)
}

// InitTelemetryFromEnv initializes telemetry based on environment variables.
// This is called at startup by the main function.
func InitTelemetryFromEnv() (func(), error) {
	cfg := OTelConfigFromEnv()
	return InitOTel(context.Background(), cfg)
}

// WrapErrorWithOTel wraps an error with telemetry attributes.
func WrapErrorWithOTel(err error, spanName string) error {
	if err == nil {
		return nil
	}
	_, span := otel.Tracer("jabline").Start(context.Background(), spanName)
	defer span.End()
	span.RecordError(err)
	span.SetAttributes(attribute.String("error.message", err.Error()))
	return err
}

// OTelEnabled checks if OpenTelemetry is configured and enabled.
func OTelEnabled() bool {
	return os.Getenv("JABLINE_OTEL_ENABLED") == "1" || os.Getenv("JABLINE_OTEL_ENABLED") == "true"
}

// SetOTelAttributesFromMap sets span attributes from a Jabline hash/object.
func SetOTelAttributesFromMap(attrs map[string]string) []interface{} {
	result := make([]interface{}, 0, len(attrs))
	for k, v := range attrs {
		result = append(result, attribute.String(k, v))
	}
	return result
}

// OTelBatchAttributes converts a map to OTel attributes in batches.
func OTelBatchAttributes(m map[string]string) []interface{} {
	attrs := make([]interface{}, 0, len(m))
	for k, v := range m {
		switch {
		case v == "true" || v == "false":
			attrs = append(attrs, attribute.Bool(k, v == "true"))
		case strings.Contains(v, "."):
			var f float64
			if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
				attrs = append(attrs, attribute.Float64(k, f))
			} else {
				attrs = append(attrs, attribute.String(k, v))
			}
		default:
			attrs = append(attrs, attribute.String(k, v))
		}
	}
	return attrs
}
