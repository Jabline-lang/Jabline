package ast

// Node represents any node in the AST
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents any statement node in the AST
type Statement interface {
	Node
	statementNode()
}

// Expression represents any expression node in the AST
type Expression interface {
	Node
	expressionNode()
}

// Program represents the root of the AST, containing all statements
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	out := ""
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}
