package evaluator

import (
	"time"

	"jabline/pkg/object"
)

func builtinSetTimeout(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}

	delayMs, ok := args[1].(*object.Integer)
	if !ok {
		return newError("second argument must be a number")
	}

	delay := time.Duration(delayMs.Value) * time.Millisecond

	if callback, ok := args[0].(*object.Function); ok {
		promise := GlobalEventLoop.SetTimeout(func() {
			extendedEnv := extendFunctionEnv(callback, []object.Object{})
			Eval(callback.Body, extendedEnv)
		}, delay)
		return promise
	}

	if arrowCallback, ok := args[0].(*object.ArrowFunction); ok {
		promise := GlobalEventLoop.SetTimeout(func() {
			extendedEnv := extendArrowFunctionEnv(arrowCallback, []object.Object{})
			Eval(arrowCallback.Body, extendedEnv)
		}, delay)
		return promise
	}

	return newError("first argument must be a function")
}

func builtinPromiseConstructor(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	promise := object.NewPromise()

	resolveFn := &object.Builtin{
		Fn: func(resolveArgs ...object.Object) object.Object {
			if len(resolveArgs) > 0 {
				promise.Resolve(resolveArgs[0])
			} else {
				promise.Resolve(NULL)
			}
			return NULL
		},
	}

	rejectFn := &object.Builtin{
		Fn: func(rejectArgs ...object.Object) object.Object {
			if len(rejectArgs) > 0 {
				promise.Reject(rejectArgs[0])
			} else {
				promise.Reject(NULL)
			}
			return NULL
		},
	}

	if executor, ok := args[0].(*object.Function); ok {
		GlobalEventLoop.ScheduleTask(func() {
			extendedEnv := extendFunctionEnv(executor, []object.Object{resolveFn, rejectFn})
			Eval(executor.Body, extendedEnv)
		}, 0)
		return promise
	}

	if arrowExecutor, ok := args[0].(*object.ArrowFunction); ok {
		GlobalEventLoop.ScheduleTask(func() {
			extendedEnv := extendArrowFunctionEnv(arrowExecutor, []object.Object{resolveFn, rejectFn})
			Eval(arrowExecutor.Body, extendedEnv)
		}, 0)
		return promise
	}

	return newError("argument must be a function")
}

func builtinPromiseResolve(args ...object.Object) object.Object {
	var value object.Object = NULL
	if len(args) > 0 {
		value = args[0]
	}
	return object.NewResolvedPromise(value)
}

func builtinPromiseReject(args ...object.Object) object.Object {
	var reason object.Object = NULL
	if len(args) > 0 {
		reason = args[0]
	}
	return object.NewRejectedPromise(reason)
}

func builtinSleep(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	delayMs, ok := args[0].(*object.Integer)
	if !ok {
		return newError("argument must be a number")
	}

	delay := time.Duration(delayMs.Value) * time.Millisecond
	promise := object.NewPromise()

	GlobalEventLoop.SchedulePromiseTask(promise, func() {
		promise.Resolve(NULL)
	}, delay)

	return promise
}

func builtinFetch(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	url, ok := args[0].(*object.String)
	if !ok {
		return newError("argument must be a string")
	}

	promise := object.NewPromise()
	delay := time.Duration(200) * time.Millisecond

	GlobalEventLoop.SchedulePromiseTask(promise, func() {
		response := &object.Hash{
			Pairs: map[object.HashKey]object.HashPair{
				(&object.String{Value: "url"}).HashKey(): {
					Key:   &object.String{Value: "url"},
					Value: url,
				},
				(&object.String{Value: "status"}).HashKey(): {
					Key:   &object.String{Value: "status"},
					Value: &object.Integer{Value: 200},
				},
				(&object.String{Value: "data"}).HashKey(): {
					Key:   &object.String{Value: "data"},
					Value: &object.String{Value: "Mock response data"},
				},
			},
		}
		promise.Resolve(response)
	}, delay)

	return promise
}

// InitAsyncBuiltins initializes async built-in functions
func InitAsyncBuiltins() {
	// These would be added to the global builtins map
	// For now, we'll add them manually in the getBuiltin function
}
