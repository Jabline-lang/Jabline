package vm

import (
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) opService(ins code.Instructions, ip *int) error {
	nameIdx := int(code.ReadUint16(ins[*ip+1:]))
	numFields := int(code.ReadUint16(ins[*ip+3:]))
	*ip += 4

	name := vm.constants[nameIdx].(*object.String).Value
	config := make(map[string]object.Object)

	for i := 0; i < numFields; i++ {
		val := vm.pop()
		key := vm.pop()
		keyStr := key.(*object.String).Value
		config[keyStr] = val
	}

	service := &object.Service{Name: name, Config: config}
	return vm.push(service)
}
