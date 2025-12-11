package stdlib

import (
	"jabline/pkg/object"
	"math"
	"math/rand"
)

var MathBuiltins = []struct {
	Name    string
	Builtin *object.Builtin
}{
	{"abs", &object.Builtin{Fn: mathAbs}},
	{"sqrt", &object.Builtin{Fn: mathSqrt}},
	{"pow", &object.Builtin{Fn: mathPow}},
	{"sin", &object.Builtin{Fn: mathSin}},
	{"cos", &object.Builtin{Fn: mathCos}},
	{"tan", &object.Builtin{Fn: mathTan}},
	{"random", &object.Builtin{Fn: mathRandom}},
}

func mathAbs(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		if arg.Value < 0 {
			return &object.Integer{Value: -arg.Value}
		}
		return arg
	case *object.Float:
		return &object.Float{Value: math.Abs(arg.Value)}
	default:
		return newError("argument to `abs` must be INTEGER or FLOAT, got %T", args[0])
	}
}

func mathSqrt(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	switch arg := args[0].(type) {
	case *object.Integer:
		return &object.Float{Value: math.Sqrt(float64(arg.Value))}
	case *object.Float:
		return &object.Float{Value: math.Sqrt(arg.Value)}
	default:
		return newError("argument to `sqrt` must be INTEGER or FLOAT, got %T", args[0])
	}
}

func mathPow(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	
	var base float64
	var exp float64
	
	switch arg := args[0].(type) {
	case *object.Integer: base = float64(arg.Value)
	case *object.Float: base = arg.Value
	default: return newError("first argument to `pow` must be INTEGER or FLOAT")
	}
	
	switch arg := args[1].(type) {
	case *object.Integer: exp = float64(arg.Value)
	case *object.Float: exp = arg.Value
	default: return newError("second argument to `pow` must be INTEGER or FLOAT")
	}
	
	return &object.Float{Value: math.Pow(base, exp)}
}

func mathSin(args ...object.Object) object.Object {
	if len(args) != 1 { return newError("wrong number of args") }
	val := toFloat(args[0])
	if val == nil { return newError("arg must be number") }
	return &object.Float{Value: math.Sin(*val)}
}

func mathCos(args ...object.Object) object.Object {
	if len(args) != 1 { return newError("wrong number of args") }
	val := toFloat(args[0])
	if val == nil { return newError("arg must be number") }
	return &object.Float{Value: math.Cos(*val)}
}

func mathTan(args ...object.Object) object.Object {
	if len(args) != 1 { return newError("wrong number of args") }
	val := toFloat(args[0])
	if val == nil { return newError("arg must be number") }
	return &object.Float{Value: math.Tan(*val)}
}

func mathRandom(args ...object.Object) object.Object {

	return &object.Float{Value: rand.Float64()}
}

func toFloat(obj object.Object) *float64 {
	switch arg := obj.(type) {
	case *object.Integer:
		v := float64(arg.Value)
		return &v
	case *object.Float:
		return &arg.Value
	}
	return nil
}
