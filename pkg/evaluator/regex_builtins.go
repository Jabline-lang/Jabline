package evaluator

import (
	"regexp"
	"strconv"
	"strings"

	"jabline/pkg/object"
)

var RegexBuiltins = map[string]*object.Builtin{
	"regex": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
			}

			pattern, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `regex` must be STRING, got %T", args[0])
			}

			flags := ""
			if len(args) == 2 {
				flagsArg, ok := args[1].(*object.String)
				if !ok {
					return newError("second argument to `regex` must be STRING, got %T", args[1])
				}
				flags = flagsArg.Value
			}

			finalPattern := applyRegexFlags(pattern.Value, flags)

			re, err := regexp.Compile(finalPattern)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			pairs := make(map[object.HashKey]object.HashPair)

			patternKey := (&object.String{Value: "pattern"}).HashKey()
			pairs[patternKey] = object.HashPair{
				Key:   &object.String{Value: "pattern"},
				Value: &object.String{Value: re.String()},
			}

			sourceKey := (&object.String{Value: "source"}).HashKey()
			pairs[sourceKey] = object.HashPair{
				Key:   &object.String{Value: "source"},
				Value: &object.String{Value: pattern.Value},
			}

			flagsKey := (&object.String{Value: "flags"}).HashKey()
			pairs[flagsKey] = object.HashPair{
				Key:   &object.String{Value: "flags"},
				Value: &object.String{Value: flags},
			}

			return &object.Hash{Pairs: pairs}
		},
	},

	"match": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			text, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `match` must be STRING, got %T", args[0])
			}

			pattern, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `match` must be STRING, got %T", args[1])
			}

			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			matches := re.FindStringSubmatch(text.Value)
			if matches == nil {
				return NULL
			}

			elements := make([]object.Object, len(matches))
			for i, match := range matches {
				elements[i] = &object.String{Value: match}
			}

			return &object.Array{Elements: elements}
		},
	},

	"matchAll": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			text, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `matchAll` must be STRING, got %T", args[0])
			}

			pattern, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `matchAll` must be STRING, got %T", args[1])
			}

			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			allMatches := re.FindAllStringSubmatch(text.Value, -1)
			if allMatches == nil {
				return &object.Array{Elements: []object.Object{}}
			}

			elements := make([]object.Object, len(allMatches))
			for i, matches := range allMatches {
				matchElements := make([]object.Object, len(matches))
				for j, match := range matches {
					matchElements[j] = &object.String{Value: match}
				}
				elements[i] = &object.Array{Elements: matchElements}
			}

			return &object.Array{Elements: elements}
		},
	},

	"test": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			text, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `test` must be STRING, got %T", args[0])
			}

			pattern, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `test` must be STRING, got %T", args[1])
			}

			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			matched := re.MatchString(text.Value)
			return nativeBoolToRegexObject(matched)
		},
	},

	"replace": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}

			text, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `replace` must be STRING, got %T", args[0])
			}

			pattern, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `replace` must be STRING, got %T", args[1])
			}

			replacement, ok := args[2].(*object.String)
			if !ok {
				return newError("third argument to `replace` must be STRING, got %T", args[2])
			}

			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			result := re.ReplaceAllString(text.Value, replacement.Value)
			return &object.String{Value: result}
		},
	},

	"replaceAll": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}

			text, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `replaceAll` must be STRING, got %T", args[0])
			}

			pattern, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `replaceAll` must be STRING, got %T", args[1])
			}

			replacement, ok := args[2].(*object.String)
			if !ok {
				return newError("third argument to `replaceAll` must be STRING, got %T", args[2])
			}

			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			result := re.ReplaceAllString(text.Value, replacement.Value)
			return &object.String{Value: result}
		},
	},

	"split": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments. got=%d, want=2 or 3", len(args))
			}

			text, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `split` must be STRING, got %T", args[0])
			}

			pattern, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `split` must be STRING, got %T", args[1])
			}

			limit := -1
			if len(args) == 3 {
				limitArg, ok := args[2].(*object.Integer)
				if !ok {
					return newError("third argument to `split` must be INTEGER, got %T", args[2])
				}
				limit = int(limitArg.Value)
			}

			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			parts := re.Split(text.Value, limit)
			elements := make([]object.Object, len(parts))
			for i, part := range parts {
				elements[i] = &object.String{Value: part}
			}

			return &object.Array{Elements: elements}
		},
	},

	"find": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			text, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `find` must be STRING, got %T", args[0])
			}

			pattern, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `find` must be STRING, got %T", args[1])
			}

			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			loc := re.FindStringIndex(text.Value)
			if loc == nil {
				return &object.Array{Elements: []object.Object{}}
			}

			elements := make([]object.Object, 2)
			elements[0] = &object.Integer{Value: int64(loc[0])}
			elements[1] = &object.Integer{Value: int64(loc[1])}

			return &object.Array{Elements: elements}
		},
	},

	"findAll": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments. got=%d, want=2 or 3", len(args))
			}

			text, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `findAll` must be STRING, got %T", args[0])
			}

			pattern, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `findAll` must be STRING, got %T", args[1])
			}

			limit := -1
			if len(args) == 3 {
				limitArg, ok := args[2].(*object.Integer)
				if !ok {
					return newError("third argument to `findAll` must be INTEGER, got %T", args[2])
				}
				limit = int(limitArg.Value)
			}

			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			matches := re.FindAllString(text.Value, limit)
			if matches == nil {
				return &object.Array{Elements: []object.Object{}}
			}

			elements := make([]object.Object, len(matches))
			for i, match := range matches {
				elements[i] = &object.String{Value: match}
			}

			return &object.Array{Elements: elements}
		},
	},
}

func applyRegexFlags(pattern, flags string) string {
	finalPattern := pattern

	for _, flag := range flags {
		switch flag {
		case 'i':
			if !strings.HasPrefix(finalPattern, "(?i)") {
				finalPattern = "(?i)" + finalPattern
			}
		case 'm':
			if !strings.HasPrefix(finalPattern, "(?m)") {
				finalPattern = "(?m)" + finalPattern
			}
		case 's':
			if !strings.HasPrefix(finalPattern, "(?s)") {
				finalPattern = "(?s)" + finalPattern
			}
		}
	}

	return finalPattern
}

func nativeBoolToRegexObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}

var CommonRegexBuiltins = map[string]*object.Builtin{
	"isEmail": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `isEmail` must be STRING, got %T", args[0])
			}

			emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			re, _ := regexp.Compile(emailRegex)
			matched := re.MatchString(str.Value)

			return nativeBoolToRegexObject(matched)
		},
	},

	"isURL": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `isURL` must be STRING, got %T", args[0])
			}

			urlRegex := `^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`
			re, _ := regexp.Compile(urlRegex)
			matched := re.MatchString(str.Value)

			return nativeBoolToRegexObject(matched)
		},
	},

	"isPhone": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `isPhone` must be STRING, got %T", args[0])
			}

			phoneRegex := `^\+?[1-9]\d{1,14}$|^(\(\d{3}\)|\d{3})[-.\s]?\d{3}[-.\s]?\d{4}$`
			re, _ := regexp.Compile(phoneRegex)
			matched := re.MatchString(str.Value)

			return nativeBoolToRegexObject(matched)
		},
	},

	"extractNumbers": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `extractNumbers` must be STRING, got %T", args[0])
			}

			numberRegex := `\d+(\.\d+)?`
			re, _ := regexp.Compile(numberRegex)
			matches := re.FindAllString(str.Value, -1)

			elements := make([]object.Object, len(matches))
			for i, match := range matches {
				if intVal, err := strconv.ParseInt(match, 10, 64); err == nil {
					elements[i] = &object.Integer{Value: intVal}
				} else if floatVal, err := strconv.ParseFloat(match, 64); err == nil {
					elements[i] = &object.Float{Value: floatVal}
				} else {
					elements[i] = &object.String{Value: match}
				}
			}

			return &object.Array{Elements: elements}
		},
	},

	"extractWords": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `extractWords` must be STRING, got %T", args[0])
			}

			wordRegex := `\b[a-zA-Z]+\b`
			re, _ := regexp.Compile(wordRegex)
			matches := re.FindAllString(str.Value, -1)

			elements := make([]object.Object, len(matches))
			for i, match := range matches {
				elements[i] = &object.String{Value: match}
			}

			return &object.Array{Elements: elements}
		},
	},
}

func InitRegexBuiltins(builtins map[string]*object.Builtin) {
	for name, builtin := range RegexBuiltins {
		builtins[name] = builtin
	}
	for name, builtin := range CommonRegexBuiltins {
		builtins[name] = builtin
	}
}
