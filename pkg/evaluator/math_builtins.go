package evaluator

import (
	"math"

	"jabline/pkg/object"
)

var MathBuiltins = map[string]*object.Builtin{
	"abs": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				value := arg.Value
				if value < 0 {
					value = -value
				}
				return &object.Integer{Value: value}
			case *object.Float:
				return &object.Float{Value: math.Abs(arg.Value)}
			default:
				return newError("argument to `abs` must be INTEGER or FLOAT, got %T", args[0])
			}
		},
	},

	"sqrt": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `sqrt` must be INTEGER or FLOAT, got %T", args[0])
			}

			if value < 0 {
				return newError("cannot compute square root of negative number")
			}

			result := math.Sqrt(value)
			return &object.Float{Value: result}
		},
	},

	"pow": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			var base, exp float64

			switch arg := args[0].(type) {
			case *object.Integer:
				base = float64(arg.Value)
			case *object.Float:
				base = arg.Value
			default:
				return newError("first argument to `pow` must be INTEGER or FLOAT, got %T", args[0])
			}

			switch arg := args[1].(type) {
			case *object.Integer:
				exp = float64(arg.Value)
			case *object.Float:
				exp = arg.Value
			default:
				return newError("second argument to `pow` must be INTEGER or FLOAT, got %T", args[1])
			}

			result := math.Pow(base, exp)
			return &object.Float{Value: result}
		},
	},

	"sin": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `sin` must be INTEGER or FLOAT, got %T", args[0])
			}

			result := math.Sin(value)
			return &object.Float{Value: result}
		},
	},

	"cos": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `cos` must be INTEGER or FLOAT, got %T", args[0])
			}

			result := math.Cos(value)
			return &object.Float{Value: result}
		},
	},

	"tan": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `tan` must be INTEGER or FLOAT, got %T", args[0])
			}

			result := math.Tan(value)
			return &object.Float{Value: result}
		},
	},

	"asin": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `asin` must be INTEGER or FLOAT, got %T", args[0])
			}

			if value < -1 || value > 1 {
				return newError("argument to `asin` must be between -1 and 1")
			}

			result := math.Asin(value)
			return &object.Float{Value: result}
		},
	},

	"acos": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `acos` must be INTEGER or FLOAT, got %T", args[0])
			}

			if value < -1 || value > 1 {
				return newError("argument to `acos` must be between -1 and 1")
			}

			result := math.Acos(value)
			return &object.Float{Value: result}
		},
	},

	"atan": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `atan` must be INTEGER or FLOAT, got %T", args[0])
			}

			result := math.Atan(value)
			return &object.Float{Value: result}
		},
	},

	"log": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `log` must be INTEGER or FLOAT, got %T", args[0])
			}

			if value <= 0 {
				return newError("argument to `log` must be positive")
			}

			result := math.Log(value)
			return &object.Float{Value: result}
		},
	},

	"log10": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `log10` must be INTEGER or FLOAT, got %T", args[0])
			}

			if value <= 0 {
				return newError("argument to `log10` must be positive")
			}

			result := math.Log10(value)
			return &object.Float{Value: result}
		},
	},

	"log2": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `log2` must be INTEGER or FLOAT, got %T", args[0])
			}

			if value <= 0 {
				return newError("argument to `log2` must be positive")
			}

			result := math.Log2(value)
			return &object.Float{Value: result}
		},
	},

	"exp": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var value float64
			switch arg := args[0].(type) {
			case *object.Integer:
				value = float64(arg.Value)
			case *object.Float:
				value = arg.Value
			default:
				return newError("argument to `exp` must be INTEGER or FLOAT, got %T", args[0])
			}

			result := math.Exp(value)
			return &object.Float{Value: result}
		},
	},

	"floor": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return arg
			case *object.Float:
				result := math.Floor(arg.Value)
				return &object.Integer{Value: int64(result)}
			default:
				return newError("argument to `floor` must be INTEGER or FLOAT, got %T", args[0])
			}
		},
	},

	"ceil": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return arg
			case *object.Float:
				result := math.Ceil(arg.Value)
				return &object.Integer{Value: int64(result)}
			default:
				return newError("argument to `ceil` must be INTEGER or FLOAT, got %T", args[0])
			}
		},
	},

	"round": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return arg
			case *object.Float:
				result := math.Round(arg.Value)
				return &object.Integer{Value: int64(result)}
			default:
				return newError("argument to `round` must be INTEGER or FLOAT, got %T", args[0])
			}
		},
	},

	"min": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want at least 2", len(args))
			}

			var minVal float64
			var isFloat bool

			for i, arg := range args {
				var value float64
				switch a := arg.(type) {
				case *object.Integer:
					value = float64(a.Value)
				case *object.Float:
					value = a.Value
					isFloat = true
				default:
					return newError("argument %d to `min` must be INTEGER or FLOAT, got %T", i, arg)
				}

				if i == 0 || value < minVal {
					minVal = value
				}
			}

			if isFloat {
				return &object.Float{Value: minVal}
			}
			return &object.Integer{Value: int64(minVal)}
		},
	},

	"max": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want at least 2", len(args))
			}

			var maxVal float64
			var isFloat bool

			for i, arg := range args {
				var value float64
				switch a := arg.(type) {
				case *object.Integer:
					value = float64(a.Value)
				case *object.Float:
					value = a.Value
					isFloat = true
				default:
					return newError("argument %d to `max` must be INTEGER or FLOAT, got %T", i, arg)
				}

				if i == 0 || value > maxVal {
					maxVal = value
				}
			}

			if isFloat {
				return &object.Float{Value: maxVal}
			}
			return &object.Integer{Value: int64(maxVal)}
		},
	},

	"PI": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			return &object.Float{Value: math.Pi}
		},
	},

	"E": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			return &object.Float{Value: math.E}
		},
	},

	"radians": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var degrees float64
			switch arg := args[0].(type) {
			case *object.Integer:
				degrees = float64(arg.Value)
			case *object.Float:
				degrees = arg.Value
			default:
				return newError("argument to `radians` must be INTEGER or FLOAT, got %T", args[0])
			}

			radians := degrees * (math.Pi / 180.0)
			return &object.Float{Value: radians}
		},
	},

	"degrees": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			var radians float64
			switch arg := args[0].(type) {
			case *object.Integer:
				radians = float64(arg.Value)
			case *object.Float:
				radians = arg.Value
			default:
				return newError("argument to `degrees` must be INTEGER or FLOAT, got %T", args[0])
			}

			degrees := radians * (180.0 / math.Pi)
			return &object.Float{Value: degrees}
		},
	},

	"random": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want 0, 1, or 2", len(args))
			}

			if len(args) == 0 {
				return &object.Float{Value: math.Abs(math.Sin(float64(len(args)+1)*12345.6789)) * 1000000}
			}

			if len(args) == 1 {
				maxArg, ok := args[0].(*object.Integer)
				if !ok {
					return newError("argument to `random` must be INTEGER, got %T", args[0])
				}

				if maxArg.Value <= 0 {
					return newError("argument to `random` must be positive")
				}

				seed := float64(maxArg.Value) * 9876.54321
				random := math.Abs(math.Sin(seed)) * 1000000
				result := int64(random) % maxArg.Value
				return &object.Integer{Value: result}
			}

			minArg, ok := args[0].(*object.Integer)
			if !ok {
				return newError("first argument to `random` must be INTEGER, got %T", args[0])
			}

			maxArg, ok := args[1].(*object.Integer)
			if !ok {
				return newError("second argument to `random` must be INTEGER, got %T", args[1])
			}

			if minArg.Value >= maxArg.Value {
				return newError("min must be less than max")
			}

			seed := float64(minArg.Value+maxArg.Value) * 1357.2468
			random := math.Abs(math.Sin(seed)) * 1000000
			result := minArg.Value + (int64(random) % (maxArg.Value - minArg.Value))
			return &object.Integer{Value: result}
		},
	},

	"factorial": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			n, ok := args[0].(*object.Integer)
			if !ok {
				return newError("argument to `factorial` must be INTEGER, got %T", args[0])
			}

			if n.Value < 0 {
				return newError("argument to `factorial` must be non-negative")
			}

			if n.Value > 20 {
				return newError("factorial of numbers greater than 20 is too large")
			}

			result := int64(1)
			for i := int64(2); i <= n.Value; i++ {
				result *= i
			}

			return &object.Integer{Value: result}
		},
	},
}

func InitMathBuiltins(builtins map[string]*object.Builtin) {
	for name, builtin := range MathBuiltins {
		builtins[name] = builtin
	}
}
