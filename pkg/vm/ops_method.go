package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) opRegisterMethod(ins code.Instructions, ip *int) error {
	structNameIdx := int(code.ReadUint16(ins[*ip+1:]))
	methodNameIdx := int(code.ReadUint16(ins[*ip+3:]))
	*ip += 4

	structNameObj := vm.constants[structNameIdx]
	methodNameObj := vm.constants[methodNameIdx]

	structName := structNameObj.(*object.String).Value
	methodName := methodNameObj.(*object.String).Value

	methodClosure := vm.pop()
	closure, ok := methodClosure.(*object.Closure)
	if !ok {
		return fmt.Errorf("method body must be closure, got %s", methodClosure.Type())
	}

	if vm.methods[structName] == nil {
		vm.methods[structName] = make(map[string]*object.Closure)
	}
	vm.methods[structName][methodName] = closure

	return nil
}
