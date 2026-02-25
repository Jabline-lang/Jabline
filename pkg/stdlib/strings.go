package stdlib

import (
	"jabline/pkg/object"
	"regexp"
	"strings"
)

var StringBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"upper", &object.Builtin{Fn: strUpper}},
	{"lower", &object.Builtin{Fn: strLower}},
	{"contains", &object.Builtin{Fn: strContains}},
	{"trim", &object.Builtin{Fn: strTrim}},
	{"split", &object.Builtin{Fn: strSplit}},
	{"match", &object.Builtin{Fn: strMatch}},
	{"replace", &object.Builtin{Fn: strReplace}},
	{"regex_replace", &object.Builtin{Fn: strRegexReplace}},
	{"startsWith", &object.Builtin{Fn: strStartsWith}},
	{"endsWith", &object.Builtin{Fn: strEndsWith}},
	{"join", &object.Builtin{Fn: strJoin}},
	{"slice", &object.Builtin{Fn: strSlice}},
	{"indexOf", &object.Builtin{Fn: strIndexOf}},
}

func strSlice(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("wrong args. want=2 or 3")
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("first arg must be string")
	}

	startObj, ok1 := args[1].(*object.Integer)
	if !ok1 {
		return newError("start index must be integer")
	}
	start := int(startObj.Value)
	if start < 0 {
		start = 0
	}
	if start > len(s.Value) {
		start = len(s.Value)
	}

	end := len(s.Value)
	if len(args) == 3 {
		endObj, ok2 := args[2].(*object.Integer)
		if !ok2 {
			return newError("end index must be integer")
		}
		end = int(endObj.Value)
		if end < start {
			end = start
		}
		if end > len(s.Value) {
			end = len(s.Value)
		}
	}

	return &object.String{Value: s.Value[start:end]}
}

func strUpper(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("arg must be string")
	}
	return &object.String{Value: strings.ToUpper(s.Value)}
}

func strLower(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("arg must be string")
	}
	return &object.String{Value: strings.ToLower(s.Value)}
}

func strContains(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args")
	}
	s1, ok1 := args[0].(*object.String)
	s2, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("args must be strings")
	}

	if strings.Contains(s1.Value, s2.Value) {
		return &object.Boolean{Value: true}
	}
	return &object.Boolean{Value: false}
}

func strIndexOf(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args. want=2")
	}
	s1, ok1 := args[0].(*object.String)
	s2, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("args must be strings")
	}

	return &object.Integer{Value: int64(strings.Index(s1.Value, s2.Value))}
}

func strTrim(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("arg must be string")
	}
	return &object.String{Value: strings.TrimSpace(s.Value)}
}

func strSplit(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args")
	}
	s, ok := args[0].(*object.String)
	sep, ok2 := args[1].(*object.String)
	if !ok || !ok2 {
		return newError("args must be strings")
	}

	parts := strings.Split(s.Value, sep.Value)
	elements := make([]object.Object, len(parts))
	for i, p := range parts {
		elements[i] = &object.String{Value: p}
	}
	return &object.Array{Elements: elements}
}

func strReplace(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("wrong args")
	}
	text, ok := args[0].(*object.String)
	pattern, ok2 := args[1].(*object.String)
	repl, ok3 := args[2].(*object.String)
	if !ok || !ok2 || !ok3 {
		return newError("args must be strings")
	}

	newText := strings.ReplaceAll(text.Value, pattern.Value, repl.Value)
	return &object.String{Value: newText}
}

func strMatch(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args")
	}
	pattern, ok := args[0].(*object.String)
	text, ok2 := args[1].(*object.String)
	if !ok || !ok2 {
		return newError("args must be strings")
	}

	matched, err := regexp.MatchString(pattern.Value, text.Value)
	if err != nil {
		return newError("regex error: %s", err)
	}
	return &object.Boolean{Value: matched}
}

func strRegexReplace(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("wrong args")
	}
	pattern, ok := args[0].(*object.String)
	text, ok2 := args[1].(*object.String)
	repl, ok3 := args[2].(*object.String)
	if !ok || !ok2 || !ok3 {
		return newError("args must be strings")
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regex error: %s", err)
	}

	newText := re.ReplaceAllString(text.Value, repl.Value)
	return &object.String{Value: newText}
}

func strStartsWith(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args")
	}
	s1, ok1 := args[0].(*object.String)
	s2, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("args must be strings")
	}

	return &object.Boolean{Value: strings.HasPrefix(s1.Value, s2.Value)}
}

func strEndsWith(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args")
	}
	s1, ok1 := args[0].(*object.String)
	s2, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("args must be strings")
	}

	return &object.Boolean{Value: strings.HasSuffix(s1.Value, s2.Value)}
}

func strJoin(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong args")
	}
	arr, ok1 := args[0].(*object.Array)
	sep, ok2 := args[1].(*object.String)
	if !ok1 || !ok2 {
		return newError("first arg must be array, second must be string")
	}

	var elements []string
	for _, el := range arr.Elements {
		if s, ok := el.(*object.String); ok {
			elements = append(elements, s.Value)
		} else {
			elements = append(elements, el.Inspect())
		}
	}
	return &object.String{Value: strings.Join(elements, sep.Value)}
}
