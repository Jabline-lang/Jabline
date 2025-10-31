package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
)

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.ArrowFunction:
		extendedEnv := extendArrowFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return evaluated
	case *object.AsyncFunction:
		promise := object.NewPromise()
		GlobalEventLoop.SchedulePromiseTask(promise, func() {
			extendedEnv := extendAsyncFunctionEnv(fn, args)
			evaluated := Eval(fn.Body, extendedEnv)
			result := unwrapReturnValue(evaluated)
			if isError(result) {
				promise.Reject(result)
			} else {
				promise.Resolve(result)
			}
		}, 0)
		return promise
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %T", fn)
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	var env *object.Environment
	if fn.IsClosureCreated {
		env = object.NewClosureEnvironment(fn.Env)
		for name, capturedVar := range fn.CapturedVars {
			env.CaptureVariable(name, capturedVar)
		}
	} else {
		env = object.NewEnclosedEnvironment(fn.Env)
	}

	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		} else {
			env.Set(param.Value, NULL)
		}
	}

	return env
}

func extendAsyncFunctionEnv(fn *object.AsyncFunction, args []object.Object) *object.Environment {
	var env *object.Environment
	if fn.IsClosureCreated {
		env = object.NewClosureEnvironment(fn.Env)
		for name, capturedVar := range fn.CapturedVars {
			env.CaptureVariable(name, capturedVar)
		}
	} else {
		env = object.NewEnclosedEnvironment(fn.Env)
	}

	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		} else {
			env.Set(param.Value, NULL)
		}
	}

	return env
}

func extendArrowFunctionEnv(fn *object.ArrowFunction, args []object.Object) *object.Environment {
	var env *object.Environment
	if fn.IsClosureCreated {
		env = object.NewClosureEnvironment(fn.Env)
		for name, capturedVar := range fn.CapturedVars {
			env.CaptureVariable(name, capturedVar)
		}
	} else {
		env = object.NewEnclosedEnvironment(fn.Env)
	}

	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		} else {
			env.Set(param.Value, NULL)
		}
	}

	return env
}

func analyzeClosureRequirements(node ast.Node, currentEnv *object.Environment) []string {
	requiredVars := make([]string, 0)
	visitedVars := make(map[string]bool)

	var analyzeNode func(ast.Node)
	analyzeNode = func(n ast.Node) {
		switch node := n.(type) {
		case *ast.Identifier:
			varName := node.Value
			if !visitedVars[varName] {
				if env := currentEnv.FindVariableEnvironment(varName); env != nil && env != currentEnv {
					requiredVars = append(requiredVars, varName)
					visitedVars[varName] = true
				}
			}
		case *ast.Program:
			for _, stmt := range node.Statements {
				analyzeNode(stmt)
			}
		case *ast.BlockStatement:
			for _, stmt := range node.Statements {
				analyzeNode(stmt)
			}
		case *ast.ExpressionStatement:
			analyzeNode(node.Expression)
		case *ast.ReturnStatement:
			if node.ReturnValue != nil {
				analyzeNode(node.ReturnValue)
			}
		case *ast.LetStatement:
			if node.Value != nil {
				analyzeNode(node.Value)
			}
		case *ast.AssignmentStatement:
			analyzeNode(node.Left)
			analyzeNode(node.Value)
		case *ast.IfExpression:
			analyzeNode(node.Condition)
			analyzeNode(node.Consequence)
			if node.Alternative != nil {
				analyzeNode(node.Alternative)
			}
		case *ast.InfixExpression:
			analyzeNode(node.Left)
			analyzeNode(node.Right)
		case *ast.PrefixExpression:
			analyzeNode(node.Right)
		case *ast.CallExpression:
			analyzeNode(node.Function)
			for _, arg := range node.Arguments {
				analyzeNode(arg)
			}
		case *ast.FunctionLiteral:
			analyzeNode(node.Body)
		case *ast.ArrowFunction:
			analyzeNode(node.Body)
		case *ast.ArrayLiteral:
			for _, elem := range node.Elements {
				analyzeNode(elem)
			}
		case *ast.HashLiteral:
			for key, value := range node.Pairs {
				analyzeNode(key)
				analyzeNode(value)
			}
		case *ast.IndexExpression:
			analyzeNode(node.Left)
			analyzeNode(node.Index)
		case *ast.WhileStatement:
			analyzeNode(node.Condition)
			analyzeNode(node.Body)
		case *ast.ForStatement:
			if node.Init != nil {
				analyzeNode(node.Init)
			}
			if node.Condition != nil {
				analyzeNode(node.Condition)
			}
			if node.Update != nil {
				analyzeNode(node.Update)
			}
			analyzeNode(node.Body)
		}
	}

	analyzeNode(node)
	return requiredVars
}

func createClosureIfNeeded(fn object.Object, env *object.Environment) object.Object {
	switch function := fn.(type) {
	case *object.Function:
		requiredVars := analyzeClosureRequirements(function.Body, env)
		if len(requiredVars) > 0 {
			return function.CreateClosure(env, requiredVars)
		}
		return function
	case *object.ArrowFunction:
		requiredVars := analyzeClosureRequirements(function.Body, env)
		if len(requiredVars) > 0 {
			return function.CreateClosure(env, requiredVars)
		}
		return function
	case *object.AsyncFunction:
		requiredVars := analyzeClosureRequirements(function.Body, env)
		if len(requiredVars) > 0 {
			return function.CreateClosure(env, requiredVars)
		}
		return function
	default:
		return fn
	}
}

func evaluateNestedFunction(node *ast.FunctionLiteral, env *object.Environment) object.Object {
	function := &object.Function{
		Parameters: node.Parameters,
		Env:        env,
		Body:       node.Body,
	}
	return createClosureIfNeeded(function, env)
}

func evaluateNestedArrowFunction(node *ast.ArrowFunction, env *object.Environment) object.Object {
	arrowFunction := &object.ArrowFunction{
		Parameters: node.Parameters,
		Env:        env,
		Body:       node.Body,
	}
	return createClosureIfNeeded(arrowFunction, env)
}

func evaluateNestedAsyncFunction(node *ast.AsyncFunctionLiteral, env *object.Environment) object.Object {
	asyncFunction := &object.AsyncFunction{
		Parameters: node.Parameters,
		Env:        env,
		Body:       node.Body,
	}
	return createClosureIfNeeded(asyncFunction, env)
}

func applyFunctionToArray(fn object.Object, arr *object.Array) object.Object {
	results := make([]object.Object, 0, len(arr.Elements))
	for _, element := range arr.Elements {
		result := applyFunction(fn, []object.Object{element})
		if isError(result) {
			return result
		}
		results = append(results, result)
	}
	return &object.Array{Elements: results}
}

func filterArrayWithFunction(fn object.Object, arr *object.Array) object.Object {
	results := make([]object.Object, 0)
	for _, element := range arr.Elements {
		result := applyFunction(fn, []object.Object{element})
		if isError(result) {
			return result
		}
		if isTruthy(result) {
			results = append(results, element)
		}
	}
	return &object.Array{Elements: results}
}

func reduceArrayWithFunction(fn object.Object, arr *object.Array, initialValue object.Object) object.Object {
	accumulator := initialValue
	for _, element := range arr.Elements {
		result := applyFunction(fn, []object.Object{accumulator, element})
		if isError(result) {
			return result
		}
		accumulator = result
	}
	return accumulator
}
