package evaluator

import (
	"jabline/pkg/ast"
	"jabline/pkg/object"
)

// applyFunction applies a function with given arguments, now with closure support
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.ArrowFunction:
		extendedEnv := extendArrowFunctionEnv(fn, args)
		// Arrow functions have implicit return
		evaluated := Eval(fn.Body, extendedEnv)
		return evaluated
	case *object.AsyncFunction:
		// Async functions always return a Promise
		promise := object.NewPromise()

		// Execute the async function in the event loop
		GlobalEventLoop.SchedulePromiseTask(promise, func() {
			extendedEnv := extendAsyncFunctionEnv(fn, args)
			evaluated := Eval(fn.Body, extendedEnv)
			result := unwrapReturnValue(evaluated)

			// If the result is an error, reject the promise
			if isError(result) {
				promise.Reject(result)
			} else {
				promise.Resolve(result)
			}
		}, 0) // Execute immediately

		return promise
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %T", fn)
	}
}

// extendFunctionEnv creates a new environment for function execution with closure support
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	// Crear entorno con soporte para closures
	var env *object.Environment
	if fn.IsClosureCreated {
		// Si es un closure, crear entorno especial que tenga acceso a variables capturadas
		env = object.NewClosureEnvironment(fn.Env)

		// Aplicar variables capturadas al entorno
		for name, capturedVar := range fn.CapturedVars {
			env.CaptureVariable(name, capturedVar)
		}
	} else {
		env = object.NewEnclosedEnvironment(fn.Env)
	}

	// Bind parameters
	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		} else {
			env.Set(param.Value, NULL)
		}
	}

	return env
}

// extendAsyncFunctionEnv creates a new environment for async function execution with closure support
func extendAsyncFunctionEnv(fn *object.AsyncFunction, args []object.Object) *object.Environment {
	// Crear entorno con soporte para closures
	var env *object.Environment
	if fn.IsClosureCreated {
		// Si es un closure, crear entorno especial que tenga acceso a variables capturadas
		env = object.NewClosureEnvironment(fn.Env)

		// Aplicar variables capturadas al entorno
		for name, capturedVar := range fn.CapturedVars {
			env.CaptureVariable(name, capturedVar)
		}
	} else {
		env = object.NewEnclosedEnvironment(fn.Env)
	}

	// Bind parameters
	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		} else {
			env.Set(param.Value, NULL)
		}
	}

	return env
}

// extendArrowFunctionEnv creates a new environment for arrow function execution with closure support
func extendArrowFunctionEnv(fn *object.ArrowFunction, args []object.Object) *object.Environment {
	// Crear entorno con soporte para closures
	var env *object.Environment
	if fn.IsClosureCreated {
		// Si es un closure, crear entorno especial que tenga acceso a variables capturadas
		env = object.NewClosureEnvironment(fn.Env)

		// Aplicar variables capturadas al entorno
		for name, capturedVar := range fn.CapturedVars {
			env.CaptureVariable(name, capturedVar)
		}
	} else {
		env = object.NewEnclosedEnvironment(fn.Env)
	}

	// Bind parameters
	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		} else {
			env.Set(param.Value, NULL)
		}
	}

	return env
}

// analyzeClosureRequirements analiza qué variables del entorno externo necesita una función
// para determinar qué variables deben ser capturadas en un closure
func analyzeClosureRequirements(node ast.Node, currentEnv *object.Environment) []string {
	requiredVars := make([]string, 0)
	visitedVars := make(map[string]bool)

	// Función helper para analizar nodos recursivamente
	var analyzeNode func(ast.Node)
	analyzeNode = func(n ast.Node) {
		switch node := n.(type) {
		case *ast.Identifier:
			varName := node.Value
			// Solo considerar si la variable no está ya en la lista
			if !visitedVars[varName] {
				// Verificar si la variable existe en un entorno externo
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
			// Analizar el cuerpo de la función anidada
			analyzeNode(node.Body)

		case *ast.ArrowFunction:
			// Analizar el cuerpo de la arrow function anidada
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

			// Agregar más casos según sea necesario
		}
	}

	analyzeNode(node)
	return requiredVars
}

// createClosureIfNeeded crea un closure si la función necesita variables del entorno externo
func createClosureIfNeeded(fn object.Object, env *object.Environment) object.Object {
	switch function := fn.(type) {
	case *object.Function:
		// Analizar qué variables necesita la función
		requiredVars := analyzeClosureRequirements(function.Body, env)

		if len(requiredVars) > 0 {
			// Crear closure con variables capturadas
			return function.CreateClosure(env, requiredVars)
		}
		return function

	case *object.ArrowFunction:
		// Analizar qué variables necesita la arrow function
		requiredVars := analyzeClosureRequirements(function.Body, env)

		if len(requiredVars) > 0 {
			// Crear closure con variables capturadas
			return function.CreateClosure(env, requiredVars)
		}
		return function

	case *object.AsyncFunction:
		// Analizar qué variables necesita la función async
		requiredVars := analyzeClosureRequirements(function.Body, env)

		if len(requiredVars) > 0 {
			// Crear closure con variables capturadas
			return function.CreateClosure(env, requiredVars)
		}
		return function

	default:
		return fn
	}
}

// evaluateNestedFunction evalúa una función anidada y crea closures automáticamente
func evaluateNestedFunction(node *ast.FunctionLiteral, env *object.Environment) object.Object {
	params := node.Parameters
	body := node.Body

	// Crear función básica
	function := &object.Function{
		Parameters: params,
		Env:        env,
		Body:       body,
	}

	// Verificar si necesita ser un closure
	return createClosureIfNeeded(function, env)
}

// evaluateNestedArrowFunction evalúa una arrow function anidada y crea closures automáticamente
func evaluateNestedArrowFunction(node *ast.ArrowFunction, env *object.Environment) object.Object {
	params := node.Parameters
	body := node.Body

	// Crear arrow function básica
	arrowFunction := &object.ArrowFunction{
		Parameters: params,
		Env:        env,
		Body:       body,
	}

	// Verificar si necesita ser un closure
	return createClosureIfNeeded(arrowFunction, env)
}

// evaluateNestedAsyncFunction evalúa una función async anidada y crea closures automáticamente
func evaluateNestedAsyncFunction(node *ast.AsyncFunctionLiteral, env *object.Environment) object.Object {
	params := node.Parameters
	body := node.Body

	// Crear async function básica
	asyncFunction := &object.AsyncFunction{
		Parameters: params,
		Env:        env,
		Body:       body,
	}

	// Verificar si necesita ser un closure
	return createClosureIfNeeded(asyncFunction, env)
}

// Higher-order function helpers

// applyFunctionToArray aplica una función a cada elemento de un array
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

// filterArrayWithFunction filtra un array usando una función predicado
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

// reduceArrayWithFunction reduce un array a un valor usando una función
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
