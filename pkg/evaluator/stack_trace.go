package evaluator

import (
	"fmt"
	"strings"

	"jabline/pkg/ast"
	"jabline/pkg/object"
)

type StackFrame struct {
	FunctionName string
	Line         int
	Column       int
	Filename     string
	SourceLine   string
}

type CallStack struct {
	Frames []StackFrame
}

func NewCallStack() *CallStack {
	return &CallStack{
		Frames: make([]StackFrame, 0),
	}
}

func (cs *CallStack) Push(functionName string, line, column int, filename, sourceLine string) {
	frame := StackFrame{
		FunctionName: functionName,
		Line:         line,
		Column:       column,
		Filename:     filename,
		SourceLine:   sourceLine,
	}
	cs.Frames = append(cs.Frames, frame)
}

func (cs *CallStack) Pop() {
	if len(cs.Frames) > 0 {
		cs.Frames = cs.Frames[:len(cs.Frames)-1]
	}
}

func (cs *CallStack) Current() *StackFrame {
	if len(cs.Frames) == 0 {
		return nil
	}
	return &cs.Frames[len(cs.Frames)-1]
}

func (cs *CallStack) Depth() int {
	return len(cs.Frames)
}

func (cs *CallStack) Clear() {
	cs.Frames = cs.Frames[:0]
}

func (cs *CallStack) FormatStackTrace(useColors bool) string {
	if len(cs.Frames) == 0 {
		return ""
	}

	var result strings.Builder

	if useColors {
		result.WriteString(ColorRed + ColorBold)
	}
	result.WriteString("Stack Trace:\n")
	if useColors {
		result.WriteString(ColorReset)
	}

	for i := len(cs.Frames) - 1; i >= 0; i-- {
		frame := cs.Frames[i]

		if useColors {
			result.WriteString(ColorBlue + ColorDim)
		}
		result.WriteString(fmt.Sprintf("  at %s", frame.FunctionName))

		if frame.Filename != "" {
			result.WriteString(fmt.Sprintf(" (%s", frame.Filename))
			if frame.Line > 0 {
				result.WriteString(fmt.Sprintf(":%d", frame.Line))
				if frame.Column > 0 {
					result.WriteString(fmt.Sprintf(":%d", frame.Column))
				}
			}
			result.WriteString(")")
		}

		if useColors {
			result.WriteString(ColorReset)
		}
		result.WriteString("\n")

		if frame.SourceLine != "" {
			if useColors {
				result.WriteString(ColorDim)
			}
			result.WriteString(fmt.Sprintf("    %s\n", strings.TrimSpace(frame.SourceLine)))
			if useColors {
				result.WriteString(ColorReset)
			}
		}
	}

	return result.String()
}

type StackEnhancedError struct {
	Message    string
	StackTrace *CallStack
	ErrorType  string
	Line       int
	Column     int
	Filename   string
}

func NewStackEnhancedError(message, errorType string, stack *CallStack) *StackEnhancedError {
	stackCopy := &CallStack{
		Frames: make([]StackFrame, len(stack.Frames)),
	}
	copy(stackCopy.Frames, stack.Frames)

	err := &StackEnhancedError{
		Message:    message,
		StackTrace: stackCopy,
		ErrorType:  errorType,
	}

	if current := stack.Current(); current != nil {
		err.Line = current.Line
		err.Column = current.Column
		err.Filename = current.Filename
	}

	return err
}

func (ee *StackEnhancedError) Format(useColors bool) string {
	var result strings.Builder

	if useColors {
		result.WriteString(ColorRed + ColorBold)
	}

	result.WriteString(fmt.Sprintf("❌ %s: %s", ee.ErrorType, ee.Message))

	if ee.Filename != "" {
		result.WriteString(fmt.Sprintf(" at %s", ee.Filename))
		if ee.Line > 0 {
			result.WriteString(fmt.Sprintf(":%d", ee.Line))
			if ee.Column > 0 {
				result.WriteString(fmt.Sprintf(":%d", ee.Column))
			}
		}
	}

	if useColors {
		result.WriteString(ColorReset)
	}
	result.WriteString("\n")

	if ee.StackTrace != nil && ee.StackTrace.Depth() > 0 {
		result.WriteString("\n")
		result.WriteString(ee.StackTrace.FormatStackTrace(useColors))
	}

	return result.String()
}

var GlobalCallStack = NewCallStack()

func WithStackFrame(functionName string, line, column int, filename, sourceLine string, fn func() object.Object) object.Object {
	GlobalCallStack.Push(functionName, line, column, filename, sourceLine)
	defer GlobalCallStack.Pop()

	if GlobalCallStack.Depth() > 100 {
		return &object.Error{
			Message: "Stack overflow - maximum call depth exceeded",
		}
	}

	return fn()
}

func NewRuntimeError(message string) *object.Error {
	if GlobalCallStack.Depth() > 0 {
		enhanced := NewStackEnhancedError(message, "RUNTIME_ERROR", GlobalCallStack)
		return &object.Error{
			Message: enhanced.Format(true),
		}
	}

	return &object.Error{Message: message}
}

func GetFunctionName(node ast.Node) string {
	switch n := node.(type) {
	case *ast.FunctionLiteral:
		return "<anonymous>"
	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok {
			return ident.Value
		}
		return "<call>"
	case *ast.Identifier:
		return n.Value
	default:
		return "<unknown>"
	}
}

func ExtractLineInfo(node ast.Node) (int, int) {
	return 0, 0
}

func AddBuiltinFrame(functionName string) {
	GlobalCallStack.Push(fmt.Sprintf("builtin:%s", functionName), 0, 0, "<builtin>", "")
}

func RemoveBuiltinFrame() {
	GlobalCallStack.Pop()
}

func SafeEval(node ast.Node, env *object.Environment, functionName string) (result object.Object) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Panic in %s: %v", functionName, r)
			result = NewRuntimeError(message)
		}
	}()

	line, column := ExtractLineInfo(node)
	return WithStackFrame(functionName, line, column, "", "", func() object.Object {
		return Eval(node, env)
	})
}

type DebugInfo struct {
	CurrentFunction string
	StackDepth      int
	Variables       map[string]object.Object
	Line            int
	Column          int
}

func GetDebugInfo(env *object.Environment) DebugInfo {
	info := DebugInfo{
		StackDepth: GlobalCallStack.Depth(),
		Variables:  make(map[string]object.Object),
	}

	if current := GlobalCallStack.Current(); current != nil {
		info.CurrentFunction = current.FunctionName
		info.Line = current.Line
		info.Column = current.Column
	}

	if env != nil {
		vars := env.GetAll()
		for name, value := range vars {
			info.Variables[name] = value
		}
	}

	return info
}

func FormatDebugInfo(info DebugInfo, useColors bool) string {
	var result strings.Builder

	if useColors {
		result.WriteString(ColorCyan + ColorBold)
	}
	result.WriteString("🔍 Debug Info:\n")
	if useColors {
		result.WriteString(ColorReset)
	}

	result.WriteString(fmt.Sprintf("  Function: %s\n", info.CurrentFunction))
	result.WriteString(fmt.Sprintf("  Location: line %d, column %d\n", info.Line, info.Column))
	result.WriteString(fmt.Sprintf("  Stack Depth: %d\n", info.StackDepth))

	if len(info.Variables) > 0 {
		result.WriteString("  Variables:\n")
		for name, value := range info.Variables {
			if useColors {
				result.WriteString(ColorGreen)
			}
			result.WriteString(fmt.Sprintf("    %s", name))
			if useColors {
				result.WriteString(ColorReset)
			}
			result.WriteString(fmt.Sprintf(" = %s (%s)\n", value.Inspect(), value.Type()))
		}
	}

	return result.String()
}
