//go:build windows

package stdlib

import (
	"jabline/pkg/object"
	"syscall"
	"unsafe"
)

func ffiLoad(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `ffi_load` must be STRING, got %s", args[0].Type())
	}

	h, err := syscall.LoadLibrary(path.Value)
	if err != nil {
		return newError("failed to load library %s: %s", path.Value, err)
	}

	return &object.Integer{Value: int64(h)}
}

func getFuncPtr(handle uintptr, name string) (uintptr, error) {
	return syscall.GetProcAddress(syscall.Handle(handle), name)
}

func stringToPtr(s string) unsafe.Pointer {
    ptr, _ := syscall.BytePtrFromString(s)
    return unsafe.Pointer(ptr)
}
