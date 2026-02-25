package ast

import (
	"strings"

	"jabline/pkg/token"
)

type TypeExpression struct {
	Token     token.Token
	Value     string
	Arguments []*TypeExpression // Para Genéricos: Array[int] -> Base: "Array", Arguments: ["int"]
}

func (te *TypeExpression) expressionNode()      {}
func (te *TypeExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TypeExpression) String() string {
	if len(te.Arguments) == 0 {
		return te.Value
	}
	var out strings.Builder
	out.WriteString(te.Value)
	out.WriteString("[")
	args := []string{}
	for _, arg := range te.Arguments {
		args = append(args, arg.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString("]")
	return out.String()
}

type StructStatement struct {
	Token          token.Token
	Name           *Identifier
	TypeParameters []*Identifier
	Fields         map[string]*TypeExpression
}

func (ss *StructStatement) statementNode()       {}
func (ss *StructStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *StructStatement) String() string {
	var out strings.Builder
	out.WriteString(ss.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(ss.Name.String())
	if len(ss.TypeParameters) > 0 {
		out.WriteString("[")
		params := []string{}
		for _, p := range ss.TypeParameters {
			params = append(params, p.String())
		}
		out.WriteString(strings.Join(params, ", "))
		out.WriteString("]")
	}
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
	Name   Expression
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
