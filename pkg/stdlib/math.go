package stdlib

import (
	"jabline/pkg/object"
	"math"
	"math/rand"
)

var MathBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"abs", &object.Builtin{Fn: mathAbs}},
	{"sqrt", &object.Builtin{Fn: mathSqrt}},
	{"pow", &object.Builtin{Fn: mathPow}},
	{"sin", &object.Builtin{Fn: mathSin}},
	{"cos", &object.Builtin{Fn: mathCos}},
	{"tan", &object.Builtin{Fn: mathTan}},
	{"random", &object.Builtin{Fn: mathRandom}},
	{"max", &object.Builtin{Fn: mathMax}},
	{"min", &object.Builtin{Fn: mathMin}},
	{"floor", &object.Builtin{Fn: mathFloor}},
	{"ceil", &object.Builtin{Fn: mathCeil}},
	{"round", &object.Builtin{Fn: mathRound}},
	{"log", &object.Builtin{Fn: mathLog}},
	{"log10", &object.Builtin{Fn: mathLog10}},
	{"exp", &object.Builtin{Fn: mathExp}},
	{"atan2", &object.Builtin{Fn: mathAtan2}},
	{"asin", &object.Builtin{Fn: mathAsin}},
	{"acos", &object.Builtin{Fn: mathAcos}},
	{"hypot", &object.Builtin{Fn: mathHypot}},
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
	case *object.Integer:
		base = float64(arg.Value)
	case *object.Float:
		base = arg.Value
	default:
		return newError("first argument to `pow` must be INTEGER or FLOAT")
	}

	switch arg := args[1].(type) {
	case *object.Integer:
		exp = float64(arg.Value)
	case *object.Float:
		exp = arg.Value
	default:
		return newError("second argument to `pow` must be INTEGER or FLOAT")
	}

	return &object.Float{Value: math.Pow(base, exp)}
}

func mathSin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Float{Value: math.Sin(*val)}
}

func mathCos(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Float{Value: math.Cos(*val)}
}

func mathTan(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
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

func mathMax(args ...object.Object) object.Object {
	if len(args) == 0 {
		return newError("wrong number of args")
	}
	var max float64 = -math.MaxFloat64
	for _, arg := range args {
		val := toFloat(arg)
		if val == nil {
			return newError("args must be numbers")
		}
		max = math.Max(max, *val)
	}
	return &object.Float{Value: max}
}

func mathMin(args ...object.Object) object.Object {
	if len(args) == 0 {
		return newError("wrong number of args")
	}
	var min float64 = math.MaxFloat64
	for _, arg := range args {
		val := toFloat(arg)
		if val == nil {
			return newError("args must be numbers")
		}
		min = math.Min(min, *val)
	}
	return &object.Float{Value: min}
}

func mathFloor(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Integer{Value: int64(math.Floor(*val))}
}

func mathCeil(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Integer{Value: int64(math.Ceil(*val))}
}

func mathRound(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Integer{Value: int64(math.Round(*val))}
}

func mathLog(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Float{Value: math.Log(*val)}
}

func mathLog10(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Float{Value: math.Log10(*val)}
}

func mathExp(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Float{Value: math.Exp(*val)}
}

func mathAtan2(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of args")
	}
	y := toFloat(args[0])
	x := toFloat(args[1])
	if y == nil || x == nil {
		return newError("args must be numbers")
	}
	return &object.Float{Value: math.Atan2(*y, *x)}
}

func mathAsin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Float{Value: math.Asin(*val)}
}

func mathAcos(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of args")
	}
	val := toFloat(args[0])
	if val == nil {
		return newError("arg must be number")
	}
	return &object.Float{Value: math.Acos(*val)}
}

func mathHypot(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of args")
	}
	p := toFloat(args[0])
	q := toFloat(args[1])
	if p == nil || q == nil {
		return newError("args must be numbers")
	}
	return &object.Float{Value: math.Hypot(*p, *q)}
}
