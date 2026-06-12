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
		return &object.Error{Message: "stack underflow in VM.pop()"}
	}
	vm.sp--
	o := vm.stack[vm.sp]
	vm.stack[vm.sp] = nil
	return o
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedElement() object.Object {
	if vm.sp < 0 || vm.sp >= len(vm.stack) {
		return nil
	}
	return vm.stack[vm.sp]
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

// Exported accessors for DAP
func (vm *VM) FramesIndex() int {
	return vm.framesIndex
}

func (vm *VM) FrameAt(i int) *Frame {
	if i < 0 || i >= vm.framesIndex {
		return nil
	}
	return vm.frames[i]
}

func (vm *VM) CurrentFrame() *Frame {
	return vm.currentFrame()
}

func (vm *VM) StackAt(i int) object.Object {
	if i < 0 || i >= vm.sp {
		return nil
	}
	return vm.stack[i]
}

func (vm *VM) SP() int {
	return vm.sp
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

func (vm *VM) pushHandler(catchIP int, finallyIP int) {
	handler := ExceptionHandler{
		CatchIP:    catchIP,
		FinallyIP:  finallyIP,
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
