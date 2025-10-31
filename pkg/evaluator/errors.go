package evaluator

import (
	"fmt"
	"strings"

	"jabline/pkg/ast"
	"jabline/pkg/object"
	"jabline/pkg/token"
)

type ErrorContext struct {
	Line   int
	Column int
	Source string
	Token  token.Token
}

type EnhancedError struct {
	Message     string
	Context     *ErrorContext
	Suggestions []string
	ErrorType   string
}

func (ee *EnhancedError) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[%s] %s", ee.ErrorType, ee.Message))

	if ee.Context != nil {
		if ee.Context.Line > 0 {
			sb.WriteString(fmt.Sprintf(" (line %d", ee.Context.Line))
			if ee.Context.Column > 0 {
				sb.WriteString(fmt.Sprintf(", column %d", ee.Context.Column))
			}
			sb.WriteString(")")
		}

		if ee.Context.Source != "" {
			sb.WriteString(fmt.Sprintf("\n  --> %s", ee.Context.Source))
		}
	}

	if len(ee.Suggestions) > 0 {
		sb.WriteString("\n\nSuggestions:")
		for _, suggestion := range ee.Suggestions {
			sb.WriteString(fmt.Sprintf("\n  - %s", suggestion))
		}
	}

	return sb.String()
}

func newEnhancedError(errorType, message string, context *ErrorContext, suggestions ...string) *object.Error {
	ee := &EnhancedError{
		Message:     message,
		Context:     context,
		Suggestions: suggestions,
		ErrorType:   errorType,
	}

	return &object.Error{Message: ee.Error()}
}

func newTypeError(expected, actual string, context *ErrorContext) *object.Error {
	message := fmt.Sprintf("Type error: expected %s, got %s", expected, actual)
	suggestions := []string{
		fmt.Sprintf("Convert the value to %s", expected),
		"Check the type of your variables",
	}

	return newEnhancedError("TYPE_ERROR", message, context, suggestions...)
}

func newNameError(name string, context *ErrorContext) *object.Error {
	message := fmt.Sprintf("Name error: identifier '%s' not found", name)
	suggestions := []string{
		"Check if the variable is declared before use",
		"Verify the spelling of the identifier",
		"Make sure the variable is in the correct scope",
	}

	return newEnhancedError("NAME_ERROR", message, context, suggestions...)
}

func newIndexError(index, length int64, context *ErrorContext) *object.Error {
	message := fmt.Sprintf("Index error: index %d out of bounds for length %d", index, length)
	suggestions := []string{
		fmt.Sprintf("Use an index between 0 and %d", length-1),
		"Check if the array/string is empty before accessing",
		"Use len() to get the size before indexing",
	}

	return newEnhancedError("INDEX_ERROR", message, context, suggestions...)
}

func newKeyError(key string, context *ErrorContext) *object.Error {
	message := fmt.Sprintf("Key error: key '%s' not found in hash", key)
	suggestions := []string{
		"Check if the key exists using contains()",
		"Verify the spelling of the key",
		"Use keys() to see available keys",
	}

	return newEnhancedError("KEY_ERROR", message, context, suggestions...)
}

func newArithmeticError(operation, reason string, context *ErrorContext) *object.Error {
	message := fmt.Sprintf("Arithmetic error: %s - %s", operation, reason)
	suggestions := []string{
		"Check for division by zero",
		"Ensure operands are numeric types",
		"Verify the range of your numbers",
	}

	return newEnhancedError("ARITHMETIC_ERROR", message, context, suggestions...)
}

func newSyntaxError(expected, found string, context *ErrorContext) *object.Error {
	message := fmt.Sprintf("Syntax error: expected %s, found %s", expected, found)
	suggestions := []string{
		"Check your syntax against language documentation",
		"Look for missing brackets, parentheses, or semicolons",
		"Verify proper nesting of blocks",
	}

	return newEnhancedError("SYNTAX_ERROR", message, context, suggestions...)
}

func newRuntimeError(message string, context *ErrorContext) *object.Error {
	suggestions := []string{
		"Check the values and types of your variables",
		"Verify function arguments and return values",
		"Look for logical errors in your code",
	}

	return newEnhancedError("RUNTIME_ERROR", message, context, suggestions...)
}

func newFunctionError(funcName, issue string, context *ErrorContext) *object.Error {
	message := fmt.Sprintf("Function error: %s - %s", funcName, issue)
	suggestions := []string{
		"Check the number of arguments passed",
		"Verify argument types match expected parameters",
		"Ensure the function is defined before calling",
	}

	return newEnhancedError("FUNCTION_ERROR", message, context, suggestions...)
}

func newErrorWithContext(message string, node ast.Node) *object.Error {
	context := &ErrorContext{}

	if node != nil {

		context.Source = node.String()
	}

	return newRuntimeError(message, context)
}

func newDivisionByZeroError(context *ErrorContext) *object.Error {
	return newArithmeticError("division by zero", "cannot divide by zero", context)
}

func newUnsupportedOperationError(operation, leftType, rightType string, context *ErrorContext) *object.Error {
	message := fmt.Sprintf("Unsupported operation: %s between %s and %s", operation, leftType, rightType)
	suggestions := []string{
		"Check if the operation is supported for these types",
		"Convert operands to compatible types",
		"Use appropriate operators for each data type",
	}

	return newEnhancedError("OPERATION_ERROR", message, context, suggestions...)
}

func newArgumentError(funcName string, expected, got int, context *ErrorContext) *object.Error {
	message := fmt.Sprintf("Argument error: %s expects %d arguments, got %d", funcName, expected, got)
	suggestions := []string{
		fmt.Sprintf("Pass exactly %d arguments to %s", expected, funcName),
		"Check the function documentation for required parameters",
		"Remove extra arguments or add missing ones",
	}

	return newEnhancedError("ARGUMENT_ERROR", message, context, suggestions...)
}

func getTypeName(obj object.Object) string {
	if obj == nil {
		return "null"
	}
	switch obj.Type() {
	case object.INTEGER_OBJ:
		return "INTEGER"
	case object.FLOAT_OBJ:
		return "FLOAT"
	case object.STRING_OBJ:
		return "STRING"
	case object.BOOLEAN_OBJ:
		return "BOOLEAN"
	case object.ARRAY_OBJ:
		return "ARRAY"
	case object.HASH_OBJ:
		return "HASH"
	case object.NULL_OBJ:
		return "NULL"
	default:
		return string(obj.Type())
	}
}

func contextFromToken(tok token.Token) *ErrorContext {
	return &ErrorContext{
		Token:  tok,
		Source: tok.Literal,
	}
}

func getOperatorName(op string) string {
	switch op {
	case "+":
		return "addition"
	case "-":
		return "subtraction"
	case "*":
		return "multiplication"
	case "/":
		return "division"
	case "%":
		return "modulo"
	case "==":
		return "equality comparison"
	case "!=":
		return "inequality comparison"
	case "<":
		return "less than comparison"
	case ">":
		return "greater than comparison"
	case "<=":
		return "less than or equal comparison"
	case ">=":
		return "greater than or equal comparison"
	case "&&":
		return "logical AND"
	case "||":
		return "logical OR"
	case "!":
		return "logical NOT"
	default:
		return fmt.Sprintf("'%s' operation", op)
	}
}
