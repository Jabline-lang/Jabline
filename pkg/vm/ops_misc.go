package vm

import (
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
		return vm.newRuntimeError("uncaught exception: %s", exception.Inspect())
	}
	handler := vm.popHandler()
	vm.framesIndex = handler.FrameIndex
	vm.sp = handler.StackSP
	vm.push(exception)

	vm.currentFrame().ip = handler.CatchIP - 1

	return nil
}
