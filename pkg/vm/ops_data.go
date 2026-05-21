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

	err := vm.executeCall(numArgs)
	if err != nil {
		return err
	}

	// For builtins: executeCall pops args and func, pushes result. vm.sp is OK.
	// For closures: opReturnValue pushes result AFTER func. Stack [func, res].
	// We need to distinguish builtins vs closures?
	// But executeCall doesn't return info.

	// Wait! Builtins behavior.
	// executeCall for builtin: vm.sp = vm.sp - numArgs - 1. vm.push(result).
	// This overwrites func_obj position with result.
	// So stack is [result].

	// If opReturnValue for closure: vm.push(returnValue).
	// Stack is [func_obj, result].

	// So builtins and closures behave differently regarding stack!
	// This is the root cause.

	// I should make builtins behave like closures: push result after func_obj.
	// Or make closures behave like builtins: overwrite func_obj.

	// I already tried making closures behave like builtins (overwrite).
	// And it caused stack underflow.

	// If I use the current opReturnValue (push after), I must make builtins do the same.
	// Or check the type of callee? But callee is gone.

	// Let's modify executeCall for builtins to NOT overwrite func_obj, but push after.
	// vm.sp = vm.sp - numArgs. (points to func_obj + 1).
	// vm.push(result).

	// Then opCall can always do vm.stack[vm.sp-2] = vm.stack[vm.sp-1]; vm.sp--.

	// Let's modify executeCall for builtins in pkg/vm/executor.go.
	return nil
}

func (vm *VM) opReturnValue() error {

	frame := vm.popFrame()

	// If this was the last (top-level) frame, leave the return value where it is on the stack.
	if vm.framesIndex == 0 {
		ReleaseFrame(frame)
		return nil
	}

	returnValue := vm.pop()

	vm.sp = frame.basePointer - 1

	vm.stack[vm.sp] = returnValue

	vm.sp++

	ReleaseFrame(frame)

	return nil

}

func (vm *VM) opReturn() error {

	frame := vm.popFrame()

	if vm.framesIndex == 0 {
		ReleaseFrame(frame)
		return nil
	}

	vm.sp = frame.basePointer - 1

	vm.stack[vm.sp] = Null

	vm.sp++

	ReleaseFrame(frame)

	return nil

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
