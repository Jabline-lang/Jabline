package stdlib

import (
	"fmt"
	"jabline/pkg/object"
	"strconv"
)

var Registry = []struct {
	Name   string
	Object object.Object
}{

	{"len", &object.Builtin{Fn: lenFunc}},
	{"type", &object.Builtin{Fn: typeFunc}},
	{"toString", &object.Builtin{Fn: toStringFunc}},
	{"parseInt", &object.Builtin{Fn: parseIntFunc}},
	{"parseFloat", &object.Builtin{Fn: parseFloatFunc}},

	// Numeric Type Constructors
	{"int8", &object.Builtin{Fn: int8Func}},
	{"int16", &object.Builtin{Fn: int16Func}},
	{"int32", &object.Builtin{Fn: int32Func}},
	{"int64", &object.Builtin{Fn: int64Func}},
	{"uint8", &object.Builtin{Fn: uint8Func}},
	{"uint16", &object.Builtin{Fn: uint16Func}},
	{"uint32", &object.Builtin{Fn: uint32Func}},
	{"uint64", &object.Builtin{Fn: uint64Func}},
	{"float32", &object.Builtin{Fn: float32Func}},
	{"float64", &object.Builtin{Fn: float64Func}},

	{"echo", &object.Builtin{Fn: printlnFunc}},
	{"set", &object.Builtin{Fn: setFunc}}, // Add set

	{"push", &object.Builtin{Fn: pushFunc}},
	{"pop", &object.Builtin{Fn: popFunc}},
	{"rest", &object.Builtin{Fn: restFunc}},
	{"first", &object.Builtin{Fn: firstFunc}},
	{"last", &object.Builtin{Fn: lastFunc}},
	{"keys", &object.Builtin{Fn: keysFunc}},
	{"values", &object.Builtin{Fn: valuesFunc}},
	{"is_error", &object.Builtin{Fn: isErrorFunc}},
	{"Error", &object.Builtin{Fn: errorFunc}}, // Native Constructor
}

func errorFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	switch arg := args[0].(type) {
	case *object.String:
		return &object.Error{Message: arg.Value}
	default:
		return &object.Error{Message: arg.Inspect()}
	}
}

func isErrorFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Boolean{Value: false}
	}
	_, ok := args[0].(*object.Error)
	return &object.Boolean{Value: ok}
}

func init() {
	// ... (rest is same)
}

// ... (other funcs)

func setFunc(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("wrong number of arguments. got=%d, want=3", len(args))
	}

	switch obj := args[0].(type) {
	case *object.Hash:
		key, ok := args[1].(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", args[1].Type())
		}
		// Hashable doesn't give the original object back easily if we need to store it as Key
		// But args[1] IS the object.
		obj.Pairs[key.HashKey()] = object.HashPair{Key: args[1], Value: args[2]}
		return args[2]

	case *object.Array:
		index, ok := args[1].(*object.Integer)
		if !ok {
			return newError("index to set must be INTEGER, got %s", args[1].Type())
		}
		idx := index.Value
		if idx < 0 || idx >= int64(len(obj.Elements)) {
			return &object.Null{} // Or error
		}
		obj.Elements[idx] = args[2]
		return args[2]
	}

	return newError("argument to `set` not supported, got %T", args[0])
}

var GlobalModules map[string]*object.Hash

func init() {
	// Register Concurrency builtins globally (channels, etc.)
	Registry = append(Registry, ConcurrencyBuiltins...)

	GlobalModules = make(map[string]*object.Hash)
	nativeModules := []string{"_strings", "_math", "_json", "_os", "_fs", "_http"}
	for _, modName := range nativeModules {
		if modHash := GetNativeModule(modName); modHash != nil {
			globalName := modName[1:]
			GlobalModules[globalName] = modHash
		}
	}

	// Register Global Modules (like fs, math, os, etc.) in the global Registry
	for name, obj := range GlobalModules {
		Registry = append(Registry, struct {
			Name   string
			Object object.Object
		}{name, obj})
	}
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
	last := arr.Elements[len(arr.Elements)-1]
	arr.Elements = arr.Elements[:len(arr.Elements)-1]
	return last
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
	switch arg := args[0].(type) {
	case *object.Array:
		length := len(arg.Elements)
		if length > 0 {
			newElements := make([]object.Object, length-1)
			copy(newElements, arg.Elements[1:length])
			return &object.Array{Elements: newElements}
		}
		return &object.Null{}
	case *object.String:
		length := len(arg.Value)
		if length > 0 {
			return &object.String{Value: arg.Value[1:]}
		}
		return &object.Null{}
	default:
		return newError("argument to `rest` must be ARRAY or STRING, got %T", args[0])
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// --- Numeric Type Constructors ---

func int8Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.Int8{Value: int8(arg.Value)}
	case *object.Float:
		return &object.Int8{Value: int8(arg.Value)}
	default:
		return newError("argument to int8 not supported, got %s", arg.Type())
	}
}

func int16Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.Int16{Value: int16(arg.Value)}
	case *object.Float:
		return &object.Int16{Value: int16(arg.Value)}
	default:
		return newError("argument to int16 not supported, got %s", arg.Type())
	}
}

func int32Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.Int32{Value: int32(arg.Value)}
	case *object.Float:
		return &object.Int32{Value: int32(arg.Value)}
	default:
		return newError("argument to int32 not supported, got %s", arg.Type())
	}
}

func int64Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.Int64{Value: int64(arg.Value)}
	case *object.Float:
		return &object.Int64{Value: int64(arg.Value)}
	default:
		return newError("argument to int64 not supported, got %s", arg.Type())
	}
}

func uint8Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.UInt8{Value: uint8(arg.Value)}
	case *object.Float:
		return &object.UInt8{Value: uint8(arg.Value)}
	default:
		return newError("argument to uint8 not supported, got %s", arg.Type())
	}
}

func uint16Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.UInt16{Value: uint16(arg.Value)}
	case *object.Float:
		return &object.UInt16{Value: uint16(arg.Value)}
	default:
		return newError("argument to uint16 not supported, got %s", arg.Type())
	}
}

func uint32Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.UInt32{Value: uint32(arg.Value)}
	case *object.Float:
		return &object.UInt32{Value: uint32(arg.Value)}
	default:
		return newError("argument to uint32 not supported, got %s", arg.Type())
	}
}

func uint64Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.UInt64{Value: uint64(arg.Value)}
	case *object.Float:
		return &object.UInt64{Value: uint64(arg.Value)}
	default:
		return newError("argument to uint64 not supported, got %s", arg.Type())
	}
}

func float32Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.Float32{Value: float32(arg.Value)}
	case *object.Float:
		return &object.Float32{Value: float32(arg.Value)}
	default:
		return newError("argument to float32 not supported, got %s", arg.Type())
	}
}

func float64Func(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.Float64{Value: float64(arg.Value)}
	case *object.Float:
		return &object.Float64{Value: float64(arg.Value)}
	default:
		return newError("argument to float64 not supported, got %s", arg.Type())
	}
}

func keysFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	hash, ok := args[0].(*object.Hash)
	if !ok {
		return newError("argument to `keys` must be HASH, got %s", args[0].Type())
	}

	elements := make([]object.Object, 0, len(hash.Pairs))
	for _, pair := range hash.Pairs {
		elements = append(elements, pair.Key)
	}

	return &object.Array{Elements: elements}
}

func valuesFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	hash, ok := args[0].(*object.Hash)
	if !ok {
		return newError("argument to `values` must be HASH, got %s", args[0].Type())
	}

	elements := make([]object.Object, 0, len(hash.Pairs))
	for _, pair := range hash.Pairs {
		elements = append(elements, pair.Value)
	}

	return &object.Array{Elements: elements}
}
