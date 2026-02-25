package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	if (left.Type() == object.STRING_OBJ || right.Type() == object.STRING_OBJ) && op == code.OpAdd {
		return vm.executeBinaryStringOperation(op, left, right)
	}

	if left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ ||
		(left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ) ||
		(left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ) {
		return vm.executeBinaryFloatOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", left.Type(), right.Type())
}

func (vm *VM) extractFloat64(obj object.Object) (float64, bool) {
	switch o := obj.(type) {
	case *object.Float:
		return o.Value, true
	case *object.Integer:
		return float64(o.Value), true
	default:
		return 0, false
	}
}

func (vm *VM) executeBinaryFloatOperation(op code.Opcode, left, right object.Object) error {
	leftVal, ok1 := vm.extractFloat64(left)
	rightVal, ok2 := vm.extractFloat64(right)

	if !ok1 || !ok2 {
		return fmt.Errorf("unsupported types for float binary operation: %s %s", left.Type(), right.Type())
	}

	var result float64

	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		if rightVal == 0 {
			return fmt.Errorf("division by zero")
		}
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unknown float operator: %d", op)
	}

	return vm.push(&object.Float{Value: result})
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		if rightValue == 0 {
			return fmt.Errorf("division by zero")
		}
		result = leftValue / rightValue
	case code.OpMod:
		if rightValue == 0 {
			return fmt.Errorf("division by zero")
		}
		result = leftValue % rightValue
	case code.OpBitAnd:
		result = leftValue & rightValue
	case code.OpBitOr:
		result = leftValue | rightValue
	case code.OpBitXor:
		result = leftValue ^ rightValue
	case code.OpShiftLeft:
		result = leftValue << rightValue
	case code.OpShiftRight:
		result = leftValue >> rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	var leftVal, rightVal string

	if left.Type() == object.STRING_OBJ {
		leftVal = left.(*object.String).Value
	} else {
		leftVal = left.Inspect()
	}

	if right.Type() == object.STRING_OBJ {
		rightVal = right.(*object.String).Value
	} else {
		rightVal = right.Inspect()
	}

	return vm.push(&object.String{Value: leftVal + rightVal})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	if left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ {
		return vm.executeFloatComparison(op, left, right)
	}

	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return vm.executeStringComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeFloatComparison(op code.Opcode, left, right object.Object) error {
	leftVal, ok1 := vm.extractFloat64(left)
	rightVal, ok2 := vm.extractFloat64(right)

	if !ok1 || !ok2 {
		return fmt.Errorf("unsupported types for float comparison: %s %s", left.Type(), right.Type())
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(leftVal != rightVal))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObj(leftVal > rightVal))
	default:
		return fmt.Errorf("unknown float comparison operator: %d", op)
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(leftValue != rightValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObj(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeStringComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(leftValue != rightValue))
	default:
		return fmt.Errorf("unknown string operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	switch op := operand.(type) {
	case *object.Integer:
		return vm.push(&object.Integer{Value: -op.Value})
	case *object.Float:
		return vm.push(&object.Float{Value: -op.Value})
	default:
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
}

func (vm *VM) executeBitNotOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for bitwise not: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	// In Go, ^x is bitwise not (complement).
	return vm.push(&object.Integer{Value: ^value})
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func nativeBoolToBooleanObj(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}
