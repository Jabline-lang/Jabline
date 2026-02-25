package ast

import (
	"fmt"

	"jabline/pkg/token"
)

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Type  *TypeExpression
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	typeStr := ""
	if ls.Type != nil {
		typeStr = ": " + ls.Type.String()
	}
	if ls.Value != nil {
		return fmt.Sprintf("%s%s = %s;", ls.Name.String(), typeStr, ls.Value.String())
	}
	return fmt.Sprintf("%s%s = ;", ls.Name.String(), typeStr)
}

type ConstStatement struct {
	Token token.Token
	Name  *Identifier
	Type  *TypeExpression
	Value Expression
}

func (cs *ConstStatement) statementNode()       {}
func (cs *ConstStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ConstStatement) String() string {
	typeStr := ""
	if cs.Type != nil {
		typeStr = ": " + cs.Type.String()
	}
	if cs.Value != nil {
		return fmt.Sprintf("const %s%s = %s;", cs.Name.String(), typeStr, cs.Value.String())
	}
	return fmt.Sprintf("const %s%s = ;", cs.Name.String(), typeStr)
}

type EchoStatement struct {
	Token  token.Token
	Values []Expression
}

func (es *EchoStatement) statementNode()       {}
func (es *EchoStatement) TokenLiteral() string { return es.Token.Literal }
func (es *EchoStatement) String() string {
	out := "echo("
	for i, v := range es.Values {
		out += v.String()
		if i < len(es.Values)-1 {
			out += ", "
		}
	}
	out += ");"
	return out
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	out := ""
	for _, s := range bs.Statements {
		out += s.String()
	}
	return out
}

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	if rs.ReturnValue != nil {
		return fmt.Sprintf("%s %s;", rs.TokenLiteral(), rs.ReturnValue.String())
	}
	return fmt.Sprintf("%s;", rs.TokenLiteral())
}

type AssignmentStatement struct {
	Token token.Token
	Left  Expression
	Value Expression
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignmentStatement) String() string {
	return fmt.Sprintf("%s = %s;", as.Left.String(), as.Value.String())
}

type BreakStatement struct {
	Token token.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break;" }

type ContinueStatement struct {
	Token token.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue;" }
