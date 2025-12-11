package lsp

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/token"
	"os"
	"path/filepath"
	"net/url"
	"strings"

	"jabline/pkg/lexer"
	"jabline/pkg/parser"

	"github.com/tliron/glsp/protocol_3_16"
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
	Program *ast.Program
	Errors  []string
	Symbols *SymbolTable
	currentScope *Scope
	Workspace    *WorkspaceSymbolStore
	FileURI      string
}

func NewSemanticAnalyzer(program *ast.Program, ws *WorkspaceSymbolStore, fileURI string) *SemanticAnalyzer {

	rootScope := NewScope(nil, program)
	
	sa := &SemanticAnalyzer{
		Program: program,
		Symbols: &SymbolTable{RootScope: rootScope},
		currentScope: rootScope,
		Workspace: ws,
		FileURI: fileURI,
	}

	return sa
}

func (sa *SemanticAnalyzer) Analyze() {

	for _, stmt := range sa.Program.Statements {
		sa.walk(stmt)
	}
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

		sa.currentScope = oldScope // Exit scope

	case *ast.LetStatement:
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindVariable, "any", n.Name.Token, n.Name)
		if n.Value != nil {
			sa.walk(n.Value)
		}

	case *ast.ConstStatement:
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindConstant, "any", n.Name.Token, n.Name)
		if n.Value != nil {
			sa.walk(n.Value)
		}

	case *ast.FunctionStatement:
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindFunction, "fn", n.Name.Token, n.Name)

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
		sa.declareSymbol(n.Name.Value, protocol.SymbolKindStruct, "struct", n.Name.Token, n.Name)

	case *ast.ExpressionStatement:
		sa.walk(n.Expression)
	case *ast.ReturnStatement:
		if n.ReturnValue != nil {
			sa.walk(n.ReturnValue)
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
	case *ast.CallExpression:
		sa.walk(n.Function)
		for _, arg := range n.Arguments {
			sa.walk(arg)
		}
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
			case *ast.ImportStatement:
				logger.Info(fmt.Sprintf("SemanticAnalyzer: Encountered ImportStatement for module: %s", n.ModuleName.Value))
				
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
					}
				}
			}

func (sa *SemanticAnalyzer) resolveModulePathToURI(modulePath string) string {

	currentFilePath, err := url.PathUnescape(strings.TrimPrefix(sa.FileURI, "file://"))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to unescape file URI %s: %v", sa.FileURI, err))
		return ""
	}
	
	currentDir := filepath.Dir(currentFilePath)

	resolvedPath := filepath.Join(currentDir, modulePath)
	
	if filepath.Ext(resolvedPath) == "" {
		resolvedPath += ".jb"
	}

	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		logger.Warning(fmt.Sprintf("Module file not found: %s (resolved from %s)", resolvedPath, modulePath))
		return ""
	}

	return "file://" + url.PathEscape(resolvedPath)
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
				Start: protocol.Position{Line: uint32(importStmt.ModuleName.Token.Line-1), Character: uint32(importStmt.ModuleName.Token.Column-1)},
				End:   protocol.Position{Line: uint32(importStmt.ModuleName.Token.Line-1), Character: uint32(importStmt.ModuleName.Token.Column-1 + len(importStmt.ModuleName.Token.Literal))},
			},
			Definition: importedDocInfo.Program,
			Container: sa.currentScope,
		}
		sa.currentScope.Set(moduleSym)

	case ast.IMPORT_NAMESPACE:

		namespaceName := importStmt.NamespaceAlias.Value
		
		namespaceSym := &Symbol{
			Name: namespaceName,
			Kind: protocol.SymbolKindModule,
			Type: "module",
			Location: protocol.Range{
				Start: protocol.Position{Line: uint32(importStmt.NamespaceAlias.Token.Line-1), Character: uint32(importStmt.NamespaceAlias.Token.Column-1)},
				End:   protocol.Position{Line: uint32(importStmt.NamespaceAlias.Token.Line-1), Character: uint32(importStmt.NamespaceAlias.Token.Column-1 + len(importStmt.NamespaceAlias.Token.Literal))},
			},
			Definition: importedDocInfo.Program,
			Container: sa.currentScope,
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
					Start: protocol.Position{Line: uint32(item.Name.Token.Line-1), Character: uint32(item.Name.Token.Column-1)},
					End:   protocol.Position{Line: uint32(item.Name.Token.Line-1), Character: uint32(item.Name.Token.Column-1 + len(item.Name.Token.Literal))},
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
				Start: protocol.Position{Line: uint32(importStmt.DefaultImport.Token.Line-1), Character: uint32(importStmt.DefaultImport.Token.Column-1)},
				End:   protocol.Position{Line: uint32(importStmt.DefaultImport.Token.Line-1), Character: uint32(importStmt.DefaultImport.Token.Column-1 + len(importStmt.DefaultImport.Token.Literal))},
			},
			Definition: importedDocInfo.Program,
			Container: sa.currentScope,
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
					Start: protocol.Position{Line: uint32(item.Name.Token.Line-1), Character: uint32(item.Name.Token.Column-1)},
					End:   protocol.Position{Line: uint32(item.Name.Token.Line-1), Character: uint32(item.Name.Token.Column-1 + len(item.Name.Token.Literal))},
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
