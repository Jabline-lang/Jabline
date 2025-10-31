package ast

import "jabline/pkg/token"

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
	return "(" + oce.Left.String() + "?." + oce.Right.String() + ")"
}
