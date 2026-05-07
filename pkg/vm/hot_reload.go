package vm

import (
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) Reload(instructions code.Instructions, constants []object.Object) {
	// Update Constants
	vm.constants = constants

	// Update the main frame (frame 0) instructions
	if len(vm.frames) > 0 && vm.frames[0] != nil {
		vm.frames[0].cl.Fn.Instructions = instructions
	}

	// Update the Types registry with new Structs and Interfaces from constants
	// (Keeping old ones if they are not redefined)
	for _, c := range constants {
		if s, ok := c.(*object.Struct); ok {
			vm.Types[s.Name] = s
		} else if i, ok := c.(*object.Interface); ok {
			vm.Types[i.Name] = i
		}
	}

	// Note: Methods are updated via OpRegisterMethod when the new instructions run.
	// But we might want to clear or update the current methods map too.
	// For now, new methods will just overwrite old ones in vm.methods.
}
