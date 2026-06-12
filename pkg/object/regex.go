package object

import (
	"fmt"
	"regexp"
)

type Regex struct {
	Value   *regexp.Regexp
	Pattern string
}

func (r *Regex) Type() ObjectType { return REGEX_OBJ }
func (r *Regex) Inspect() string {
	return fmt.Sprintf("/%s/", r.Pattern)
}
