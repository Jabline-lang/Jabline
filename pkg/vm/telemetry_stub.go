//go:build !otel

package vm

import (
	"context"
	"os"
)

// OTelConfig configures OpenTelemetry export.
type OTelConfig struct {
	Enabled      bool
	ServiceName  string
	OTLPEndpoint string
	Insecure     bool
	TraceSampler float64
}

// DefaultOTelConfig returns sensible defaults.
func DefaultOTelConfig() OTelConfig {
	return OTelConfig{Enabled: false}
}

// OTelConfigFromEnv reads configuration from environment variables.
func OTelConfigFromEnv() OTelConfig {
	return DefaultOTelConfig()
}

// InitOTel initializes the OpenTelemetry SDK (stub - no-op).
func InitOTel(ctx context.Context, cfg OTelConfig) (func(), error) {
	return func() {}, nil
}

// StartOTelTrace starts a new trace span (stub).
func StartOTelTrace(ctx context.Context, name string, attrs ...interface{}) (context.Context, func()) {
	return ctx, func() {}
}

// OTelRecordMetric records a metric value (stub).
func OTelRecordMetric(meterName, metricName string, value int64, attrs ...interface{}) {}

// InitTelemetryFromEnv initializes telemetry from env (stub).
func InitTelemetryFromEnv() (func(), error) {
	return func() {}, nil
}

// WrapErrorWithOTel wraps an error with telemetry (stub).
func WrapErrorWithOTel(err error, spanName string) error {
	return err
}

// OTelEnabled checks if OpenTelemetry is enabled (stub).
func OTelEnabled() bool {
	return os.Getenv("JABLINE_OTEL_ENABLED") == "1" || os.Getenv("JABLINE_OTEL_ENABLED") == "true"
}

// SetOTelAttributesFromMap converts a string map to OTel attributes (stub).
func SetOTelAttributesFromMap(attrs map[string]string) []interface{} {
	return nil
}

// OTelBatchAttributes converts a map to OTel attributes in batches (stub).
func OTelBatchAttributes(m map[string]string) []interface{} {
	return nil
}
