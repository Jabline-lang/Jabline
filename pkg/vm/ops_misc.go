package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) opTry(ins code.Instructions, ip *int) {
	catchPos := int(code.ReadUint16(ins[*ip+1:]))
	finallyPos := int(code.ReadUint16(ins[*ip+3:]))
	*ip += 4
	vm.pushHandler(catchPos, finallyPos)
}

func (vm *VM) opEndTry() {
	vm.popHandler()
}

func (vm *VM) opFinally() {
	// Push a finally handler marker to be executed on exception unwind
	// The finally block has already been compiled inline after the catch block.
	// This is just a marker for the VM to know we're entering a finally block.
	// No stack manipulation needed.
}

func (vm *VM) opEndFinally() {
	if vm.LastError == nil {
		return
	}

	if len(vm.handlers) == 0 {
		return
	}
	handler := vm.handlers[len(vm.handlers)-1]
	vm.handlers = vm.handlers[:len(vm.handlers)-1]

	vm.sp = handler.StackSP
	vm.framesIndex = handler.FrameIndex

	if handler.CatchIP > 0 {
		exception := vm.PendingException
		if exception == nil {
			exception = vm.LastError
		}
		vm.LastError = nil
		vm.PendingException = nil
		vm.stack[vm.sp] = exception
		vm.sp++
		vm.currentFrame().ip = handler.CatchIP - 1
	} else {
		vm.PendingException = nil
	}
}

func (vm *VM) opThrow() error {
	exception := vm.pop()

	if len(vm.handlers) == 0 {
		return fmt.Errorf("uncaught exception: %s", exception.Inspect())
	}

	handler := vm.handlers[len(vm.handlers)-1]

	// Unwind stack
	vm.sp = handler.StackSP
	vm.framesIndex = handler.FrameIndex

	if handler.FinallyIP > 0 {
		// Keep handler on stack — OpEndFinally will pop it and jump to catch
		vm.PendingException = exception
		vm.stack[vm.sp] = exception
		vm.sp++
		vm.currentFrame().ip = handler.FinallyIP - 1

		if errObj, ok := exception.(*object.Error); ok {
			vm.LastError = errObj
		} else {
			vm.LastError = &object.Error{Message: exception.Inspect()}
		}
	} else if handler.CatchIP > 0 {
		// Pop the handler — catch will handle it directly
		vm.handlers = vm.handlers[:len(vm.handlers)-1]
		vm.stack[vm.sp] = exception
		vm.sp++
		vm.currentFrame().ip = handler.CatchIP - 1
		if errObj, ok := exception.(*object.Error); ok {
			vm.LastError = errObj
		} else {
			vm.LastError = &object.Error{Message: exception.Inspect()}
		}
	} else {
		vm.handlers = vm.handlers[:len(vm.handlers)-1]
		return fmt.Errorf("uncaught exception: %s", exception.Inspect())
	}

	return nil
}
