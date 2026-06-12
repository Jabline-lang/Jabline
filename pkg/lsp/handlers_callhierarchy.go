package lsp

import (
	"jabline/pkg/ast"
	"jabline/pkg/token"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func buildCallHierarchyItem(sym *Symbol, uri string) *protocol.CallHierarchyItem {
	if sym == nil {
		return nil
	}
	dt := declTokenFromSymbol(sym)
	if dt == nil {
		return nil
	}
	start, end := tokToPos(*dt)
	detail := sym.Type
	return &protocol.CallHierarchyItem{
		Name:           sym.Name,
		Kind:           sym.Kind,
		Detail:         &detail,
		URI:            uri,
		Range:          protocol.Range{Start: start, End: end},
		SelectionRange: protocol.Range{Start: start, End: end},
	}
}

func textDocumentPrepareCallHierarchy(context *glsp.Context, params *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	line := int(params.Position.Line) + 1
	col := int(params.Position.Character) + 1

	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil || docInfo.SymbolTable == nil {
		return nil, nil
	}

	path := FindPathToNode(docInfo.Program, line, col)
	if len(path) == 0 {
		return nil, nil
	}

	node := path[len(path)-1]
	ident, ok := node.(*ast.Identifier)
	if !ok {
		return nil, nil
	}

	scope := FindCurrentScope(docInfo, path)
	if scope == nil {
		return nil, nil
	}

	sym := scope.Get(ident.Value)
	if sym == nil || sym.Definition == nil {
		return nil, nil
	}

	switch sym.Definition.(type) {
	case *ast.FunctionStatement, *ast.AsyncFunctionStatement:
		item := buildCallHierarchyItem(sym, params.TextDocument.URI)
		if item == nil {
			return nil, nil
		}
		return []protocol.CallHierarchyItem{*item}, nil
	}

	return nil, nil
}

func callHierarchyIncomingCalls(context *glsp.Context, params *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	targetName := params.Item.Name
	var calls []protocol.CallHierarchyIncomingCall

	workspaceStore.Mutex.RLock()
	defer workspaceStore.Mutex.RUnlock()

	for uri, docInfo := range workspaceStore.Documents {
		if docInfo == nil || docInfo.Program == nil {
			continue
		}

		var walk func(node ast.Node)
		walk = func(node ast.Node) {
			if node == nil {
				return
			}
			switch n := node.(type) {
			case *ast.CallExpression:
				if ident, ok2 := n.Function.(*ast.Identifier); ok2 && ident.Value == targetName {
					callerScope := findEnclosingFunction(docInfo, n)
					if callerScope != nil {
						for _, callerSym := range callerScope.Symbols {
							if callerSym.Definition != nil {
								if _, isFn := callerSym.Definition.(*ast.FunctionStatement); isFn {
									item := buildCallHierarchyItem(callerSym, uri)
									if item != nil {
										start, end := tokToPos(ident.Token)
										calls = append(calls, protocol.CallHierarchyIncomingCall{
											From:       *item,
											FromRanges: []protocol.Range{{Start: start, End: end}},
										})
									}
									break
								}
							}
						}
					}
				}
				for _, arg := range n.Arguments {
					walk(arg)
				}
			case *ast.ExpressionStatement:
				walk(n.Expression)
			case *ast.BlockStatement:
				for _, stmt := range n.Statements {
					walk(stmt)
				}
			case *ast.Program:
				for _, stmt := range n.Statements {
					walk(stmt)
				}
			default:
			}
		}
		walk(docInfo.Program)
	}

	return calls, nil
}

func callHierarchyOutgoingCalls(context *glsp.Context, params *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	targetName := params.Item.Name
	var calls []protocol.CallHierarchyOutgoingCall

	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.Item.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil || docInfo.SymbolTable == nil {
		return nil, nil
	}

	var targetFnBody *ast.BlockStatement
	var walkScope func(s *Scope) *ast.BlockStatement
	walkScope = func(s *Scope) *ast.BlockStatement {
		for _, sym := range s.Symbols {
			if sym.Name == targetName {
				switch d := sym.Definition.(type) {
				case *ast.FunctionStatement:
					return d.Body
				case *ast.AsyncFunctionStatement:
					return d.Body
				}
			}
		}
		for _, child := range s.Children {
			if found := walkScope(child); found != nil {
				return found
			}
		}
		return nil
	}
	targetFnBody = walkScope(docInfo.SymbolTable.RootScope)

	if targetFnBody == nil {
		return nil, nil
	}

	workspaceStore.Mutex.RLock()
	defer workspaceStore.Mutex.RUnlock()

	var walk func(node ast.Node)
	walk = func(node ast.Node) {
		if node == nil {
			return
		}
		switch n := node.(type) {
		case *ast.CallExpression:
			if ident, ok2 := n.Function.(*ast.Identifier); ok2 {
				calleeScope := findSymbolScope(docInfo, ident.Value)
				if calleeScope != nil {
					if calleeSym := calleeScope.Get(ident.Value); calleeSym != nil {
						item := buildCallHierarchyItem(calleeSym, params.Item.URI)
						if item != nil {
							start, end := tokToPos(ident.Token)
							calls = append(calls, protocol.CallHierarchyOutgoingCall{
								To:         *item,
								FromRanges: []protocol.Range{{Start: start, End: end}},
							})
						}
					}
				}
			}
			for _, arg := range n.Arguments {
				walk(arg)
			}
		case *ast.ExpressionStatement:
			walk(n.Expression)
		case *ast.BlockStatement:
			for _, stmt := range n.Statements {
				walk(stmt)
			}
		default:
		}
	}
	walk(targetFnBody)

	return calls, nil
}

func findEnclosingFunction(docInfo *DocumentSemanticInfo, node ast.Node) *Scope {
	if docInfo == nil || docInfo.SymbolTable == nil {
		return nil
	}

	var result *Scope
	var search func(s *Scope, target ast.Node) *Scope
	search = func(s *Scope, target ast.Node) *Scope {
		for _, sym := range s.Symbols {
			if sym.Definition == nil {
				continue
			}
			var fnBody *ast.BlockStatement
			switch d := sym.Definition.(type) {
			case *ast.FunctionStatement:
				fnBody = d.Body
			case *ast.AsyncFunctionStatement:
				fnBody = d.Body
			}
			if fnBody != nil {
				if nodeBelongsToBlock(fnBody, target) {
					return s
				}
			}
		}
		for _, child := range s.Children {
			if found := search(child, target); found != nil {
				return found
			}
		}
		return nil
	}
	result = search(docInfo.SymbolTable.RootScope, node)
	return result
}

func findSymbolScope(docInfo *DocumentSemanticInfo, name string) *Scope {
	if docInfo == nil || docInfo.SymbolTable == nil {
		return nil
	}
	var search func(s *Scope) *Scope
	search = func(s *Scope) *Scope {
		if _, ok := s.Symbols[name]; ok {
			return s
		}
		for _, child := range s.Children {
			if found := search(child); found != nil {
				return found
			}
		}
		return nil
	}
	return search(docInfo.SymbolTable.RootScope)
}

func nodeBelongsToBlock(block *ast.BlockStatement, target ast.Node) bool {
	if block == nil {
		return false
	}
	found := false
	var walk func(node ast.Node)
	walk = func(node ast.Node) {
		if found {
			return
		}
		if node == target {
			found = true
			return
		}
		if node == nil {
			return
		}
		switch n := node.(type) {
		case *ast.BlockStatement:
			for _, stmt := range n.Statements {
				walk(stmt)
				if found {
					return
				}
			}
		case *ast.ExpressionStatement:
			walk(n.Expression)
		case *ast.CallExpression:
			walk(n.Function)
			for _, arg := range n.Arguments {
				walk(arg)
			}
		case *ast.LetStatement:
			if n.Value != nil {
				walk(n.Value)
			}
		case *ast.ConstStatement:
			if n.Value != nil {
				walk(n.Value)
			}
		case *ast.ReturnStatement:
			if n.ReturnValue != nil {
				walk(n.ReturnValue)
			}
		case *ast.IfExpression:
			walk(n.Condition)
			walk(n.Consequence)
			if n.Alternative != nil {
				walk(n.Alternative)
			}
		case *ast.WhileStatement:
			walk(n.Condition)
			walk(n.Body)
		case *ast.ForStatement:
			if n.Init != nil {
				walk(n.Init)
			}
			if n.Condition != nil {
				walk(n.Condition)
			}
			if n.Update != nil {
				walk(n.Update)
			}
			walk(n.Body)
		case *ast.ForEachStatement:
			walk(n.Iterable)
			walk(n.Variable)
			walk(n.Body)
		case *ast.InfixExpression:
			walk(n.Left)
			walk(n.Right)
		case *ast.PrefixExpression:
			walk(n.Right)
		case *ast.Identifier:
		default:
		}
	}
	walk(block)
	return found
}

func getSymbolToken(sym *Symbol) *token.Token {
	if sym.Definition == nil {
		return nil
	}
	switch d := sym.Definition.(type) {
	case *ast.Identifier:
		return &d.Token
	case *ast.FunctionStatement:
		return &d.Name.Token
	case *ast.AsyncFunctionStatement:
		return &d.Name.Token
	case *ast.LetStatement:
		if d.Name != nil {
			return &d.Name.Token
		}
		return &d.Token
	case *ast.ConstStatement:
		if d.Name != nil {
			return &d.Name.Token
		}
		return &d.Token
	}
	return nil
}

func kindToSymbolKind(sym *Symbol) protocol.SymbolKind {
	if sym != nil {
		return sym.Kind
	}
	return protocol.SymbolKindVariable
}

func declDetail(sym *Symbol) string {
	if sym == nil {
		return ""
	}
	switch sym.Definition.(type) {
	case *ast.FunctionStatement, *ast.AsyncFunctionStatement:
		return sym.Type
	}
	return ""
}
