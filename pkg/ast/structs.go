package ast

import (
	"strings"

	"jabline/pkg/token"
)

type TypeExpression struct {
	Token token.Token
	Value string
}

func (te *TypeExpression) expressionNode()      {}
func (te *TypeExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TypeExpression) String() string       { return te.Value }

type StructStatement struct {
	Token  token.Token
	Name   *Identifier
	Fields map[string]*TypeExpression
}

func (ss *StructStatement) statementNode()       {}
func (ss *StructStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *StructStatement) String() string {
	var out strings.Builder
	out.WriteString(ss.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(ss.Name.String())
	out.WriteString(" { ")

	fields := []string{}
	for name, typeExpr := range ss.Fields {
		fields = append(fields, name+": "+typeExpr.String())
	}
	out.WriteString(strings.Join(fields, ", "))
	out.WriteString(" }")
	return out.String()
}

type StructLiteral struct {
	Token  token.Token
	Name   *Identifier
	Fields map[string]Expression
}

func (sl *StructLiteral) expressionNode()      {}
func (sl *StructLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StructLiteral) String() string {
	var out strings.Builder
	out.WriteString(sl.Name.String())
	out.WriteString(" { ")

	fields := []string{}
	for name, value := range sl.Fields {
		fields = append(fields, name+": "+value.String())
	}
	out.WriteString(strings.Join(fields, ", "))
	out.WriteString(" }")
	return out.String()
}
