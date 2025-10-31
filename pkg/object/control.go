package object

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Break struct{}

func (b *Break) Type() ObjectType { return BREAK_OBJ }
func (b *Break) Inspect() string  { return "break" }

type Continue struct{}

func (c *Continue) Type() ObjectType { return CONTINUE_OBJ }
func (c *Continue) Inspect() string  { return "continue" }

type Exception struct {
	Message string
	Value   Object
}

func (e *Exception) Type() ObjectType { return EXCEPTION_OBJ }
func (e *Exception) Inspect() string {
	if e.Value != nil {
		return "EXCEPTION: " + e.Value.Inspect()
	}
	return "EXCEPTION: " + e.Message
}
