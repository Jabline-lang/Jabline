package ast

import (
	"strings"

	"jabline/pkg/token"
)

// TypeExpression represents a type annotation like string, int, float, bool
type TypeExpression struct {
	Token token.Token // token for the type
	Value string      // the type name: "string", "int", "float", "bool"
}

func (te *TypeExpression) expressionNode()      {}
func (te *TypeExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TypeExpression) String() string       { return te.Value }

// StructStatement represents struct declarations like struct Person { name: string, age: int }
type StructStatement struct {
	Token  token.Token // token 'struct'
	Name   *Identifier
	Fields map[string]*TypeExpression // nombre del campo -> tipo
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

// StructLiteral represents struct instantiation like Person { name: "Alice", age: 30 }
type StructLiteral struct {
	Token  token.Token // token '{'
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
