//go:build !windows && !wasm && !wasip1

package stdlib

import (
	"jabline/pkg/object"
	"unsafe"

	"github.com/ebitengine/purego"
)

func ffiLoad(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `ffi_load` must be STRING, got %s", args[0].Type())
	}

	h, err := purego.Dlopen(path.Value, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return newError("failed to load library %s: %s", path.Value, err)
	}

	return &object.Integer{Value: int64(h)}
}

func getFuncPtr(handle uintptr, name string) (uintptr, error) {
	return purego.Dlsym(handle, name)
}

func stringToPtr(s string) unsafe.Pointer {
    return unsafe.Pointer(purego.ByteSliceFromString(s))
}
