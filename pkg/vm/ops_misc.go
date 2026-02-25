package vm

import (
	"fmt"
	"jabline/pkg/code"
)

func (vm *VM) opTry(ins code.Instructions, ip *int) {
	catchPos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	vm.pushHandler(catchPos)
}

func (vm *VM) opEndTry() {
	vm.popHandler()
}

func (vm *VM) opThrow() error {
	exception := vm.pop()

	if len(vm.handlers) == 0 {
		return fmt.Errorf("uncaught exception: %s", exception.Inspect())
	}

	handler := vm.handlers[len(vm.handlers)-1]
	vm.handlers = vm.handlers[:len(vm.handlers)-1]

	// Unwind stack
	vm.sp = handler.StackSP
	vm.framesIndex = handler.FrameIndex
	vm.stack[vm.sp] = exception // Push exception back for catch block
	vm.sp++

	// Jump to catch block (offset by -1 because loop increments it)
	vm.currentFrame().ip = handler.CatchIP - 1

	return nil
}
