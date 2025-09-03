package evaluator

import (
	"fmt"

	"jabline/pkg/ast"
	"jabline/pkg/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// nativeBoolToPyBoolean converts a native Go bool to a Boolean object
func nativeBoolToPyBoolean(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// isTruthy determines if an object is truthy in the language
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

// isError checks if an object is an error or exception
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ || obj.Type() == object.EXCEPTION_OBJ
	}
	return false
}

// isBreak checks if an object is a break statement
func isBreak(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.BREAK_OBJ
	}
	return false
}

// isContinue checks if an object is a continue statement
func isContinue(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.CONTINUE_OBJ
	}
	return false
}

// isNullish checks if an object is null (for nullish coalescing operator)
func isNullish(obj object.Object) bool {
	return obj == NULL || obj.Type() == object.NULL_OBJ
}

// newError creates a new error object with formatted message
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// objectToString converts any object to its string representation
func objectToString(obj object.Object) string {
	switch obj := obj.(type) {
	case *object.String:
		return obj.Value
	case *object.Integer:
		return fmt.Sprintf("%d", obj.Value)
	case *object.Float:
		return fmt.Sprintf("%g", obj.Value)
	case *object.Boolean:
		return fmt.Sprintf("%t", obj.Value)
	case *object.Null:
		return "null"
	default:
		return obj.Inspect()
	}
}

// evalExpressions evaluates a slice of expressions and returns their values
func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	result := []object.Object{}

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// unwrapReturnValue unwraps a return value if present, otherwise returns the object as-is
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	// Don't unwrap exceptions - they should propagate as-is
	if exception, ok := obj.(*object.Exception); ok {
		return exception
	}
	return obj
}
