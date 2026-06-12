package stdlib

import (
	"regexp"

	"jabline/pkg/object"
)

func init() {
	NativeModuleRegistry["_regex"] = RegexBuiltins
	NativeModulePrefixes["_regex"] = "regex_"
}

var RegexBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"regex_match", &object.Builtin{Fn: regexMatch}},
	{"regex_find", &object.Builtin{Fn: regexFind}},
	{"regex_find_all", &object.Builtin{Fn: regexFindAll}},
	{"regex_replace", &object.Builtin{Fn: regexReplace}},
	{"regex_replace_all", &object.Builtin{Fn: regexReplaceAll}},
	{"regex_split", &object.Builtin{Fn: regexSplit}},
	{"regex_compile", &object.Builtin{Fn: regexCompile}},
	{"regex_test", &object.Builtin{Fn: regexTest}},
}

func regexCompile(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("regex_compile expects 1 argument, got %d", len(args))
	}
	pattern, ok := args[0].(*object.String)
	if !ok {
		return newError("regex_compile expects string, got %s", args[0].Type())
	}

	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return newError("regex_compile error: %s", err.Error())
	}

	return &object.Regex{Value: re, Pattern: pattern.Value}
}

func regexMatch(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("regex_match expects 2 arguments, got %d", len(args))
	}

	var re *regexp.Regexp
	var input string

	switch a := args[0].(type) {
	case *object.Regex:
		re = a.Value
	case *object.String:
		var err error
		re, err = regexp.Compile(a.Value)
		if err != nil {
			return newError("regex_match error: %s", err.Error())
		}
	default:
		return newError("regex_match expects regex or string pattern, got %s", args[0].Type())
	}

	inputStr, ok := args[1].(*object.String)
	if !ok {
		return newError("regex_match expects string input, got %s", args[1].Type())
	}
	input = inputStr.Value

	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return &object.Array{Elements: []object.Object{}}
	}

	elements := make([]object.Object, len(matches))
	for i, m := range matches {
		elements[i] = &object.String{Value: m}
	}
	return &object.Array{Elements: elements}
}

func regexFind(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("regex_find expects 2 arguments, got %d", len(args))
	}

	re, input := getRegexAndInput(args)
	if re == nil {
		return newError("regex_find expects pattern and string")
	}

	match := re.FindString(input)
	if match == "" {
		return object.NullObj
	}
	return &object.String{Value: match}
}

func regexFindAll(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("regex_find_all expects 2 or 3 arguments, got %d", len(args))
	}

	re, input := getRegexAndInput(args)
	if re == nil {
		return newError("regex_find_all expects pattern and string")
	}

	n := -1
	if len(args) == 3 {
		if nObj, ok := args[2].(*object.Integer); ok {
			n = int(nObj.Value)
		}
	}

	matches := re.FindAllString(input, n)
	elements := make([]object.Object, len(matches))
	for i, m := range matches {
		elements[i] = &object.String{Value: m}
	}
	return &object.Array{Elements: elements}
}

func regexReplace(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("regex_replace expects 3 arguments, got %d", len(args))
	}

	re, input := getRegexAndInput(args)
	if re == nil {
		return newError("regex_replace expects pattern and string")
	}

	repl, ok := args[2].(*object.String)
	if !ok {
		return newError("regex_replace expects replacement string, got %s", args[2].Type())
	}

	result := re.ReplaceAllString(input, repl.Value)
	return &object.String{Value: result}
}

func regexReplaceAll(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("regex_replace_all expects 3 arguments, got %d", len(args))
	}

	re, input := getRegexAndInput(args)
	if re == nil {
		return newError("regex_replace_all expects pattern and string")
	}

	repl, ok := args[2].(*object.String)
	if !ok {
		return newError("regex_replace_all expects replacement string, got %s", args[2].Type())
	}

	result := re.ReplaceAllLiteralString(input, repl.Value)
	return &object.String{Value: result}
}

func regexSplit(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("regex_split expects 2 or 3 arguments, got %d", len(args))
	}

	re, input := getRegexAndInput(args)
	if re == nil {
		return newError("regex_split expects pattern and string")
	}

	n := -1
	if len(args) == 3 {
		if nObj, ok := args[2].(*object.Integer); ok {
			n = int(nObj.Value)
		}
	}

	parts := re.Split(input, n)
	elements := make([]object.Object, len(parts))
	for i, p := range parts {
		elements[i] = &object.String{Value: p}
	}
	return &object.Array{Elements: elements}
}

func regexTest(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("regex_test expects 2 arguments, got %d", len(args))
	}

	re, input := getRegexAndInput(args)
	if re == nil {
		return newError("regex_test expects pattern and string")
	}

	return &object.Boolean{Value: re.MatchString(input)}
}

func getRegexAndInput(args []object.Object) (*regexp.Regexp, string) {
	var re *regexp.Regexp

	switch a := args[0].(type) {
	case *object.Regex:
		re = a.Value
	case *object.String:
		var err error
		re, err = regexp.Compile(a.Value)
		if err != nil {
			return nil, ""
		}
	default:
		return nil, ""
	}

	inputStr, ok := args[1].(*object.String)
	if !ok {
		return nil, ""
	}

	return re, inputStr.Value
}

func regexEscape(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("regex_escape expects 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("regex_escape expects string, got %s", args[0].Type())
	}
	return &object.String{Value: regexp.QuoteMeta(s.Value)}
}

func init() {
	// Register additional helper
	RegexBuiltins = append(RegexBuiltins, struct {
		Name   string
		Object object.Object
	}{"regex_escape", &object.Builtin{Fn: regexEscape}})
}
