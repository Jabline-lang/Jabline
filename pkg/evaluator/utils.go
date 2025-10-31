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

func nativeBoolToPyBoolean(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

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

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ || obj.Type() == object.EXCEPTION_OBJ
	}
	return false
}

func isBreak(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.BREAK_OBJ
	}
	return false
}

func isContinue(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.CONTINUE_OBJ
	}
	return false
}

func isNullish(obj object.Object) bool {
	return obj == NULL || obj.Type() == object.NULL_OBJ
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

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

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	if exception, ok := obj.(*object.Exception); ok {
		return exception
	}
	return obj
}

func isClosure(obj object.Object) bool {
	switch fn := obj.(type) {
	case *object.Function:
		return fn.IsClosureCreated
	case *object.ArrowFunction:
		return fn.IsClosureCreated
	case *object.AsyncFunction:
		return fn.IsClosureCreated
	default:
		return false
	}
}

func getClosureCapturedVars(obj object.Object) map[string]object.Object {
	switch fn := obj.(type) {
	case *object.Function:
		if fn.IsClosureCreated {
			return fn.CapturedVars
		}
	case *object.ArrowFunction:
		if fn.IsClosureCreated {
			return fn.CapturedVars
		}
	case *object.AsyncFunction:
		if fn.IsClosureCreated {
			return fn.CapturedVars
		}
	}
	return nil
}

func isCallable(obj object.Object) bool {
	switch obj.Type() {
	case object.FUNCTION_OBJ, object.ARROW_FUNCTION_OBJ, object.ASYNC_FUNCTION_OBJ, object.BUILTIN_OBJ:
		return true
	default:
		return false
	}
}

func createClosureEnvironment(outerEnv *object.Environment, capturedVars map[string]object.Object) *object.Environment {
	env := object.NewClosureEnvironment(outerEnv)

	if capturedVars != nil {
		for name, obj := range capturedVars {
			env.CaptureVariable(name, obj)
		}
	}

	return env
}

func extractFreeVariables(node ast.Node, definedVars map[string]bool) []string {
	freeVars := make([]string, 0)
	visited := make(map[string]bool)

	var extract func(ast.Node)
	extract = func(n ast.Node) {
		switch node := n.(type) {
		case *ast.Identifier:
			varName := node.Value
			if !definedVars[varName] && !visited[varName] {
				if !isBuiltinIdentifier(varName) {
					freeVars = append(freeVars, varName)
					visited[varName] = true
				}
			}

		case *ast.Program:
			for _, stmt := range node.Statements {
				extract(stmt)
			}

		case *ast.BlockStatement:
			for _, stmt := range node.Statements {
				extract(stmt)
			}

		case *ast.LetStatement:
			if node.Value != nil {
				extract(node.Value)
			}
			definedVars[node.Name.Value] = true

		case *ast.ConstStatement:
			if node.Value != nil {
				extract(node.Value)
			}
			definedVars[node.Name.Value] = true

		case *ast.FunctionStatement:
			definedVars[node.Name.Value] = true

		case *ast.ExpressionStatement:
			extract(node.Expression)

		case *ast.ReturnStatement:
			if node.ReturnValue != nil {
				extract(node.ReturnValue)
			}

		case *ast.AssignmentStatement:
			extract(node.Left)
			extract(node.Value)

		case *ast.IfExpression:
			extract(node.Condition)
			extract(node.Consequence)
			if node.Alternative != nil {
				extract(node.Alternative)
			}

		case *ast.InfixExpression:
			extract(node.Left)
			extract(node.Right)

		case *ast.PrefixExpression:
			extract(node.Right)

		case *ast.CallExpression:
			extract(node.Function)
			for _, arg := range node.Arguments {
				extract(arg)
			}

		case *ast.ArrayLiteral:
			for _, elem := range node.Elements {
				extract(elem)
			}

		case *ast.HashLiteral:
			for key, value := range node.Pairs {
				extract(key)
				extract(value)
			}

		case *ast.IndexExpression:
			extract(node.Left)
			extract(node.Index)
		}
	}

	extract(node)
	return freeVars
}

func isBuiltinIdentifier(name string) bool {
	builtins := []string{
		"len", "first", "last", "rest", "push", "puts", "echo", "type",
		"str", "int", "float", "bool", "array", "hash", "keys", "values",
		"indexOf", "join", "split", "trim", "upper", "lower", "replace",
		"substr", "contains", "startsWith", "endsWith", "reverse",
		"sort", "filter", "map", "reduce", "forEach", "find", "some", "every",
		"min", "max", "sum", "avg", "round", "floor", "ceil", "abs",
		"sqrt", "pow", "log", "exp", "sin", "cos", "tan", "random",
		"time", "now", "sleep", "print", "println", "read", "write",
		"null", "true", "false", "undefined",
	}

	for _, builtin := range builtins {
		if name == builtin {
			return true
		}
	}
	return false
}

func cloneObject(obj object.Object) object.Object {
	switch o := obj.(type) {
	case *object.Integer:
		return &object.Integer{Value: o.Value}
	case *object.Float:
		return &object.Float{Value: o.Value}
	case *object.String:
		return &object.String{Value: o.Value}
	case *object.Boolean:
		return &object.Boolean{Value: o.Value}
	case *object.Null:
		return NULL
	case *object.Array:
		elements := make([]object.Object, len(o.Elements))
		copy(elements, o.Elements)
		return &object.Array{Elements: elements}
	case *object.Hash:
		pairs := make(map[object.HashKey]object.HashPair)
		for k, v := range o.Pairs {
			pairs[k] = v
		}
		return &object.Hash{Pairs: pairs}
	default:
		return obj
	}
}
