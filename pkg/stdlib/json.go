package stdlib

import (
	"encoding/json"
	"jabline/pkg/object"
)

var JSONBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"json_parse", &object.Builtin{Fn: jsonParse}},
	{"json_stringify", &object.Builtin{Fn: jsonStringify}},
	{"json_pretty", &object.Builtin{Fn: jsonPretty}},
}

func jsonParse(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return newError("arg must be string")
	}

	var data interface{}
	err := json.Unmarshal([]byte(s.Value), &data)
	if err != nil {
		return newError("json error: %s", err)
	}

	return goToJabline(data)
}

func jsonStringify(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}

	data := jablineToGo(args[0])
	bytes, err := json.Marshal(data)
	if err != nil {
		return newError("json error: %s", err)
	}

	return &object.String{Value: string(bytes)}
}

func jsonPretty(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}

	data := jablineToGo(args[0])
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return newError("json error: %s", err)
	}

	return &object.String{Value: string(bytes)}
}

func goToJabline(val interface{}) object.Object {
	switch v := val.(type) {
	case string:
		return &object.String{Value: v}
	case float64:
		if v == float64(int64(v)) {
			return &object.Integer{Value: int64(v)}
		}
		return &object.Float{Value: v}
	case int:
		return &object.Integer{Value: int64(v)}
	case int64:
		return &object.Integer{Value: v}
	case bool:
		if v {
			return &object.Boolean{Value: true}
		}
		return &object.Boolean{Value: false}
	case nil:
		return &object.Null{}
	case map[string]interface{}:
		pairs := make(map[object.HashKey]object.HashPair)
		for k, val := range v {
			key := &object.String{Value: k}
			value := goToJabline(val)
			pairs[key.HashKey()] = object.HashPair{Key: key, Value: value}
		}
		return &object.Hash{Pairs: pairs}
	case []interface{}:
		elements := make([]object.Object, len(v))
		for i, val := range v {
			elements[i] = goToJabline(val)
		}
		return &object.Array{Elements: elements}
	}
	return &object.Null{}
}

func jablineToGo(obj object.Object) interface{} {
	switch o := obj.(type) {
	case *object.Integer:
		return o.Value
	case *object.Float:
		return o.Value
	case *object.String:
		return o.Value
	case *object.Boolean:
		return o.Value
	case *object.Null:
		return nil
	case *object.Array:
		list := make([]interface{}, len(o.Elements))
		for i, el := range o.Elements {
			list[i] = jablineToGo(el)
		}
		return list
	case *object.Hash:
		m := make(map[string]interface{})
		for _, pair := range o.Pairs {
			key, ok := pair.Key.(*object.String)
			if ok {
				m[key.Value] = jablineToGo(pair.Value)
			}
		}
		return m
	}
	return nil
}
