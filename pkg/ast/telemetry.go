package ast

import (
	"bytes"
	"jabline/pkg/token"
)

type MeterStatement struct {
	Token    token.Token // the 'meter' token
	Name     Expression  // e.g., "requests_total"
	Operator string      // e.g., "++"
}

func (ms *MeterStatement) statementNode()       {}
func (ms *MeterStatement) TokenLiteral() string { return ms.Token.Literal }
func (ms *MeterStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ms.TokenLiteral() + " ")
	out.WriteString(ms.Name.String())
	out.WriteString(ms.Operator + ";")

	return out.String()
}

type TraceStatement struct {
	Token token.Token // the 'trace' token
	Name  Expression  // e.g., "db_query"
	Body  *BlockStatement
}

func (ts *TraceStatement) statementNode()       {}
func (ts *TraceStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TraceStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ts.TokenLiteral() + " ")
	out.WriteString(ts.Name.String() + " ")
	out.WriteString(ts.Body.String())

	return out.String()
}

func (node *MeterStatement) GetToken() token.Token { return node.Token }

func (node *TraceStatement) GetToken() token.Token { return node.Token }
