package evaluator

import (
	"fmt"
	"strings"

	"jabline/pkg/object"
)

// DebugBuiltins contains all debugging-related built-in functions
var DebugBuiltins = map[string]*object.Builtin{
	"debug": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) == 0 {
				return newError("wrong number of arguments. got=%d, want at least 1", len(args))
			}

			// Format debug output
			var parts []string
			for _, arg := range args {
				parts = append(parts, arg.Inspect())
			}

			message := strings.Join(parts, " ")

			// Create formatted debug output
			formatter := NewErrorFormatter(true, false)
			debugInfo := FormattedError{
				Level:      ErrorLevelInfo,
				Message:    message,
				Line:       0,
				Column:     0,
				Filename:   "<debug>",
				SourceLine: "",
				Suggestion: "",
			}

			output := formatter.FormatError(debugInfo)
			fmt.Print(output)

			return NULL
		},
	},

	"debugger": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}

			// Get current debug info
			env := object.NewEnvironment() // This would be passed from context in real implementation
			info := GetDebugInfo(env)

			// Format and display debug information
			debugOutput := FormatDebugInfo(info, true)
			fmt.Print(debugOutput)

			// Show current stack trace
			if GlobalCallStack.Depth() > 0 {
				fmt.Print("\n")
				fmt.Print(GlobalCallStack.FormatStackTrace(true))
			}

			fmt.Print("\n🔍 Debugger paused. Press Enter to continue...\n")

			// In a real debugger, this would pause execution and wait for input
			// For now, we just show the info
			var input string
			fmt.Scanln(&input)

			return NULL
		},
	},

	"stackTrace": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}

			// Get current stack trace
			if GlobalCallStack.Depth() == 0 {
				fmt.Print("No stack trace available\n")
				return NULL
			}

			stackTrace := GlobalCallStack.FormatStackTrace(true)
			fmt.Print(stackTrace)

			return NULL
		},
	},

	"vars": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}

			// Get current environment variables
			env := object.NewEnvironment() // This would be passed from context in real implementation
			vars := env.GetAll()

			if len(vars) == 0 {
				fmt.Print("No variables in current scope\n")
				return NULL
			}

			fmt.Print("📋 Current Variables:\n")
			for name, value := range vars {
				fmt.Printf("  %s%s%s = %s (%s%s%s)\n",
					ColorGreen, name, ColorReset,
					value.Inspect(),
					ColorBlue, value.Type(), ColorReset)
			}

			return NULL
		},
	},

	"typeof": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			obj := args[0]
			typeName := obj.Type()

			// Enhanced type information
			switch o := obj.(type) {
			case *object.Array:
				return &object.String{Value: fmt.Sprintf("array[%d]", len(o.Elements))}
			case *object.Hash:
				return &object.String{Value: fmt.Sprintf("hash[%d]", len(o.Pairs))}
			case *object.String:
				return &object.String{Value: fmt.Sprintf("string[%d]", len(o.Value))}
			case *object.Function:
				paramCount := len(o.Parameters)
				return &object.String{Value: fmt.Sprintf("function(%d)", paramCount)}
			case *object.Builtin:
				return &object.String{Value: "builtin"}
			default:
				return &object.String{Value: string(typeName)}
			}
		},
	},

	"inspect": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			obj := args[0]

			// Create detailed inspection
			var result strings.Builder

			result.WriteString(fmt.Sprintf("%sType:%s %s\n", ColorCyan, ColorReset, obj.Type()))
			result.WriteString(fmt.Sprintf("%sValue:%s %s\n", ColorCyan, ColorReset, obj.Inspect()))

			switch o := obj.(type) {
			case *object.Array:
				result.WriteString(fmt.Sprintf("%sLength:%s %d\n", ColorCyan, ColorReset, len(o.Elements)))
				result.WriteString(fmt.Sprintf("%sElements:%s\n", ColorCyan, ColorReset))
				for i, elem := range o.Elements {
					if i < 10 { // Limit to first 10 elements
						result.WriteString(fmt.Sprintf("  [%d] = %s (%s)\n", i, elem.Inspect(), elem.Type()))
					} else if i == 10 {
						result.WriteString("  ... (more elements)\n")
					}
				}
			case *object.Hash:
				result.WriteString(fmt.Sprintf("%sSize:%s %d\n", ColorCyan, ColorReset, len(o.Pairs)))
				result.WriteString(fmt.Sprintf("%sKeys:%s\n", ColorCyan, ColorReset))
				count := 0
				for _, pair := range o.Pairs {
					if count < 10 { // Limit to first 10 pairs
						result.WriteString(fmt.Sprintf("  %s = %s (%s)\n",
							pair.Key.Inspect(), pair.Value.Inspect(), pair.Value.Type()))
						count++
					} else if count == 10 {
						result.WriteString("  ... (more keys)\n")
						break
					}
				}
			case *object.String:
				result.WriteString(fmt.Sprintf("%sLength:%s %d\n", ColorCyan, ColorReset, len(o.Value)))
				if len(o.Value) > 100 {
					result.WriteString(fmt.Sprintf("%sPreview:%s %s...\n", ColorCyan, ColorReset, o.Value[:100]))
				}
			case *object.Function:
				result.WriteString(fmt.Sprintf("%sParameters:%s %d\n", ColorCyan, ColorReset, len(o.Parameters)))
				if len(o.Parameters) > 0 {
					var params []string
					for _, param := range o.Parameters {
						params = append(params, param.Value)
					}
					result.WriteString(fmt.Sprintf("  (%s)\n", strings.Join(params, ", ")))
				}
			}

			fmt.Print(result.String())
			return NULL
		},
	},

	"trace": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want at least 1", len(args))
			}

			// Get message
			var parts []string
			for _, arg := range args {
				parts = append(parts, arg.Inspect())
			}
			message := strings.Join(parts, " ")

			// Get current location info
			currentFrame := GlobalCallStack.Current()
			location := "<unknown>"
			if currentFrame != nil {
				location = fmt.Sprintf("%s:%d:%d", currentFrame.Filename, currentFrame.Line, currentFrame.Column)
			}

			// Format trace output
			fmt.Printf("%s🔍 TRACE%s [%s] %s\n",
				ColorPurple+ColorBold, ColorReset, location, message)

			return NULL
		},
	},

	"benchmark": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			// This is a placeholder - in a real implementation,
			// this would measure execution time of a function
			nameArg, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument must be STRING, got %T", args[0])
			}

			// The second argument should be a function to benchmark
			// For now, we just return a mock result
			fmt.Printf("%s⏱️  BENCHMARK%s %s: <timing would be measured here>\n",
				ColorYellow+ColorBold, ColorReset, nameArg.Value)

			return NULL
		},
	},

	"assert": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
			}

			// Check assertion
			condition := args[0]
			message := "Assertion failed"

			if len(args) == 2 {
				if msgArg, ok := args[1].(*object.String); ok {
					message = msgArg.Value
				}
			}

			// Evaluate condition
			var passed bool
			switch c := condition.(type) {
			case *object.Boolean:
				passed = c.Value
			case *object.Null:
				passed = false
			default:
				passed = true // Any non-null, non-false value is truthy
			}

			if !passed {
				// Create assertion error
				errorMsg := fmt.Sprintf("❌ ASSERTION FAILED: %s", message)

				// Show stack trace for failed assertions
				if GlobalCallStack.Depth() > 0 {
					fmt.Printf("%s%s%s\n", ColorRed+ColorBold, errorMsg, ColorReset)
					fmt.Print(GlobalCallStack.FormatStackTrace(true))
				} else {
					fmt.Printf("%s%s%s\n", ColorRed+ColorBold, errorMsg, ColorReset)
				}

				return newError(message)
			}

			fmt.Printf("%s✅ ASSERTION PASSED%s\n", ColorGreen+ColorBold, ColorReset)
			return NULL
		},
	},
}

// InitDebugBuiltins initializes debug built-ins in the global builtins map
func InitDebugBuiltins(builtins map[string]*object.Builtin) {
	for name, builtin := range DebugBuiltins {
		builtins[name] = builtin
	}
}
