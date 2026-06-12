package ast

import (
	"jabline/pkg/token"
)

type RetryStatement struct {
	Token      token.Token
	Attempts   Expression
	RetryBlock *BlockStatement
	CatchBlock *BlockStatement
	CatchParam *Identifier
}

func (rs *RetryStatement) statementNode()       {}
func (rs *RetryStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RetryStatement) String() string {
	out := "retry (" + rs.Attempts.String() + ") " + rs.RetryBlock.String()
	if rs.CatchBlock != nil {
		out += " catch"
		if rs.CatchParam != nil {
			out += "(" + rs.CatchParam.String() + ")"
		}
		out += " " + rs.CatchBlock.String()
	}
	return out
}

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

type DoWhileStatement struct {
	Token     token.Token
	Body      *BlockStatement
	Condition Expression
}

func (ds *DoWhileStatement) statementNode()       {}
func (ds *DoWhileStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DoWhileStatement) String() string {
	return "do " + ds.Body.String() + " while (" + ds.Condition.String() + ")"
}

type DeferStatement struct {
	Token     token.Token
	Call      Expression
}

func (ds *DeferStatement) statementNode()       {}
func (ds *DeferStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DeferStatement) String() string {
	return "defer " + ds.Call.String()
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
	CatchType  *TypeExpression
	Finally    *BlockStatement
}

func (ts *TryStatement) statementNode()       {}
func (ts *TryStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TryStatement) String() string {
	out := "try " + ts.TryBlock.String()
	if ts.CatchBlock != nil {
		out += " catch"
		if ts.CatchParam != nil {
			out += "(" + ts.CatchParam.String()
			if ts.CatchType != nil {
				out += ": " + ts.CatchType.String()
			}
			out += ")"
		}
		out += " " + ts.CatchBlock.String()
	}
	if ts.Finally != nil {
		out += " finally " + ts.Finally.String()
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

type MatchStatement struct {
	Token      token.Token
	Expression Expression
	Cases      []*MatchCase
}

func (ms *MatchStatement) statementNode()       {}
func (ms *MatchStatement) TokenLiteral() string { return ms.Token.Literal }
func (ms *MatchStatement) String() string {
	out := "match (" + ms.Expression.String() + ") {"
	for _, c := range ms.Cases {
		out += c.String()
	}
	out += "}"
	return out
}

type MatchCase struct {
	Token      token.Token
	Pattern    Expression // Can be literal, identifier (type), or array (structural)
	IsDefault  bool
	Statements []Statement
}

func (mc *MatchCase) statementNode()       {}
func (mc *MatchCase) TokenLiteral() string { return mc.Token.Literal }
func (mc *MatchCase) String() string {
	out := ""
	if mc.IsDefault {
		out += "default"
	} else {
		out += "case " + mc.Pattern.String()
	}
	out += ":"
	for _, stmt := range mc.Statements {
	out += stmt.String()
	}
	return out
}

type SelectStatement struct {
	Token       token.Token
	Cases       []*SelectCase
	DefaultCase *DefaultClause
}

func (ss *SelectStatement) statementNode()       {}
func (ss *SelectStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SelectStatement) String() string {
	out := "select {"
	for _, c := range ss.Cases {
		out += c.String()
	}
	if ss.DefaultCase != nil {
		out += ss.DefaultCase.String()
	}
	out += "}"
	return out
}

type SelectCase struct {
	Token  token.Token
	IsSend bool // true for send (ch <- val), false for receive (<-ch)
	// For send: Channel is the expression before <-, Value is the expression after <-
	// For recv: Channel is the expression after <-
	// BindingName is set for case x := <-ch
	Channel      Expression
	Value        Expression // only for send
	BindingName  string     // only for recv bindings
	BindingIsNew bool       // true for := (new binding), false for = (assignment)
	Statements   []Statement
}

func (sc *SelectCase) statementNode()       {}
func (sc *SelectCase) TokenLiteral() string { return sc.Token.Literal }
func (sc *SelectCase) String() string {
	out := "case "
	if sc.IsSend {
		out += sc.Channel.String() + " <- " + sc.Value.String()
	} else {
		if sc.BindingName != "" {
			if sc.BindingIsNew {
				out += sc.BindingName + " := "
			} else {
				out += sc.BindingName + " = "
			}
		}
		out += "<- " + sc.Channel.String()
	}
	out += ":"
	for _, stmt := range sc.Statements {
		out += stmt.String()
	}
	return out
}

func (node *RetryStatement) GetToken() token.Token { return node.Token }
func (node *IfExpression) GetToken() token.Token { return node.Token }
func (node *DoWhileStatement) GetToken() token.Token { return node.Token }
func (node *DeferStatement) GetToken() token.Token { return node.Token }
func (node *WhileStatement) GetToken() token.Token { return node.Token }
func (node *ForStatement) GetToken() token.Token { return node.Token }
func (node *ForEachStatement) GetToken() token.Token { return node.Token }
func (node *TryStatement) GetToken() token.Token { return node.Token }
func (node *ThrowStatement) GetToken() token.Token { return node.Token }
func (node *SwitchStatement) GetToken() token.Token { return node.Token }
func (node *CaseClause) GetToken() token.Token { return node.Token }
func (node *DefaultClause) GetToken() token.Token { return node.Token }
func (node *MatchStatement) GetToken() token.Token { return node.Token }
func (node *MatchCase) GetToken() token.Token { return node.Token }
func (node *SelectStatement) GetToken() token.Token { return node.Token }
func (node *SelectCase) GetToken() token.Token { return node.Token }
