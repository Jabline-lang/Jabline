package vm

import (
	"fmt"
	"jabline/pkg/object"
)

func (vm *VM) push(o object.Object) error {
	// Grow the stack dynamically if needed (doubles capacity each time)
	if vm.sp >= len(vm.stack) {
		if len(vm.stack) >= MaxStackSize {
			return fmt.Errorf("stack overflow: exceeded maximum stack depth of %d", MaxStackSize)
		}
		newSize := len(vm.stack) * 2
		if newSize > MaxStackSize {
			newSize = MaxStackSize
		}
		newStack := make([]object.Object, newSize)
		copy(newStack, vm.stack)
		vm.stack = newStack
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	if vm.sp == 0 {
		panic("stack underflow in VM.pop()")
	}
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedElement() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) error {
	// Grow the frame slice dynamically if needed
	if vm.framesIndex >= len(vm.frames) {
		if len(vm.frames) >= MaxFrames {
			return fmt.Errorf("stack overflow: maximum recursion depth of %d exceeded", MaxFrames)
		}
		newSize := len(vm.frames) * 2
		if newSize > MaxFrames {
			newSize = MaxFrames
		}
		newFrames := make([]*Frame, newSize)
		copy(newFrames, vm.frames)
		vm.frames = newFrames
	}
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
	return nil
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	frame := vm.frames[vm.framesIndex]
	if frame.savedGlobals != nil {
		vm.globals = frame.savedGlobals
	}
	if frame.savedConstants != nil {
		vm.constants = frame.savedConstants
	}
	return frame // We will let the caller ReleaseFrame
}

func (vm *VM) pushHandler(catchIP int) {
	handler := ExceptionHandler{
		CatchIP:    catchIP,
		StackSP:    vm.sp,
		FrameIndex: vm.framesIndex,
	}
	vm.handlers = append(vm.handlers, handler)
}

func (vm *VM) popHandler() ExceptionHandler {
	handler := vm.handlers[len(vm.handlers)-1]
	vm.handlers = vm.handlers[:len(vm.handlers)-1]
	return handler
}
