package lsp

import (
	"fmt"
	"jabline/pkg/ast"
	jfmt "jabline/pkg/fmt"
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
	currentScope := FindCurrentScope(docInfo, path)

	switch n := node.(type) {
	case *ast.Identifier:
		if currentScope != nil {
			symbol := currentScope.Get(n.Value)
			if symbol != nil {
				switch symbol.Kind {
				case protocol.SymbolKindFunction:
					doc := ""
					if bd, ok := BuiltinDocs[symbol.Name]; ok {
						doc = "\n\n" + bd.Description
					}
					content = fmt.Sprintf("```jabline\n%s\n```\n_Function_%s", symbol.Type, doc)
				case protocol.SymbolKindConstant:
					content = fmt.Sprintf("```jabline\nconst %s: %s\n```", symbol.Name, symbol.Type)
				case protocol.SymbolKindModule:
					content = fmt.Sprintf("```jabline\nmodule %s\n```\n_Imported module_", symbol.Name)
				case protocol.SymbolKindStruct:
					content = fmt.Sprintf("```jabline\nstruct %s\n```\n_Struct type_", symbol.Name)
				case protocol.SymbolKindInterface:
					content = fmt.Sprintf("```jabline\ninterface %s\n```\n_Interface type_", symbol.Name)
				default:
					content = fmt.Sprintf("```jabline\nlet %s: %s\n```", symbol.Name, symbol.Type)
				}
			} else {
				content = fmt.Sprintf("⚠️ **Undefined**: `%s`", n.Value)
			}
		}
	case *ast.LetStatement:
		sym := currentScope.Get(n.Name.Value)
		doc := "any"
		if sym != nil {
			doc = sym.Type
		}
		content = fmt.Sprintf("```jabline\nlet %s: %s\n```", n.Name.Value, doc)
	case *ast.ConstStatement:
		sym := currentScope.Get(n.Name.Value)
		doc := "any"
		if sym != nil {
			doc = sym.Type
		}
		content = fmt.Sprintf("```jabline\nconst %s: %s\n```", n.Name.Value, doc)
	case *ast.FunctionStatement:
		sig := buildFnSignature(n.Name.Value, n.Parameters, n.ReturnType, false)
		content = fmt.Sprintf("```jabline\n%s\n```\n_Function declaration_", sig)
	case *ast.AsyncFunctionStatement:
		sig := buildFnSignature(n.Name.Value, n.Parameters, n.ReturnType, true)
		content = fmt.Sprintf("```jabline\n%s\n```\n_Async function declaration_", sig)
	case *ast.FunctionLiteral:
		varName := ""
		for i := len(path) - 2; i >= 0; i-- {
			if ls, ok2 := path[i].(*ast.LetStatement); ok2 {
				varName = ls.Name.Value
				break
			}
			if cs, ok2 := path[i].(*ast.ConstStatement); ok2 {
				varName = cs.Name.Value
				break
			}
		}
		sig := buildFnSignature(varName, n.Parameters, n.ReturnType, false)
		label := "_Anonymous function_"
		if varName != "" {
			label = "_Function assigned to `" + varName + "`_"
		}
		content = fmt.Sprintf("```jabline\n%s\n```\n%s", sig, label)
	case *ast.AsyncFunctionLiteral:
		varName := ""
		for i := len(path) - 2; i >= 0; i-- {
			if ls, ok2 := path[i].(*ast.LetStatement); ok2 {
				varName = ls.Name.Value
				break
			}
		}
		sig := buildFnSignature(varName, n.Parameters, n.ReturnType, true)
		content = fmt.Sprintf("```jabline\n%s\n```\n_Async function_", sig)
	case *ast.ArrowFunction:
		sig := buildFnSignature("", n.Parameters, n.ReturnType, false)
		content = fmt.Sprintf("```jabline\n%s => ...\n```\n_Arrow function_", sig)
	case *ast.StructStatement:
		fieldLines := ""
		for fname, typeExpr := range n.Fields {
			fieldLines += fmt.Sprintf("\n    %s: %s", fname, typeExpr.String())
		}
		content = fmt.Sprintf("```jabline\nstruct %s {%s\n}\n```\n_Struct definition_", n.Name.Value, fieldLines)
	case *ast.InterfaceStatement:
		content = fmt.Sprintf("```jabline\ninterface %s\n```\n_Interface definition_", n.Name.Value)
	case *ast.IntegerLiteral:
		content = fmt.Sprintf("```jabline\n%d\n```\n_int_", n.Value)
	case *ast.FloatLiteral:
		content = fmt.Sprintf("```jabline\n%g\n```\n_float_", n.Value)
	case *ast.Boolean:
		content = fmt.Sprintf("```jabline\n%t\n```\n_bool_", n.Value)
	case *ast.StringLiteral:
		preview := n.Value
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}
		content = fmt.Sprintf("```jabline\n\"%s\"\n```\n_string_", preview)
	case *ast.TemplateLiteral:
		content = "```jabline\n`...`\n```\n_template string_"
	case *ast.Null:
		content = "```jabline\nnull\n```\n_null value_"
	case *ast.ArrayLiteral:
		elemCount := len(n.Elements)
		content = fmt.Sprintf("```jabline\nArray[%d]\n```\n_Array literal with %d element(s)_", elemCount, elemCount)
	case *ast.HashLiteral:
		pairCount := len(n.Pairs)
		content = fmt.Sprintf("```jabline\nHash { %d pair(s) }\n```\n_Hash/object literal_", pairCount)
	case *ast.ArrayIndexExpression:
		content = "```jabline\nexpr[index]\n```\n_Array index access_"
	case *ast.IndexExpression:
		content = "```jabline\nexpr.member\n```\n_Member access_"
	case *ast.AwaitExpression:
		content = "```jabline\nawait expr\n```\n_Awaits a Promise result_"
	case *ast.SpawnExpression:
		content = "```jabline\nspawn fn(...)\n```\n_Spawns an async process_"
	default:
		return nil, nil
	}

	if content == "" {
		return nil, nil
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

	currentScope := FindCurrentScope(docInfo, path)
	if currentScope == nil {
		return nil, nil
	}

	symbol := currentScope.Get(ident.Value)
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
			case *ast.AsyncFunctionStatement:
				startTok = n.Token
				endTok = n.Body.Token
			case *ast.StructStatement:
				startTok = n.Token
				endTok = n.Name.Token
			case *ast.InterfaceStatement:
				startTok = n.Token
				endTok = n.Name.Token
			default:
				continue
			}

			startLine := uint32(startTok.Line - 1)
			startCol := uint32(startTok.Column - 1)
			endLine := uint32(endTok.Line - 1)
			endCol := uint32(endTok.Column - 1 + len(endTok.Literal))

			rng := protocol.Range{
				Start: protocol.Position{Line: startLine, Character: startCol},
				End:   protocol.Position{Line: endLine, Character: endCol},
			}

			selectionStartLine := uint32(sym.Location.Start.Line)
			selectionStartCol := uint32(sym.Location.Start.Character)
			selectionEndLine := uint32(sym.Location.End.Line)
			selectionEndCol := uint32(sym.Location.End.Character)

			// Safety check: ensure selectionRange is within rng
			if selectionStartLine < startLine || (selectionStartLine == startLine && selectionStartCol < startCol) {
				rng.Start.Line = selectionStartLine
				rng.Start.Character = selectionStartCol
			}
			if selectionEndLine > endLine || (selectionEndLine == endLine && selectionEndCol > endCol) {
				rng.End.Line = selectionEndLine
				rng.End.Character = selectionEndCol
			}

			selectionRng := protocol.Range{
				Start: protocol.Position{Line: selectionStartLine, Character: selectionStartCol},
				End:   protocol.Position{Line: selectionEndLine, Character: selectionEndCol},
			}

			children := []protocol.DocumentSymbol{}

			for _, childScope := range scope.Children {
				isChildOfSym := childScope.Node == sym.Definition
				if !isChildOfSym && sym.Kind == protocol.SymbolKindFunction {
					if fs, ok := sym.Definition.(*ast.FunctionStatement); ok {
						isChildOfSym = childScope.Node == fs.Body
					} else if afs, ok := sym.Definition.(*ast.AsyncFunctionStatement); ok {
						isChildOfSym = childScope.Node == afs.Body
					}
				}
				if isChildOfSym {
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
	var items []protocol.CompletionItem
	snippetFormat := protocol.InsertTextFormatSnippet

	// 1. Keyword/structure snippets with tab stops
	for _, snip := range KeywordSnippets {
		s := snip
		items = append(items, protocol.CompletionItem{
			Label:            s.Label,
			Kind:             ptr(protocol.CompletionItemKindKeyword),
			Detail:           ptr(s.Detail),
			InsertText:       ptr(s.InsertText),
			InsertTextFormat: &snippetFormat,
		})
	}

	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if ok && docInfo != nil && docInfo.Program != nil && docInfo.SymbolTable != nil {
		line := int(params.Position.Line) + 1
		col := int(params.Position.Character) + 1
		path := FindPathToNode(docInfo.Program, line, col)
		currentScope := FindCurrentScope(docInfo, path)

		visitedSymbols := make(map[string]bool)
		for scope := currentScope; scope != nil; scope = scope.Parent {
			for _, sym := range scope.Symbols {
				if visitedSymbols[sym.Name] {
					continue
				}
				visitedSymbols[sym.Name] = true

				item := protocol.CompletionItem{
					Label:  sym.Name,
					Kind:   ptr(protocol.CompletionItemKind(sym.Kind)),
					Detail: ptr(sym.Type),
				}

				// Add documentation for builtins
				if doc, isBuiltin := BuiltinDocs[sym.Name]; isBuiltin {
					item.Documentation = &protocol.MarkupContent{
						Kind:  protocol.MarkupKindMarkdown,
						Value: fmt.Sprintf("**%s**\n\n%s", doc.Signature, doc.Description),
					}
					// Functions get a snippet with ()
					if sym.Kind == protocol.SymbolKindFunction {
						item.InsertText = ptr(sym.Name + "($1)")
						item.InsertTextFormat = &snippetFormat
					}
				} else if sym.Kind == protocol.SymbolKindFunction && sym.Definition != nil {
					// User-defined function: add signature as doc
					item.Documentation = &protocol.MarkupContent{
						Kind:  protocol.MarkupKindMarkdown,
						Value: fmt.Sprintf("```jabline\n%s\n```", sym.Type),
					}
					item.InsertText = ptr(sym.Name + "($1)")
					item.InsertTextFormat = &snippetFormat
				}

				items = append(items, item)
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

	content, err := os.ReadFile(params.TextDocument.URI[len("file://"):])
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
						Label:      label,
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
func textDocumentFormatting(context *glsp.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil {
		return nil, nil
	}

	formatted := jfmt.Format(docInfo.Program)

	// Calculate range of entire document
	lines := strings.Split(docInfo.Content, "\n")
	lineCount := len(lines)
	lastLineLen := 0
	if lineCount > 0 {
		lastLineLen = len(lines[lineCount-1])
	}

	return []protocol.TextEdit{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: uint32(lineCount - 1), Character: uint32(lastLineLen)},
			},
			NewText: formatted,
		},
	}, nil
}
