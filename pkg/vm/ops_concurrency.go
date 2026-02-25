package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func ExecuteClosureBridge(closureObj object.Object, args []object.Object) object.Object {
	callee, ok := closureObj.(*object.Closure)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("bridge expected closure, got %s", closureObj.Type())}
	}

	// Create a new VM for this execution (isolated request)
	// We need constants and globals.
	// Since this is a static method, how do we get constants?
	// We assume constants are part of the closure (Captured).
	// Actually, Closure struct HAS Constants and Globals!

	newVM := &VM{
		constants:   callee.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     callee.Globals, // Use captured globals
		frames:      make([]*Frame, MaxFrames),
		framesIndex: 0,
		// Loader and Filename are harder to get, assume defaults or Closure should carry them?
		// For now, empty filename is fine.
	}

	// Push a dummy object at stack[0] so that OpReturnValue has a place to write to
	// when it does vm.stack[basePointer-1] = result.
	newVM.push(Null)

	// Push arguments
	for _, arg := range args {
		if newVM.sp >= StackSize {
			return &object.Error{Message: "stack overflow in bridge"}
		}
		newVM.stack[newVM.sp] = arg
		newVM.sp++
	}

	// Setup Frame
	// basePointer is 1 because stack[0] is the dummy/return slot
	frame := NewFrame(callee, 1)
	newVM.pushFrame(frame)
	newVM.sp = frame.basePointer + callee.Fn.NumLocals

	// Run
	err := newVM.Run()
	if err != nil {
		return &object.Error{Message: err.Error()}
	}

	// Return Result
	// OpReturnValue put the result at basePointer-1, which is index 0.
	return newVM.stack[0]
}

func (vm *VM) executeAsyncCall(callee *object.Closure, numArgs int) object.Object {
	// Args are on stack at vm.sp-numArgs to vm.sp
	// We need to copy them

	args := make([]object.Object, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = vm.stack[vm.sp-numArgs+i]
	}

	resultChan := make(chan object.Object, 1)
	chanObj := &object.Channel{Value: resultChan}

	constants := vm.constants
	filename := vm.filename
	loader := vm.loader
	globals := vm.globals // Capture globals from current VM

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("async task panicked: %v", r)
				}
				resultChan <- &object.Error{Message: err.Error()}
			}
		}()

		// Manually set up the new VM for executing the specific closure
		asyncVM := &VM{
			constants:   constants,
			stack:       make([]object.Object, StackSize),
			sp:          0,
			globals:     globals, // Use captured globals
			frames:      make([]*Frame, MaxFrames),
			framesIndex: 0, // Start with 0 frames, we'll push one
			filename:    filename,
			loader:      loader,
		}

		// Push the arguments onto the asyncVM's stack
		// The arguments are [id, jobs, results]
		for i := 0; i < numArgs; i++ {
			asyncVM.stack[i] = args[i]
		}
		asyncVM.sp = numArgs // Stack pointer is now past the arguments

		// Create a new frame for the closure
		// basePointer should be 0 because args are already on stack and serve as the base
		asyncFrame := NewFrame(callee, asyncVM.sp-numArgs) // basePointer for the arguments
		asyncVM.pushFrame(asyncFrame)                      // Push this new frame

		// Update stack pointer for the asyncVM to reflect the new frame and its locals
		asyncVM.sp = asyncFrame.basePointer + callee.Fn.NumLocals

		err := asyncVM.Run() // Run this specific function in its own VM
		if err != nil {
			resultChan <- &object.Error{Message: err.Error()}
			close(resultChan)
			return
		}

		var result object.Object = Null
		// The result should be at the top of the stack when the function finishes
		if asyncVM.sp > asyncFrame.basePointer {
			result = asyncVM.stack[asyncVM.sp-1]
		}
		resultChan <- result
		close(resultChan)
	}()

	return chanObj
}

func (vm *VM) opSpawn(ins code.Instructions, ip *int) error {
	numArgs := int(ins[*ip+1])
	*ip += 1

	calleePos := vm.sp - 1 - numArgs
	callee := vm.stack[calleePos]

	args := make([]object.Object, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = vm.stack[calleePos+1+i]
	}

	// Pop from current stack
	vm.sp = calleePos

	// Create result channel
	resultChan := make(chan object.Object, 1)
	chanObj := &object.Channel{Value: resultChan}

	// Capture constants and loader context
	constants := vm.constants
	filename := vm.filename
	loader := vm.loader

	go func() {
		// Create new VM
		newVM := NewWithLoader(code.Instructions{}, constants, filename, loader)

		// Push callee and args
		newVM.push(callee)
		for _, arg := range args {
			newVM.push(arg)
		}

		// Setup call
		// We manually invoke executeCall to set up the frame
		err := newVM.executeCall(numArgs)
		if err != nil {
			resultChan <- &object.Error{Message: err.Error()}
			close(resultChan)
			return
		}

		// Run
		err = newVM.Run()
		if err != nil {
			resultChan <- &object.Error{Message: err.Error()}
			close(resultChan)
			return
		}

		// Result
		var result object.Object = Null
		if newVM.sp > 0 {
			result = newVM.stack[newVM.sp-1]
		}

		resultChan <- result
		close(resultChan)
	}()

	return vm.push(chanObj)
}

func (vm *VM) opAwait() error {
	obj := vm.pop()

	switch ch := obj.(type) {
	case *object.Channel:
		val, ok := <-ch.Value
		if !ok {
			return vm.push(Null) // Channel was closed or empty
		}
		return vm.push(val)

	case *object.RemoteChannel:
		val, err := ch.Receive()
		if err != nil {
			return fmt.Errorf("remote channel receive error: %s", err)
		}
		return vm.push(val)

	default:
		return fmt.Errorf("can only await on a channel, got %s", obj.Type())
	}
}

func (vm *VM) opSendChannel() error {
	val := vm.pop()
	chObj := vm.pop()

	ch, ok := chObj.(*object.Channel)
	if !ok {
		return fmt.Errorf("send to non-channel type: %T", chObj)
	}

	ch.Value <- val

	// channel expression evaluates to the sent value
	return vm.push(val)
}

func (vm *VM) opRecvChannel() error {
	chObj := vm.pop()

	ch, ok := chObj.(*object.Channel)
	if !ok {
		return fmt.Errorf("receive from non-channel type: %T", chObj)
	}

	val, ok := <-ch.Value
	if !ok {
		// channel is closed, push Null
		return vm.push(Null)
	}

	return vm.push(val)
}
