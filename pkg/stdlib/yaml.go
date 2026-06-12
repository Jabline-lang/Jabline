package stdlib

import (
	"strings"

	"jabline/pkg/object"
	"gopkg.in/yaml.v3"
)

func init() {
	NativeModuleRegistry["_yaml"] = YAMLBuiltins
	NativeModulePrefixes["_yaml"] = "yaml_"
}

var YAMLBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"yaml_parse", &object.Builtin{Fn: yamlParse}},
	{"yaml_encode", &object.Builtin{Fn: yamlEncode}},
	{"yaml_parse_file", &object.Builtin{Fn: yamlParseFile}},
}

func yamlParse(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("yaml_parse expects 1 argument, got %d", len(args))
	}

	s, ok := args[0].(*object.String)
	if !ok {
		return newError("yaml_parse expects string, got %s", args[0].Type())
	}

	var data interface{}
	if err := yaml.Unmarshal([]byte(s.Value), &data); err != nil {
		return newError("yaml_parse error: %s", err.Error())
	}

	return yamlToObject(data)
}

func yamlEncode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("yaml_encode expects 1 argument, got %d", len(args))
	}

	goVal := objectToGoValue(args[0])
	data, err := yaml.Marshal(goVal)
	if err != nil {
		return newError("yaml_encode error: %s", err.Error())
	}

	return &object.String{Value: strings.TrimSpace(string(data))}
}

func yamlParseFile(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("yaml_parse_file expects 1 argument, got %d", len(args))
	}

	path, ok := args[0].(*object.String)
	if !ok {
		return newError("yaml_parse_file expects string, got %s", args[0].Type())
	}

	content, err := readFile(path.Value)
	if err != nil {
		return newError("yaml_parse_file: %s", err.Error())
	}

	return yamlParse(&object.String{Value: string(content)})
}

func yamlToObject(v interface{}) object.Object {
	switch val := v.(type) {
	case map[string]interface{}:
		pairs := make(map[object.HashKey]object.HashPair)
		for k, v := range val {
			key := &object.String{Value: k}
			pairs[key.HashKey()] = object.HashPair{
				Key:   key,
				Value: yamlToObject(v),
			}
		}
		return &object.Hash{Pairs: pairs}
	case []interface{}:
		elements := make([]object.Object, len(val))
		for i, v := range val {
			elements[i] = yamlToObject(v)
		}
		return &object.Array{Elements: elements}
	case string:
		return &object.String{Value: val}
	case int:
		return &object.Integer{Value: int64(val)}
	case int64:
		return &object.Integer{Value: val}
	case float64:
		return &object.Float{Value: val}
	case bool:
		return &object.Boolean{Value: val}
	case nil:
		return object.NullObj
	default:
		return &object.String{Value: ""}
	}
}

func objectToGoValue(obj object.Object) interface{} {
	switch o := obj.(type) {
	case *object.Hash:
		m := make(map[string]interface{})
		for _, pair := range o.Pairs {
			if key, ok := pair.Key.(*object.String); ok {
				m[key.Value] = objectToGoValue(pair.Value)
			}
		}
		return m
	case *object.Array:
		arr := make([]interface{}, len(o.Elements))
		for i, e := range o.Elements {
			arr[i] = objectToGoValue(e)
		}
		return arr
	case *object.String:
		return o.Value
	case *object.Integer:
		return o.Value
	case *object.Float:
		return o.Value
	case *object.Boolean:
		return o.Value
	case *object.Null:
		return nil
	default:
		return o.Inspect()
	}
}
