package stdlib

import (
	"jabline/pkg/object"
	"os"
)

var OSBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"exit", &object.Builtin{Fn: osExit}},
	{"getenv", &object.Builtin{Fn: osGetenv}},
	{"setenv", &object.Builtin{Fn: osSetenv}},
	{"getwd", &object.Builtin{Fn: osGetwd}},
	{"mkdir", &object.Builtin{Fn: osMkdir}},
	{"remove", &object.Builtin{Fn: osRemove}},
	{"rename", &object.Builtin{Fn: osRename}},
	{"stat", &object.Builtin{Fn: osStat}},
	{"chmod", &object.Builtin{Fn: osChmod}},
	{"tempDir", &object.Builtin{Fn: osTempDir}},
}

func osExit(args ...object.Object) object.Object {
	code := 0
	if len(args) > 0 {
		if intObj, ok := args[0].(*object.Integer); ok {
			code = int(intObj.Value)
		}
	}
	os.Exit(code)
	return &object.Null{}
}

func osGetenv(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("arg must be string")
	}

	val := os.Getenv(key.Value)
	if val == "" {
		return &object.Null{}
	}
	return &object.String{Value: val}
}

func osSetenv(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args")
	}
	key, ok1 := args[0].(*object.String)
	value, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("args must be strings")
	}

	err := os.Setenv(key.Value, value.Value)
	if err != nil {
		return newError("failed to setenv: %s", err)
	}
	return &object.Boolean{Value: true}
}

func osGetwd(args ...object.Object) object.Object {
	if len(args) != 0 {
		return newError("wrong args. want=0")
	}

	dir, err := os.Getwd()
	if err != nil {
		return newError("failed to getwd: %s", err)
	}
	return &object.String{Value: dir}
}

func osMkdir(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `mkdir` must be STRING, got %s", args[0].Type())
	}

	perm := os.FileMode(0755)
	if len(args) == 2 {
		p, ok := args[1].(*object.Integer)
		if !ok {
			return newError("permission argument to `mkdir` must be INTEGER, got %s", args[1].Type())
		}
		perm = os.FileMode(p.Value)
	}

	err := os.MkdirAll(path.Value, perm)
	if err != nil {
		return newError("failed to mkdir: %s", err)
	}
	return &object.Boolean{Value: true}
}

func osRemove(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `remove` must be STRING, got %s", args[0].Type())
	}

	err := os.RemoveAll(path.Value)
	if err != nil {
		return newError("failed to remove: %s", err)
	}
	return &object.Boolean{Value: true}
}

func osRename(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	oldPath, ok1 := args[0].(*object.String)
	newPath, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("arguments to `rename` must be STRING")
	}

	err := os.Rename(oldPath.Value, newPath.Value)
	if err != nil {
		return newError("failed to rename: %s", err)
	}
	return &object.Boolean{Value: true}
}

func osStat(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return newError("argument to `stat` must be STRING, got %s", args[0].Type())
	}

	info, err := os.Stat(path.Value)
	if err != nil {
		return &object.Null{} // Or error? Let's return Null for non-existent but that's handled by err.
		// For consistency with typical script languages, null if not found.
	}

	pairs := make(map[object.HashKey]object.HashPair)
	add := func(k string, v object.Object) {
		ks := &object.String{Value: k}
		pairs[ks.HashKey()] = object.HashPair{Key: ks, Value: v}
	}

	add("name", &object.String{Value: info.Name()})
	add("size", &object.Integer{Value: info.Size()})
	add("is_dir", &object.Boolean{Value: info.IsDir()})
	add("mode", &object.Integer{Value: int64(info.Mode())})
	add("mod_time", &object.Integer{Value: info.ModTime().Unix()})

	return &object.Hash{Pairs: pairs}
}

func osChmod(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	path, ok1 := args[0].(*object.String)
	mode, ok2 := args[1].(*object.Integer)
	if !ok1 || !ok2 {
		return newError("arguments to `chmod` must be STRING and INTEGER")
	}

	err := os.Chmod(path.Value, os.FileMode(mode.Value))
	if err != nil {
		return newError("failed to chmod: %s", err)
	}
	return &object.Boolean{Value: true}
}

func osTempDir(args ...object.Object) object.Object {
	return &object.String{Value: os.TempDir()}
}
