package vm

import (
	"jabline/pkg/code"
	"jabline/pkg/object"
	"jabline/pkg/stdlib"
)

func (vm *VM) opSetGlobal(ins code.Instructions, ip *int) {
	globalIndex := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	vm.globals.Set(globalIndex, vm.pop())
}

func (vm *VM) opGetGlobal(ins code.Instructions, ip *int) error {
	globalIndex := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	return vm.push(vm.globals.Get(globalIndex))
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
	return vm.push(definition.Object)
}

func (vm *VM) opCurrentClosure() error {
	currentClosure := vm.currentFrame().cl
	return vm.push(currentClosure)
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

func (vm *VM) opSetFree(ins code.Instructions, ip *int) {
	freeIndex := int(ins[*ip+1])
	*ip += 1
	currentClosure := vm.currentFrame().cl
	currentClosure.Free[freeIndex] = vm.pop()
}

func (vm *VM) opIncLocal(ins code.Instructions, ip *int) error {
	localIndex := int(ins[*ip+1])
	*ip += 1
	frame := vm.currentFrame()
	obj := vm.stack[frame.basePointer+localIndex]
	if i, ok := obj.(*object.Integer); ok {
		vm.stack[frame.basePointer+localIndex] = object.NewInteger(i.Value + 1)
		return nil
	} else if f, ok := obj.(*object.Float); ok {
		vm.stack[frame.basePointer+localIndex] = &object.Float{Value: f.Value + 1.0}
		return nil
	}
	return vm.newRuntimeError("attempted to fast-increment non-numeric local variable")
}

func (vm *VM) opDecLocal(ins code.Instructions, ip *int) error {
	localIndex := int(ins[*ip+1])
	*ip += 1
	frame := vm.currentFrame()
	obj := vm.stack[frame.basePointer+localIndex]
	if i, ok := obj.(*object.Integer); ok {
		vm.stack[frame.basePointer+localIndex] = object.NewInteger(i.Value - 1)
		return nil
	} else if f, ok := obj.(*object.Float); ok {
		vm.stack[frame.basePointer+localIndex] = &object.Float{Value: f.Value - 1.0}
		return nil
	}
	return vm.newRuntimeError("attempted to fast-decrement non-numeric local variable")
}

func (vm *VM) opIncGlobal(ins code.Instructions, ip *int) error {
	globalIndex := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	obj := vm.globals.Get(globalIndex)
	if i, ok := obj.(*object.Integer); ok {
		vm.globals.Set(globalIndex, object.NewInteger(i.Value+1))
		return nil
	} else if f, ok := obj.(*object.Float); ok {
		vm.globals.Set(globalIndex, &object.Float{Value: f.Value + 1.0})
		return nil
	}
	return vm.newRuntimeError("attempted to fast-increment non-numeric global variable")
}

func (vm *VM) opDecGlobal(ins code.Instructions, ip *int) error {
	globalIndex := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	obj := vm.globals.Get(globalIndex)
	if i, ok := obj.(*object.Integer); ok {
		vm.globals.Set(globalIndex, object.NewInteger(i.Value-1))
		return nil
	} else if f, ok := obj.(*object.Float); ok {
		vm.globals.Set(globalIndex, &object.Float{Value: f.Value - 1.0})
		return nil
	}
	return vm.newRuntimeError("attempted to fast-decrement non-numeric global variable")
}
