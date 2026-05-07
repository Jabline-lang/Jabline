package lsp

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/token"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"jabline/pkg/stdlib"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

type SymbolKind = protocol.SymbolKind

type Symbol struct {
	Name       string
	Kind       SymbolKind
	Type       string
	Location   protocol.Range
	Definition ast.Node
	Container  *Scope
	References []protocol.Location
}

type Scope struct {
	Parent   *Scope
	Symbols  map[string]*Symbol
	Children []*Scope
	Node     ast.Node
}

func NewScope(parent *Scope, node ast.Node) *Scope {
	return &Scope{
		Parent:  parent,
		Symbols: make(map[string]*Symbol),
		Node:    node,
	}
}

func (s *Scope) Get(name string) *Symbol {
	if symbol, ok := s.Symbols[name]; ok {
		return symbol
	}

	if s.Parent != nil {
		return s.Parent.Get(name)
	}

	return nil
}

func (s *Scope) Set(symbol *Symbol) error {
	if _, ok := s.Symbols[symbol.Name]; ok {
		return fmt.Errorf("symbol '%s' already declared in this scope", symbol.Name)
	}
	s.Symbols[symbol.Name] = symbol
	symbol.Container = s
	return nil
}

type SymbolTable struct {
	RootScope *Scope
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{}
}

type SemanticAnalyzer struct {
	Program      *ast.Program
	Errors       []string
	Symbols      *SymbolTable
	currentScope *Scope
	Workspace    *WorkspaceSymbolStore
	FileURI      string
}

func NewSemanticAnalyzer(program *ast.Program, ws *WorkspaceSymbolStore, fileURI string) *SemanticAnalyzer {

	rootScope := NewScope(nil, program)

	// Pre-populate builtins so they don't trigger "undefined variable"
	for _, builtin := range stdlib.Registry {
		sig := "builtin fn"
		if doc, ok := BuiltinDocs[builtin.Name]; ok {
			sig = doc.Signature
		}
		symbol := &Symbol{
			Name:       builtin.Name,
			Kind:       protocol.SymbolKindFunction,
			Type:       sig,
			Location:   protocol.Range{},
			Definition: nil,
		}
		rootScope.Set(symbol)
	}

	// Pre-populate global test functions so they don't trigger "undefined variable"
	testBuiltins := []string{"describe", "it", "assertEqual", "assertNotEqual", "assertTrue", "assertFalse"}
	for _, name := range testBuiltins {
		symbol := &Symbol{
			Name:       name,
			Kind:       protocol.SymbolKindFunction,
			Type:       "test fn",
			Location:   protocol.Range{},
			Definition: nil,
		}
		rootScope.Set(symbol)
	}

	sa := &SemanticAnalyzer{
		Program:      program,
		Symbols:      &SymbolTable{RootScope: rootScope},
		currentScope: rootScope,
		Workspace:    ws,
		FileURI:      fileURI,
	}

	return sa
}

func (sa *SemanticAnalyzer) Analyze() {
	for _, stmt := range sa.Program.Statements {
		sa.walk(stmt)
	}
}

// buildFnSignature builds a human-readable function signature string.
func buildFnSignature(name string, params []*ast.Identifier, returnType *ast.TypeExpression, isAsync bool) string {
	prefix := "fn"
	if isAsync {
		prefix = "async fn"
	}
	var sb strings.Builder
	sb.WriteString(prefix)
	if name != "" {
		sb.WriteString(" ")
		sb.WriteString(name)
	}
	sb.WriteString("(")
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.Value
	}
	sb.WriteString(strings.Join(parts, ", "))
	sb.WriteString(")")
	if returnType != nil {
		sb.WriteString(": ")
		sb.WriteString(returnType.String())
	}
	return sb.String()
}

func (sa *SemanticAnalyzer) walk(node ast.Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.Program:

	case *ast.BlockStatement:
		oldScope := sa.currentScope
		newScope := NewScope(oldScope, n)
		oldScope.Children = append(oldScope.Children, newScope)
		sa.currentScope = newScope

		for _, stmt := range n.Statements {
			sa.walk(stmt)
		}

		sa.currentScope = oldScope

	case *ast.LetStatement:
		valType := "any"
		if n.Value != nil {
			valType = sa.inferType(n.Value)
			sa.walk(n.Value)
		}
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindVariable, valType, n.Name.Token, n)

	case *ast.ConstStatement:
		valType := "any"
		if n.Value != nil {
			valType = sa.inferType(n.Value)
			sa.walk(n.Value)
		}
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindConstant, valType, n.Name.Token, n)

	case *ast.FunctionStatement:
		sig := buildFnSignature(n.Name.Value, n.Parameters, n.ReturnType, false)
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindFunction, sig, n.Name.Token, n)

		oldScope := sa.currentScope
		newScope := NewScope(oldScope, n)
		oldScope.Children = append(oldScope.Children, newScope)
		sa.currentScope = newScope

		for _, param := range n.Parameters {
			sa.declareSymbol(param.Value, protocol.SymbolKindVariable, "any", param.Token, param)
		}
		sa.walk(n.Body)

		sa.currentScope = oldScope

	case *ast.AsyncFunctionStatement:
		sig := buildFnSignature(n.Name.Value, n.Parameters, n.ReturnType, true)
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindFunction, sig, n.Name.Token, n)

		oldScope := sa.currentScope
		newScope := NewScope(oldScope, n)
		oldScope.Children = append(oldScope.Children, newScope)
		sa.currentScope = newScope

		for _, param := range n.Parameters {
			sa.declareSymbol(param.Value, protocol.SymbolKindVariable, "any", param.Token, param)
		}
		sa.walk(n.Body)

		sa.currentScope = oldScope

	case *ast.FunctionLiteral:
		oldScope := sa.currentScope
		newScope := NewScope(oldScope, n)
		oldScope.Children = append(oldScope.Children, newScope)
		sa.currentScope = newScope

		for _, param := range n.Parameters {
			sa.declareSymbol(param.Value, protocol.SymbolKindVariable, "any", param.Token, param)
		}
		sa.walk(n.Body)

		sa.currentScope = oldScope

	case *ast.AsyncFunctionLiteral:
		oldScope := sa.currentScope
		newScope := NewScope(oldScope, n)
		oldScope.Children = append(oldScope.Children, newScope)
		sa.currentScope = newScope

		for _, param := range n.Parameters {
			sa.declareSymbol(param.Value, protocol.SymbolKindVariable, "any", param.Token, param)
		}
		sa.walk(n.Body)

		sa.currentScope = oldScope

	case *ast.ArrowFunction:
		oldScope := sa.currentScope
		newScope := NewScope(oldScope, n)
		oldScope.Children = append(oldScope.Children, newScope)
		sa.currentScope = newScope

		for _, param := range n.Parameters {
			sa.declareSymbol(param.Value, protocol.SymbolKindVariable, "any", param.Token, param)
		}
		sa.walk(n.Body)

		sa.currentScope = oldScope

	case *ast.StructStatement:
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindStruct, "struct", n.Name.Token, n)

	case *ast.InterfaceStatement:
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindInterface, "interface", n.Name.Token, n)

	case *ast.ExpressionStatement:
		sa.walk(n.Expression)

	case *ast.ReturnStatement:
		if n.ReturnValue != nil {
			sa.walk(n.ReturnValue)
		}

	case *ast.AssignmentStatement:
		sa.walk(n.Left)
		sa.walk(n.Value)

	case *ast.EchoStatement:
		for _, v := range n.Values {
			sa.walk(v)
		}

	case *ast.IfExpression:
		sa.walk(n.Condition)
		sa.walk(n.Consequence)
		if n.Alternative != nil {
			sa.walk(n.Alternative)
		}

	case *ast.InfixExpression:
		sa.walk(n.Left)
		sa.walk(n.Right)

	case *ast.PrefixExpression:
		sa.walk(n.Right)

	case *ast.PostfixExpression:
		sa.walk(n.Left)

	case *ast.TernaryExpression:
		sa.walk(n.Condition)
		sa.walk(n.TrueValue)
		sa.walk(n.FalseValue)

	case *ast.NullishCoalescingExpression:
		sa.walk(n.Left)
		sa.walk(n.Right)

	case *ast.OptionalChainingExpression:
		sa.walk(n.Left)
		sa.walk(n.Right)

	case *ast.CallExpression:
		sa.walk(n.Function)
		for _, arg := range n.Arguments {
			sa.walk(arg)
		}
		// Arity Check Diagnostics
		if ident, ok := n.Function.(*ast.Identifier); ok {
			sym := sa.currentScope.Get(ident.Value)
			if sym != nil && sym.Definition != nil {
				if fs, isFunc := sym.Definition.(*ast.FunctionStatement); isFunc {
					expected := len(fs.Parameters)
					actual := len(n.Arguments)
					if expected != actual {
						sa.Errors = append(sa.Errors, fmt.Sprintf("line %d, column %d: expected %d arguments, got %d", n.Token.Line, n.Token.Column, expected, actual))
					}
				}
			}
		}

	case *ast.AwaitExpression:
		sa.walk(n.Value)

	case *ast.SpawnExpression:
		sa.walk(n.Call)

	case *ast.ArrayLiteral:
		for _, el := range n.Elements {
			sa.walk(el)
		}

	case *ast.HashLiteral:
		for k, v := range n.Pairs {
			sa.walk(k)
			sa.walk(v)
		}

	case *ast.ArrayIndexExpression:
		sa.walk(n.Left)
		sa.walk(n.Index)

	case *ast.IndexExpression:
		sa.walk(n.Left)
		sa.walk(n.Index)

	case *ast.WhileStatement:
		sa.walk(n.Condition)
		oldScope := sa.currentScope
		newScope := NewScope(oldScope, n.Body)
		oldScope.Children = append(oldScope.Children, newScope)
		sa.currentScope = newScope
		sa.walk(n.Body)
		sa.currentScope = oldScope

	case *ast.ForStatement:
		oldScope := sa.currentScope
		newScope := NewScope(oldScope, n)
		oldScope.Children = append(oldScope.Children, newScope)
		sa.currentScope = newScope

		if n.Init != nil {
			sa.walk(n.Init)
		}
		if n.Condition != nil {
			sa.walk(n.Condition)
		}
		sa.walk(n.Body)
		if n.Update != nil {
			sa.walk(n.Update)
		}
		sa.currentScope = oldScope

	case *ast.ForEachStatement:
		sa.walk(n.Iterable)
		oldScope := sa.currentScope
		newScope := NewScope(oldScope, n)
		oldScope.Children = append(oldScope.Children, newScope)
		sa.currentScope = newScope
		sa.declareSymbol(n.Variable.Value, protocol.SymbolKindVariable, "any", n.Variable.Token, n.Variable)
		sa.walk(n.Body)
		sa.currentScope = oldScope

	case *ast.TryStatement:
		sa.walk(n.TryBlock)
		if n.CatchBlock != nil {
			oldScope := sa.currentScope
			newScope := NewScope(oldScope, n.CatchBlock)
			oldScope.Children = append(oldScope.Children, newScope)
			sa.currentScope = newScope
			if n.CatchParam != nil {
				sa.declareSymbol(n.CatchParam.Value, protocol.SymbolKindVariable, "Error", n.CatchParam.Token, n.CatchParam)
			}
			sa.walk(n.CatchBlock)
			sa.currentScope = oldScope
		}

	case *ast.ThrowStatement:
		sa.walk(n.Value)

	case *ast.SwitchStatement:
		sa.walk(n.Expression)
		for _, c := range n.Cases {
			sa.walk(c.Value)
			for _, s := range c.Statements {
				sa.walk(s)
			}
		}
		if n.DefaultCase != nil {
			for _, s := range n.DefaultCase.Statements {
				sa.walk(s)
			}
		}

	case *ast.ImportStatement:
		if logger != nil {
			logger.Info(fmt.Sprintf("SemanticAnalyzer: Encountered ImportStatement for module: %s", n.ModuleName.Value))
		}

		if strings.HasPrefix(n.ModuleName.Value, "std/") || strings.HasPrefix(n.ModuleName.Value, "sys/") {
			alias := ""
			if n.NamespaceAlias != nil {
				alias = n.NamespaceAlias.Value
			} else if n.DefaultImport != nil {
				alias = n.DefaultImport.Value
			} else {
				alias = strings.TrimPrefix(n.ModuleName.Value, "std/")
				alias = strings.TrimPrefix(alias, "sys/")
				if idx := strings.LastIndex(alias, "/"); idx != -1 {
					alias = alias[idx+1:]
				}
			}

			if alias != "" {
				sa.declareSymbol(alias, protocol.SymbolKindModule, "module", n.ModuleName.Token, nil)
			}
			return
		}

		moduleURI := sa.resolveModulePathToURI(n.ModuleName.Value)
		if moduleURI != "" {
			importedDoc := sa.processImportedModule(moduleURI)
			if importedDoc != nil {
				sa.integrateImportedSymbols(n, importedDoc)
			}
		}

	case *ast.Identifier:
		symbol := sa.currentScope.Get(n.Value)
		if symbol != nil {
			refLocation := protocol.Location{
				URI: sa.FileURI,
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(n.Token.Line - 1), Character: uint32(n.Token.Column - 1)},
					End:   protocol.Position{Line: uint32(n.Token.Line - 1), Character: uint32(n.Token.Column - 1 + len(n.Token.Literal))},
				},
			}
			symbol.References = append(symbol.References, refLocation)
		} else {
			// Real-time Semantic Diagnostic: Undefined Variable!
			sa.Errors = append(sa.Errors, fmt.Sprintf("line %d, column %d: undefined symbol '%s'", n.Token.Line, n.Token.Column, n.Value))
		}
	}
}

// inferType tries to determine the type of an AST node without compiling.
func (sa *SemanticAnalyzer) inferType(node ast.Node) string {
	if node == nil {
		return "any"
	}
	switch n := node.(type) {
	case *ast.IntegerLiteral:
		return "int"
	case *ast.FloatLiteral:
		return "float"
	case *ast.StringLiteral:
		return "string"
	case *ast.TemplateLiteral:
		return "string"
	case *ast.Boolean:
		return "bool"
	case *ast.Null:
		return "null"
	case *ast.ArrayLiteral:
		if len(n.Elements) > 0 {
			elemType := sa.inferType(n.Elements[0])
			return fmt.Sprintf("Array[%s]", elemType)
		}
		return "Array"
	case *ast.HashLiteral:
		return "Hash"
	case *ast.FunctionLiteral:
		return buildFnSignature("", n.Parameters, n.ReturnType, false)
	case *ast.AsyncFunctionLiteral:
		return buildFnSignature("", n.Parameters, n.ReturnType, true)
	case *ast.ArrowFunction:
		return buildFnSignature("", n.Parameters, n.ReturnType, false)
	case *ast.StructLiteral:
		if ident, ok := n.Name.(*ast.Identifier); ok {
			return ident.Value
		}
		return "struct"
	case *ast.CallExpression:
		// Attempt to resolve expected return type from function statement
		if ident, ok := n.Function.(*ast.Identifier); ok {
			sym := sa.currentScope.Get(ident.Value)
			if sym != nil && sym.Definition != nil {
				if fs, isFunc := sym.Definition.(*ast.FunctionStatement); isFunc {
					if fs.ReturnType != nil {
						return fs.ReturnType.Value
					}
				}
				if afs, isAsyncFunc := sym.Definition.(*ast.AsyncFunctionStatement); isAsyncFunc {
					if afs.ReturnType != nil {
						return "Promise[" + afs.ReturnType.Value + "]"
					}
					return "Promise"
				}
			}
		}
		return "any"
	case *ast.AwaitExpression:
		inner := sa.inferType(n.Value)
		if strings.HasPrefix(inner, "Promise[") && strings.HasSuffix(inner, "]") {
			return inner[8 : len(inner)-1]
		}
		return "any"
	case *ast.Identifier:
		sym := sa.currentScope.Get(n.Value)
		if sym != nil {
			return sym.Type
		}
		return "any"
	case *ast.InfixExpression:
		leftType := sa.inferType(n.Left)
		rightType := sa.inferType(n.Right)
		if leftType == "float" || rightType == "float" {
			if (leftType == "int" || leftType == "float") && (rightType == "int" || rightType == "float") {
				return "float"
			}
		}
		if leftType == "int" && rightType == "int" {
			return "int"
		}
		if leftType == "string" || rightType == "string" {
			if n.Operator == "+" {
				return "string"
			}
		}
		switch n.Operator {
		case "==", "!=", "<", ">", "<=", ">=", "&&", "||":
			return "bool"
		}
		return leftType
	case *ast.PrefixExpression:
		if n.Operator == "!" {
			return "bool"
		}
		return sa.inferType(n.Right)
	case *ast.TernaryExpression:
		return sa.inferType(n.TrueValue)
	case *ast.NullishCoalescingExpression:
		return sa.inferType(n.Left)
	case *ast.ArrayIndexExpression:
		// We can't know the element type statically in most cases
		arrType := sa.inferType(n.Left)
		if strings.HasPrefix(arrType, "Array[") && strings.HasSuffix(arrType, "]") {
			return arrType[6 : len(arrType)-1]
		}
		return "any"
	}
	return "any"
}

func (sa *SemanticAnalyzer) resolveModulePathToURI(modulePath string) string {
	currentFilePath, err := url.PathUnescape(strings.TrimPrefix(sa.FileURI, "file://"))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to unescape file URI %s: %v", sa.FileURI, err))
		return ""
	}

	currentDir := filepath.Dir(currentFilePath)
	cwd, _ := os.Getwd()

	// List of paths to search, similar to ModuleLoader
	searchPaths := []string{
		currentDir,
		filepath.Join(cwd, "lib"),
		filepath.Join(cwd, "modules"),
		cwd,
	}

	filename := modulePath
	if filepath.Ext(filename) == "" {
		filename += ".jb"
	}

	for _, path := range searchPaths {
		resolvedPath := filepath.Join(path, filename)
		// Try file
		if info, err := os.Stat(resolvedPath); err == nil && !info.IsDir() {
			return "file://" + resolvedPath
		}
		// Try directory/main.jb
		dirPath := filepath.Join(path, modulePath)
		mainPath := filepath.Join(dirPath, "main.jb")
		if info, err := os.Stat(mainPath); err == nil && !info.IsDir() {
			return "file://" + mainPath
		}
	}

	logger.Warning(fmt.Sprintf("Module file not found: %s", modulePath))
	return ""
}

func (sa *SemanticAnalyzer) processImportedModule(moduleURI string) *DocumentSemanticInfo {

	sa.Workspace.Mutex.RLock()
	docInfo, ok := sa.Workspace.Documents[moduleURI]
	sa.Workspace.Mutex.RUnlock()

	if ok && docInfo != nil {
		return docInfo
	}

	moduleFilePath, err := url.PathUnescape(strings.TrimPrefix(moduleURI, "file://"))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to unescape module URI %s: %v", moduleURI, err))
		return nil
	}
	content, err := os.ReadFile(moduleFilePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to read module file %s: %v", moduleFilePath, err))
		return nil
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	moduleSA := NewSemanticAnalyzer(program, sa.Workspace, moduleURI)
	moduleSA.Analyze()

	sa.Errors = append(sa.Errors, moduleSA.Errors...)

	newDocInfo := &DocumentSemanticInfo{
		Program:     program,
		SymbolTable: moduleSA.Symbols,
	}

	sa.Workspace.Mutex.Lock()
	sa.Workspace.Documents[moduleURI] = newDocInfo
	sa.Workspace.Mutex.Unlock()

	return newDocInfo
}

func (sa *SemanticAnalyzer) integrateImportedSymbols(importStmt *ast.ImportStatement, importedDocInfo *DocumentSemanticInfo) {
	sa.Workspace.Mutex.RLock()
	defer sa.Workspace.Mutex.RUnlock()

	if importedDocInfo == nil || importedDocInfo.SymbolTable == nil || importedDocInfo.SymbolTable.RootScope == nil {
		return
	}

	switch importStmt.ImportType {
	case ast.IMPORT_DEFAULT:
		moduleName := strings.TrimSuffix(filepath.Base(importedDocInfo.URI), ".jb")
		if importStmt.DefaultImport != nil {

			moduleName = importStmt.DefaultImport.Value
		}

		moduleSym := &Symbol{
			Name: moduleName,
			Kind: protocol.SymbolKindModule,
			Type: "module",
			Location: protocol.Range{
				Start: protocol.Position{Line: uint32(importStmt.ModuleName.Token.Line - 1), Character: uint32(importStmt.ModuleName.Token.Column - 1)},
				End:   protocol.Position{Line: uint32(importStmt.ModuleName.Token.Line - 1), Character: uint32(importStmt.ModuleName.Token.Column - 1 + len(importStmt.ModuleName.Token.Literal))},
			},
			Definition: importedDocInfo.Program,
			Container:  sa.currentScope,
		}
		sa.currentScope.Set(moduleSym)

	case ast.IMPORT_NAMESPACE:

		namespaceName := importStmt.NamespaceAlias.Value

		namespaceSym := &Symbol{
			Name: namespaceName,
			Kind: protocol.SymbolKindModule,
			Type: "module",
			Location: protocol.Range{
				Start: protocol.Position{Line: uint32(importStmt.NamespaceAlias.Token.Line - 1), Character: uint32(importStmt.NamespaceAlias.Token.Column - 1)},
				End:   protocol.Position{Line: uint32(importStmt.NamespaceAlias.Token.Line - 1), Character: uint32(importStmt.NamespaceAlias.Token.Column - 1 + len(importStmt.NamespaceAlias.Token.Literal))},
			},
			Definition: importedDocInfo.Program,
			Container:  sa.currentScope,
		}
		sa.currentScope.Set(namespaceSym)

		namespaceScope := NewScope(sa.currentScope, namespaceSym.Definition)

		for _, sym := range importedDocInfo.SymbolTable.RootScope.Symbols {

			newSym := *sym
			newSym.Container = namespaceScope
			namespaceScope.Set(&newSym)
		}
		sa.currentScope.Children = append(sa.currentScope.Children, namespaceScope)

	case ast.IMPORT_NAMED:
		for _, item := range importStmt.NamedImports {
			originalName := item.Name.Value
			aliasName := originalName
			if item.Alias != nil {
				aliasName = item.Alias.Value
			}

			symbol := importedDocInfo.SymbolTable.RootScope.Get(originalName)
			if symbol != nil {
				newSym := *symbol
				newSym.Name = aliasName
				newSym.Location = protocol.Range{
					Start: protocol.Position{Line: uint32(item.Name.Token.Line - 1), Character: uint32(item.Name.Token.Column - 1)},
					End:   protocol.Position{Line: uint32(item.Name.Token.Line - 1), Character: uint32(item.Name.Token.Column - 1 + len(item.Name.Token.Literal))},
				}
				newSym.Container = sa.currentScope
				sa.currentScope.Set(&newSym)
			} else {
				sa.Errors = append(sa.Errors, fmt.Sprintf("line %d, column %d: symbol '%s' not found in module '%s'", item.Name.Token.Line, item.Name.Token.Column, originalName, importStmt.ModuleName.Value))
			}
		}

	case ast.IMPORT_MIXED:

		moduleName := strings.TrimSuffix(filepath.Base(importedDocInfo.URI), ".jb")
		if importStmt.DefaultImport != nil {
			moduleName = importStmt.DefaultImport.Value
		}
		moduleSym := &Symbol{
			Name: moduleName,
			Kind: protocol.SymbolKindModule,
			Type: "module",
			Location: protocol.Range{
				Start: protocol.Position{Line: uint32(importStmt.DefaultImport.Token.Line - 1), Character: uint32(importStmt.DefaultImport.Token.Column - 1)},
				End:   protocol.Position{Line: uint32(importStmt.DefaultImport.Token.Line - 1), Character: uint32(importStmt.DefaultImport.Token.Column - 1 + len(importStmt.DefaultImport.Token.Literal))},
			},
			Definition: importedDocInfo.Program,
			Container:  sa.currentScope,
		}
		sa.currentScope.Set(moduleSym)

		for _, item := range importStmt.NamedImports {
			originalName := item.Name.Value
			aliasName := originalName
			if item.Alias != nil {
				aliasName = item.Alias.Value
			}

			symbol := importedDocInfo.SymbolTable.RootScope.Get(originalName)
			if symbol != nil {
				newSym := *symbol
				newSym.Name = aliasName
				newSym.Location = protocol.Range{
					Start: protocol.Position{Line: uint32(item.Name.Token.Line - 1), Character: uint32(item.Name.Token.Column - 1)},
					End:   protocol.Position{Line: uint32(item.Name.Token.Line - 1), Character: uint32(item.Name.Token.Column - 1 + len(item.Name.Token.Literal))},
				}
				newSym.Container = sa.currentScope
				sa.currentScope.Set(&newSym)
			} else {
				sa.Errors = append(sa.Errors, fmt.Sprintf("line %d, column %d: symbol '%s' not found in module '%s'", item.Name.Token.Line, item.Name.Token.Column, originalName, importStmt.ModuleName.Value))
			}
		}

	case ast.IMPORT_SIDE_EFFECT:
	}
}

func (sa *SemanticAnalyzer) declareSymbol(name string, kind SymbolKind, typ string, tok token.Token, definition ast.Node) {
	startLine := uint32(tok.Line - 1)
	startCol := uint32(tok.Column - 1)

	symbol := &Symbol{
		Name: name,
		Kind: kind,
		Type: typ,
		Location: protocol.Range{
			Start: protocol.Position{Line: startLine, Character: startCol},
			End:   protocol.Position{Line: startLine, Character: startCol + uint32(len(tok.Literal))},
		},
		Definition: definition,
	}

	if err := sa.currentScope.Set(symbol); err != nil {
		sa.Errors = append(sa.Errors, fmt.Sprintf("line %d, column %d: %s", tok.Line, tok.Column, err.Error()))
	}
}
