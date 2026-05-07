package ast

import (
	"strings"

	"jabline/pkg/token"
)

type TypeExpression struct {
	Token     token.Token
	Value     string
	Arguments []*TypeExpression // Para Genéricos: Array<int> -> Base: "Array", Arguments: ["int"]
}

func (te *TypeExpression) expressionNode()      {}
func (te *TypeExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TypeExpression) String() string {
	if len(te.Arguments) == 0 {
		return te.Value
	}
	var out strings.Builder
	out.WriteString(te.Value)
	out.WriteString("<")
	args := []string{}
	for _, arg := range te.Arguments {
		args = append(args, arg.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(">")
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
		out.WriteString("<")
		params := []string{}
		for _, p := range ss.TypeParameters {
			params = append(params, p.String())
		}
		out.WriteString(strings.Join(params, ", "))
		out.WriteString(">")
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

type FunctionSignature struct {
	Token      token.Token
	Name       string
	Parameters []*Identifier
	ReturnType *TypeExpression
}

func (fs *FunctionSignature) String() string {
	var out strings.Builder
	out.WriteString("(")
	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if fs.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(fs.ReturnType.String())
	}
	return out.String()
}

type InterfaceStatement struct {
	Token          token.Token
	Name           *Identifier
	TypeParameters []*Identifier
	Methods        map[string]*FunctionSignature
}

func (is *InterfaceStatement) statementNode()       {}
func (is *InterfaceStatement) TokenLiteral() string { return is.Token.Literal }
func (is *InterfaceStatement) String() string {
	var out strings.Builder
	out.WriteString(is.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(is.Name.String())
	if len(is.TypeParameters) > 0 {
		out.WriteString("<")
		params := []string{}
		for _, p := range is.TypeParameters {
			params = append(params, p.String())
		}
		out.WriteString(strings.Join(params, ", "))
		out.WriteString(">")
	}
	out.WriteString(" { ")

	methods := []string{}
	for name, sig := range is.Methods {
		methods = append(methods, name+sig.String())
	}
	out.WriteString(strings.Join(methods, ", "))
	out.WriteString(" }")
	return out.String()
}

func (node *TypeExpression) GetToken() token.Token { return node.Token }

func (node *StructStatement) GetToken() token.Token { return node.Token }

func (node *StructLiteral) GetToken() token.Token { return node.Token }

func (node *FunctionSignature) GetToken() token.Token { return node.Token }

func (node *InterfaceStatement) GetToken() token.Token { return node.Token }
