package vm

import (
	"jabline/pkg/code"
	"jabline/pkg/object"
	"fmt"
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
	if err != nil { return err }
	vm.sp = vm.sp - numElements
	return vm.push(hash)
}

func (vm *VM) opIndex() error {
	index := vm.pop()
	left := vm.pop()
	return vm.executeIndexExpression(left, index)
}

func (vm *VM) opCall(ins code.Instructions, ip *int) error {
	numArgs := int(ins[*ip+1])
	*ip += 1
	return vm.executeCall(numArgs)
}

func (vm *VM) opReturnValue() error {
	returnValue := vm.pop()
	frame := vm.popFrame()
	vm.sp = frame.basePointer - 1
	return vm.push(returnValue)
}

func (vm *VM) opReturn() error {
	frame := vm.popFrame()
	vm.sp = frame.basePointer - 1
	return vm.push(Null)
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
	structDef, ok := structObj.(*object.Struct)
	if !ok {
		return fmt.Errorf("instance creation on non-struct: %s", structObj.Inspect())
	}
	
	instance := &object.Instance{
		StructName: structDef.Name,
		Fields:     fields,
	}
	return vm.push(instance)
}
