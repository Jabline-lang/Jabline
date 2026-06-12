package vm

import (
	"fmt"
	"jabline/pkg/object"
	"jabline/pkg/stdlib"
)

var arrayBuiltinMethods = map[string]*object.Builtin{
	"map":     {Fn: arrayMapBuiltin, Name: "map"},
	"filter":  {Fn: arrayFilterBuiltin, Name: "filter"},
	"reduce":  {Fn: arrayReduceBuiltin, Name: "reduce"},
	"forEach": {Fn: arrayForEachBuiltin, Name: "forEach"},
	"some":    {Fn: arraySomeBuiltin, Name: "some"},
	"every":   {Fn: arrayEveryBuiltin, Name: "every"},
	"find":    {Fn: arrayFindBuiltin, Name: "find"},
	"includes": {Fn: arrayIncludesBuiltin, Name: "includes"},
}

func (vm *VM) getArrayMethod(left, index object.Object) (object.Object, error) {
	methodName := index.(*object.String).Value
	if method, ok := arrayBuiltinMethods[methodName]; ok {
		return method, nil
	}
	return nil, fmt.Errorf("array has no method '%s'", methodName)
}

func arrayMapBuiltin(args ...object.Object) object.Object {
	if len(args) < 2 {
		return &object.Error{Message: "array.map() requires a callback function"}
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("array.map() receiver must be an array, got %s", args[0].Type())}
	}
	callback := args[1]
	result := make([]object.Object, len(arr.Elements))
	for i, elem := range arr.Elements {
		res := stdlib.Executor(callback, []object.Object{elem, object.NewInteger(int64(i)), arr})
		if errObj, isErr := res.(*object.Error); isErr {
			return errObj
		}
		result[i] = res
	}
	return &object.Array{Elements: result}
}

func arrayFilterBuiltin(args ...object.Object) object.Object {
	if len(args) < 2 {
		return &object.Error{Message: "array.filter() requires a callback function"}
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("array.filter() receiver must be an array, got %s", args[0].Type())}
	}
	callback := args[1]
	var result []object.Object
	for i, elem := range arr.Elements {
		res := stdlib.Executor(callback, []object.Object{elem, object.NewInteger(int64(i)), arr})
		if errObj, isErr := res.(*object.Error); isErr {
			return errObj
		}
		if isTruthy(res) {
			result = append(result, elem)
		}
	}
	if result == nil {
		return &object.Array{Elements: []object.Object{}}
	}
	return &object.Array{Elements: result}
}

func arrayReduceBuiltin(args ...object.Object) object.Object {
	if len(args) < 3 {
		return &object.Error{Message: "array.reduce() requires a callback function and initial value"}
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("array.reduce() receiver must be an array, got %s", args[0].Type())}
	}
	callback := args[1]
	acc := args[2]
	for i, elem := range arr.Elements {
		res := stdlib.Executor(callback, []object.Object{acc, elem, object.NewInteger(int64(i)), arr})
		if errObj, isErr := res.(*object.Error); isErr {
			return errObj
		}
		acc = res
	}
	return acc
}

func arrayForEachBuiltin(args ...object.Object) object.Object {
	if len(args) < 2 {
		return &object.Error{Message: "array.forEach() requires a callback function"}
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("array.forEach() receiver must be an array, got %s", args[0].Type())}
	}
	callback := args[1]
	for i, elem := range arr.Elements {
		res := stdlib.Executor(callback, []object.Object{elem, object.NewInteger(int64(i)), arr})
		if errObj, isErr := res.(*object.Error); isErr {
			return errObj
		}
	}
	return Null
}

func arraySomeBuiltin(args ...object.Object) object.Object {
	if len(args) < 2 {
		return &object.Error{Message: "array.some() requires a callback function"}
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("array.some() receiver must be an array, got %s", args[0].Type())}
	}
	callback := args[1]
	for i, elem := range arr.Elements {
		res := stdlib.Executor(callback, []object.Object{elem, object.NewInteger(int64(i)), arr})
		if errObj, isErr := res.(*object.Error); isErr {
			return errObj
		}
		if isTruthy(res) {
			return True
		}
	}
	return False
}

func arrayEveryBuiltin(args ...object.Object) object.Object {
	if len(args) < 2 {
		return &object.Error{Message: "array.every() requires a callback function"}
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("array.every() receiver must be an array, got %s", args[0].Type())}
	}
	callback := args[1]
	for i, elem := range arr.Elements {
		res := stdlib.Executor(callback, []object.Object{elem, object.NewInteger(int64(i)), arr})
		if errObj, isErr := res.(*object.Error); isErr {
			return errObj
		}
		if !isTruthy(res) {
			return False
		}
	}
	return True
}

func arrayFindBuiltin(args ...object.Object) object.Object {
	if len(args) < 2 {
		return &object.Error{Message: "array.find() requires a callback function"}
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("array.find() receiver must be an array, got %s", args[0].Type())}
	}
	callback := args[1]
	for i, elem := range arr.Elements {
		res := stdlib.Executor(callback, []object.Object{elem, object.NewInteger(int64(i)), arr})
		if errObj, isErr := res.(*object.Error); isErr {
			return errObj
		}
		if isTruthy(res) {
			return elem
		}
	}
	return Null
}

func arrayIncludesBuiltin(args ...object.Object) object.Object {
	if len(args) < 2 {
		return &object.Error{Message: "array.includes() requires a value to search for"}
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("array.includes() receiver must be an array, got %s", args[0].Type())}
	}
	target := args[1]
	for _, elem := range arr.Elements {
		if elem.Inspect() == target.Inspect() {
			return True
		}
	}
	return False
}
