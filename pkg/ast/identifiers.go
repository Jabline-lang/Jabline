package ast

import "jabline/pkg/token"

type Identifier struct {
	Token token.Token
	Value string
	Type  *TypeExpression
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string {
	if i.Type != nil {
		return i.Value + ": " + i.Type.String()
	}
	return i.Value
}
