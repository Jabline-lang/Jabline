package evaluator

import (
	"fmt"

	"jabline/pkg/ast"
	"jabline/pkg/object"
)

// evalProgram evaluates a program (list of statements)
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

// evalBlockStatement evaluates a block of statements
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

// evalLetStatement evaluates a let statement (variable declaration)
func evalLetStatement(node *ast.LetStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}
	env.Set(node.Name.Value, val)
	return val
}

// evalConstStatement evaluates a const statement (constant declaration)
func evalConstStatement(node *ast.ConstStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}

	// Try to set as constant
	result := env.SetConstant(node.Name.Value, val)
	if result == nil {
		return newError("identifier '%s' already declared", node.Name.Value)
	}

	return val
}

// evalEchoStatement evaluates an echo statement (print to stdout)
func evalEchoStatement(node *ast.EchoStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}
	fmt.Println(val.Inspect())
	return NULL
}

// evalReturnStatement evaluates a return statement
func evalReturnStatement(node *ast.ReturnStatement, env *object.Environment) object.Object {
	val := Eval(node.ReturnValue, env)
	if isError(val) {
		return val
	}
	return &object.ReturnValue{Value: val}
}

// evalExpressionStatement evaluates an expression statement
func evalExpressionStatement(node *ast.ExpressionStatement, env *object.Environment) object.Object {
	return Eval(node.Expression, env)
}
