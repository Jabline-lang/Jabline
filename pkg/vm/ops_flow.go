package vm

import (
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) opJumpNotTruthy(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.pop()
	if !isTruthy(condition) {
		*ip = pos - 1
	}
}

func (vm *VM) opJump(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	*ip = pos - 1
}

func (vm *VM) opJumpNotNull(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.StackTop() // Peek
	if condition != Null {
		*ip = pos - 1
	}
}

func (vm *VM) opJumpNotTruthyKeep(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.StackTop() // Peek
	if !isTruthy(condition) {
		*ip = pos - 1
	}
}

func (vm *VM) opJumpTruthyKeep(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.StackTop() // Peek
	if isTruthy(condition) {
		*ip = pos - 1
	}
}

func (vm *VM) opJumpIfEqual(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	// Expects stack: [..., target, value]
	// Checks if value == target. Pops value. Target remains on stack for next check?
	// No, switch usually pops both or peeks target.
	// Standard implementation: Peek target, Pop value.
	value := vm.pop()
	target := vm.StackTop()
	
	match := false
	switch t := target.(type) {
	case *object.Integer:
		if v, ok := value.(*object.Integer); ok { match = t.Value == v.Value }
	case *object.String:
		if v, ok := value.(*object.String); ok { match = t.Value == v.Value }
	case *object.Boolean:
		if v, ok := value.(*object.Boolean); ok { match = t.Value == v.Value }
	}
	
	if match {
		*ip = pos - 1
	}
}

func (vm *VM) opJumpIfTrue(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.pop()
	if isTruthy(condition) {
		*ip = pos - 1
	}
}
