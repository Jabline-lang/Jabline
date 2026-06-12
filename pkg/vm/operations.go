package vm

import (
	"fmt"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (vm *VM) isIntegerType(t object.ObjectType) bool {
	switch t {
	case object.INTEGER_OBJ, object.INT8_OBJ, object.INT16_OBJ, object.INT32_OBJ, object.INT64_OBJ,
		object.UINT8_OBJ, object.UINT16_OBJ, object.UINT32_OBJ, object.UINT64_OBJ:
		return true
	}
	return false
}

func (vm *VM) extractInt64(obj object.Object) (int64, bool) {
	switch o := obj.(type) {
	case *object.Integer:
		return o.Value, true
	case *object.Int8:
		return int64(o.Value), true
	case *object.Int16:
		return int64(o.Value), true
	case *object.Int32:
		return int64(o.Value), true
	case *object.Int64:
		return o.Value, true
	case *object.UInt8:
		return int64(o.Value), true
	case *object.UInt16:
		return int64(o.Value), true
	case *object.UInt32:
		return int64(o.Value), true
	case *object.UInt64:
		return int64(o.Value), true
	default:
		return 0, false
	}
}

func (vm *VM) resultIntegerType(left, right object.Object) object.ObjectType {
	lt := left.Type()
	rt := right.Type()
	if lt == object.INTEGER_OBJ || rt == object.INTEGER_OBJ {
		return object.INTEGER_OBJ
	}
	return lt
}

func (vm *VM) pushInteger(value int64, typ object.ObjectType) error {
	switch typ {
	case object.INTEGER_OBJ:
		return vm.push(object.NewInteger(value))
	case object.INT8_OBJ:
		return vm.push(&object.Int8{Value: int8(value)})
	case object.INT16_OBJ:
		return vm.push(&object.Int16{Value: int16(value)})
	case object.INT32_OBJ:
		return vm.push(&object.Int32{Value: int32(value)})
	case object.INT64_OBJ:
		return vm.push(&object.Int64{Value: value})
	case object.UINT8_OBJ:
		return vm.push(&object.UInt8{Value: uint8(value)})
	case object.UINT16_OBJ:
		return vm.push(&object.UInt16{Value: uint16(value)})
	case object.UINT32_OBJ:
		return vm.push(&object.UInt32{Value: uint32(value)})
	case object.UINT64_OBJ:
		return vm.push(&object.UInt64{Value: uint64(value)})
	default:
		return vm.push(object.NewInteger(value))
	}
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if vm.isIntegerType(left.Type()) && vm.isIntegerType(right.Type()) {
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
		if val, ok := vm.extractInt64(obj); ok {
			return float64(val), true
		}
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
	case code.OpAdd, code.OpFloatAdd:
		result = leftVal + rightVal
	case code.OpSub, code.OpFloatSub:
		result = leftVal - rightVal
	case code.OpMul, code.OpFloatMul:
		result = leftVal * rightVal
	case code.OpDiv, code.OpFloatDiv:
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
	leftValue, _ := vm.extractInt64(left)
	rightValue, _ := vm.extractInt64(right)

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

	return vm.pushInteger(result, vm.resultIntegerType(left, right))
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

	return vm.push(object.NewString(leftVal + rightVal))
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if vm.isIntegerType(left.Type()) && vm.isIntegerType(right.Type()) {
		return vm.executeIntegerComparison(op, left, right)
	}

	if left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ {
		return vm.executeFloatComparison(op, left, right)
	}

	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return vm.executeStringComparison(op, left, right)
	}

	if left.Type() == object.DATETIME_OBJ && right.Type() == object.DATETIME_OBJ {
		return vm.executeDateTimeComparison(op, left, right)
	}

	if left.Type() == object.REGEX_OBJ && right.Type() == object.REGEX_OBJ {
		return vm.executeRegexComparison(op, left, right)
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
	leftValue, _ := vm.extractInt64(left)
	rightValue, _ := vm.extractInt64(right)

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

func (vm *VM) executeDateTimeComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.DateTime).Time
	rightValue := right.(*object.DateTime).Time

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(leftValue.Equal(rightValue)))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(!leftValue.Equal(rightValue)))
	default:
		return fmt.Errorf("unknown datetime operator: %d", op)
	}
}

func (vm *VM) executeRegexComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Regex).Pattern
	rightValue := right.(*object.Regex).Pattern

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(leftValue != rightValue))
	default:
		return fmt.Errorf("unknown regex operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	// Use value-based comparison for booleans
	if b, ok := operand.(*object.Boolean); ok {
		if b.Value {
			return vm.push(False)
		}
		return vm.push(True)
	}

	switch operand {
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if vm.isIntegerType(operand.Type()) {
		val, _ := vm.extractInt64(operand)
		return vm.pushInteger(-val, operand.Type())
	}

	switch op := operand.(type) {
	case *object.Float:
		return vm.push(&object.Float{Value: -op.Value})
	default:
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
}

func (vm *VM) executeBitNotOperator() error {
	operand := vm.pop()

	if !vm.isIntegerType(operand.Type()) {
		return fmt.Errorf("unsupported type for bitwise not: %s", operand.Type())
	}

	value, _ := vm.extractInt64(operand)
	return vm.pushInteger(^value, operand.Type())
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
