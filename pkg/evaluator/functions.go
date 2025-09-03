package evaluator

import (
	"jabline/pkg/object"
)

// applyFunction applies a function with given arguments
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

// extendFunctionEnv creates a new environment for function execution
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

// extendAsyncFunctionEnv creates a new environment for async function execution
func extendAsyncFunctionEnv(fn *object.AsyncFunction, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

// extendArrowFunctionEnv creates a new environment for arrow function execution
func extendArrowFunctionEnv(fn *object.ArrowFunction, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}
