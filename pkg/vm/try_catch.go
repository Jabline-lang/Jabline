package vm

import (
	"jabline/pkg/code"
)

// FinallyHandler tracks the finally block IP for cleanup execution.
type FinallyHandler struct {
	FinallyIP int
	StackSP   int
}

// opTryExt records finally offset if present.
func (vm *VM) opTryExt(ins code.Instructions, ip *int, hasFinally bool) {
	catchIP := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2

	var finallyIP int
	if hasFinally {
		finallyIP = int(code.ReadUint16(ins[*ip+1:]))
		*ip += 2
	}

	handler := ExceptionHandler{
		CatchIP:    catchIP,
		StackSP:    vm.sp,
		FrameIndex: vm.framesIndex,
	}
	vm.handlers = append(vm.handlers, handler)

	if hasFinally {
		vm.finallyHandlers = append(vm.finallyHandlers, FinallyHandler{
			FinallyIP: finallyIP,
			StackSP:   vm.sp,
		})
	}
}

// opEndTryExt pops the exception handler (and finally handler if present).
func (vm *VM) opEndTryExt(hasFinally bool) {
	if len(vm.handlers) > 0 {
		vm.handlers = vm.handlers[:len(vm.handlers)-1]
	}
	if hasFinally && len(vm.finallyHandlers) > 0 {
		vm.finallyHandlers = vm.finallyHandlers[:len(vm.finallyHandlers)-1]
	}
}

// handleNativeErrorExtended supports finally blocks during exception unwinding.
func (vm *VM) handleNativeErrorExtended(msg string) error {
	for i := len(vm.finallyHandlers) - 1; i >= 0; i-- {
		handler := vm.finallyHandlers[i]
		vm.currentFrame().ip = handler.FinallyIP
	}
	return vm.handleNativeError(msg)
}

// HandleFinallyOnReturn ensures finally blocks execute when a function returns.
func (vm *VM) HandleFinallyOnReturn() {
	for len(vm.finallyHandlers) > 0 {
		handler := vm.finallyHandlers[len(vm.finallyHandlers)-1]
		vm.finallyHandlers = vm.finallyHandlers[:len(vm.finallyHandlers)-1]
		vm.currentFrame().ip = handler.FinallyIP
	}
}
