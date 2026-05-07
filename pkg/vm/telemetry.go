package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"time"
)

type Telemetry struct {
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

	vm.Telemetry.Metrics[name.Value]++
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
	// For now, we print the trace. Later we could send it to a collector.
	fmt.Printf("[TRACE] %s: %v\n", span.Name, duration)

	return nil
}

func (vm *VM) PrintTelemetry() {
	if vm.Telemetry == nil {
		return
	}

	fmt.Println("\n--- Jabline Telemetry ---")
	if len(vm.Telemetry.Metrics) > 0 {
		fmt.Println("Metrics:")
		for name, val := range vm.Telemetry.Metrics {
			fmt.Printf("  %s: %d\n", name, val)
		}
	}
	fmt.Println("-------------------------")
}
