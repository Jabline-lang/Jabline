//go:build wasm || wasip1

package stdlib

import (
	"jabline/pkg/object"
	"unsafe"
)

var FFIBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"__native_ffi_load", &object.Builtin{Fn: ffiLoad}},
	{"__native_ffi_call", &object.Builtin{Fn: ffiCall}},
	{"__native_ffi_call_array", &object.Builtin{Fn: ffiCallArray}},
}

func ffiLoad(args ...object.Object) object.Object {
	return newError("FFI is not supported on WebAssembly")
}

func ffiCall(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("wrong number of arguments. got=%d, want at least 3 (handle, name, retType)", len(args))
	}
	return newError("FFI is not supported on WebAssembly")
}

func ffiCallArray(args ...object.Object) object.Object {
	if len(args) != 4 {
		return newError("wrong number of arguments. got=%d, want=4 (handle, name, retType, argsArray)", len(args))
	}
	return newError("FFI is not supported on WebAssembly")
}

func getFuncPtr(handle uintptr, name string) (uintptr, error) {
	return 0, nil
}

func stringToPtr(s string) unsafe.Pointer {
	return nil
}
