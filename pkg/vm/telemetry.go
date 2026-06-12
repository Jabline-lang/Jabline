package vm

import (
	"context"
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"sync"
	"time"
)

type Telemetry struct {
	mu      sync.Mutex
	Metrics map[string]int64
	Spans   []Span
}

type Span struct {
	Name      string
	StartTime time.Time
}

func NewTelemetry() *Telemetry {
	return &Telemetry{
		Metrics: make(map[string]int64),
		Spans:   []Span{},
	}
}

func (vm *VM) opMetricInc(ins code.Instructions, ip *int) error {
	constIdx := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2

	nameObj := vm.constants[constIdx]
	name, ok := nameObj.(*object.String)
	if !ok {
		return fmt.Errorf("metric name must be a string, got %s", nameObj.Type())
	}

	if vm.Telemetry == nil {
		vm.Telemetry = NewTelemetry()
	}

	vm.Telemetry.mu.Lock()
	vm.Telemetry.Metrics[name.Value]++
	vm.Telemetry.mu.Unlock()

	if OTelEnabled() {
		OTelRecordMetric("jabline", name.Value, 1)
	}

	return nil
}

func (vm *VM) opTraceStart(ins code.Instructions, ip *int) error {
	constIdx := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2

	nameObj := vm.constants[constIdx]
	name, ok := nameObj.(*object.String)
	if !ok {
		return fmt.Errorf("trace name must be a string, got %s", nameObj.Type())
	}

	if vm.Telemetry == nil {
		vm.Telemetry = NewTelemetry()
	}

	span := Span{
		Name:      name.Value,
		StartTime: time.Now(),
	}
	vm.Telemetry.Spans = append(vm.Telemetry.Spans, span)

	if OTelEnabled() {
		ctx := context.Background()
		if vm.otelCtx != nil {
			ctx = vm.otelCtx
		}
		newCtx, endFn := StartOTelTrace(ctx, name.Value)
		vm.otelCtx = newCtx
		vm.otelSpanEnd = append(vm.otelSpanEnd, endFn)
	}

	return nil
}

func (vm *VM) opTraceEnd() error {
	if vm.Telemetry == nil || len(vm.Telemetry.Spans) == 0 {
		return fmt.Errorf("trace end called without matching start")
	}

	// Pop the last span and calculate duration
	index := len(vm.Telemetry.Spans) - 1
	span := vm.Telemetry.Spans[index]
	vm.Telemetry.Spans = vm.Telemetry.Spans[:index]

	duration := time.Since(span.StartTime)
	fmt.Printf("[TRACE] %s: %v\n", span.Name, duration)

	if OTelEnabled() && len(vm.otelSpanEnd) > 0 {
		endFn := vm.otelSpanEnd[len(vm.otelSpanEnd)-1]
		vm.otelSpanEnd = vm.otelSpanEnd[:len(vm.otelSpanEnd)-1]
		endFn()
	}

	return nil
}

// InitGlobalOTel initializes OpenTelemetry from environment variables.
// Should be called once at startup. Returns a cleanup function.
func InitGlobalOTel() (func(), error) {
	cleanup, err := InitTelemetryFromEnv()
	if err != nil {
		return nil, err
	}
	globalOTelCleanup = cleanup
	return cleanup, nil
}

// ShutdownOTel flushes and shuts down the global OTel provider.
func ShutdownOTel() {
	if globalOTelCleanup != nil {
		globalOTelCleanup()
		globalOTelCleanup = nil
	}
}

func (vm *VM) PrintTelemetry() {
	if vm.Telemetry == nil {
		return
	}

	fmt.Println("\n--- Jabline Telemetry ---")
	vm.Telemetry.mu.Lock()
	if len(vm.Telemetry.Metrics) > 0 {
		fmt.Println("Metrics:")
		for name, val := range vm.Telemetry.Metrics {
			fmt.Printf("  %s: %d\n", name, val)
		}
	}
	vm.Telemetry.mu.Unlock()
	fmt.Println("-------------------------")
}
