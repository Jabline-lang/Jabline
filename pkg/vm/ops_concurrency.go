package vm

import (
	"context"
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/log"
	"jabline/pkg/object"
	"jabline/pkg/sandbox"
	"sync"
)

// MaxConcurrentSpawns limits the number of active goroutines created by spawn
// to prevent fork bombs and out-of-memory errors.
// Can be changed at runtime (e.g., jabline.MaxConcurrentSpawns = 50000).
var MaxConcurrentSpawns = 10000

var spawnLimiterOnce sync.Once
var spawnLimiter chan struct{}

func getSpawnLimiter() chan struct{} {
	spawnLimiterOnce.Do(func() {
		spawnLimiter = make(chan struct{}, MaxConcurrentSpawns)
	})
	return spawnLimiter
}

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
		ReleaseFrame(frame)
		return &object.Error{Message: err.Error()}
	}
	newVM.sp = frame.basePointer + callee.Fn.NumLocals

	if err := newVM.Run(); err != nil {
		if newVM.framesIndex > 0 && newVM.frames[0] == frame {
			newVM.frames[0] = nil
			newVM.framesIndex = 0
			ReleaseFrame(frame)
		}
		return &object.Error{Message: err.Error()}
	}
	return newVM.stack[0]
}

func (vm *VM) executeAsyncCall(callee *object.Closure, numArgs int) object.Object {
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

	// Deep copy maps to avoid data races between goroutines
	methodsCopy := make(map[string]map[string]*object.Closure, len(vm.methods))
	for k, v := range vm.methods {
		innerCopy := make(map[string]*object.Closure, len(v))
		for mk, mv := range v {
			innerCopy[mk] = mv
		}
		methodsCopy[k] = innerCopy
	}
	typesCopy := make(map[string]object.Object)
	for k, v := range vm.Types {
		typesCopy[k] = v
	}

		go func() {
			lim := getSpawnLimiter()
			lim <- struct{}{}
			defer func() { <-lim }()

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("async task panicked: %v", r)
					}
					select {
					case resultChan <- &object.Error{Message: err.Error()}:
					case <-childCtx.Done():
					}
				}
			}()

			asyncVM := &VM{
				constants:   constants,
				stack:       make([]object.Object, InitialStackSize),
				sp:          0,
				globals:     GlobalStoreFromSlice(vm.globals.Snapshot()),
			frames:      make([]*Frame, InitialFrames),
			framesIndex: 0,
			filename:    filename,
			loader:      loader,
			Ctx:         childCtx,
			Cancel:      cancel,
			methods:     methodsCopy,
			Types:       typesCopy,
			Sandbox:     vm.Sandbox,
		}

		asyncVM.stack[0] = Null
		asyncVM.sp = 1

		for i := 0; i < numArgs; i++ {
			asyncVM.stack[1+i] = args[i]
		}
		asyncVM.sp = 1 + numArgs

		asyncFrame := NewFrame(callee, 1)
		if err := asyncVM.pushFrame(asyncFrame); err != nil {
			ReleaseFrame(asyncFrame)
			select {
			case resultChan <- &object.Error{Message: err.Error()}:
			case <-childCtx.Done():
			}
			close(resultChan)
			return
		}
		asyncVM.sp = asyncFrame.basePointer + callee.Fn.NumLocals

		err := asyncVM.Run()
		if asyncVM.framesIndex > 0 && asyncVM.frames[0] == asyncFrame {
			asyncVM.frames[0] = nil
			asyncVM.framesIndex = 0
			ReleaseFrame(asyncFrame)
		}
		if err != nil {
			select {
			case resultChan <- &object.Error{Message: err.Error()}:
			case <-childCtx.Done():
			}
			close(resultChan)
			return
		}

		result := asyncVM.stack[0]
		select {
		case resultChan <- result:
		case <-childCtx.Done():
		}
		close(resultChan)
	}()

	return chanObj
}

func (vm *VM) opSpawn(ins code.Instructions, ip *int) error {
	if err := vm.CheckPermission(sandbox.PermSpawn, "spawn"); err != nil {
		return err
	}

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
		newVM.globals = GlobalStoreFromSlice(vm.globals.Snapshot()) // Isolated copy of globals
		newVM.Sandbox = vm.Sandbox // Inherit sandbox policy

		// Setup supervisor recovery for this spawn worker
		defer func() {
			// Release main frame to avoid memory leak
			if newVM.framesIndex > 0 && newVM.frames[0] != nil {
				mainFrame := newVM.frames[0]
				newVM.frames[0] = nil
				newVM.framesIndex = 0
				ReleaseFrame(mainFrame)
			}
		}()
		defer func() {
			if r := recover(); r != nil {
				// We don't crash the host VM. Instead, we emit a Supervisor error.
				log.Error("Supervisor background process panicked", "panic", r)

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
		select {
		case val, ok := <-ch.Value:
			if !ok {
				return vm.push(Null)
			}
			return vm.push(val)
		case <-vm.Ctx.Done():
			if ch.Cancel != nil {
				ch.Cancel()
			}
			return vm.push(Null)
		}

	case *object.RemoteChannel:
		val, err := ch.Receive()
		if err != nil {
			return fmt.Errorf("remote channel receive error: %s", err)
		}
		return vm.push(val)

	case *object.Promise:
		switch ch.State {
		case object.RESOLVED:
			return vm.push(ch.Value)
		case object.REJECTED:
			return vm.push(ch.Reason)
		default:
			return vm.push(Null)
		}

	default:
		return fmt.Errorf("can only await on a channel or promise, got %s", obj.Type())
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

func (vm *VM) opNextItem() error {
	indexObj := vm.pop()
	iterable := vm.pop()

	// Determine index for array/string iteration
	index := int64(0)
	if idx, ok := indexObj.(*object.Integer); ok {
		index = idx.Value
	}

	switch it := iterable.(type) {
	case *object.Array:
		if index < int64(len(it.Elements)) {
			val := it.Elements[index]
			vm.push(val)
			nextIndex := &object.Integer{Value: index + 1}
			vm.push(nextIndex)
			return vm.push(True)
		}
		// Exhausted
		vm.push(Null)
		vm.push(Null)
		return vm.push(False)

	case *object.String:
		runes := []rune(it.Value)
		if index < int64(len(runes)) {
			val := &object.String{Value: string(runes[index])}
			vm.push(val)
			nextIndex := &object.Integer{Value: index + 1}
			vm.push(nextIndex)
			return vm.push(True)
		}
		// Exhausted
		vm.push(Null)
		vm.push(Null)
		return vm.push(False)

	case *object.Channel:
		val, open := <-it.Value
		if open {
			vm.push(val)
			vm.push(&object.Integer{Value: 0})
			return vm.push(True)
		}
		// Channel closed
		vm.push(Null)
		vm.push(Null)
		return vm.push(False)

	default:
		return fmt.Errorf("cannot iterate over %s", iterable.Type())
	}
}
