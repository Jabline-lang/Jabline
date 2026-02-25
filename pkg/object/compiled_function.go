package object

import (
	"fmt"
	"jabline/pkg/code"
)

type CompiledFunction struct {
	Instructions   code.Instructions
	NumLocals      int
	NumParameters  int
	SourceMap      code.SourceMap
	IsAsync        bool
	Name           string
	TypeParameters []string
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	name := cf.Name
	if name == "" {
		name = "<anonymous>"
	}
	return fmt.Sprintf("CompiledFunction[%p, %s]", cf, name)
}
