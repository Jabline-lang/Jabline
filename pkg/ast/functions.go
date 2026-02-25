package ast

import (
	"strings"

	"jabline/pkg/token"
)

type FunctionLiteral struct {
	Token          token.Token
	TypeParameters []*Identifier
	Parameters     []*Identifier
	ReturnType     *TypeExpression
	Body           *BlockStatement
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
	if len(fl.TypeParameters) > 0 {
		out.WriteString("[")
		tparams := []string{}
		for _, tp := range fl.TypeParameters {
			tparams = append(tparams, tp.String())
		}
		out.WriteString(strings.Join(tparams, ", "))
		out.WriteString("]")
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if fl.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(fl.ReturnType.String())
	}
	out.WriteString(" ")
	out.WriteString(fl.Body.String())
	return out.String()
}

type CallExpression struct {
	Token         token.Token
	Function      Expression
	TypeArguments []*TypeExpression
	Arguments     []Expression
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
	if len(ce.TypeArguments) > 0 {
		out.WriteString("[")
		targs := []string{}
		for _, ta := range ce.TypeArguments {
			targs = append(targs, ta.String())
		}
		out.WriteString(strings.Join(targs, ", "))
		out.WriteString("]")
	}
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

type FunctionStatement struct {
	Token          token.Token
	ReceiverName   *Identifier // The 'l' in (l Libro)
	ReceiverType   *Identifier // The 'Libro' in (l Libro)
	Name           *Identifier
	TypeParameters []*Identifier
	Parameters     []*Identifier
	ReturnType     *TypeExpression
	Body           *BlockStatement
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

	if fs.ReceiverName != nil && fs.ReceiverType != nil {
		out.WriteString("(")
		out.WriteString(fs.ReceiverName.String())
		out.WriteString(" ")
		out.WriteString(fs.ReceiverType.String())
		out.WriteString(") ")
	}

	out.WriteString(fs.Name.String())
	if len(fs.TypeParameters) > 0 {
		out.WriteString("[")
		tparams := []string{}
		for _, tp := range fs.TypeParameters {
			tparams = append(tparams, tp.String())
		}
		out.WriteString(strings.Join(tparams, ", "))
		out.WriteString("]")
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if fs.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(fs.ReturnType.String())
	}
	out.WriteString(" ")
	out.WriteString(fs.Body.String())
	return out.String()
}

type ArrowFunction struct {
	Token      token.Token
	Parameters []*Identifier
	ReturnType *TypeExpression
	Body       Expression
}

func (af *ArrowFunction) expressionNode()      {}
func (af *ArrowFunction) TokenLiteral() string { return af.Token.Literal }
func (af *ArrowFunction) String() string {
	var out strings.Builder
	params := []string{}
	for _, p := range af.Parameters {
		params = append(params, p.String())
	}

	if len(af.Parameters) == 1 {
		out.WriteString(af.Parameters[0].String())
	} else {

		out.WriteString("(")
		out.WriteString(strings.Join(params, ", "))
		out.WriteString(")")
	}

	out.WriteString(" => ")
	if af.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(af.ReturnType.String())
		out.WriteString(" ")
	}
	out.WriteString(af.Body.String())
	return out.String()
}

type AsyncFunctionStatement struct {
	Token          token.Token
	Name           *Identifier
	TypeParameters []*Identifier
	Parameters     []*Identifier
	ReturnType     *TypeExpression
	Body           *BlockStatement
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
	if len(afs.TypeParameters) > 0 {
		out.WriteString("[")
		tparams := []string{}
		for _, tp := range afs.TypeParameters {
			tparams = append(tparams, tp.String())
		}
		out.WriteString(strings.Join(tparams, ", "))
		out.WriteString("]")
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if afs.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(afs.ReturnType.String())
	}
	out.WriteString(" ")
	out.WriteString(afs.Body.String())
	return out.String()
}

type AsyncFunctionLiteral struct {
	Token          token.Token
	TypeParameters []*Identifier
	Parameters     []*Identifier
	ReturnType     *TypeExpression
	Body           *BlockStatement
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
	if len(afl.TypeParameters) > 0 {
		out.WriteString("[")
		tparams := []string{}
		for _, tp := range afl.TypeParameters {
			tparams = append(tparams, tp.String())
		}
		out.WriteString(strings.Join(tparams, ", "))
		out.WriteString("]")
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if afl.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(afl.ReturnType.String())
	}
	out.WriteString(" ")
	out.WriteString(afl.Body.String())
	return out.String()
}

type AwaitExpression struct {
	Token token.Token
	Value Expression
}

func (ae *AwaitExpression) expressionNode()      {}
func (ae *AwaitExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AwaitExpression) String() string {
	return "await " + ae.Value.String()
}
