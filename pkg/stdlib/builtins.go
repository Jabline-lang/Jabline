package stdlib

import (
	"bufio"
	"fmt"
	"jabline/pkg/object"
	"os"
	"strconv"
)

// Global test stats for the current test run
var (
	TestPassed = 0
	TestFailed = 0
	TestTotal  = 0
)

func ResetTestResults() {
	TestPassed = 0
	TestFailed = 0
	TestTotal = 0
}

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
	{"print", &object.Builtin{Fn: printFunc}},
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
	{"panic", &object.Builtin{Fn: panicFunc}},
	{"cancel", &object.Builtin{Fn: cancelProcessFunc}},
	{"input", &object.Builtin{Fn: inputFunc}},
	{"recover", &object.Builtin{Fn: recoverFunc}},
	{"gc", &object.Builtin{Fn: runtimeGC}},
	{"memoryStats", &object.Builtin{Fn: runtimeMemoryStats}},
	{"__register_test_results", &object.Builtin{Fn: registerTestResultsFunc}},
	// Math builtins
	{"abs",     &object.Builtin{Fn: mathAbs}},
	{"sqrt",    &object.Builtin{Fn: mathSqrt}},
	{"pow",     &object.Builtin{Fn: mathPow}},
	{"sin",     &object.Builtin{Fn: mathSin}},
	{"cos",     &object.Builtin{Fn: mathCos}},
	{"tan",     &object.Builtin{Fn: mathTan}},
	{"random",  &object.Builtin{Fn: mathRandom}},
	{"max",     &object.Builtin{Fn: mathMax}},
	{"min",     &object.Builtin{Fn: mathMin}},
	{"floor",   &object.Builtin{Fn: mathFloor}},
	{"ceil",    &object.Builtin{Fn: mathCeil}},
	{"round",   &object.Builtin{Fn: mathRound}},
	{"log",     &object.Builtin{Fn: mathLog}},
	{"log10",   &object.Builtin{Fn: mathLog10}},
	{"exp",     &object.Builtin{Fn: mathExp}},
	{"atan2",   &object.Builtin{Fn: mathAtan2}},
	{"asin",    &object.Builtin{Fn: mathAsin}},
	{"acos",    &object.Builtin{Fn: mathAcos}},
	{"hypot",   &object.Builtin{Fn: mathHypot}},

	// Array builtins
	{"append",  &object.Builtin{Fn: appendFunc}},
	{"copy",    &object.Builtin{Fn: copyFunc}},

	{"clone", &object.Builtin{Fn: cloneFunc}},
	{"delete", &object.Builtin{Fn: deleteFunc}},
	{"range",   &object.Builtin{Fn: rangeFunc}},

	// JSON globals (always available)
	{"parse",           &object.Builtin{Fn: jsonParse}},
	{"stringify",       &object.Builtin{Fn: jsonStringify}},
}

func cancelProcessFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	ch, ok := args[0].(*object.Channel)
	if !ok {
		return newError("argument to `cancel` must be a process, got %s", args[0].Type())
	}
	if ch.Cancel != nil {
		ch.Cancel()
	}
	return &object.Null{}
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

func recoverFunc(args ...object.Object) object.Object {
	return &object.Null{} // Placeholder, VM handles this by name
}

func init() {
	for i := range Registry {
		if b, ok := Registry[i].Object.(*object.Builtin); ok {
			b.Name = Registry[i].Name
		}
	}

	// Register Concurrency builtins globally (channels, etc.)
	Registry = append(Registry, ConcurrencyBuiltins...)
	Registry = append(Registry, CryptoBuiltins...)
	Registry = append(Registry, FFIBuiltins...)
}

func panicFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Panic{Message: "panic called with wrong number of arguments"}
	}
	return &object.Panic{Message: args[0].Inspect()}
}

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

var GlobalModules = make(map[string]*object.Hash)

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

func inputFunc(args ...object.Object) object.Object {
	if len(args) > 1 {
		return newError("wrong number of arguments. got=%d, want=0 or 1", len(args))
	}

	if len(args) == 1 {
		fmt.Print(args[0].Inspect())
	}

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return &object.String{Value: scanner.Text()}
	}

	if err := scanner.Err(); err != nil {
		return newError("error reading from stdin: %s", err)
	}

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

func appendFunc(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("wrong number of arguments. got=%d, want>=2", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("first argument to `append` must be ARRAY, got %T", args[0])
	}
	newElements := make([]object.Object, len(arr.Elements)+len(args)-1)
	copy(newElements, arr.Elements)
	for i := 1; i < len(args); i++ {
		newElements[len(arr.Elements)+i-1] = args[i]
	}
	return &object.Array{Elements: newElements}
}

func copyFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return newError("argument to `copy` must be ARRAY, got %T", args[0])
	}
	newElements := make([]object.Object, len(arr.Elements))
	copy(newElements, arr.Elements)
	return &object.Array{Elements: newElements}
}

func deleteFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	hash, ok := args[0].(*object.Hash)
	if !ok {
		return newError("first argument to `delete` must be HASH, got %s", args[0].Type())
	}
	key, ok := args[1].(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", args[1].Type())
	}
	delete(hash.Pairs, key.HashKey())
	return &object.Null{}
}

func rangeFunc(args ...object.Object) object.Object {
	if len(args) == 1 {
		end, ok := args[0].(*object.Integer)
		if !ok {
			return newError("argument to `range` must be INTEGER, got %s", args[0].Type())
		}
		elements := make([]object.Object, end.Value)
		for i := int64(0); i < end.Value; i++ {
			elements[i] = &object.Integer{Value: i}
		}
		return &object.Array{Elements: elements}
	}
	if len(args) == 2 {
		start, ok1 := args[0].(*object.Integer)
		end, ok2 := args[1].(*object.Integer)
		if !ok1 || !ok2 {
			return newError("arguments to `range` must be INTEGERs, got %s, %s", args[0].Type(), args[1].Type())
		}
		if end.Value <= start.Value {
			return &object.Array{Elements: []object.Object{}}
		}
		size := end.Value - start.Value
		elements := make([]object.Object, size)
		for i := int64(0); i < size; i++ {
			elements[i] = &object.Integer{Value: start.Value + i}
		}
		return &object.Array{Elements: elements}
	}
	return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
}

func cloneFunc(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Array:
		newElements := make([]object.Object, len(arg.Elements))
		copy(newElements, arg.Elements)
		return &object.Array{Elements: newElements}
	case *object.Hash:
		newPairs := make(map[object.HashKey]object.HashPair)
		for k, v := range arg.Pairs {
			newPairs[k] = v
		}
		return &object.Hash{Pairs: newPairs}
	case *object.Instance:
		newFields := make(map[string]object.Object)
		for k, v := range arg.Fields {
			newFields[k] = v
		}
		return &object.Instance{StructName: arg.StructName, Fields: newFields}
	default:
		return args[0]
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
func registerTestResultsFunc(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}

	passed, ok1 := args[0].(*object.Integer)
	failed, ok2 := args[1].(*object.Integer)

	if !ok1 || !ok2 {
		return newError("arguments to `__register_test_results` must be integers")
	}

	TestPassed = int(passed.Value)
	TestFailed = int(failed.Value)
	TestTotal = TestPassed + TestFailed

	return &object.Null{}
}
