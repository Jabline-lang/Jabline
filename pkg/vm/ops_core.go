package vm

import (
	"jabline/pkg/code"
)

func (vm *VM) opConstant(ins code.Instructions, ip *int) error {
	constIndex := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	return vm.push(vm.constants[constIndex])
}

func (vm *VM) opPop() {
	vm.pop()
}

func (vm *VM) opDup() error {
	obj := vm.StackTop()
	return vm.push(obj)
}

func (vm *VM) opBinary(op code.Opcode) error {
	return vm.executeBinaryOperation(op)
}

func (vm *VM) opComparison(op code.Opcode) error {
	return vm.executeComparison(op)
}

func (vm *VM) opPrefix(op code.Opcode) error {
	switch op {
	case code.OpBang:
		return vm.executeBangOperator()
	case code.OpMinus:
		return vm.executeMinusOperator()
	case code.OpBitNot:
		return vm.executeBitNotOperator()
	}
	return nil
}

func (vm *VM) opTrue() error { return vm.push(True) }
func (vm *VM) opFalse() error { return vm.push(False) }
func (vm *VM) opNull() error { return vm.push(Null) }
