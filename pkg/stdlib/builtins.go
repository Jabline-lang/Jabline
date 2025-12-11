package stdlib

import (
	"fmt"
	"jabline/pkg/object"
	"strconv"
)

var Registry = []struct {
	Name    string
	Builtin *object.Builtin
}{

	{"len", &object.Builtin{Fn: lenFunc}},
	{"type", &object.Builtin{Fn: typeFunc}},
	{"toString", &object.Builtin{Fn: toStringFunc}},
	{"parseInt", &object.Builtin{Fn: parseIntFunc}},
	{"parseFloat", &object.Builtin{Fn: parseFloatFunc}},
	
	{"echo", &object.Builtin{Fn: printlnFunc}},

	{"push", &object.Builtin{Fn: pushFunc}},
	{"pop", &object.Builtin{Fn: popFunc}},
	{"first", &object.Builtin{Fn: firstFunc}},
	{"last", &object.Builtin{Fn: lastFunc}},
	{"rest", &object.Builtin{Fn: restFunc}},
}

func init() {
	Registry = append(Registry, MathBuiltins...)
	Registry = append(Registry, StringBuiltins...)
	Registry = append(Registry, IOBuiltins...)
	Registry = append(Registry, OSBuiltins...)
}

func lenFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	case *object.Hash:
		return &object.Integer{Value: int64(len(arg.Pairs))}
	default:
		return newError("argument to `len` not supported, got %T", args[0])
	}
}

func typeFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	return &object.String{Value: string(args[0].Type())}
}

func toStringFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	return &object.String{Value: args[0].Inspect()}
}

func parseIntFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	str, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `parseInt` must be STRING, got %T", args[0])
	}
	val, err := strconv.ParseInt(str.Value, 10, 64)
	if err != nil {
		return newError("could not parse int: %s", err)
	}
	return &object.Integer{Value: val}
}

func parseFloatFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	str, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `parseFloat` must be STRING, got %T", args[0])
	}
	val, err := strconv.ParseFloat(str.Value, 64)
	if err != nil {
		return newError("could not parse float: %s", err)
	}
	return &object.Float{Value: val}
}

func printFunc(args ...object.Object) object.Object {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg.Inspect())
	}
	return &object.Null{}
}

func printlnFunc(args ...object.Object) object.Object {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg.Inspect())
	}
	fmt.Println()
	return &object.Null{}
}

func pushFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("argument to `push` must be ARRAY, got %T", args[0])
	}
	newElements := make([]object.Object, len(arr.Elements)+1)
	copy(newElements, arr.Elements)
	newElements[len(arr.Elements)] = args[1]
	return &object.Array{Elements: newElements}
}

func popFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("argument to `pop` must be ARRAY, got %T", args[0])
	}
	if len(arr.Elements) == 0 {
		return &object.Null{}
	}
	return arr.Elements[len(arr.Elements)-1]
}

func firstFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("argument to `first` must be ARRAY, got %T", args[0])
	}
	if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}
	return &object.Null{}
}

func lastFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("argument to `last` must be ARRAY, got %T", args[0])
	}
	if len(arr.Elements) > 0 {
		return arr.Elements[len(arr.Elements)-1]
	}
	return &object.Null{}
}

func restFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("argument to `rest` must be ARRAY, got %T", args[0])
	}
	length := len(arr.Elements)
	if length > 0 {
		newElements := make([]object.Object, length-1)
		copy(newElements, arr.Elements[1:length])
		return &object.Array{Elements: newElements}
	}
	return &object.Null{}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
