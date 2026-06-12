package object

import (
	"time"
)

type DateTime struct {
	Time time.Time
}

func (dt *DateTime) Type() ObjectType { return DATETIME_OBJ }
func (dt *DateTime) Inspect() string {
	return dt.Time.Format(time.RFC3339)
}
