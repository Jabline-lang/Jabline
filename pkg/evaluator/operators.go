package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
)

// evalPrefixExpression evaluates prefix expressions (!, -)
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newUnsupportedOperationError(operator, getTypeName(right), "", nil)
	}
}

// evalBangOperatorExpression evaluates the bang operator (!)
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

// evalMinusPrefixOperatorExpression evaluates the minus prefix operator (-)
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER_OBJ:
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	case object.FLOAT_OBJ:
		value := right.(*object.Float).Value
		return &object.Float{Value: -value}
	default:
		return newUnsupportedOperationError("unary minus", getTypeName(right), "", nil)
	}
}

// evalLogicalInfixExpression evaluates logical infix expressions (&& and ||)
func evalLogicalInfixExpression(node *ast.InfixExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	// Short-circuit evaluation for logical operators
	if node.Operator == "&&" {
		if !isTruthy(left) {
			return FALSE
		}
		// Evaluate right side only if left is truthy
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return nativeBoolToPyBoolean(isTruthy(right))
	} else if node.Operator == "||" {
		if isTruthy(left) {
			return TRUE
		}
		// Evaluate right side only if left is falsy
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return nativeBoolToPyBoolean(isTruthy(right))
	}

	return newError("unknown operator: %s", node.Operator)
}

// evalInfixExpression evaluates infix expressions (+, -, *, /, ==, !=, <, >, etc.)
func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case (left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ) ||
		(left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ):
		return evalMixedNumericInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToPyBoolean(left == right)
	case operator == "!=":
		return nativeBoolToPyBoolean(left != right)
	case operator == "+":
		return evalMixedInfixExpression(operator, left, right)
	default:
		return newUnsupportedOperationError(getOperatorName(operator), getTypeName(left), getTypeName(right), nil)
	}
}

// evalIntegerInfixExpression evaluates infix expressions on integers
func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newDivisionByZeroError(nil)
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newArithmeticError("modulo by zero", "cannot perform modulo with zero", nil)
		}
		return &object.Integer{Value: leftVal % rightVal}
	case "<":
		return nativeBoolToPyBoolean(leftVal < rightVal)
	case ">":
		return nativeBoolToPyBoolean(leftVal > rightVal)
	case "<=":
		return nativeBoolToPyBoolean(leftVal <= rightVal)
	case ">=":
		return nativeBoolToPyBoolean(leftVal >= rightVal)
	case "==":
		return nativeBoolToPyBoolean(leftVal == rightVal)
	case "!=":
		return nativeBoolToPyBoolean(leftVal != rightVal)
	default:
		return newUnsupportedOperationError(getOperatorName(operator), "INTEGER", "INTEGER", nil)
	}
}

// evalFloatInfixExpression evaluates infix expressions on floats
func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0.0 {
			return newDivisionByZeroError(nil)
		}
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToPyBoolean(leftVal < rightVal)
	case ">":
		return nativeBoolToPyBoolean(leftVal > rightVal)
	case "<=":
		return nativeBoolToPyBoolean(leftVal <= rightVal)
	case ">=":
		return nativeBoolToPyBoolean(leftVal >= rightVal)
	case "==":
		return nativeBoolToPyBoolean(leftVal == rightVal)
	case "!=":
		return nativeBoolToPyBoolean(leftVal != rightVal)
	default:
		return newUnsupportedOperationError(getOperatorName(operator), "FLOAT", "FLOAT", nil)
	}
}

// evalMixedNumericInfixExpression evaluates infix expressions between integers and floats
func evalMixedNumericInfixExpression(operator string, left, right object.Object) object.Object {
	var leftVal, rightVal float64

	// Convert both operands to float64
	if left.Type() == object.INTEGER_OBJ {
		leftVal = float64(left.(*object.Integer).Value)
	} else {
		leftVal = left.(*object.Float).Value
	}

	if right.Type() == object.INTEGER_OBJ {
		rightVal = float64(right.(*object.Integer).Value)
	} else {
		rightVal = right.(*object.Float).Value
	}

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0.0 {
			return newDivisionByZeroError(nil)
		}
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToPyBoolean(leftVal < rightVal)
	case ">":
		return nativeBoolToPyBoolean(leftVal > rightVal)
	case "<=":
		return nativeBoolToPyBoolean(leftVal <= rightVal)
	case ">=":
		return nativeBoolToPyBoolean(leftVal >= rightVal)
	case "==":
		return nativeBoolToPyBoolean(leftVal == rightVal)
	case "!=":
		return nativeBoolToPyBoolean(leftVal != rightVal)
	default:
		return newUnsupportedOperationError(getOperatorName(operator), getTypeName(left), getTypeName(right), nil)
	}
}

// evalStringInfixExpression evaluates infix expressions on strings
func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToPyBoolean(leftVal == rightVal)
	case "!=":
		return nativeBoolToPyBoolean(leftVal != rightVal)
	default:
		return newUnsupportedOperationError(getOperatorName(operator), getTypeName(left), getTypeName(right), nil)
	}
}

// evalMixedInfixExpression evaluates infix expressions on mixed types (for concatenation)
func evalMixedInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newUnsupportedOperationError(getOperatorName(operator), getTypeName(left), getTypeName(right), nil)
	}

	// Convert both operands to strings for concatenation
	leftStr := objectToString(left)
	rightStr := objectToString(right)

	return &object.String{Value: leftStr + rightStr}
}

// evalPostfixExpression evaluates postfix expressions (++, --)
func evalPostfixExpression(node *ast.PostfixExpression, env *object.Environment) object.Object {
	switch leftNode := node.Left.(type) {
	case *ast.Identifier:
		return evalIdentifierPostfix(leftNode, node.Operator, env)
	case *ast.ArrayIndexExpression:
		return evalArrayIndexPostfix(leftNode, node.Operator, env)
	case *ast.IndexExpression:
		return evalFieldAccessPostfix(leftNode, node.Operator, env)
	default:
		return newError("postfix operator not supported on %T", node.Left)
	}
}

// evalIdentifierPostfix evaluates postfix expressions on identifiers
func evalIdentifierPostfix(identifier *ast.Identifier, operator string, env *object.Environment) object.Object {
	obj, ok := env.Get(identifier.Value)
	if !ok {
		return newError("identifier not found: %s", identifier.Value)
	}

	switch obj.Type() {
	case object.INTEGER_OBJ:
		integer := obj.(*object.Integer)
		originalValue := integer.Value

		switch operator {
		case "++":
			env.Set(identifier.Value, &object.Integer{Value: originalValue + 1})
			return &object.Integer{Value: originalValue} // Return original value (post-increment)
		case "--":
			env.Set(identifier.Value, &object.Integer{Value: originalValue - 1})
			return &object.Integer{Value: originalValue} // Return original value (post-decrement)
		default:
			return newUnsupportedOperationError("postfix "+operator, getTypeName(obj), "", nil)
		}
	case object.FLOAT_OBJ:
		float := obj.(*object.Float)
		originalValue := float.Value

		switch operator {
		case "++":
			env.Set(identifier.Value, &object.Float{Value: originalValue + 1.0})
			return &object.Float{Value: originalValue} // Return original value (post-increment)
		case "--":
			env.Set(identifier.Value, &object.Float{Value: originalValue - 1.0})
			return &object.Float{Value: originalValue} // Return original value (post-decrement)
		default:
			return newUnsupportedOperationError("postfix "+operator, getTypeName(obj), "", nil)
		}
	default:
		return newTypeError("INTEGER or FLOAT", getTypeName(obj), nil)
	}
}
