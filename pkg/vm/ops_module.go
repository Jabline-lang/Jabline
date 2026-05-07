package vm

import (
	"fmt"
	"path/filepath"
	"strings"

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

	moduleName := pathStr.Value

	// Internal module encapsulation: programmers should not use _modules.
	// Only the Jabline Standard Library can consume and wrap them.
	if strings.HasPrefix(moduleName, "_") {
		isInternal := strings.HasPrefix(vm.filename, "std/") ||
			strings.Contains(filepath.ToSlash(vm.filename), "internal/embedded")

		if !isInternal {
			return fmt.Errorf("import error: The module '%s' is internal and private to Jabline. Please use the modules in 'std/*'", moduleName)
		}
	}

	module, err := vm.loader.Load(moduleName)
	if err != nil {
		return err
	}

	return vm.push(module)
}
