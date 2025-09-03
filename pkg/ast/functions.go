package ast

import (
	"strings"

	"jabline/pkg/token"
)

// FunctionLiteral represents function literals like fn(x, y) { return x + y; }
type FunctionLiteral struct {
	Token      token.Token // token 'fn'
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out strings.Builder
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// CallExpression represents function calls like add(1, 2)
type CallExpression struct {
	Token     token.Token // token '('
	Function  Expression  // Identifier o FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out strings.Builder
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// FunctionStatement represents function declarations like fn add(x, y) { return x + y; }
type FunctionStatement struct {
	Token      token.Token // token 'fn'
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var out strings.Builder
	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fs.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(fs.Name.String())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// ArrowFunction represents arrow function literals like (x, y) => x + y or x => x * 2
type ArrowFunction struct {
	Token      token.Token // token '(' or identifier for single param
	Parameters []*Identifier
	Body       Expression // Arrow functions have expression bodies
}

func (af *ArrowFunction) expressionNode()      {}
func (af *ArrowFunction) TokenLiteral() string { return af.Token.Literal }
func (af *ArrowFunction) String() string {
	var out strings.Builder
	params := []string{}
	for _, p := range af.Parameters {
		params = append(params, p.String())
	}

	// Single parameter without parentheses
	if len(af.Parameters) == 1 {
		out.WriteString(af.Parameters[0].String())
	} else {
		// Multiple or zero parameters with parentheses
		out.WriteString("(")
		out.WriteString(strings.Join(params, ", "))
		out.WriteString(")")
	}

	out.WriteString(" => ")
	out.WriteString(af.Body.String())
	return out.String()
}

// AsyncFunctionStatement represents async function declarations like async fn getData() { ... }
type AsyncFunctionStatement struct {
	Token      token.Token // token 'async'
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (afs *AsyncFunctionStatement) statementNode()       {}
func (afs *AsyncFunctionStatement) TokenLiteral() string { return afs.Token.Literal }
func (afs *AsyncFunctionStatement) String() string {
	var out strings.Builder
	params := []string{}
	for _, p := range afs.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(afs.TokenLiteral())
	out.WriteString(" fn ")
	out.WriteString(afs.Name.String())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(afs.Body.String())
	return out.String()
}

// AsyncFunctionLiteral represents async function literals like async fn(x, y) { ... }
type AsyncFunctionLiteral struct {
	Token      token.Token // token 'async'
	Parameters []*Identifier
	Body       *BlockStatement
}

func (afl *AsyncFunctionLiteral) expressionNode()      {}
func (afl *AsyncFunctionLiteral) TokenLiteral() string { return afl.Token.Literal }
func (afl *AsyncFunctionLiteral) String() string {
	var out strings.Builder
	params := []string{}
	for _, p := range afl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(afl.TokenLiteral())
	out.WriteString(" fn(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(afl.Body.String())
	return out.String()
}

// AwaitExpression represents await expressions like await somePromise
type AwaitExpression struct {
	Token token.Token // token 'await'
	Value Expression  // the expression to await
}

func (ae *AwaitExpression) expressionNode()      {}
func (ae *AwaitExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AwaitExpression) String() string {
	return "await " + ae.Value.String()
}
