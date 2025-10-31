package evaluator

import (
	"fmt"

	"jabline/pkg/ast"
	"jabline/pkg/object"
)

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		case *object.Exception:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object = NULL

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_OBJ || rt == object.ERROR_OBJ || rt == object.BREAK_OBJ || rt == object.CONTINUE_OBJ || rt == object.EXCEPTION_OBJ {
				return result
			}
		}
	}

	return result
}

func evalLetStatement(node *ast.LetStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}
	env.Set(node.Name.Value, val)
	return val
}

func evalConstStatement(node *ast.ConstStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}

	result := env.SetConstant(node.Name.Value, val)
	if result == nil {
		return newError("identifier '%s' already declared", node.Name.Value)
	}

	return val
}

func evalEchoStatement(node *ast.EchoStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}
	fmt.Println(val.Inspect())
	return NULL
}

func evalReturnStatement(node *ast.ReturnStatement, env *object.Environment) object.Object {
	val := Eval(node.ReturnValue, env)
	if isError(val) {
		return val
	}
	return &object.ReturnValue{Value: val}
}

func evalExpressionStatement(node *ast.ExpressionStatement, env *object.Environment) object.Object {
	return Eval(node.Expression, env)
}
