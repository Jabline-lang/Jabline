package vm

import (
	"context"
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"os"
)

// MaxConcurrentSpawns limits the number of active goroutines created by spawn
// to prevent fork bombs and out-of-memory errors.
const MaxConcurrentSpawns = 10000

var spawnLimiter = make(chan struct{}, MaxConcurrentSpawns)

func ExecuteClosureBridge(closureObj object.Object, args []object.Object) object.Object {
	callee, ok := closureObj.(*object.Closure)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("bridge expected closure, got %s", closureObj.Type())}
	}

	// Concurrencia nativa: si hay un VM principal activo, usar Fork()
	// para heredar globals, methods y Types.
	if GlobalVM != nil {
		fork := GlobalVM.Fork()
		return fork.RunClosure(callee, args)
	}

	// Fallback: crear un VM minimal con los datos del closure capturado
	constants := callee.Constants
	if constants == nil {
		constants = []object.Object{}
	}
	globals := callee.Globals
	if globals == nil {
		globals = make([]object.Object, GlobalsSize)
	}

	newVM := &VM{
		constants:   constants,
		stack:       make([]object.Object, InitialStackSize),
		sp:          0,
		globals:     GlobalStoreFromSlice(globals),
		frames:      make([]*Frame, InitialFrames),
		framesIndex: 0,
		loader:      GlobalLoader,
		Ctx:         context.Background(),
	}
	newVM.push(Null)
	for _, arg := range args {
		if newVM.sp >= StackSize {
			return &object.Error{Message: "stack overflow in bridge"}
		}
		newVM.stack[newVM.sp] = arg
		newVM.sp++
	}
	frame := NewFrame(callee, 1)
	if err := newVM.pushFrame(frame); err != nil {
		return &object.Error{Message: err.Error()}
	}
	newVM.sp = frame.basePointer + callee.Fn.NumLocals

	if err := newVM.Run(); err != nil {
		return &object.Error{Message: err.Error()}
	}
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
	childCtx, cancel := context.WithCancel(vm.Ctx)
	chanObj := &object.Channel{Value: resultChan, Cancel: cancel}

	constants := vm.constants
	filename := vm.filename
	loader := vm.loader
	globals := vm.globals // Capture globals from current VM

	go func() {
		spawnLimiter <- struct{}{} // Acquire semaphore slot
		defer func() { <-spawnLimiter }() // Release slot

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
			stack:       make([]object.Object, InitialStackSize),
			sp:          0,
			globals:     globals, // Use captured globals
			frames:      make([]*Frame, InitialFrames),
			framesIndex: 0, // Start with 0 frames, we'll push one
			filename:    filename,
			loader:      loader,
			Ctx:         childCtx,
			methods:     vm.methods,
			Types:       vm.Types,
		}

		// Push a dummy object at stack[0] as the 'function' slot for OpReturnValue to overwrite
		asyncVM.stack[0] = Null
		asyncVM.sp = 1

		// Push the arguments onto the asyncVM's stack
		for i := 0; i < numArgs; i++ {
			asyncVM.stack[1+i] = args[i]
		}
		asyncVM.sp = 1 + numArgs // Stack pointer is now past dummy + arguments

		// Create a new frame for the closure
		// basePointer should be 1 because args are after the dummy at 0
		asyncFrame := NewFrame(callee, 1) // basePointer for the arguments
		if err := asyncVM.pushFrame(asyncFrame); err != nil {
			resultChan <- &object.Error{Message: err.Error()}
			close(resultChan)
			return
		}

		// Update stack pointer for the asyncVM to reflect the new frame and its locals
		asyncVM.sp = asyncFrame.basePointer + callee.Fn.NumLocals

		err := asyncVM.Run() // Run this specific function in its own VM
		if err != nil {
			resultChan <- &object.Error{Message: err.Error()}
			close(resultChan)
			return
		}

		var result object.Object = Null
		// OpReturnValue puts the result at basePointer-1 (index 0)
		result = asyncVM.stack[0]
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

	// Context for child VM
	childCtx, cancel := context.WithCancel(vm.Ctx)
	chanObj := &object.Channel{Value: resultChan, Cancel: cancel}

	// Capture constants and loader context
	constants := vm.constants
	filename := vm.filename
	loader := vm.loader

	go func() {
		// Create new VM
		newVM := NewWithLoader(code.Instructions{}, constants, filename, loader)
		newVM.Ctx = childCtx
		newVM.Cancel = cancel
		newVM.globals = vm.globals // Share globals with child process

		// Setup supervisor recovery for this spawn worker
		defer func() {
			if r := recover(); r != nil {
				// We don't crash the host VM. Instead, we emit a Supervisor error.
				fmt.Fprintf(os.Stderr, "[Supervisor] Background process panicked: %v\n", r)

				// Optional: Send the error down the result channel so await doesn't hang forever
				// if it was waiting for a result from a process that just died.
				select {
				case resultChan <- &object.Error{Message: fmt.Sprintf("panic: %v", r)}:
				default:
				}
				close(resultChan)
			}
		}()

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
