package vm

import (
	"sync"

	"jabline/pkg/code"
	"jabline/pkg/object"
)

// reloadMu serializes Reload() with the main execution loop to prevent data races.
var reloadMu sync.Mutex

func (vm *VM) Reload(instructions code.Instructions, constants []object.Object) {
	reloadMu.Lock()
	defer reloadMu.Unlock()

	vm.constants = constants

	if len(vm.frames) > 0 && vm.frames[0] != nil {
		vm.frames[0].cl.Fn.Instructions = instructions
	}

	for _, c := range constants {
		if s, ok := c.(*object.Struct); ok {
			vm.Types[s.Name] = s
		} else if i, ok := c.(*object.Interface); ok {
			vm.Types[i.Name] = i
		}
	}
}
