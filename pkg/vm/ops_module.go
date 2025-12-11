package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) opImport(ins code.Instructions, ip *int) error {
	constIndex := code.ReadUint16(ins[*ip+1:])
	*ip += 2

	pathObj := vm.constants[constIndex]
	pathStr, ok := pathObj.(*object.String)
	if !ok {
		return fmt.Errorf("import path must be a string. got=%T", pathObj)
	}

	module, err := vm.loader.Load(pathStr.Value)
	if err != nil {
		return err
	}

	return vm.push(module)
}