package stdlib

import (
	"os"
	"jabline/pkg/object"
)

var OSBuiltins = []struct {
	Name    string
	Builtin *object.Builtin
}{
	{"exit", &object.Builtin{Fn: osExit}},
	{"getenv", &object.Builtin{Fn: osGetenv}},
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
	if len(args) != 1 { return newError("wrong args") }
	key, ok := args[0].(*object.String)
	if !ok { return newError("arg must be string") }
	
	val := os.Getenv(key.Value)
	if val == "" {
		return &object.Null{}
	}
	return &object.String{Value: val}
}
