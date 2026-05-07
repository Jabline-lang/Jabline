package stdlib

import (
	"jabline/pkg/object"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

// FFIBuiltins defines the native functions for Foreign Function Interface
var FFIBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"__native_ffi_load", &object.Builtin{Fn: ffiLoad}},
	{"__native_ffi_call", &object.Builtin{Fn: ffiCall}},
	{"__native_ffi_call_array", &object.Builtin{Fn: ffiCallArray}},
}

// ffiCall executes a function from a loaded library
func ffiCall(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("wrong number of arguments. got=%d, want at least 3 (handle, name, retType)", len(args))
	}

	handle, ok1 := args[0].(*object.Integer)
	funcName, ok2 := args[1].(*object.String)
	retType, ok3 := args[2].(*object.String)

	if !ok1 || !ok2 || !ok3 {
		return newError("invalid argument types for `ffi_call`")
	}

	fnPtr, err := getFuncPtr(uintptr(handle.Value), funcName.Value)
	if err != nil {
		return newError("symbol %s not found: %s", funcName.Value, err)
	}

	var callArgs []uintptr
	var keepAlive []interface{}

	for _, arg := range args[3:] {
		ptr, val, errObj := prepareFFIArg(arg)
		if errObj != nil {
			return errObj
		}
		callArgs = append(callArgs, val)
		if ptr != nil {
			keepAlive = append(keepAlive, ptr)
		}
	}

	r1, _, _ := purego.SyscallN(fnPtr, callArgs...)
	runtime.KeepAlive(keepAlive)

	return convertFFIReturn(r1, retType.Value)
}

func ffiCallArray(args ...object.Object) object.Object {
	if len(args) != 4 {
		return newError("wrong number of arguments. got=%d, want=4 (handle, name, retType, argsArray)", len(args))
	}

	handle, ok1 := args[0].(*object.Integer)
	funcName, ok2 := args[1].(*object.String)
	retType, ok3 := args[2].(*object.String)
	argsArr, ok4 := args[3].(*object.Array)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return newError("invalid argument types for `ffi_call_array`")
	}

	fnPtr, err := getFuncPtr(uintptr(handle.Value), funcName.Value)
	if err != nil {
		return newError("symbol %s not found: %s", funcName.Value, err)
	}

	var callArgs []uintptr
	var keepAlive []interface{}

	for _, arg := range argsArr.Elements {
		ptr, val, errObj := prepareFFIArg(arg)
		if errObj != nil {
			return errObj
		}
		callArgs = append(callArgs, val)
		if ptr != nil {
			keepAlive = append(keepAlive, ptr)
		}
	}

	r1, _, _ := purego.SyscallN(fnPtr, callArgs...)
	runtime.KeepAlive(keepAlive)

	return convertFFIReturn(r1, retType.Value)
}

func prepareFFIArg(arg object.Object) (unsafe.Pointer, uintptr, object.Object) {
	switch v := arg.(type) {
	case *object.Integer:
		return nil, uintptr(v.Value), nil
	case *object.Float:
		bits := *(*uintptr)(unsafe.Pointer(&v.Value))
		return nil, bits, nil
	case *object.String:
		ptr := stringToPtr(v.Value)
		return ptr, uintptr(ptr), nil
	case *object.Boolean:
		if v.Value {
			return nil, 1, nil
		}
		return nil, 0, nil
	case *object.Null:
		return nil, 0, nil
	default:
		return nil, 0, newError("unsupported FFI argument type: %s", arg.Type())
	}
}

func convertFFIReturn(r1 uintptr, retType string) object.Object {
	switch retType {
	case "int", "long", "void*":
		return &object.Integer{Value: int64(r1)}
	case "string", "char*":
		if r1 == 0 {
			return &object.Null{}
		}
		return &object.String{Value: cstringToGo(r1)}
	case "void":
		return &object.Null{}
	default:
		return &object.Integer{Value: int64(r1)}
	}
}

func cstringToGo(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	p := unsafe.Pointer(ptr)
	var length int
	for {
		if *(*byte)(unsafe.Add(p, length)) == 0 {
			break
		}
		length++
	}
	return string(unsafe.Slice((*byte)(p), length))
}
