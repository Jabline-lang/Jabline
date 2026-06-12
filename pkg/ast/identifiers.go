package ast

import "jabline/pkg/token"

type Identifier struct {
	Token        token.Token
	Value        string
	Type         *TypeExpression
	DefaultValue Expression // nil = required parameter (no default)
	Variadic     bool       // true for ...param in variadic functions
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string {
	if i.Type != nil {
		return i.Value + ": " + i.Type.String()
	}
	return i.Value
}

func (node *Identifier) GetToken() token.Token { return node.Token }
