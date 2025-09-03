package ast

import "jabline/pkg/token"

// PrefixExpression represents prefix expressions like !true, -5
type PrefixExpression struct {
	Token    token.Token // operador (!,-)
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// InfixExpression represents binary expressions like 5 + 5, x == y
type InfixExpression struct {
	Token    token.Token // operador (+, -, *, /, ==, !=, <, >)
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// PostfixExpression represents postfix expressions like x++, y--
type PostfixExpression struct {
	Token    token.Token // token '++' o '--'
	Left     Expression  // la variable
	Operator string
}

func (pe *PostfixExpression) expressionNode()      {}
func (pe *PostfixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PostfixExpression) String() string {
	return "(" + pe.Left.String() + pe.Operator + ")"
}

// ArrayIndexExpression represents array indexing like arr[0], hash["key"]
type ArrayIndexExpression struct {
	Token token.Token // token '['
	Left  Expression  // el array
	Index Expression  // el índice
}

func (aie *ArrayIndexExpression) expressionNode()      {}
func (aie *ArrayIndexExpression) TokenLiteral() string { return aie.Token.Literal }
func (aie *ArrayIndexExpression) String() string {
	return "(" + aie.Left.String() + "[" + aie.Index.String() + "])"
}

// IndexExpression represents field access like obj.field
type IndexExpression struct {
	Token token.Token // token '.'
	Left  Expression  // la estructura
	Index Expression  // el nombre del campo
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + "." + ie.Index.String() + ")"
}

// TernaryExpression represents ternary conditional expressions like condition ? trueValue : falseValue
type TernaryExpression struct {
	Token      token.Token // token '?'
	Condition  Expression  // la condición
	TrueValue  Expression  // valor si true
	FalseValue Expression  // valor si false
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TernaryExpression) String() string {
	return "(" + te.Condition.String() + " ? " + te.TrueValue.String() + " : " + te.FalseValue.String() + ")"
}

// NullishCoalescingExpression represents nullish coalescing expressions like a ?? b
type NullishCoalescingExpression struct {
	Token token.Token // token '??'
	Left  Expression  // left operand
	Right Expression  // right operand (evaluated only if left is null)
}

func (nce *NullishCoalescingExpression) expressionNode()      {}
func (nce *NullishCoalescingExpression) TokenLiteral() string { return nce.Token.Literal }
func (nce *NullishCoalescingExpression) String() string {
	return "(" + nce.Left.String() + " ?? " + nce.Right.String() + ")"
}

// OptionalChainingExpression represents optional chaining expressions like obj?.prop
type OptionalChainingExpression struct {
	Token token.Token // token '?.'
	Left  Expression  // the object
	Right Expression  // the property/method
}

func (oce *OptionalChainingExpression) expressionNode()      {}
func (oce *OptionalChainingExpression) TokenLiteral() string { return oce.Token.Literal }
func (oce *OptionalChainingExpression) String() string {
	return "(" + oce.Left.String() + "?." + oce.Right.String() + ")"
}
