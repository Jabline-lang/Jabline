package object

import "fmt"

type Channel struct {
	Value chan Object
}

func (c *Channel) Type() ObjectType { return CHANNEL_OBJ }
func (c *Channel) Inspect() string  { return fmt.Sprintf("Channel[%p]", c.Value) }
