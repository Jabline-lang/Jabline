package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"jabline/pkg/object"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Hash:
				return &object.Integer{Value: int64(len(arg.Pairs))}
			default:
				return newError("argument to `len` not supported, got %T", args[0])
			}
		},
	},

	"type": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var typeName string
			switch args[0].(type) {
			case *object.Integer:
				typeName = "INTEGER"
			case *object.Float:
				typeName = "FLOAT"
			case *object.String:
				typeName = "STRING"
			case *object.Boolean:
				typeName = "BOOLEAN"
			case *object.Array:
				typeName = "ARRAY"
			case *object.Hash:
				typeName = "HASH"
			case *object.Function:
				typeName = "FUNCTION"
			case *object.Builtin:
				typeName = "BUILTIN"
			case *object.Null:
				typeName = "NULL"
			case *object.Instance:
				typeName = "INSTANCE"
			default:
				typeName = "UNKNOWN"
			}

			return &object.String{Value: typeName}
		},
	},

	"toString": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			return &object.String{Value: objectToString(args[0])}
		},
	},

	"parseInt": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `parseInt` must be STRING, got %T", args[0])
			}

			value, err := strconv.ParseInt(str.Value, 10, 64)
			if err != nil {
				return newError("could not parse '%s' as integer", str.Value)
			}

			return &object.Integer{Value: value}
		},
	},

	"parseFloat": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `parseFloat` must be STRING, got %T", args[0])
			}

			value, err := strconv.ParseFloat(str.Value, 64)
			if err != nil {
				return newError("could not parse '%s' as float", str.Value)
			}

			return &object.Float{Value: value}
		},
	},

	"keys": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			hash, ok := args[0].(*object.Hash)
			if !ok {
				return newError("argument to `keys` must be HASH, got %T", args[0])
			}

			keys := make([]object.Object, 0, len(hash.Pairs))
			for _, pair := range hash.Pairs {
				keys = append(keys, pair.Key)
			}

			return &object.Array{Elements: keys}
		},
	},

	"values": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			hash, ok := args[0].(*object.Hash)
			if !ok {
				return newError("argument to `values` must be HASH, got %T", args[0])
			}

			values := make([]object.Object, 0, len(hash.Pairs))
			for _, pair := range hash.Pairs {
				values = append(values, pair.Value)
			}

			return &object.Array{Elements: values}
		},
	},

	"push": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `push` must be ARRAY, got %T", args[0])
			}

			newElements := make([]object.Object, len(arr.Elements)+1)
			copy(newElements, arr.Elements)
			newElements[len(arr.Elements)] = args[1]

			return &object.Array{Elements: newElements}
		},
	},

	"pop": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `pop` must be ARRAY, got %T", args[0])
			}

			if len(arr.Elements) == 0 {
				return NULL
			}

			return arr.Elements[len(arr.Elements)-1]
		},
	},

	"first": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `first` must be ARRAY, got %T", args[0])
			}

			if len(arr.Elements) == 0 {
				return NULL
			}

			return arr.Elements[0]
		},
	},

	"last": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `last` must be ARRAY, got %T", args[0])
			}

			if len(arr.Elements) == 0 {
				return NULL
			}

			return arr.Elements[len(arr.Elements)-1]
		},
	},

	"rest": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `rest` must be ARRAY, got %T", args[0])
			}

			if len(arr.Elements) == 0 {
				return &object.Array{Elements: []object.Object{}}
			}

			newElements := make([]object.Object, len(arr.Elements)-1)
			copy(newElements, arr.Elements[1:])

			return &object.Array{Elements: newElements}
		},
	},

	"upper": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `upper` must be STRING, got %T", args[0])
			}

			return &object.String{Value: strings.ToUpper(str.Value)}
		},
	},

	"lower": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `lower` must be STRING, got %T", args[0])
			}

			return &object.String{Value: strings.ToLower(str.Value)}
		},
	},

	"split": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `split` must be STRING, got %T", args[0])
			}

			separator, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `split` must be STRING, got %T", args[1])
			}

			parts := strings.Split(str.Value, separator.Value)
			elements := make([]object.Object, len(parts))
			for i, part := range parts {
				elements[i] = &object.String{Value: part}
			}

			return &object.Array{Elements: elements}
		},
	},

	"join": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `join` must be ARRAY, got %T", args[0])
			}

			separator, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `join` must be STRING, got %T", args[1])
			}

			parts := make([]string, len(arr.Elements))
			for i, elem := range arr.Elements {
				parts[i] = objectToString(elem)
			}

			return &object.String{Value: strings.Join(parts, separator.Value)}
		},
	},

	"contains": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `contains` must be STRING, got %T", args[0])
			}

			substr, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `contains` must be STRING, got %T", args[1])
			}

			result := strings.Contains(str.Value, substr.Value)
			return nativeBoolToPyBoolean(result)
		},
	},

	"replace": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `replace` must be STRING, got %T", args[0])
			}

			old, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `replace` must be STRING, got %T", args[1])
			}

			new, ok := args[2].(*object.String)
			if !ok {
				return newError("third argument to `replace` must be STRING, got %T", args[2])
			}

			result := strings.ReplaceAll(str.Value, old.Value, new.Value)
			return &object.String{Value: result}
		},
	},

	"substring": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments. got=%d, want=2 or 3", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `substring` must be STRING, got %T", args[0])
			}

			start, ok := args[1].(*object.Integer)
			if !ok {
				return newError("second argument to `substring` must be INTEGER, got %T", args[1])
			}

			startIdx := int(start.Value)
			strLen := len(str.Value)

			if startIdx < 0 || startIdx >= strLen {
				return &object.String{Value: ""}
			}

			var endIdx int
			if len(args) == 3 {
				end, ok := args[2].(*object.Integer)
				if !ok {
					return newError("third argument to `substring` must be INTEGER, got %T", args[2])
				}
				endIdx = int(end.Value)
				if endIdx > strLen {
					endIdx = strLen
				}
			} else {
				endIdx = strLen
			}

			if startIdx >= endIdx {
				return &object.String{Value: ""}
			}

			return &object.String{Value: str.Value[startIdx:endIdx]}
		},
	},

	"indexOf": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `indexOf` must be STRING, got %T", args[0])
			}

			substr, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `indexOf` must be STRING, got %T", args[1])
			}

			index := strings.Index(str.Value, substr.Value)
			return &object.Integer{Value: int64(index)}
		},
	},

	"lastIndexOf": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `lastIndexOf` must be STRING, got %T", args[0])
			}

			substr, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `lastIndexOf` must be STRING, got %T", args[1])
			}

			index := strings.LastIndex(str.Value, substr.Value)
			return &object.Integer{Value: int64(index)}
		},
	},

	"charAt": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `charAt` must be STRING, got %T", args[0])
			}

			index, ok := args[1].(*object.Integer)
			if !ok {
				return newError("second argument to `charAt` must be INTEGER, got %T", args[1])
			}

			idx := int(index.Value)
			if idx < 0 || idx >= len(str.Value) {
				return &object.String{Value: ""}
			}

			return &object.String{Value: string(str.Value[idx])}
		},
	},

	"trim": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `trim` must be STRING, got %T", args[0])
			}

			return &object.String{Value: strings.TrimSpace(str.Value)}
		},
	},

	"trimLeft": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `trimLeft` must be STRING, got %T", args[0])
			}

			return &object.String{Value: strings.TrimLeftFunc(str.Value, func(r rune) bool {
				return r == ' ' || r == '\t' || r == '\n' || r == '\r'
			})}
		},
	},

	"trimRight": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `trimRight` must be STRING, got %T", args[0])
			}

			return &object.String{Value: strings.TrimRightFunc(str.Value, func(r rune) bool {
				return r == ' ' || r == '\t' || r == '\n' || r == '\r'
			})}
		},
	},

	"startsWith": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `startsWith` must be STRING, got %T", args[0])
			}

			prefix, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `startsWith` must be STRING, got %T", args[1])
			}

			result := strings.HasPrefix(str.Value, prefix.Value)
			return nativeBoolToPyBoolean(result)
		},
	},

	"endsWith": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `endsWith` must be STRING, got %T", args[0])
			}

			suffix, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `endsWith` must be STRING, got %T", args[1])
			}

			result := strings.HasSuffix(str.Value, suffix.Value)
			return nativeBoolToPyBoolean(result)
		},
	},

	"repeat": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `repeat` must be STRING, got %T", args[0])
			}

			count, ok := args[1].(*object.Integer)
			if !ok {
				return newError("second argument to `repeat` must be INTEGER, got %T", args[1])
			}

			if count.Value < 0 {
				return &object.String{Value: ""}
			}

			return &object.String{Value: strings.Repeat(str.Value, int(count.Value))}
		},
	},

	"reverse": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `reverse` must be STRING, got %T", args[0])
			}

			runes := []rune(str.Value)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}

			return &object.String{Value: string(runes)}
		},
	},

	"padLeft": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `padLeft` must be STRING, got %T", args[0])
			}

			length, ok := args[1].(*object.Integer)
			if !ok {
				return newError("second argument to `padLeft` must be INTEGER, got %T", args[1])
			}

			pad, ok := args[2].(*object.String)
			if !ok {
				return newError("third argument to `padLeft` must be STRING, got %T", args[2])
			}

			targetLen := int(length.Value)
			if targetLen <= len(str.Value) {
				return str
			}

			padCount := targetLen - len(str.Value)
			padding := strings.Repeat(pad.Value, padCount/len(pad.Value)+1)
			padding = padding[:padCount]

			return &object.String{Value: padding + str.Value}
		},
	},

	"padRight": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `padRight` must be STRING, got %T", args[0])
			}

			length, ok := args[1].(*object.Integer)
			if !ok {
				return newError("second argument to `padRight` must be INTEGER, got %T", args[1])
			}

			pad, ok := args[2].(*object.String)
			if !ok {
				return newError("third argument to `padRight` must be STRING, got %T", args[2])
			}

			targetLen := int(length.Value)
			if targetLen <= len(str.Value) {
				return str
			}

			padCount := targetLen - len(str.Value)
			padding := strings.Repeat(pad.Value, padCount/len(pad.Value)+1)
			padding = padding[:padCount]

			return &object.String{Value: str.Value + padding}
		},
	},

	"Exception": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			message := objectToString(args[0])
			return &object.Exception{
				Message: message,
				Value:   args[0],
			}
		},
	},

	"print": {
		Fn: func(args ...object.Object) object.Object {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(arg.Inspect())
			}
			return NULL
		},
	},

	"println": {
		Fn: func(args ...object.Object) object.Object {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(arg.Inspect())
			}
			fmt.Println()
			return NULL
		},
	},
}

func isBuiltin(name string) bool {
	_, exists := builtins[name]
	if exists {
		return true
	}

	_, exists = IOBuiltins[name]
	return exists
}

func getBuiltin(name string) *object.Builtin {

	if builtin, exists := builtins[name]; exists {
		return builtin
	}

	if builtin, exists := IOBuiltins[name]; exists {
		return builtin
	}

	if builtin, exists := JSONBuiltins[name]; exists {
		return builtin
	}

	if builtin, exists := RegexBuiltins[name]; exists {
		return builtin
	}

	if builtin, exists := CommonRegexBuiltins[name]; exists {
		return builtin
	}

	if builtin, exists := MathBuiltins[name]; exists {
		return builtin
	}

	if builtin, exists := DebugBuiltins[name]; exists {
		return builtin
	}

	switch name {
	case "setTimeout":
		return &object.Builtin{Fn: builtinSetTimeout}
	case "Promise":
		return &object.Builtin{Fn: builtinPromiseConstructor}
	case "resolve":
		return &object.Builtin{Fn: builtinPromiseResolve}
	case "reject":
		return &object.Builtin{Fn: builtinPromiseReject}
	case "sleep":
		return &object.Builtin{Fn: builtinSleep}
	case "fetch":
		return &object.Builtin{Fn: builtinFetch}
	default:
		return nil
	}
}
