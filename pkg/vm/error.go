package vm

import (
	"fmt"
	"strings"
)

type CallFrame struct {
	Function string
	File     string
	Line     int
	Column   int
}

type RuntimeError struct {
	Message    string
	StackTrace []CallFrame
}

func (e *RuntimeError) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n\x1b[31;1mRuntimeError\x1b[0m: %s\n", e.Message))

	if len(e.StackTrace) > 0 {
		sb.WriteString("\nTraceback (most recent call last):\n")
		// Reverse iterate to show the most recent call at the bottom
		for i := len(e.StackTrace) - 1; i >= 0; i-- {
			frame := e.StackTrace[i]
			file := frame.File
			if file == "" {
				file = "<anonymous>"
			}
			fn := frame.Function
			if fn == "" {
				fn = "<main>"
			}

			if frame.Line > 0 {
				sb.WriteString(fmt.Sprintf("  def \x1b[33m%s()\x1b[0m in \x1b[36m%s:%d:%d\x1b[0m\n", fn, file, frame.Line, frame.Column))
			} else {
				sb.WriteString(fmt.Sprintf("  def \x1b[33m%s()\x1b[0m in \x1b[36m%s\x1b[0m\n", fn, file))
			}
		}
	} else {
		sb.WriteString("  [No stack trace available]\n")
	}

	return sb.String()
}
