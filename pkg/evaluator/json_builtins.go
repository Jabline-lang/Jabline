package evaluator

import (
	"encoding/json"
	"fmt"
	"reflect"

	"jabline/pkg/object"
)

// JSONBuiltins contains all JSON-related built-in functions
var JSONBuiltins = map[string]*object.Builtin{
	"JSON.stringify": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			jsonStr, err := objectToJSON(args[0])
			if err != nil {
				return newError("error converting to JSON: %s", err.Error())
			}

			return &object.String{Value: jsonStr}
		},
	},

	"JSON.parse": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `JSON.parse` must be STRING, got %T", args[0])
			}

			obj, err := jsonToObject(str.Value)
			if err != nil {
				return newError("error parsing JSON: %s", err.Error())
			}

			return obj
		},
	},

	"JSON.isValid": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return nativeBoolToJablineObject(false)
			}

			var temp interface{}
			err := json.Unmarshal([]byte(str.Value), &temp)
			return nativeBoolToJablineObject(err == nil)
		},
	},

	// Alias functions for backward compatibility and convenience
	"stringify": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			jsonStr, err := objectToJSON(args[0])
			if err != nil {
				return newError("error converting to JSON: %s", err.Error())
			}

			return &object.String{Value: jsonStr}
		},
	},

	"parse": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `parse` must be STRING, got %T", args[0])
			}

			obj, err := jsonToObject(str.Value)
			if err != nil {
				return newError("error parsing JSON: %s", err.Error())
			}

			return obj
		},
	},
}

// objectToJSON converts a Jabline object to JSON string
func objectToJSON(obj object.Object) (string, error) {
	switch o := obj.(type) {
	case *object.String:
		jsonBytes, err := json.Marshal(o.Value)
		return string(jsonBytes), err

	case *object.Integer:
		return fmt.Sprintf("%d", o.Value), nil

	case *object.Float:
		return fmt.Sprintf("%g", o.Value), nil

	case *object.Boolean:
		if o.Value {
			return "true", nil
		}
		return "false", nil

	case *object.Null:
		return "null", nil

	case *object.Array:
		elements := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			goValue, err := jablineObjectToGoValue(elem)
			if err != nil {
				return "", err
			}
			elements[i] = goValue
		}
		jsonBytes, err := json.Marshal(elements)
		return string(jsonBytes), err

	case *object.Hash:
		goMap := make(map[string]interface{})
		for _, pair := range o.Pairs {
			key, ok := pair.Key.(*object.String)
			if !ok {
				return "", fmt.Errorf("hash keys must be strings for JSON conversion")
			}

			value, err := jablineObjectToGoValue(pair.Value)
			if err != nil {
				return "", err
			}

			goMap[key.Value] = value
		}
		jsonBytes, err := json.Marshal(goMap)
		return string(jsonBytes), err

	default:
		return "", fmt.Errorf("cannot convert %T to JSON", obj)
	}
}

// jablineObjectToGoValue converts Jabline object to Go value for JSON marshaling
func jablineObjectToGoValue(obj object.Object) (interface{}, error) {
	switch o := obj.(type) {
	case *object.String:
		return o.Value, nil
	case *object.Integer:
		return o.Value, nil
	case *object.Float:
		return o.Value, nil
	case *object.Boolean:
		return o.Value, nil
	case *object.Null:
		return nil, nil
	case *object.Array:
		elements := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			value, err := jablineObjectToGoValue(elem)
			if err != nil {
				return nil, err
			}
			elements[i] = value
		}
		return elements, nil
	case *object.Hash:
		goMap := make(map[string]interface{})
		for _, pair := range o.Pairs {
			key, ok := pair.Key.(*object.String)
			if !ok {
				return nil, fmt.Errorf("hash keys must be strings for JSON conversion")
			}

			value, err := jablineObjectToGoValue(pair.Value)
			if err != nil {
				return nil, err
			}

			goMap[key.Value] = value
		}
		return goMap, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to Go value", obj)
	}
}

// jsonToObject converts JSON string to Jabline object
func jsonToObject(jsonStr string) (object.Object, error) {
	var value interface{}
	err := json.Unmarshal([]byte(jsonStr), &value)
	if err != nil {
		return nil, err
	}

	return goValueToJablineObject(value), nil
}

// goValueToJablineObject converts Go value from JSON to Jabline object
func goValueToJablineObject(value interface{}) object.Object {
	if value == nil {
		return NULL
	}

	switch v := value.(type) {
	case bool:
		return nativeBoolToJablineObject(v)
	case float64:
		// JSON numbers are always float64, but we try to preserve integers
		if v == float64(int64(v)) {
			return &object.Integer{Value: int64(v)}
		}
		return &object.Float{Value: v}
	case string:
		return &object.String{Value: v}
	case []interface{}:
		elements := make([]object.Object, len(v))
		for i, elem := range v {
			elements[i] = goValueToJablineObject(elem)
		}
		return &object.Array{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[object.HashKey]object.HashPair)
		for key, val := range v {
			keyObj := &object.String{Value: key}
			valueObj := goValueToJablineObject(val)
			hashKey := keyObj.HashKey()
			pairs[hashKey] = object.HashPair{
				Key:   keyObj,
				Value: valueObj,
			}
		}
		return &object.Hash{Pairs: pairs}
	default:
		// Handle other numeric types that might come from JSON
		switch reflect.TypeOf(v).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return &object.Integer{Value: reflect.ValueOf(v).Int()}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return &object.Integer{Value: int64(reflect.ValueOf(v).Uint())}
		case reflect.Float32, reflect.Float64:
			return &object.Float{Value: reflect.ValueOf(v).Float()}
		default:
			// Fallback to string representation
			return &object.String{Value: fmt.Sprintf("%v", v)}
		}
	}
}

// nativeBoolToJablineObject converts native Go bool to Jabline boolean object
func nativeBoolToJablineObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}

// Pretty print functions for debugging and development
var JSONPrettyBuiltins = map[string]*object.Builtin{
	"JSON.prettify": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
			}

			// Get the object to prettify
			obj := args[0]

			// Get indent (default to 2 spaces)
			indent := "  "
			if len(args) == 2 {
				indentArg, ok := args[1].(*object.String)
				if ok {
					indent = indentArg.Value
				} else if indentNum, ok := args[1].(*object.Integer); ok {
					// If number, create that many spaces
					spaces := ""
					for i := int64(0); i < indentNum.Value; i++ {
						spaces += " "
					}
					indent = spaces
				}
			}

			// Convert to Go value first
			goValue, err := jablineObjectToGoValue(obj)
			if err != nil {
				return newError("error converting to JSON: %s", err.Error())
			}

			// Pretty print
			jsonBytes, err := json.MarshalIndent(goValue, "", indent)
			if err != nil {
				return newError("error creating pretty JSON: %s", err.Error())
			}

			return &object.String{Value: string(jsonBytes)}
		},
	},
}

// InitJSONBuiltins initializes JSON built-ins in the global builtins map
func InitJSONBuiltins(builtins map[string]*object.Builtin) {
	for name, builtin := range JSONBuiltins {
		builtins[name] = builtin
	}
	for name, builtin := range JSONPrettyBuiltins {
		builtins[name] = builtin
	}
}
