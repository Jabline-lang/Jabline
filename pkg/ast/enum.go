package ast

import (
	"strings"

	"jabline/pkg/token"
)

// EnumStatement represents: enum Name { Variant1, Variant2, ... }
// Compiles to an immutable Hash constant: { "Variant1": 0, "Variant2": 1, ... }
type EnumStatement struct {
	Token  token.Token   // the 'enum' token
	Name   *Identifier   // enum name
	Values []*Identifier // ordered list of variant names
}

func (es *EnumStatement) statementNode()       {}
func (es *EnumStatement) TokenLiteral() string { return es.Token.Literal }
func (es *EnumStatement) String() string {
	var out strings.Builder
	out.WriteString("enum ")
	out.WriteString(es.Name.String())
	out.WriteString(" { ")
	names := make([]string, len(es.Values))
	for i, v := range es.Values {
		names[i] = v.String()
	}
	out.WriteString(strings.Join(names, ", "))
	out.WriteString(" }")
	return out.String()
}
