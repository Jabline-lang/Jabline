package vm

import "fmt"

type RuntimeError struct {
	Message string
	Line    int
	Column  int
	File    string
}

func (e *RuntimeError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("runtime error at %s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("runtime error: %s", e.Message)
}
