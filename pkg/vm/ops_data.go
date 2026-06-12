package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) opArray(ins code.Instructions, ip *int) error {
	numElements := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	array := vm.buildArray(vm.sp-numElements, vm.sp)
	vm.sp = vm.sp - numElements
	return vm.push(array)
}

func (vm *VM) opHash(ins code.Instructions, ip *int) error {
	numElements := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
	if err != nil {
		return err
	}
	vm.sp = vm.sp - numElements
	return vm.push(hash)
}

func (vm *VM) opIndex() error {
	index := vm.pop()
	left := vm.pop()
	return vm.executeIndexExpression(left, index)
}

func (vm *VM) opSetProperty() error {
	val := vm.pop()
	index := vm.pop()
	left := vm.pop()

	switch obj := left.(type) {
	case *object.Instance:
		key, ok := index.(*object.String)
		if !ok {
			return fmt.Errorf("property name must be string, got %s", index.Type())
		}
		if obj.Fields == nil {
			obj.Fields = make(map[string]object.Object)
		}
		obj.Fields[key.Value] = val
		return nil

	case *object.Service:
		key, ok := index.(*object.String)
		if !ok {
			return fmt.Errorf("property name must be string, got %s", index.Type())
		}
		if obj.Config == nil {
			obj.Config = make(map[string]object.Object)
		}
		obj.Config[key.Value] = val
		return nil

	case *object.Hash:
		key, ok := index.(object.Hashable)
		if !ok {
			return fmt.Errorf("unusable as hash key: %s", index.Type())
		}
		if obj.Pairs == nil {
			obj.Pairs = make(map[object.HashKey]object.HashPair)
		}
		obj.Pairs[key.HashKey()] = object.HashPair{Key: index, Value: val}
		return nil

	case *object.Array:
		idxObj, ok := index.(*object.Integer)
		if !ok {
			return fmt.Errorf("array index must be integer, got %s", index.Type())
		}
		idx := idxObj.Value
		if idx < 0 || idx >= int64(len(obj.Elements)) {
			return fmt.Errorf("index out of bounds: %d", idx)
		}
		obj.Elements[idx] = val
		return nil

	default:
		return fmt.Errorf("assignment not supported for %s", left.Type())
	}
}

func (vm *VM) opCall(ins code.Instructions, ip *int) error {
	numArgs := int(ins[*ip+1])
	*ip += 1

	return vm.executeCall(numArgs)
}

func (vm *VM) opReturnValue() error {
	frame := vm.popFrame()
	returnValue := vm.pop()
	if vm.framesIndex == 0 {
		vm.sp = 0
		vm.stack[0] = returnValue
		vm.sp++
		ReleaseFrame(frame)
		return nil
	}
	vm.sp = frame.basePointer - 1
	vm.stack[vm.sp] = returnValue
	vm.sp++
	ReleaseFrame(frame)
	return nil
}

func (vm *VM) opReturn() error {
	frame := vm.popFrame()
	if vm.framesIndex == 0 {
		vm.sp = 0
		vm.stack[0] = Null
		vm.sp++
		ReleaseFrame(frame)
		return nil
	}
	vm.sp = frame.basePointer - 1
	vm.stack[vm.sp] = Null
	vm.sp++
	ReleaseFrame(frame)
	return nil
}

// opReturnValueDefer handles OpReturnValue with deferred call support.
func (vm *VM) opReturnValueDefer() error {
	frame := vm.currentFrame()

	// If we're in a deferred continuation, a deferred call just returned.
	if frame.deferIndex >= 0 {
		// Pop the deferred call's return value (discard it)
		vm.pop()

		frame.deferIndex--
		if frame.deferIndex >= 0 {
			// Execute next deferred call (LIFO)
			call := frame.deferred[frame.deferIndex]
			vm.push(call.Fn)
			for _, arg := range call.Args {
				vm.push(arg)
			}
			frame.ip -= 1 // Re-dispatch OpReturnValue after call returns
			return vm.executeCall(len(call.Args))
		}

		// All deferred calls done, complete the original return
		retVal := frame.savedReturn
		frame.deferred = nil
		frame.deferIndex = -1
		frame.savedReturn = nil

		vm.popFrame() // Pop this frame
		if vm.framesIndex == 0 {
			vm.sp = 0
			vm.stack[0] = retVal
			vm.sp++
			ReleaseFrame(frame)
			return nil
		}
		vm.sp = frame.basePointer - 1
		vm.stack[vm.sp] = retVal
		vm.sp++
		ReleaseFrame(frame)
		return nil
	}

	// Normal case: check if this frame has deferred calls
	if len(frame.deferred) > 0 {
		returnValue := vm.pop()
		frame.savedReturn = returnValue
		frame.deferIndex = len(frame.deferred)

		// Execute the first deferred call (from end = LIFO)
		frame.deferIndex = len(frame.deferred) - 1
		call := frame.deferred[frame.deferIndex]
		vm.push(call.Fn)
		for _, arg := range call.Args {
			vm.push(arg)
		}
		frame.ip -= 1 // After the deferred call returns, re-dispatch OpReturnValue
		return vm.executeCall(len(call.Args))
	}

	return vm.opReturnValue()
}

// opReturnDefer handles OpReturn with deferred call support (same as opReturnValueDefer but with Null return).
func (vm *VM) opReturnDefer() error {
	frame := vm.currentFrame()

	if frame.deferIndex >= 0 {
		vm.pop()
		frame.deferIndex--
		if frame.deferIndex >= 0 {
			call := frame.deferred[frame.deferIndex]
			vm.push(call.Fn)
			for _, arg := range call.Args {
				vm.push(arg)
			}
			frame.ip -= 1
			return vm.executeCall(len(call.Args))
		}

		retVal := frame.savedReturn
		frame.deferred = nil
		frame.deferIndex = -1
		frame.savedReturn = nil

		vm.popFrame()
		if vm.framesIndex == 0 {
			vm.sp = 0
			vm.stack[0] = retVal
			vm.sp++
			ReleaseFrame(frame)
			return nil
		}
		vm.sp = frame.basePointer - 1
		vm.stack[vm.sp] = retVal
		vm.sp++
		ReleaseFrame(frame)
		return nil
	}

	if len(frame.deferred) > 0 {
		frame.savedReturn = Null
		frame.deferIndex = len(frame.deferred) - 1
		call := frame.deferred[frame.deferIndex]
		vm.push(call.Fn)
		for _, arg := range call.Args {
			vm.push(arg)
		}
		frame.ip -= 1
		return vm.executeCall(len(call.Args))
	}

	return vm.opReturn()
}

func (vm *VM) opSlice() error {
	highObj := vm.pop()
	lowObj := vm.pop()
	left := vm.pop()

	arr, ok := left.(*object.Array)
	if !ok {
		return fmt.Errorf("slice operator requires array, got %s", left.Type())
	}

	length := int64(len(arr.Elements))

	var low int64
	if _, isNull := lowObj.(*object.Null); isNull {
		low = 0
	} else if lowInt, ok := lowObj.(*object.Integer); ok {
		low = lowInt.Value
	} else {
		return fmt.Errorf("slice lower bound must be integer or null, got %s", lowObj.Type())
	}

	var high int64
	if _, isNull := highObj.(*object.Null); isNull {
		high = length
	} else if highInt, ok := highObj.(*object.Integer); ok {
		high = highInt.Value
	} else {
		return fmt.Errorf("slice upper bound must be integer or null, got %s", highObj.Type())
	}

	if low < 0 {
		low = 0
	}
	if high > length {
		high = length
	}
	if low >= high {
		return vm.push(&object.Array{Elements: []object.Object{}})
	}

	elements := make([]object.Object, high-low)
	copy(elements, arr.Elements[low:high])
	return vm.push(&object.Array{Elements: elements})
}

func (vm *VM) opBuildArrayWithSpread(ins code.Instructions, ip *int) error {
	numSlots := int(ins[*ip+1])
	*ip++
	spreadMask := code.ReadUint16(ins[*ip+1:])
	*ip += 2

	slots := make([]object.Object, numSlots)
	for i := numSlots - 1; i >= 0; i-- {
		slots[i] = vm.pop()
	}

	var elements []object.Object
	for i, slot := range slots {
		if spreadMask&(1<<i) != 0 {
			arr, ok := slot.(*object.Array)
			if !ok {
				return fmt.Errorf("spread: expected array, got %s", slot.Type())
			}
			elements = append(elements, arr.Elements...)
		} else {
			elements = append(elements, slot)
		}
	}
	return vm.push(&object.Array{Elements: elements})
}

func (vm *VM) opCallSpread(numArgs int, spreadMask uint16) error {
	slots := make([]object.Object, numArgs)
	for i := numArgs - 1; i >= 0; i-- {
		slots[i] = vm.pop()
	}

	var args []object.Object
	for i, slot := range slots {
		if spreadMask&(1<<i) != 0 {
			arr, ok := slot.(*object.Array)
			if !ok {
				return fmt.Errorf("spread argument: expected array, got %s", slot.Type())
			}
			args = append(args, arr.Elements...)
		} else {
			args = append(args, slot)
		}
	}

	fn := vm.pop()
	return vm.executeCallFn(fn, args)
}

func (vm *VM) opInstance(ins code.Instructions, ip *int) error {
	numFields := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2

	fields := make(map[string]object.Object)
	for i := 0; i < numFields; i++ {
		value := vm.pop()
		keyObj := vm.pop()
		keyStr := keyObj.(*object.String).Value
		fields[keyStr] = value
	}

	structObj := vm.pop()
	var structName string

	switch s := structObj.(type) {
	case *object.Struct:
		structName = s.Name
	case *object.InstantiatedStruct:
		structName = s.FullTypeName
	default:
		return fmt.Errorf("instance creation on non-struct: %s", structObj.Inspect())
	}

	instance := &object.Instance{
		StructName: structName,
		Fields:     fields,
	}
	return vm.push(instance)
}

func (vm *VM) opDefer(numArgs int) error {
	fn := vm.stack[vm.sp-numArgs-1]
	args := make([]object.Object, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = vm.stack[vm.sp-numArgs+i]
	}
	vm.sp = vm.sp - numArgs - 1

	frame := vm.currentFrame()
	frame.deferred = append(frame.deferred, DeferredCall{Fn: fn, Args: args})
	return nil
}
