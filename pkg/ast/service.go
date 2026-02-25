package ast

import (
	"strings"
	"jabline/pkg/token"
)

type ServiceStatement struct {
	Token   token.Token
	Name    *Identifier
	Fields  map[string]Expression // Configuration fields (e.g., port: 8080)
	Methods []*FunctionStatement  // API Endpoints
}

func (ss *ServiceStatement) statementNode()       {}
func (ss *ServiceStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *ServiceStatement) String() string {
	var out strings.Builder
	out.WriteString("service ")
	out.WriteString(ss.Name.String())
	out.WriteString(" {\n")

	for name, val := range ss.Fields {
		out.WriteString("  " + name + ": " + val.String() + "\n")
	}
	
	for _, method := range ss.Methods {
		out.WriteString(method.String() + "\n")
	}

	out.WriteString("}")
	return out.String()
}
