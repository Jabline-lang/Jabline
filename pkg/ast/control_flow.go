package ast

import (
	"jabline/pkg/token"
)

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	out := "if (" + ie.Condition.String() + ") " + ie.Consequence.String()
	if ie.Alternative != nil {

		altBlock := ie.Alternative

		if len(altBlock.Statements) == 1 {
			if exprStmt, ok := altBlock.Statements[0].(*ExpressionStatement); ok {
				if _, ok := exprStmt.Expression.(*IfExpression); ok {

					out += " else " + exprStmt.String()
					return out
				}
			}
		}

		out += " else " + ie.Alternative.String()
	}

	return out
}

type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	return "while " + ws.Condition.String() + " " + ws.Body.String()
}

type ForStatement struct {
	Token     token.Token
	Init      Statement
	Condition Expression
	Update    Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	out := "for ("
	if fs.Init != nil {
		out += fs.Init.String()
	}
	out += "; "
	if fs.Condition != nil {
		out += fs.Condition.String()
	}
	out += "; "
	if fs.Update != nil {
		out += fs.Update.String()
	}
	out += ") " + fs.Body.String()
	return out
}

type ForEachStatement struct {
	Token    token.Token
	Variable *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (fes *ForEachStatement) statementNode()       {}
func (fes *ForEachStatement) TokenLiteral() string { return fes.Token.Literal }
func (fes *ForEachStatement) String() string {
	out := "for (" + fes.Variable.String() + " in " + fes.Iterable.String() + ") "
	out += fes.Body.String()
	return out
}

type TryStatement struct {
	Token      token.Token
	TryBlock   *BlockStatement
	CatchBlock *BlockStatement
	CatchParam *Identifier
}

func (ts *TryStatement) statementNode()       {}
func (ts *TryStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TryStatement) String() string {
	out := "try " + ts.TryBlock.String()
	if ts.CatchBlock != nil {
		out += " catch"
		if ts.CatchParam != nil {
			out += "(" + ts.CatchParam.String() + ")"
		}
		out += " " + ts.CatchBlock.String()
	}
	return out
}

type ThrowStatement struct {
	Token token.Token
	Value Expression
}

func (ts *ThrowStatement) statementNode()       {}
func (ts *ThrowStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *ThrowStatement) String() string {
	return "throw " + ts.Value.String()
}

type SwitchStatement struct {
	Token       token.Token
	Expression  Expression
	Cases       []*CaseClause
	DefaultCase *DefaultClause
}

func (ss *SwitchStatement) statementNode()       {}
func (ss *SwitchStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SwitchStatement) String() string {
	out := "switch (" + ss.Expression.String() + ") {"
	for _, c := range ss.Cases {
		out += c.String()
	}
	if ss.DefaultCase != nil {
		out += ss.DefaultCase.String()
	}
	out += "}"
	return out
}

type CaseClause struct {
	Token      token.Token
	Value      Expression
	Statements []Statement
}

func (cc *CaseClause) statementNode()       {}
func (cc *CaseClause) TokenLiteral() string { return cc.Token.Literal }
func (cc *CaseClause) String() string {
	out := "case " + cc.Value.String() + ":"
	for _, stmt := range cc.Statements {
		out += stmt.String()
	}
	return out
}

type DefaultClause struct {
	Token      token.Token
	Statements []Statement
}

func (dc *DefaultClause) statementNode()       {}
func (dc *DefaultClause) TokenLiteral() string { return dc.Token.Literal }
func (dc *DefaultClause) String() string {
	out := "default:"
	for _, stmt := range dc.Statements {
		out += stmt.String()
	}
	return out
}
