package lsp

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/token"
	"os"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentHover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil {
		return nil, nil
	}

	line := int(params.Position.Line) + 1
	col := int(params.Position.Character) + 1

	path := FindPathToNode(docInfo.Program, line, col)
	if len(path) == 0 {
		return nil, nil
	}
	node := path[len(path)-1]

	var content string
	switch n := node.(type) {
	case *ast.Identifier:

		symbol := docInfo.SymbolTable.RootScope.Get(n.Value)
		if symbol != nil {
			content = fmt.Sprintf("**Identifier**: `%s` (Type: `%s`, Kind: `%v`)", symbol.Name, symbol.Type, symbol.Kind)
		} else {
			content = fmt.Sprintf("**Identifier**: `%s` (Undefined)", n.Value)
		}
	case *ast.IntegerLiteral:
		content = fmt.Sprintf("**Integer**: `%d`", n.Value)
	case *ast.Boolean:
		content = fmt.Sprintf("**Boolean**: `%t`", n.Value)
	case *ast.FunctionLiteral:
		content = "**Function Definition**"
	case *ast.LetStatement:
		content = fmt.Sprintf("**Variable Declaration**: `%s`", n.Name.Value)
	case *ast.ConstStatement:
		content = fmt.Sprintf("**Constant Declaration**: `%s`", n.Name.Value)
	case *ast.StructStatement:
		content = fmt.Sprintf("**Struct Definition**: `%s`", n.Name.Value)
	default:
		content = fmt.Sprintf("**Node**: %T\n`%s`", n, n.TokenLiteral())
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: content,
		},
	}, nil
}

func textDocumentDefinition(context *glsp.Context, params *protocol.DefinitionParams) (any, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil {
		return nil, nil
	}

	line := int(params.Position.Line) + 1
	col := int(params.Position.Character) + 1

	path := FindPathToNode(docInfo.Program, line, col)
	if len(path) == 0 {
		return nil, nil
	}

	node := path[len(path)-1]

	ident, ok := node.(*ast.Identifier)
	if !ok {
		return nil, nil
	}

	symbol := docInfo.SymbolTable.RootScope.Get(ident.Value)
	if symbol == nil || symbol.Definition == nil {
		return nil, nil
	}

	var declToken token.Token

	switch d := symbol.Definition.(type) {
	case *ast.Identifier:
	
declToken = d.Token

	case *ast.FunctionStatement:
	
declToken = d.Name.Token
	case *ast.LetStatement:
	
declToken = d.Name.Token
	case *ast.ConstStatement:
	
declToken = d.Name.Token
	case *ast.StructStatement:
	
declToken = d.Name.Token
	default:
		return nil, nil
	}

	targetLine := uint32(declToken.Line - 1)
	targetCol := uint32(declToken.Column - 1)

	return protocol.Location{
		URI: params.TextDocument.URI,
		Range: protocol.Range{
			Start: protocol.Position{Line: targetLine, Character: targetCol},
			End:   protocol.Position{Line: targetLine, Character: targetCol + uint32(len(declToken.Literal))},
		},
	}, nil
}

func textDocumentDocumentSymbol(context *glsp.Context, params *protocol.DocumentSymbolParams) (any, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil || docInfo.SymbolTable == nil {
		return nil, nil
	}

	var symbols []protocol.DocumentSymbol

	var walkScope func(scope *Scope) []protocol.DocumentSymbol
	walkScope = func(scope *Scope) []protocol.DocumentSymbol {
		var currentSymbols []protocol.DocumentSymbol
		
		for _, sym := range scope.Symbols {

			var startTok, endTok token.Token
			switch n := sym.Definition.(type) {
			case *ast.Identifier:
				startTok = n.Token
				endTok = n.Token
			case *ast.LetStatement:
				startTok = n.Token
				endTok = n.Name.Token
			case *ast.ConstStatement:
				startTok = n.Token
				endTok = n.Name.Token
			case *ast.FunctionStatement:
				startTok = n.Token
				endTok = n.Body.Token
			case *ast.StructStatement:
				startTok = n.Token
				endTok = n.Name.Token
			default:
				continue
			}

			startLine := uint32(startTok.Line - 1)
			startCol := uint32(startTok.Column - 1)
			endLine := uint32(endTok.Line - 1)
			endCol := uint32(endTok.Column + len(endTok.Literal))

			rng := protocol.Range{
				Start: protocol.Position{Line: startLine, Character: startCol},
				End:   protocol.Position{Line: endLine, Character: endCol},
			}
			selectionRng := protocol.Range{
				Start: protocol.Position{Line: uint32(sym.Location.Start.Line), Character: uint32(sym.Location.Start.Character)},
				End:   protocol.Position{Line: uint32(sym.Location.End.Line), Character: uint32(sym.Location.End.Character)},
			}
			
			children := []protocol.DocumentSymbol{}

			for _, childScope := range scope.Children {

				if childScope.Node == sym.Definition || (sym.Kind == protocol.SymbolKindFunction && childScope.Node == sym.Definition.(*ast.FunctionStatement).Body) {
					children = append(children, walkScope(childScope)...)
				}
			}

			currentSymbols = append(currentSymbols, protocol.DocumentSymbol{
				Name:           sym.Name,
				Kind:           sym.Kind,
				Range:          rng,
				SelectionRange: selectionRng,
				Children:       children,
			})
		}
		
		for _, childScope := range scope.Children {

			isHandledByParentSymbol := false
			for _, sym := range scope.Symbols {
				if childScope.Node == sym.Definition {
					isHandledByParentSymbol = true
					break
				}
			}
			if !isHandledByParentSymbol {
				currentSymbols = append(currentSymbols, walkScope(childScope)...)
			}
		}

		return currentSymbols
	}

	symbols = walkScope(docInfo.SymbolTable.RootScope)
	return symbols, nil
}

func textDocumentCompletion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {

	keywords := []string{
		"fn", "let", "const", "return", "if", "else", "true", "false", "for", "while",
		"struct", "import", "export", "null", "async", "await", "try", "catch", "throw",
	}

	var items []protocol.CompletionItem
	for _, kw := range keywords {
		k := kw
		items = append(items, protocol.CompletionItem{
			Label: k,
			Kind:  ptr(protocol.CompletionItemKindKeyword),
		})
	}

	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if ok && docInfo != nil && docInfo.Program != nil && docInfo.SymbolTable != nil {
		line := int(params.Position.Line) + 1
		col := int(params.Position.Character) + 1
		
		var currentScope *Scope
		path := FindPathToNode(docInfo.Program, line, col)

		for i := len(path) - 1; i >= 0; i-- {
			node := path[i]
			
			var foundScope *Scope
			var findScopeByNode func(s *Scope, target ast.Node) *Scope
			findScopeByNode = func(s *Scope, target ast.Node) *Scope {
				if s.Node == target {
					return s
				}
				for _, child := range s.Children {
					if found := findScopeByNode(child, target); found != nil {
						return found
					}
				}
				return nil
			}

			if docInfo.SymbolTable.RootScope.Node == node {
				foundScope = docInfo.SymbolTable.RootScope
			} else {
				foundScope = findScopeByNode(docInfo.SymbolTable.RootScope, node)
			}
			
			if foundScope != nil {
				currentScope = foundScope
				break
			}
		}
		
		if currentScope == nil {
			currentScope = docInfo.SymbolTable.RootScope
		}

		visitedSymbols := make(map[string]bool)
		for scope := currentScope; scope != nil; scope = scope.Parent {
			for _, sym := range scope.Symbols {
				if !visitedSymbols[sym.Name] {
					items = append(items, protocol.CompletionItem{
						Label: sym.Name,
						Kind:  ptr(protocol.CompletionItemKind(sym.Kind)),
						Detail: ptr(sym.Type),
					})
					visitedSymbols[sym.Name] = true
				}
			}
		}
	}

	return items, nil
}

func textDocumentSignatureHelp(context *glsp.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {

	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil {
		return nil, nil
	}

	content, err := os.ReadFile(params.TextDocument.URI[len("file://"):
	])
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to read file for signature help: %v", err))
		return nil, nil
	}
	funcContent := string(content)

	line := int(params.Position.Line) + 1
	col := int(params.Position.Character) + 1
	
	path := FindPathToNode(docInfo.Program, line, col)
	
	var callExpr *ast.CallExpression
	
	for i := len(path) - 1; i >= 0; i-- {
		if ce, ok := path[i].(*ast.CallExpression); ok {
			callExpr = ce
			break
		}
	}
	
	if callExpr == nil {
		return nil, nil
	}

	ident, ok := callExpr.Function.(*ast.Identifier)
	if !ok {
		return nil, nil
	}
	
	symbol := docInfo.SymbolTable.RootScope.Get(ident.Value)
	if symbol == nil || symbol.Definition == nil {
		return nil, nil
	}

	var label string
	var paramsInfo []protocol.ParameterInformation
	
	switch f := symbol.Definition.(type) {
	case *ast.FunctionStatement:
		label = "fn " + f.Name.Value + "("
		for i, p := range f.Parameters {
			if i > 0 {
				label += ", "
			}
			label += p.Value
			paramsInfo = append(paramsInfo, protocol.ParameterInformation{
				Label: p.Value,
			})
		}
		label += ")"
	case *ast.FunctionLiteral:

		label = "fn("
		for i, p := range f.Parameters {
			if i > 0 {
				label += ", "
			}
			label += p.Value
			paramsInfo = append(paramsInfo, protocol.ParameterInformation{
				Label: p.Value,
			})
		}
		label += ")"
	
	default:
		return nil, nil
	}

	activeParameter := uint32(0)
	
	funcIdentifierStartOffset := getTokenByteOffset(funcContent, ident.Token.Line, ident.Token.Column)
	if funcIdentifierStartOffset == -1 {
		return nil, nil
	}

	openParenIdx := -1
	for i := funcIdentifierStartOffset + len(ident.Token.Literal); i < len(funcContent); i++ {
		if funcContent[i] == '(' {
			openParenIdx = i
			break
		} else if !isWhitespaceByte(funcContent[i]) {

			return nil, nil
		}
	}

	if openParenIdx != -1 {

		cursorByteOffset := getByteOffset(funcContent, int(params.Position.Line)+1, int(params.Position.Character)+1)
		if cursorByteOffset == -1 || cursorByteOffset < openParenIdx {

			activeParameter = 0
		} else {
			argSegment := funcContent[openParenIdx:cursorByteOffset]
			activeParameter = uint32(strings.Count(argSegment, ","))
		}
	}

	if label != "" {
		return &protocol.SignatureHelp{
			Signatures: []protocol.SignatureInformation{
				{
					Label: label,
					Parameters: paramsInfo,
				},
			},
			ActiveSignature: ptr(uint32(0)),
			ActiveParameter: ptr(activeParameter),
		},
		nil
	}

	return nil, nil
}

func textDocumentReferences(context *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil {
		return nil, nil
	}

	line := int(params.Position.Line) + 1
	col := int(params.Position.Character) + 1

	path := FindPathToNode(docInfo.Program, line, col)
	if len(path) == 0 {
		return nil, nil
	}

	node := path[len(path)-1]

	ident, ok := node.(*ast.Identifier)
	if !ok {
		return nil, nil
	}

	symbol := docInfo.SymbolTable.RootScope.Get(ident.Value)
	if symbol == nil {
		return nil, nil
	}

	var references []protocol.Location
	if params.Context.IncludeDeclaration {

		var declToken token.Token
		hasDecl := false
		switch d := symbol.Definition.(type) {
		case *ast.Identifier:
		
declToken = d.Token
			hasDecl = true
		case *ast.FunctionStatement:
		
declToken = d.Name.Token
			hasDecl = true
		case *ast.LetStatement:
		
declToken = d.Name.Token
			hasDecl = true
		case *ast.ConstStatement:
		
declToken = d.Name.Token
			hasDecl = true
		case *ast.StructStatement:
		
declToken = d.Name.Token
			hasDecl = true
		}

		if hasDecl {
			startLine := uint32(declToken.Line - 1)
			startCol := uint32(declToken.Column - 1)
			references = append(references, protocol.Location{
				URI: params.TextDocument.URI,

				Range: protocol.Range{
					Start: protocol.Position{Line: startLine, Character: startCol},
					End:   protocol.Position{Line: startLine, Character: startCol + uint32(len(declToken.Literal))},
				},
			})
		}
	}

	references = append(references, symbol.References...)

	return references, nil
}

func textDocumentRename(context *glsp.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil {
		return nil, nil
	}

	line := int(params.Position.Line) + 1
	col := int(params.Position.Character) + 1

	path := FindPathToNode(docInfo.Program, line, col)
	if len(path) == 0 {
		return nil, nil
	}

	node := path[len(path)-1]

	ident, ok := node.(*ast.Identifier)
	if !ok {
		return nil, nil
	}

	symbol := docInfo.SymbolTable.RootScope.Get(ident.Value)
	if symbol == nil {
		return nil, nil
	}

	changes := make(map[string][]protocol.TextEdit)
	
	var declToken token.Token
	hasDecl := false
	switch d := symbol.Definition.(type) {
	case *ast.Identifier:
	
declToken = d.Token
		hasDecl = true
	case *ast.FunctionStatement:
	
declToken = d.Name.Token
		hasDecl = true
	case *ast.LetStatement:
	
declToken = d.Name.Token
		hasDecl = true
	case *ast.ConstStatement:
	
declToken = d.Name.Token
		hasDecl = true
	case *ast.StructStatement:
	
declToken = d.Name.Token
		hasDecl = true
	}

	if hasDecl {
		uri := params.TextDocument.URI
		startLine := uint32(declToken.Line - 1)
		startCol := uint32(declToken.Column - 1)
		edit := protocol.TextEdit{
			Range: protocol.Range{
				Start: protocol.Position{Line: startLine, Character: startCol},
				End:   protocol.Position{Line: startLine, Character: startCol + uint32(len(declToken.Literal))},
			},
			NewText: params.NewName,
		}
		changes[uri] = append(changes[uri], edit)
	}

	for _, ref := range symbol.References {
		changes[ref.URI] = append(changes[ref.URI], protocol.TextEdit{
			Range:   ref.Range,
			NewText: params.NewName,
		})
	}

	return &protocol.WorkspaceEdit{
		Changes: changes,
	},
	nil
}
