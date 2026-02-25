package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) opImport(ins code.Instructions, ip *int) error {
	// Module name is on the stack
	pathObj := vm.pop()
	pathStr, ok := pathObj.(*object.String)
	if !ok {
		return fmt.Errorf("import path must be a string. got=%T", pathObj)
	}

	module, err := vm.loader.Load(pathStr.Value)
	if err != nil {
		fmt.Println("DEBUG: Import Error:", err) // <--- Debug
		return err
	}

	return vm.push(module)
}
