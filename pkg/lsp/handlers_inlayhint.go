package lsp

import (
	"fmt"
	"jabline/pkg/ast"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentInlayHint(context *glsp.Context, params *InlayHintParams) ([]InlayHint, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil || docInfo.SymbolTable == nil {
		return nil, nil
	}

	var hints []InlayHint

	var walk func(node ast.Node, scope *Scope)
	walk = func(node ast.Node, scope *Scope) {
		if node == nil {
			return
		}
		switch n := node.(type) {
		case *ast.LetStatement:
			// Type hint: "let x = 42" → ": int" after "x"
			if n.Name != nil && n.Value != nil {
				if scope != nil {
					sym := scope.Get(n.Name.Value)
					if sym != nil && sym.Type != "" && sym.Type != "any" {
						// Position after the identifier
						col := n.Name.Token.Column + len(n.Name.Token.Literal)
						paddingLeft := true
						hints = append(hints, InlayHint{
							Position: protocol.Position{
								Line:      uint32(n.Name.Token.Line - 1),
								Character: uint32(col),
							},
							Label:        fmt.Sprintf(": %s", sym.Type),
							Kind:         (*InlayHintKind)(&[]InlayHintKind{InlayHintKindType}[0]),
							PaddingLeft:  &paddingLeft,
						})
					}
				}
			}
			if n.Value != nil {
				walk(n.Value, scope)
			}

		case *ast.ConstStatement:
			if n.Name != nil && n.Value != nil {
				if scope != nil {
					sym := scope.Get(n.Name.Value)
					if sym != nil && sym.Type != "" && sym.Type != "any" {
						col := n.Name.Token.Column + len(n.Name.Token.Literal)
						paddingLeft := true
						hints = append(hints, InlayHint{
							Position: protocol.Position{
								Line:      uint32(n.Name.Token.Line - 1),
								Character: uint32(col),
							},
							Label:        fmt.Sprintf(": %s", sym.Type),
							Kind:         (*InlayHintKind)(&[]InlayHintKind{InlayHintKindType}[0]),
							PaddingLeft:  &paddingLeft,
						})
					}
				}
			}
			if n.Value != nil {
				walk(n.Value, scope)
			}

		case *ast.CallExpression:
			// Parameter name hints: "add(x, 7)" → "a: " before x, "b: " before 7
			if fnIdent, ok2 := n.Function.(*ast.Identifier); ok2 {
				if scope != nil {
					sym := scope.Get(fnIdent.Value)
					if sym != nil && sym.Definition != nil {
						var paramNames []string
						switch d := sym.Definition.(type) {
						case *ast.FunctionStatement:
							for _, p := range d.Parameters {
								paramNames = append(paramNames, p.Value)
							}
						case *ast.FunctionLiteral:
							for _, p := range d.Parameters {
								paramNames = append(paramNames, p.Value)
							}
						}
						if len(paramNames) > 0 {
							for i, arg := range n.Arguments {
								if i >= len(paramNames) {
									break
								}
								if isSimpleArg(arg) {
									argTok := getArgToken(arg)
									if argTok != nil {
										col := uint32(argTok.Column - 1)
										paddingRight := true
										hints = append(hints, InlayHint{
											Position: protocol.Position{
												Line:  uint32(argTok.Line - 1),
												Character: col,
											},
											Label:        fmt.Sprintf("%s: ", paramNames[i]),
											Kind:         (*InlayHintKind)(&[]InlayHintKind{InlayHintKindParameter}[0]),
											PaddingRight: &paddingRight,
										})
									}
								}
							}
						}
					}
				}
			}
			for _, arg := range n.Arguments {
				walk(arg, scope)
			}

		case *ast.Program:
			for _, stmt := range n.Statements {
				walk(stmt, scope)
			}

		case *ast.ExpressionStatement:
			walk(n.Expression, scope)

		case *ast.BlockStatement:
			for _, stmt := range n.Statements {
				walk(stmt, scope)
			}

		case *ast.FunctionStatement:
			walk(n.Body, scope)

		case *ast.AsyncFunctionStatement:
			walk(n.Body, scope)

		case *ast.WhileStatement:
			walk(n.Body, scope)

		case *ast.ForStatement:
			walk(n.Body, scope)

		case *ast.ForEachStatement:
			walk(n.Body, scope)

		case *ast.IfExpression:
			if n.Consequence != nil {
				walk(n.Consequence, scope)
			}
			if n.Alternative != nil {
				walk(n.Alternative, scope)
			}

		case *ast.ReturnStatement:
			if n.ReturnValue != nil {
				walk(n.ReturnValue, scope)
			}

		case *ast.InfixExpression:
			walk(n.Left, scope)
			walk(n.Right, scope)

		case *ast.AssignmentStatement:
			walk(n.Left, scope)
			walk(n.Value, scope)

		case *ast.SwitchStatement:
			for _, c := range n.Cases {
				for _, s := range c.Statements {
					walk(s, scope)
				}
			}

		case *ast.MatchStatement:
			for _, c := range n.Cases {
				for _, s := range c.Statements {
					walk(s, scope)
				}
			}

		case *ast.TryStatement:
			if n.TryBlock != nil {
				walk(n.TryBlock, scope)
			}
			if n.CatchBlock != nil {
				walk(n.CatchBlock, scope)
			}

		case *ast.AsyncFunctionLiteral:
			walk(n.Body, scope)

		case *ast.FunctionLiteral:
			walk(n.Body, scope)

		case *ast.StructLiteral:
			for _, v := range n.Fields {
				walk(v, scope)
			}
		}
	}

	walk(docInfo.Program, docInfo.SymbolTable.RootScope)
	return hints, nil
}

func isSimpleArg(node ast.Node) bool {
	switch node.(type) {
	case *ast.Identifier, *ast.IntegerLiteral, *ast.FloatLiteral,
		*ast.StringLiteral, *ast.Boolean, *ast.Null:
		return true
	}
	return false
}

func getArgToken(node ast.Node) *struct {
	Line   int
	Column int
	Literal string
} {
	switch n := node.(type) {
	case *ast.Identifier:
		return &struct {
			Line    int
			Column  int
			Literal string
		}{n.Token.Line, n.Token.Column, n.Token.Literal}
	case *ast.IntegerLiteral:
		return &struct {
			Line    int
			Column  int
			Literal string
		}{n.Token.Line, n.Token.Column, n.Token.Literal}
	case *ast.FloatLiteral:
		return &struct {
			Line    int
			Column  int
			Literal string
		}{n.Token.Line, n.Token.Column, n.Token.Literal}
	case *ast.StringLiteral:
		return &struct {
			Line    int
			Column  int
			Literal string
		}{n.Token.Line, n.Token.Column, n.Token.Literal}
	case *ast.Boolean:
		return &struct {
			Line    int
			Column  int
			Literal string
		}{n.Token.Line, n.Token.Column, n.Token.Literal}
	case *ast.Null:
		return &struct {
			Line    int
			Column  int
			Literal string
		}{n.Token.Line, n.Token.Column, n.Token.Literal}
	}
	return nil
}
