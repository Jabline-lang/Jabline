package ast

import (
	"jabline/pkg/token"
	"strings"
)

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

type PostfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
}

func (pe *PostfixExpression) expressionNode()      {}
func (pe *PostfixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PostfixExpression) String() string {
	return "(" + pe.Left.String() + pe.Operator + ")"
}

type ArrayIndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (aie *ArrayIndexExpression) expressionNode()      {}
func (aie *ArrayIndexExpression) TokenLiteral() string { return aie.Token.Literal }
func (aie *ArrayIndexExpression) String() string {
	return "(" + aie.Left.String() + "[" + aie.Index.String() + "])"
}

type SliceExpression struct {
	Token token.Token
	Left  Expression
	Low   Expression
	High  Expression
}

func (se *SliceExpression) expressionNode()      {}
func (se *SliceExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SliceExpression) String() string {
	lowStr := ""
	if se.Low != nil {
		lowStr = se.Low.String()
	}
	highStr := ""
	if se.High != nil {
		highStr = se.High.String()
	}
	return "(" + se.Left.String() + "[" + lowStr + ":" + highStr + "])"
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + "." + ie.Index.String() + ")"
}

type TernaryExpression struct {
	Token      token.Token
	Condition  Expression
	TrueValue  Expression
	FalseValue Expression
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TernaryExpression) String() string {
	return "(" + te.Condition.String() + " ? " + te.TrueValue.String() + " : " + te.FalseValue.String() + ")"
}

type NullishCoalescingExpression struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (nce *NullishCoalescingExpression) expressionNode()      {}
func (nce *NullishCoalescingExpression) TokenLiteral() string { return nce.Token.Literal }
func (nce *NullishCoalescingExpression) String() string {
	return "(" + nce.Left.String() + " ?? " + nce.Right.String() + ")"
}

type OptionalChainingExpression struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (oce *OptionalChainingExpression) expressionNode()      {}
func (oce *OptionalChainingExpression) TokenLiteral() string { return oce.Token.Literal }
func (oce *OptionalChainingExpression) String() string {
	rightStr := oce.Right.String()
	// If right side is a string literal, strip quotes for display
	if sl, ok := oce.Right.(*StringLiteral); ok {
		rightStr = sl.Value
	}
	return "(" + oce.Left.String() + "?." + rightStr + ")"
}

type SpawnExpression struct {
	Token token.Token // The 'spawn' token
	Call  *CallExpression
}

func (se *SpawnExpression) expressionNode()      {}
func (se *SpawnExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SpawnExpression) String() string {
	return "spawn " + se.Call.String()
}

type InstantiatedExpression struct {
	Token         token.Token // The '[' token
	Left          Expression
	TypeArguments []*TypeExpression
}

func (ie *InstantiatedExpression) expressionNode()      {}
func (ie *InstantiatedExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InstantiatedExpression) String() string {
	var out strings.Builder
	out.WriteString(ie.Left.String())
	out.WriteString("<")
	args := []string{}
	for _, arg := range ie.TypeArguments {
		args = append(args, arg.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(">")
	return out.String()
}

func (node *PrefixExpression) GetToken() token.Token { return node.Token }

func (node *InfixExpression) GetToken() token.Token { return node.Token }

func (node *PostfixExpression) GetToken() token.Token { return node.Token }

func (node *ArrayIndexExpression) GetToken() token.Token { return node.Token }

func (node *SliceExpression) GetToken() token.Token { return node.Token }

func (node *IndexExpression) GetToken() token.Token { return node.Token }

func (node *TernaryExpression) GetToken() token.Token { return node.Token }

func (node *NullishCoalescingExpression) GetToken() token.Token { return node.Token }

func (node *OptionalChainingExpression) GetToken() token.Token { return node.Token }

func (node *SpawnExpression) GetToken() token.Token { return node.Token }

func (node *InstantiatedExpression) GetToken() token.Token { return node.Token }

type SpreadExpr struct {
	Token token.Token
	Right Expression
}

func (se *SpreadExpr) expressionNode()      {}
func (se *SpreadExpr) TokenLiteral() string { return se.Token.Literal }
func (se *SpreadExpr) String() string {
	return "..." + se.Right.String()
}
func (node *SpreadExpr) GetToken() token.Token { return node.Token }
