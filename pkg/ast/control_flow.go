package ast

import "jabline/pkg/token"

// IfExpression represents if-else conditional expressions
type IfExpression struct {
	Token       token.Token // 'if'
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	out := "if" + ie.Condition.String() + " " + ie.Consequence.String()
	if ie.Alternative != nil {
		out += "else " + ie.Alternative.String()
	}
	return out
}

// WhileStatement represents while loop statements
type WhileStatement struct {
	Token     token.Token // 'while'
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	return "while " + ws.Condition.String() + " " + ws.Body.String()
}

// ForStatement represents for loop statements
type ForStatement struct {
	Token     token.Token // 'for'
	Init      Statement   // inicialización: i = 0
	Condition Expression  // condición: i < 10
	Update    Statement   // actualización: i = i + 1
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

// ForEachStatement represents foreach loop statements
type ForEachStatement struct {
	Token    token.Token // token 'for'
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

// TryStatement represents try-catch statements
type TryStatement struct {
	Token      token.Token // 'try'
	TryBlock   *BlockStatement
	CatchBlock *BlockStatement
	CatchParam *Identifier // parameter for caught exception
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

// ThrowStatement represents throw statements
type ThrowStatement struct {
	Token token.Token // 'throw'
	Value Expression
}

func (ts *ThrowStatement) statementNode()       {}
func (ts *ThrowStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *ThrowStatement) String() string {
	return "throw " + ts.Value.String()
}

// SwitchStatement represents switch statements
type SwitchStatement struct {
	Token       token.Token // 'switch'
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

// CaseClause represents individual case clauses in switch statements
type CaseClause struct {
	Token      token.Token // 'case'
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

// DefaultClause represents the default clause in switch statements
type DefaultClause struct {
	Token      token.Token // 'default'
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
