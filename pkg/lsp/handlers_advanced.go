package lsp

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/token"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var semanticTokenTypes = []string{
	"namespace", "type", "class", "enum", "interface",
	"struct", "typeParameter", "parameter", "variable", "property",
	"enumMember", "function", "method", "macro", "keyword",
	"modifier", "comment", "string", "number", "regexp",
	"operator",
}

var semanticTokenModifiers = []string{
	"declaration", "definition", "readonly", "static",
	"deprecated", "abstract", "async", "modification",
	"documentation", "defaultLibrary",
}

func tokToPos(tok token.Token) (protocol.Position, protocol.Position) {
	start := protocol.Position{
		Line:      uint32(tok.Line - 1),
		Character: uint32(tok.Column - 1),
	}
	end := protocol.Position{
		Line:      uint32(tok.Line - 1),
		Character: uint32(tok.Column - 1 + len(tok.Literal)),
	}
	return start, end
}

func declTokenFromSymbol(symbol *Symbol) *token.Token {
	if symbol == nil || symbol.Definition == nil {
		return nil
	}
	switch d := symbol.Definition.(type) {
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
	case *ast.StructStatement:
		return &d.Name.Token
	case *ast.InterfaceStatement:
		return &d.Name.Token
	case *ast.EnumStatement:
		return &d.Name.Token
	case *ast.ServiceStatement:
		return &d.Name.Token
	}
	return nil
}

func getDocInfoAndPath(context *glsp.Context, uri string, line, col int) (*DocumentSemanticInfo, []ast.Node, bool) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[uri]
	workspaceStore.Mutex.RUnlock()
	if !ok || docInfo == nil || docInfo.Program == nil {
		return nil, nil, false
	}
	path := FindPathToNode(docInfo.Program, line, col)
	if len(path) == 0 {
		return docInfo, nil, false
	}
	return docInfo, path, true
}

func findDefinitionLocation(ident *ast.Identifier, docInfo *DocumentSemanticInfo, uri string) (*protocol.Location, error) {
	scope := FindCurrentScope(docInfo, []ast.Node{ident})
	if scope == nil {
		return nil, nil
	}
	symbol := scope.Get(ident.Value)
	if symbol == nil {
		return nil, nil
	}
	dt := declTokenFromSymbol(symbol)
	if dt == nil {
		return nil, nil
	}
	start, end := tokToPos(*dt)
	return &protocol.Location{
		URI:   uri,
		Range: protocol.Range{Start: start, End: end},
	}, nil
}

func tokenTypeToIndex(t string) uint32 {
	for i, s := range semanticTokenTypes {
		if s == t {
			return uint32(i)
		}
	}
	return 0
}

func estimateBlockEnd(block *ast.BlockStatement) uint32 {
	if block == nil {
		return 0
	}
	if len(block.Statements) > 0 {
		last := block.Statements[len(block.Statements)-1]
		tok := last.GetToken()
		return uint32(tok.Line)
	}
	return uint32(block.Token.Line + 1)
}

// ─────────────────────────────────────────────────────────────────────
// textDocument/typeDefinition
// ─────────────────────────────────────────────────────────────────────

func textDocumentTypeDefinition(context *glsp.Context, params *protocol.TypeDefinitionParams) (any, error) {
	line := int(params.Position.Line) + 1
	col := int(params.Position.Character) + 1
	docInfo, path, ok := getDocInfoAndPath(context, params.TextDocument.URI, line, col)
	if !ok {
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

	symbol := scope.Get(ident.Value)
	if symbol == nil || symbol.Definition == nil {
		return nil, nil
	}

	typeDef := symbol.Type
	if typeDef == "" || typeDef == "any" {
		return nil, nil
	}

	typeScope := scope.Get(typeDef)
	if typeScope == nil {
		switch typeDef {
		case "int", "string", "bool", "float", "null", "INTEGER", "STRING", "BOOLEAN", "FLOAT":
			return nil, nil
		}
		return nil, nil
	}

	dt := declTokenFromSymbol(typeScope)
	if dt == nil {
		return nil, nil
	}

	start, end := tokToPos(*dt)
	return &protocol.Location{
		URI:   params.TextDocument.URI,
		Range: protocol.Range{Start: start, End: end},
	}, nil
}

// ─────────────────────────────────────────────────────────────────────
// textDocument/implementation
// ─────────────────────────────────────────────────────────────────────

func textDocumentImplementation(context *glsp.Context, params *protocol.ImplementationParams) (any, error) {
	line := int(params.Position.Line) + 1
	col := int(params.Position.Character) + 1
	docInfo, path, ok := getDocInfoAndPath(context, params.TextDocument.URI, line, col)
	if !ok {
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
	symbol := scope.Get(ident.Value)
	if symbol == nil || symbol.Definition == nil {
		return nil, nil
	}

	if _, isInterface := symbol.Definition.(*ast.InterfaceStatement); !isInterface {
		return nil, nil
	}

	var results []protocol.Location
	workspaceStore.Mutex.RLock()
	for uri, doc := range workspaceStore.Documents {
		if doc == nil || doc.SymbolTable == nil {
			continue
		}
		var searchScope func(s *Scope)
		searchScope = func(s *Scope) {
			for _, sym := range s.Symbols {
				if sym.Definition == nil {
					continue
				}
				if fs, ok2 := sym.Definition.(*ast.FunctionStatement); ok2 {
					if fs.ReceiverType != nil && fs.ReceiverType.Value == ident.Value {
						dt := declTokenFromSymbol(sym)
						if dt != nil {
							start, end := tokToPos(*dt)
							results = append(results, protocol.Location{
								URI:   uri,
								Range: protocol.Range{Start: start, End: end},
							})
						}
					}
				}
			}
			for _, child := range s.Children {
				searchScope(child)
			}
		}
		searchScope(doc.SymbolTable.RootScope)
	}
	workspaceStore.Mutex.RUnlock()
	return results, nil
}

// ─────────────────────────────────────────────────────────────────────
// textDocument/prepareRename
// ─────────────────────────────────────────────────────────────────────

func textDocumentPrepareRename(context *glsp.Context, params *protocol.PrepareRenameParams) (any, error) {
	line := int(params.Position.Line) + 1
	col := int(params.Position.Character) + 1
	docInfo, path, ok := getDocInfoAndPath(context, params.TextDocument.URI, line, col)
	if !ok {
		return nil, fmt.Errorf("cannot rename: no document info")
	}
	node := path[len(path)-1]
	ident, ok := node.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("cannot rename: not an identifier")
	}
	scope := FindCurrentScope(docInfo, path)
	if scope == nil {
		return nil, fmt.Errorf("cannot rename: no scope found")
	}
	symbol := scope.Get(ident.Value)
	if symbol == nil {
		return nil, fmt.Errorf("cannot rename: undefined symbol '%s'", ident.Value)
	}
	start, end := tokToPos(ident.Token)
	return &protocol.Range{Start: start, End: end}, nil
}

// ─────────────────────────────────────────────────────────────────────
// textDocument/codeAction
// ─────────────────────────────────────────────────────────────────────

func textDocumentCodeAction(context *glsp.Context, params *protocol.CodeActionParams) (any, error) {
	var actions []protocol.CodeAction

	uri := params.TextDocument.URI
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[uri]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil {
		return actions, nil
	}

	for _, diag := range params.Context.Diagnostics {
		if diag.Source != nil && *diag.Source == lsName {
			errMsg := diag.Message

			if strings.HasPrefix(errMsg, "undefined symbol") {
				symName := extractSymbolName(errMsg)
				if symName != "" {
					actions = append(actions, protocol.CodeAction{
						Title: fmt.Sprintf("Create variable '%s'", symName),
						Kind:  ptr(protocol.CodeActionKindQuickFix),
						Edit: &protocol.WorkspaceEdit{
							Changes: map[string][]protocol.TextEdit{
								uri: {{
									Range: protocol.Range{
										Start: protocol.Position{Line: diag.Range.Start.Line, Character: 0},
										End:   protocol.Position{Line: diag.Range.Start.Line, Character: 0},
									},
									NewText: fmt.Sprintf("let %s = null;\n", symName),
								}},
							},
						},
						IsPreferred: ptr(true),
					})
				}
			}
		}
	}

	return actions, nil
}

func extractSymbolName(msg string) string {
	if idx := strings.LastIndex(msg, "'"); idx >= 0 {
		if start := strings.LastIndex(msg[:idx], "'"); start >= 0 {
			return msg[start+1 : idx]
		}
	}
	return ""
}

// ─────────────────────────────────────────────────────────────────────
// textDocument/foldingRange
// ─────────────────────────────────────────────────────────────────────

func textDocumentFoldingRange(context *glsp.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil {
		return nil, nil
	}

	var ranges []protocol.FoldingRange
	seen := make(map[string]bool)

	addRange := func(startLine, endLine uint32, kind string) {
		if endLine <= startLine {
			return
		}
		key := fmt.Sprintf("%d-%d", startLine, endLine)
		if seen[key] {
			return
		}
		seen[key] = true
		k := kind
		ranges = append(ranges, protocol.FoldingRange{
			StartLine: startLine,
			EndLine:   endLine,
			Kind:      &k,
		})
	}

	var walk func(node ast.Node)
	walk = func(node ast.Node) {
		if node == nil {
			return
		}
		switch n := node.(type) {
		case *ast.BlockStatement:
			if len(n.Statements) > 0 {
				addRange(uint32(n.Token.Line-1), estimateBlockEnd(n)-1, "region")
			}
			for _, s := range n.Statements {
				walk(s)
			}
		case *ast.FunctionStatement:
			if n.Body != nil {
				addRange(uint32(n.Body.Token.Line-1), estimateBlockEnd(n.Body)-1, "region")
				walk(n.Body)
			}
		case *ast.AsyncFunctionStatement:
			if n.Body != nil {
				addRange(uint32(n.Body.Token.Line-1), estimateBlockEnd(n.Body)-1, "region")
				walk(n.Body)
			}
		case *ast.IfExpression:
			if n.Consequence != nil {
				addRange(uint32(n.Consequence.Token.Line-1), estimateBlockEnd(n.Consequence)-1, "region")
				walk(n.Consequence)
			}
			if n.Alternative != nil {
				walk(n.Alternative)
			}
		case *ast.WhileStatement:
			if n.Body != nil {
				addRange(uint32(n.Body.Token.Line-1), estimateBlockEnd(n.Body)-1, "region")
				walk(n.Body)
			}
		case *ast.ForStatement:
			if n.Body != nil {
				addRange(uint32(n.Body.Token.Line-1), estimateBlockEnd(n.Body)-1, "region")
				walk(n.Body)
			}
		case *ast.ForEachStatement:
			if n.Body != nil {
				addRange(uint32(n.Body.Token.Line-1), estimateBlockEnd(n.Body)-1, "region")
				walk(n.Body)
			}
		case *ast.TryStatement:
			if n.TryBlock != nil {
				addRange(uint32(n.TryBlock.Token.Line-1), estimateBlockEnd(n.TryBlock)-1, "region")
				walk(n.TryBlock)
			}
			if n.CatchBlock != nil {
				addRange(uint32(n.CatchBlock.Token.Line-1), estimateBlockEnd(n.CatchBlock)-1, "region")
				walk(n.CatchBlock)
			}
		case *ast.SwitchStatement:
			for _, c := range n.Cases {
				for _, s := range c.Statements {
					walk(s)
				}
			}
		case *ast.MatchStatement:
			for _, c := range n.Cases {
				for _, s := range c.Statements {
					walk(s)
				}
			}
		case *ast.RetryStatement:
			if n.RetryBlock != nil {
				addRange(uint32(n.RetryBlock.Token.Line-1), estimateBlockEnd(n.RetryBlock)-1, "region")
				walk(n.RetryBlock)
			}
		case *ast.Program:
			for _, s := range n.Statements {
				walk(s)
			}
		case *ast.ExpressionStatement:
			walk(n.Expression)
		case *ast.LetStatement:
			if n.Value != nil {
				walk(n.Value)
			}
		case *ast.ConstStatement:
			if n.Value != nil {
				walk(n.Value)
			}
		case *ast.AssignmentStatement:
			walk(n.Left)
			walk(n.Value)
		case *ast.ImportStatement:
		case *ast.ExportStatement:
		default:
		}
	}

	walk(docInfo.Program)
	return ranges, nil
}

// ─────────────────────────────────────────────────────────────────────
// textDocument/semanticTokens
// ─────────────────────────────────────────────────────────────────────

func textDocumentSemanticTokensFull(context *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.Program == nil {
		return &protocol.SemanticTokens{Data: []uint32{}}, nil
	}

	var data []uint32
	var prevLine, prevCol uint32

	emit := func(line, col, length uint32, tokenType uint32, modifiers uint32) {
		if len(data) == 0 {
			data = append(data, line, col, length, tokenType, modifiers)
		} else if line > prevLine {
			data = append(data, line-prevLine, col, length, tokenType, modifiers)
		} else {
			data = append(data, line-prevLine, col-prevCol, length, tokenType, modifiers)
		}
		prevLine = line
		prevCol = col
	}

	var walk func(node ast.Node)
	walk = func(node ast.Node) {
		if node == nil {
			return
		}
		switch n := node.(type) {
		case *ast.Identifier:
			sline := uint32(n.Token.Line - 1)
			scol := uint32(n.Token.Column - 1)
			length := uint32(len(n.Token.Literal))
			emit(sline, scol, length, tokenTypeToIndex("variable"), 0)

		case *ast.MatchStatement:
			sline := uint32(n.Token.Line - 1)
			scol := uint32(n.Token.Column - 1)
			length := uint32(len(n.Token.Literal))
			emit(sline, scol, length, tokenTypeToIndex("keyword"), 0)

		case *ast.IntegerLiteral:
			sline := uint32(n.Token.Line - 1)
			scol := uint32(n.Token.Column - 1)
			length := uint32(len(n.Token.Literal))
			emit(sline, scol, length, tokenTypeToIndex("number"), 0)

		case *ast.FloatLiteral:
			sline := uint32(n.Token.Line - 1)
			scol := uint32(n.Token.Column - 1)
			length := uint32(len(n.Token.Literal))
			emit(sline, scol, length, tokenTypeToIndex("number"), 0)

		case *ast.StringLiteral:
			sline := uint32(n.Token.Line - 1)
			scol := uint32(n.Token.Column - 1)
			length := uint32(len(n.Token.Literal))
			emit(sline, scol, length, tokenTypeToIndex("string"), 0)

		case *ast.TemplateLiteral:
			sline := uint32(n.Token.Line - 1)
			scol := uint32(n.Token.Column - 1)
			length := uint32(len(n.Token.Literal))
			emit(sline, scol, length, tokenTypeToIndex("string"), 0)

		case *ast.Boolean:
			sline := uint32(n.Token.Line - 1)
			scol := uint32(n.Token.Column - 1)
			length := uint32(len(n.Token.Literal))
			emit(sline, scol, length, tokenTypeToIndex("keyword"), 0)

		case *ast.Null:
			sline := uint32(n.Token.Line - 1)
			scol := uint32(n.Token.Column - 1)
			length := uint32(len(n.Token.Literal))
			emit(sline, scol, length, tokenTypeToIndex("keyword"), 0)

		case *ast.LetStatement:
			if n.Name != nil {
				sline := uint32(n.Name.Token.Line - 1)
				scol := uint32(n.Name.Token.Column - 1)
				length := uint32(len(n.Name.Token.Literal))
				emit(sline, scol, length, tokenTypeToIndex("variable"), 1)
			}
			if n.Value != nil {
				walk(n.Value)
			}

		case *ast.ConstStatement:
			if n.Name != nil {
				sline := uint32(n.Name.Token.Line - 1)
				scol := uint32(n.Name.Token.Column - 1)
				length := uint32(len(n.Name.Token.Literal))
				emit(sline, scol, length, tokenTypeToIndex("variable"), 1)
			}
			if n.Value != nil {
				walk(n.Value)
			}

		case *ast.FunctionStatement:
			if n.Name != nil {
				sline := uint32(n.Name.Token.Line - 1)
				scol := uint32(n.Name.Token.Column - 1)
				length := uint32(len(n.Name.Token.Literal))
				emit(sline, scol, length, tokenTypeToIndex("function"), 1)
			}
			if n.Body != nil {
				walk(n.Body)
			}

		case *ast.AsyncFunctionStatement:
			if n.Name != nil {
				sline := uint32(n.Name.Token.Line - 1)
				scol := uint32(n.Name.Token.Column - 1)
				length := uint32(len(n.Name.Token.Literal))
				emit(sline, scol, length, tokenTypeToIndex("function"), 1)
			}
			if n.Body != nil {
				walk(n.Body)
			}

		case *ast.StructStatement:
			if n.Name != nil {
				sline := uint32(n.Name.Token.Line - 1)
				scol := uint32(n.Name.Token.Column - 1)
				length := uint32(len(n.Name.Token.Literal))
				emit(sline, scol, length, tokenTypeToIndex("struct"), 1)
			}

		case *ast.InterfaceStatement:
			if n.Name != nil {
				sline := uint32(n.Name.Token.Line - 1)
				scol := uint32(n.Name.Token.Column - 1)
				length := uint32(len(n.Name.Token.Literal))
				emit(sline, scol, length, tokenTypeToIndex("interface"), 1)
			}

		case *ast.EnumStatement:
			if n.Name != nil {
				sline := uint32(n.Name.Token.Line - 1)
				scol := uint32(n.Name.Token.Column - 1)
				length := uint32(len(n.Name.Token.Literal))
				emit(sline, scol, length, tokenTypeToIndex("enum"), 1)
			}

		case *ast.ServiceStatement:
			if n.Name != nil {
				sline := uint32(n.Name.Token.Line - 1)
				scol := uint32(n.Name.Token.Column - 1)
				length := uint32(len(n.Name.Token.Literal))
				emit(sline, scol, length, tokenTypeToIndex("namespace"), 1)
			}

		case *ast.BlockStatement:
			for _, stmt := range n.Statements {
				walk(stmt)
			}

		case *ast.Program:
			for _, stmt := range n.Statements {
				walk(stmt)
			}

		case *ast.ExpressionStatement:
			walk(n.Expression)

		default:
		}
	}

	walk(docInfo.Program)
	return &protocol.SemanticTokens{Data: data}, nil
}

// ─────────────────────────────────────────────────────────────────────
// workspace/symbol
// ─────────────────────────────────────────────────────────────────────

func workspaceSymbol(context *glsp.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	var symbols []protocol.SymbolInformation
	query := strings.ToLower(params.Query)

	workspaceStore.Mutex.RLock()
	defer workspaceStore.Mutex.RUnlock()

	for uri, docInfo := range workspaceStore.Documents {
		if docInfo == nil || docInfo.SymbolTable == nil {
			continue
		}

		var walkScope func(scope *Scope, container string)
		walkScope = func(scope *Scope, container string) {
			for _, sym := range scope.Symbols {
				if query != "" && !strings.Contains(strings.ToLower(sym.Name), query) {
					continue
				}

				si := protocol.SymbolInformation{
					Name: sym.Name,
					Kind: sym.Kind,
					Location: protocol.Location{
						URI: uri,
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(sym.Location.Start.Line),
								Character: uint32(sym.Location.Start.Character),
							},
							End: protocol.Position{
								Line:      uint32(sym.Location.End.Line),
								Character: uint32(sym.Location.End.Character),
							},
						},
					},
				}
				if container != "" {
					si.ContainerName = &container
				}
				symbols = append(symbols, si)
			}

			for _, child := range scope.Children {
				var childContainer string
				for _, sym := range scope.Symbols {
					if child.Node == sym.Definition {
						childContainer = sym.Name
						break
					}
				}
				walkScope(child, childContainer)
			}
		}

		walkScope(docInfo.SymbolTable.RootScope, "")
	}

	return symbols, nil
}

// ─────────────────────────────────────────────────────────────────────
// completionItem/resolve
// ─────────────────────────────────────────────────────────────────────

func textDocumentCompletionResolve(context *glsp.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	if params.Detail != nil && *params.Detail != "" {
		if doc, ok := BuiltinDocs[params.Label]; ok {
			params.Documentation = &protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: fmt.Sprintf("**%s**\n\n%s", doc.Signature, doc.Description),
			}
		}
	}
	return params, nil
}
