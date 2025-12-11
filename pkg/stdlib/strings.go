package stdlib

import (
	"jabline/pkg/object"
	"strings"
)

var StringBuiltins = []struct {
	Name    string
	Builtin *object.Builtin
}{
	{"upper", &object.Builtin{Fn: strUpper}},
	{"lower", &object.Builtin{Fn: strLower}},
	{"contains", &object.Builtin{Fn: strContains}},
}

func strUpper(args ...object.Object) object.Object {
	if len(args) != 1 { return newError("wrong args") }
	s, ok := args[0].(*object.String)
	if !ok { return newError("arg must be string") }
	return &object.String{Value: strings.ToUpper(s.Value)}
}

func strLower(args ...object.Object) object.Object {
	if len(args) != 1 { return newError("wrong args") }
	s, ok := args[0].(*object.String)
	if !ok { return newError("arg must be string") }
	return &object.String{Value: strings.ToLower(s.Value)}
}

func strContains(args ...object.Object) object.Object {
	if len(args) != 2 { return newError("wrong args") }
	s1, ok1 := args[0].(*object.String)
	s2, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 { return newError("args must be strings") }
	
	if strings.Contains(s1.Value, s2.Value) {
		return &object.Boolean{Value: true}
	}
	return &object.Boolean{Value: false}
}
