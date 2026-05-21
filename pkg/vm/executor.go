package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
	"strings"
)

func (vm *VM) executeCall(numArgs int) error {
	calleePos := vm.sp - numArgs - 1
	if calleePos < 0 {
		return fmt.Errorf("runtime error: stack underflow for callee. vm.sp=%d, numArgs=%d", vm.sp, numArgs)
	}
	callee := vm.stack[calleePos]

	switch callee := callee.(type) {
	case *object.BoundMethod:
		// Replace the BoundMethod on the stack with its receiver.
		// The receiver will now be at calleePos, and the arguments follow.
		// The executeCallClosure will then treat the receiver as the first argument (local 0).
		vm.stack[calleePos] = callee.Receiver
		// The underlying closure expects (Receiver + Args).
		// So expected params = numArgs + 1
		return vm.executeCallClosure(callee.Function, numArgs+1, nil)
	case *object.Closure:
		return vm.executeCallClosure(callee, numArgs, nil)
	case *object.InstantiatedFunction:
		return vm.executeCallClosure(callee.Closure, numArgs, callee.TypeArgs)
	case *object.Builtin:
		return vm.executeCallBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function: %T", callee)
	}
}

func (vm *VM) executeCallBuiltin(callee *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	if callee.Name == "recover" {
		result := vm.LastError
		vm.LastError = nil // Clear after recover
		vm.sp = vm.sp - numArgs - 1
		if result != nil {
			vm.push(result)
		} else {
			vm.push(Null)
		}
		return nil
	}

	result := callee.Fn(args...)
	
	// Error handling
	if errObj, ok := result.(*object.Error); ok {
		// Builtins returning explicitly Error() objects means we throw natively
		err := vm.handleNativeError(errObj.Message)
		if err == nil {
			return nil
		}
		return err
	}

	if p, ok := result.(*object.Panic); ok {
		err := vm.handleNativeError(p.Message)
		if err == nil {
			return nil // Exception handled, IP already updated
		}
		return err
	}
	vm.sp = vm.sp - numArgs - 1

	if result != nil {
		vm.push(result)
	} else {
		vm.push(Null)
	}
	return nil
}

func (vm *VM) executeCallClosure(cl *object.Closure, numArgs int, typeArgs map[string]string) error {
	if cl.Fn.IsAsync {
		// Async functions return a Channel immediately
		resultChannel := vm.executeAsyncCall(cl, numArgs)
		vm.sp = vm.sp - numArgs - 1 // Pop function and arguments
		return vm.push(resultChannel)
	}

	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", cl.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	for k, v := range typeArgs {
		frame.TypeArgs[k] = v
	}
	if cl.Globals != nil {
		frame.savedGlobals = vm.globals
		vm.globals = GlobalStoreFromSlice(cl.Globals)
	}
	if cl.Constants != nil {
		frame.savedConstants = vm.constants
		vm.constants = cl.Constants
	}
	if err := vm.pushFrame(frame); err != nil {
		return err
	}

	vm.sp = frame.basePointer + cl.Fn.NumLocals
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

	closure := &object.Closure{
		Fn:        function,
		Free:      free,
		Globals:   vm.globals.Snapshot(),
		Constants: vm.constants,
	}
	return vm.push(closure)
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	obj, err := vm.getProperty(left, index)
	if err != nil {
		return err
	}
	return vm.push(obj)
}

func (vm *VM) getProperty(left, index object.Object) (object.Object, error) {
	if left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ {
		return vm.getArrayIndex(left, index)
	}
	if left.Type() == object.HASH_OBJ {
		return vm.getHashIndex(left, index)
	}
	if left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ {
		return vm.getStringIndex(left, index)
	}
	if left.Type() == object.INSTANCE_OBJ && index.Type() == object.STRING_OBJ {
		return vm.getInstanceIndex(left, index)
	}
	if left.Type() == object.SERVICE_OBJ && index.Type() == object.STRING_OBJ {
		return vm.getServiceIndex(left, index)
	}
	if left.Type() == object.ERROR_OBJ && index.Type() == object.STRING_OBJ {
		return vm.getErrorIndex(left, index)
	}
	if left.Type() == object.PANIC_OBJ && index.Type() == object.STRING_OBJ {
		return vm.getPanicIndex(left, index)
	}
	return nil, fmt.Errorf("index operator not supported: %s", left.Type())
}

func (vm *VM) getStringIndex(str, index object.Object) (object.Object, error) {
	s := str.(*object.String).Value
	idx := index.(*object.Integer).Value
	max := int64(len(s) - 1)

	if idx < 0 || idx > max {
		return nil, fmt.Errorf("runtime error: index out of bounds: %d for string of length %d", idx, len(s))
	}

	return &object.String{Value: string(s[idx])}, nil
}

func (vm *VM) getArrayIndex(array, index object.Object) (object.Object, error) {
	arrayObj := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)

	if idx < 0 || idx > max {
		return nil, fmt.Errorf("runtime error: index out of bounds: %d for array of length %d", idx, len(arrayObj.Elements))
	}

	return arrayObj.Elements[idx], nil
}

func (vm *VM) getHashIndex(hash, index object.Object) (object.Object, error) {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return nil, fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return Null, nil
	}

	return pair.Value, nil
}

func (vm *VM) getInstanceIndex(instance, index object.Object) (object.Object, error) {
	instObj := instance.(*object.Instance)
	fieldName := index.(*object.String).Value

	val, ok := instObj.Fields[fieldName]
	if ok {
		return val, nil
	}

	// Try Method Lookup
	if methods, ok := vm.methods[instObj.StructName]; ok {
		if methodClosure, ok := methods[fieldName]; ok {
			boundMethod := &object.BoundMethod{
				Receiver: instance,
				Function: methodClosure,
			}
			return boundMethod, nil
		}
	}

	return nil, fmt.Errorf("field or method '%s' not found in instance of '%s'", fieldName, instObj.StructName)
}

func (vm *VM) getServiceIndex(service, index object.Object) (object.Object, error) {
	serviceObj := service.(*object.Service)
	fieldName := index.(*object.String).Value

	if fieldName == "start" {
		return &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				return vm.StartService(serviceObj)
			},
		}, nil
	}

	val, ok := serviceObj.Config[fieldName]
	if ok {
		return val, nil
	}

	if methods, ok := vm.methods[serviceObj.Name]; ok {
		if methodClosure, ok := methods[fieldName]; ok {
			boundMethod := &object.BoundMethod{
				Receiver: service,
				Function: methodClosure,
			}
			return boundMethod, nil
		}
	}

	return nil, fmt.Errorf("field or method '%s' not found in service '%s'", fieldName, serviceObj.Name)
}

func (vm *VM) getErrorIndex(errObj, index object.Object) (object.Object, error) {
	e := errObj.(*object.Error)
	fieldName := index.(*object.String).Value

	if fieldName == "Message" {
		return &object.String{Value: e.Message}, nil
	}

	return nil, fmt.Errorf("field '%s' not found in Error object", fieldName)
}

func (vm *VM) getPanicIndex(panicObj, index object.Object) (object.Object, error) {
	p := panicObj.(*object.Panic)
	fieldName := index.(*object.String).Value

	if fieldName == "Message" {
		return &object.String{Value: p.Message}, nil
	}

	return nil, fmt.Errorf("field '%s' not found in Panic object", fieldName)
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

func (vm *VM) opInstantiate(ins code.Instructions, ip *int) error {
	numTypes := int(ins[*ip+1])
	*ip += 1

	typeArgs := make([]string, numTypes)
	for i := numTypes - 1; i >= 0; i-- {
		typeArgs[i] = vm.pop().(*object.String).Value
	}

	obj := vm.pop()

	switch o := obj.(type) {
	case *object.Struct:
		typeArgsMap := make(map[string]string)
		for i, tp := range o.TypeParameters {
			if i < len(typeArgs) {
				typeArgsMap[tp] = typeArgs[i]
			}
		}
		fullTypeName := o.Name + "[" + strings.Join(typeArgs, ", ") + "]"
		return vm.push(&object.InstantiatedStruct{
			Struct:       o,
			FullTypeName: fullTypeName,
		})
	case *object.Closure:
		typeArgsMap := make(map[string]string)
		for i, tp := range o.Fn.TypeParameters {
			if i < len(typeArgs) {
				typeArgsMap[tp] = typeArgs[i]
			}
		}
		fullTypeName := o.Fn.Name + "[" + strings.Join(typeArgs, ", ") + "]"
		return vm.push(&object.InstantiatedFunction{
			Closure:      o,
			TypeArgs:     typeArgsMap,
			FullTypeName: fullTypeName,
		})
	default:
		return fmt.Errorf("instantiation on non-struct/func: %s", obj.Inspect())
	}
}

func (vm *VM) opCheckType(ins code.Instructions, ip *int) error {
	typeIdx := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2

	// Peek at the value on top of the stack (don't pop it!)
	val := vm.stack[vm.sp-1]

	expectedTypeStr := vm.constants[typeIdx].(*object.String).Value
	frame := vm.currentFrame()

	// Resolver tipo si es un parámetro genérico
	if frame.TypeArgs != nil {
		if resolved, ok := frame.TypeArgs[expectedTypeStr]; ok {
			expectedTypeStr = resolved
		}
	}

	return vm.executeIsType(val, expectedTypeStr)
}

func (vm *VM) executeIsType(val object.Object, expectedTypeStr string) error {
	actualTypeStr := string(val.Type())

	// Special mapping for standard names vs internal ObjectType
	if expectedTypeStr == "int" && actualTypeStr == "INTEGER" {
		return nil
	}
	if expectedTypeStr == "string" && actualTypeStr == "STRING" {
		return nil
	}
	if expectedTypeStr == "bool" && actualTypeStr == "BOOLEAN" {
		return nil
	}
	if expectedTypeStr == "float" && actualTypeStr == "FLOAT" {
		return nil
	}

	// Dynamic Interface Duck Typing Check
	if typeObj, isType := vm.Types[expectedTypeStr]; isType {
		if iface, isInterface := typeObj.(*object.Interface); isInterface {
			if err := vm.checkInterfaceCompliance(val, iface); err != nil {
				return err
			}
			return nil
		}
	}

	// Complex types like Arrays, Maps, Functions could be checked further.
	// For basic struct instance checking:
	if actualTypeStr == "INSTANCE" {
		inst := val.(*object.Instance)
		if inst.StructName == expectedTypeStr {
			return nil
		}
		return fmt.Errorf("type error: expected type %s, got instance of %s", expectedTypeStr, inst.StructName)
	}

	if expectedTypeStr != string(actualTypeStr) {
		return fmt.Errorf("type error: expected type %s, got %s", expectedTypeStr, actualTypeStr)
	}

	return nil
}

func (vm *VM) checkInterfaceCompliance(val object.Object, iface *object.Interface) error {
	inst, ok := val.(*object.Instance)
	if !ok {
		return fmt.Errorf("type error: expected object implementing interface '%s', but got base type %s", iface.Name, val.Type())
	}

	methods, hasMethods := vm.methods[inst.StructName]

	for methodName, requiredMethod := range iface.Methods {
		// Ensure the method exists
		if !hasMethods {
			return fmt.Errorf("type error: object '%s' implements no methods, missing '%s' for interface '%s'", inst.StructName, methodName, iface.Name)
		}
		actualClosure, methodExists := methods[methodName]
		if !methodExists {
			return fmt.Errorf("type error: object '%s' missing required method '%s' for interface '%s'", inst.StructName, methodName, iface.Name)
		}

		// Ensure parameter count matches (methods implicitly have 'self' as first param, so subtract 1)
		actualParams := actualClosure.Fn.NumParameters - 1
		requiredParams := len(requiredMethod.Parameters)
		if actualParams != requiredParams {
			return fmt.Errorf("type error: method '%s' in '%s' requires %d explicit parameters, but interface '%s' expects %d", methodName, inst.StructName, actualParams, iface.Name, requiredParams)
		}
	}

	return nil
}
