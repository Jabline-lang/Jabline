package object

import (
	"context"
	"fmt"
)

type Channel struct {
	Value  chan Object
	Cancel context.CancelFunc
}

func (c *Channel) Type() ObjectType { return CHANNEL_OBJ }
func (c *Channel) Inspect() string  { return fmt.Sprintf("Channel[%p]", c.Value) }
