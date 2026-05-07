package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) opConstant(ins code.Instructions, ip *int) error {
	constIndex := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2
	return vm.push(vm.constants[constIndex])
}

func (vm *VM) opConstant8(ins code.Instructions, ip *int) error {
	constIndex := int(ins[*ip+1])
	*ip += 1
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

func (vm *VM) opTrue() error  { return vm.push(True) }
func (vm *VM) opFalse() error { return vm.push(False) }
func (vm *VM) opNull() error  { return vm.push(Null) }

func (vm *VM) opFloatBinary(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	return vm.executeBinaryFloatOperation(op, left, right)
}

func (vm *VM) opIntToFloat() error {
	obj := vm.pop()
	if integer, ok := obj.(*object.Integer); ok {
		return vm.push(&object.Float{Value: float64(integer.Value)})
	}
	return fmt.Errorf("OpIntToFloat: expected INTEGER, got %s", obj.Type())
}

func (vm *VM) opIsType(ins code.Instructions, ip *int) error {
	typeIndex := int(code.ReadUint16(ins[*ip+1:]))
	*ip += 2

	expectedTypeObj := vm.constants[typeIndex]
	expectedTypeStr := expectedTypeObj.(*object.String).Value

	val := vm.pop()
	// Use ExecuteIsType logic but returning bool instead of error
	err := vm.executeIsType(val, expectedTypeStr)
	if err == nil {
		return vm.push(True)
	}
	return vm.push(False)
}
