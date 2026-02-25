package object

import (
	"fmt"
)

type Closure struct {
	Fn       *CompiledFunction
	Free     []Object
	Globals  []Object // Capture globals context
	Constants []Object // Capture constants context
}

func (c *Closure) Type() ObjectType { return CLOSURE_OBJ }
func (c *Closure) Inspect() string {
	return fmt.Sprintf("Closure[%p]", c)
}
