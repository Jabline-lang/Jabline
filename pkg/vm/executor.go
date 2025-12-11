package vm

import (
	"fmt"
	"jabline/pkg/object"
)

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	
	switch callee := callee.(type) {
	case *object.Closure:
		if numArgs != callee.Fn.NumParameters {
			return fmt.Errorf("wrong number of arguments: want=%d, got=%d", 
				callee.Fn.NumParameters, numArgs)
		}
		
		frame := NewFrame(callee, vm.sp-numArgs)
		vm.pushFrame(frame)
		
		vm.sp = frame.basePointer + callee.Fn.NumLocals

	case *object.Builtin:
		args := vm.stack[vm.sp-numArgs : vm.sp]

		result := callee.Fn(args...)
		vm.sp = vm.sp - numArgs - 1

		if result != nil {
			vm.push(result)
		} else {
			vm.push(Null)
		}
		
	default:
		return fmt.Errorf("calling non-function: %T", callee)
	}
	return nil
}

func (vm *VM) pushClosure(constIndex, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree

	closure := &object.Closure{Fn: function, Free: free}
	return vm.push(closure)
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	if left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ {
		return vm.executeArrayIndex(left, index)
	}
	if left.Type() == object.HASH_OBJ {
		return vm.executeHashIndex(left, index)
	}
	if left.Type() == object.INSTANCE_OBJ && index.Type() == object.STRING_OBJ {
		return vm.executeInstanceIndex(left, index)
	}
	return fmt.Errorf("index operator not supported: %s", left.Type())
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObj := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)
	
	if idx < 0 || idx > max {
		return vm.push(Null)
	}
	
	return vm.push(arrayObj.Elements[idx])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}
	
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}
	
	return vm.push(pair.Value)
}

func (vm *VM) executeInstanceIndex(instance, index object.Object) error {
	instObj := instance.(*object.Instance)
	fieldName := index.(*object.String).Value
	
	val, ok := instObj.Fields[fieldName]
	if !ok {
		return fmt.Errorf("field '%s' not found in instance of '%s'", fieldName, instObj.StructName)
	}
	return vm.push(val)
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	
	for i := 0; i < len(elements); i++ {
		elements[i] = vm.stack[startIndex+i]
	}
	
	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)
	
	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		
		pair := object.HashPair{Key: key, Value: value}
		
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}
		
		hashedPairs[hashKey.HashKey()] = pair
	}
	
	return &object.Hash{Pairs: hashedPairs}, nil
}
