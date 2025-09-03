package ast

import "jabline/pkg/token"

// ImportStatement represents import statements like "import math" or "import { add, subtract } from 'utils'"
type ImportStatement struct {
	Token      token.Token // 'import'
	ModuleName *StringLiteral
	ImportList []*Identifier // for selective imports
	Alias      *Identifier   // for aliased imports
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	out := "import "

	if len(is.ImportList) > 0 {
		out += "{ "
		for i, item := range is.ImportList {
			if i > 0 {
				out += ", "
			}
			out += item.String()
		}
		out += " } from "
	}

	if is.Alias != nil {
		out += is.ModuleName.String() + " as " + is.Alias.String()
	} else {
		out += is.ModuleName.String()
	}

	return out
}

// ExportStatement represents export statements like "export let name = 'value'" or "export fn add(a, b) { ... }"
type ExportStatement struct {
	Token     token.Token // 'export'
	Statement Statement   // the statement being exported
}

func (es *ExportStatement) statementNode()       {}
func (es *ExportStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExportStatement) String() string {
	return "export " + es.Statement.String()
}

// Module represents a module with its exports
type Module struct {
	Name    string
	Exports map[string]interface{}
	Path    string
}
