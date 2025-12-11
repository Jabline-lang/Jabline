package vm

import (
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) opJump(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	vm.currentFrame().ip = pos - 1
	*ip = pos - 1
}

func (vm *VM) opJumpNotTruthy(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.pop()
	if !isTruthy(condition) {
		vm.currentFrame().ip = pos - 1
		*ip = pos - 1
	} else {
		vm.currentFrame().ip = *ip
	}
}

func (vm *VM) opJumpNotTruthyKeep(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.StackTop()
	if !isTruthy(condition) {
		vm.currentFrame().ip = pos - 1
		*ip = pos - 1
	} else {
		vm.currentFrame().ip = *ip
	}
}

func (vm *VM) opJumpTruthyKeep(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.StackTop()
	if isTruthy(condition) {
		vm.currentFrame().ip = pos - 1
		*ip = pos - 1
	} else {
		vm.currentFrame().ip = *ip
	}
}

func (vm *VM) opJumpNotNull(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.StackTop()
	if condition.Type() != object.NULL_OBJ {
		vm.currentFrame().ip = pos - 1
		*ip = pos - 1
	} else {
		vm.currentFrame().ip = *ip
	}
}

func (vm *VM) opJumpIfEqual(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	right := vm.pop()
	left := vm.pop()
	
	equal := false
	if left.Type() == right.Type() {
		if left.Type() == object.INTEGER_OBJ {
			equal = left.(*object.Integer).Value == right.(*object.Integer).Value
		} else if left.Type() == object.STRING_OBJ {
			equal = left.(*object.String).Value == right.(*object.String).Value
		} else if left.Type() == object.BOOLEAN_OBJ {
			equal = left.(*object.Boolean).Value == right.(*object.Boolean).Value
		} else {
			equal = left == right
		}
	}
	if equal {
		vm.currentFrame().ip = pos - 1
		*ip = pos - 1
	} else {
		vm.currentFrame().ip = *ip
	}
}

func (vm *VM) opJumpIfTrue(ins code.Instructions, ip *int) {
	pos := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	condition := vm.pop()
	if isTruthy(condition) {
		vm.currentFrame().ip = pos - 1
		*ip = pos - 1
	} else {
		vm.currentFrame().ip = *ip
	}
}
