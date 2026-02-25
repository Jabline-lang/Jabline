package stdlib

import (
	"jabline/pkg/object"
)

var TypesBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"int8", &object.Builtin{Fn: toInt8}},
	{"int16", &object.Builtin{Fn: toInt16}},
	{"int32", &object.Builtin{Fn: toInt32}},
	{"int64", &object.Builtin{Fn: toInt64}},
	{"uint8", &object.Builtin{Fn: toUint8}},
	{"uint16", &object.Builtin{Fn: toUint16}},
	{"uint32", &object.Builtin{Fn: toUint32}},
	{"uint64", &object.Builtin{Fn: toUint64}}, // Returns Int64 but bit-representation is uint64 (reinterpreted)
	{"float32", &object.Builtin{Fn: toFloat32}},
	{"float64", &object.Builtin{Fn: toFloat64}},
}

func getInt(arg object.Object) (int64, *object.Error) {
	switch v := arg.(type) {
	case *object.Integer:
		return v.Value, nil
	case *object.Float:
		return int64(v.Value), nil
	default:
		return 0, newError("expected number, got %s", arg.Type())
	}
}

func toInt8(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	val, err := getInt(args[0])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(int8(val))}
}

func toInt16(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	val, err := getInt(args[0])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(int16(val))}
}

func toInt32(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	val, err := getInt(args[0])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(int32(val))}
}

func toInt64(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	val, err := getInt(args[0])
	if err != nil {
		return err
	}
	return &object.Integer{Value: val}
}

func toUint8(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	val, err := getInt(args[0])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(uint8(val))}
}

func toUint16(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	val, err := getInt(args[0])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(uint16(val))}
}

func toUint32(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	val, err := getInt(args[0])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(uint32(val))}
}

// Mimic uint64 by storing bits in int64. Arithmetic might act signed in Jabline,
// but bits are preserved for storage/bitwise ops.
func toUint64(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	val, err := getInt(args[0])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(uint64(val))}
}

func toFloat32(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	var val float64
	switch v := args[0].(type) {
	case *object.Integer:
		val = float64(v.Value)
	case *object.Float:
		val = v.Value
	default:
		return newError("expected number")
	}
	return &object.Float{Value: float64(float32(val))}
}

func toFloat64(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong args")
	}
	switch v := args[0].(type) {
	case *object.Integer:
		return &object.Float{Value: float64(v.Value)}
	case *object.Float:
		return v
	default:
		return newError("expected number")
	}
}
