package stdlib

import (
	"jabline/pkg/object"
	"os"
	"strings"
)

// EnvBuiltins — Native environment variable functions for _env module
var EnvBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"get", &object.Builtin{Fn: envGet}},
	{"set", &object.Builtin{Fn: envSet}},
	{"all", &object.Builtin{Fn: envAll}},
	{"require", &object.Builtin{Fn: envRequire}},
	{"unset", &object.Builtin{Fn: envUnset}},
}

// envGet(key: string, default?: string) -> string | null
func envGet(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("env.get: expects 1 or 2 arguments, got %d", len(args))
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("env.get: key must be a string")
	}

	val := os.Getenv(key.Value)
	if val == "" {
		if len(args) == 2 {
			return args[1] // return default
		}
		return &object.Null{}
	}
	return &object.String{Value: val}
}

// envSet(key: string, value: string) -> bool
func envSet(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("env.set: expects 2 arguments, got %d", len(args))
	}
	key, ok1 := args[0].(*object.String)
	val, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("env.set: key and value must be strings")
	}
	if err := os.Setenv(key.Value, val.Value); err != nil {
		return newError("env.set: %s", err)
	}
	return &object.Boolean{Value: true}
}

// envAll() -> Hash of all env vars
func envAll(args ...object.Object) object.Object {
	if len(args) != 0 {
		return newError("env.all: expects no arguments")
	}
	pairs := make(map[object.HashKey]object.HashPair)
	for _, e := range os.Environ() {
		idx := strings.Index(e, "=")
		if idx < 0 {
			continue
		}
		k := e[:idx]
		v := e[idx+1:]
		ks := &object.String{Value: k}
		pairs[ks.HashKey()] = object.HashPair{
			Key:   ks,
			Value: &object.String{Value: v},
		}
	}
	return &object.Hash{Pairs: pairs}
}

// envRequire(key: string) -> string (panics if missing)
func envRequire(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("env.require: expects 1 argument, got %d", len(args))
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("env.require: key must be a string")
	}
	val := os.Getenv(key.Value)
	if val == "" {
		return newError("env.require: required environment variable '%s' is not set", key.Value)
	}
	return &object.String{Value: val}
}

// envUnset(key: string) -> bool
func envUnset(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("env.unset: expects 1 argument, got %d", len(args))
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("env.unset: key must be a string")
	}
	if err := os.Unsetenv(key.Value); err != nil {
		return newError("env.unset: %s", err)
	}
	return &object.Boolean{Value: true}
}
