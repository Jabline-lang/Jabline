package object

import (
	"fmt"
	"jabline/pkg/code"
)

type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int
	NumParameters int
	SourceMap     code.SourceMap
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}
