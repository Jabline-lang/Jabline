package vm

import (
	"jabline/pkg/code"
	"jabline/pkg/stdlib"
)

func (vm *VM) opSetGlobal(ins code.Instructions, ip *int) {
	globalIndex := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	vm.globals[globalIndex] = vm.pop()
}

func (vm *VM) opGetGlobal(ins code.Instructions, ip *int) error {
	globalIndex := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	return vm.push(vm.globals[globalIndex])
}

func (vm *VM) opSetLocal(ins code.Instructions, ip *int) {
	localIndex := int(ins[*ip+1])
	*ip += 1
	frame := vm.currentFrame()
	vm.stack[frame.basePointer+localIndex] = vm.pop()
}

func (vm *VM) opGetLocal(ins code.Instructions, ip *int) error {
	localIndex := int(ins[*ip+1])
	*ip += 1
	frame := vm.currentFrame()
	return vm.push(vm.stack[frame.basePointer+localIndex])
}

func (vm *VM) opGetBuiltin(ins code.Instructions, ip *int) error {
	builtinIndex := int(ins[*ip+1])
	*ip += 1
	definition := stdlib.Registry[builtinIndex]
	return vm.push(definition.Builtin)
}

func (vm *VM) opClosure(ins code.Instructions, ip *int) error {
	constIndex := int(code.ReadUint16(ins[*ip+1:]))
	numFree := int(ins[*ip+3])
	*ip += 3
	return vm.pushClosure(constIndex, numFree)
}

func (vm *VM) opGetFree(ins code.Instructions, ip *int) error {
	freeIndex := int(ins[*ip+1])
	*ip += 1
	currentClosure := vm.currentFrame().cl
	return vm.push(currentClosure.Free[freeIndex])
}
