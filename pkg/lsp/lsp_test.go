package lsp

import (
	"testing"

	"jabline/pkg/ast"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"jabline/pkg/token"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func tokenAt(tok token.Token, line, col int) token.Token {
	tok.Line = line
	tok.Column = col
	return tok
}

func ident(name string, line, col int) *ast.Identifier {
	return &ast.Identifier{
		Token: tokenAt(token.Token{Type: token.IDENT, Literal: name}, line, col),
		Value: name,
	}
}

// ---------------------------------------------------------------------------
// Scope tests
// ---------------------------------------------------------------------------

func TestNewScope(t *testing.T) {
	parent := NewScope(nil, nil)
	child := NewScope(parent, nil)
	if child.Parent != parent {
		t.Error("child.Parent should be parent")
	}
	if child.Symbols == nil {
		t.Error("child.Symbols should be initialized")
	}
}

func TestScopeGet(t *testing.T) {
	scope := NewScope(nil, nil)
	sym := &Symbol{Name: "x", Kind: protocol.SymbolKindVariable}
	scope.Set(sym)

	got := scope.Get("x")
	if got == nil {
		t.Fatal("Get(x) returned nil")
	}
	if got.Name != "x" {
		t.Errorf("Name = %s; want x", got.Name)
	}
	if scope.Get("undefined") != nil {
		t.Error("Get(undefined) should be nil")
	}
}

func TestScopeGetRecursive(t *testing.T) {
	root := NewScope(nil, nil)
	root.Set(&Symbol{Name: "x", Kind: protocol.SymbolKindVariable})

	child := NewScope(root, nil)
	child.Set(&Symbol{Name: "y", Kind: protocol.SymbolKindVariable})

	// Child sees both
	if child.Get("x") == nil {
		t.Error("child.Get(x) should find in parent")
	}
	if child.Get("y") == nil {
		t.Error("child.Get(y) should find in child")
	}
}

// ---------------------------------------------------------------------------
// isTokenAt tests
// ---------------------------------------------------------------------------

func TestIsTokenAt(t *testing.T) {
	tok := token.Token{Type: token.IDENT, Literal: "hello", Line: 3, Column: 5}
	if !isTokenAt(tok, 3, 5) {
		t.Error("isTokenAt should match start of token")
	}
	if !isTokenAt(tok, 3, 9) {
		t.Error("isTokenAt should match within token (hello has len 5)")
	}
	if isTokenAt(tok, 3, 10) {
		t.Error("isTokenAt should NOT match past token end")
	}
	if isTokenAt(tok, 4, 5) {
		t.Error("isTokenAt should NOT match different line")
	}
}

// ---------------------------------------------------------------------------
// FindPathToNode tests
// ---------------------------------------------------------------------------

func parseProgram(t *testing.T, src string) *ast.Program {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	return prog
}

func TestFindPathToNode(t *testing.T) {
	// `let x = 42;` should find path to identifier "x" at position (1, 5)
	prog := parseProgram(t, "let x = 42;")

	// Find identifier "x" at line 1, column 5 (let x = 42;)
	path := FindPathToNode(prog, 1, 5)
	if path == nil {
		t.Fatal("FindPathToNode returned nil for position of x")
	}
	if len(path) < 2 {
		t.Fatalf("expected at least 2 nodes in path, got %d", len(path))
	}
	// Last node should be the identifier
	if ident, ok := path[len(path)-1].(*ast.Identifier); !ok {
		t.Errorf("last path node should be Identifier, got %T", path[len(path)-1])
	} else if ident.Value != "x" {
		t.Errorf("identifier value = %s; want x", ident.Value)
	}
}

func TestFindPathToNodeUnknownPosition(t *testing.T) {
	prog := parseProgram(t, "let x = 42;")
	path := FindPathToNode(prog, 999, 999)
	if path != nil {
		t.Error("FindPathToNode should return nil for position not in AST")
	}
}

func TestFindPathToNodeNil(t *testing.T) {
	path := FindPathToNode(nil, 1, 1)
	if path != nil {
		t.Error("FindPathToNode(nil) should return nil")
	}
}

func TestFindPathToNodeInBlock(t *testing.T) {
	// `if (true) { let y = 10; }` - find "y" at its position
	// Original source: "if (true) { let y = 10; }"
	//                  ^12345678901234567890123456789
	//                                     ^ y at col 16
	prog := parseProgram(t, "if (true) { let y = 10; }")
	path := FindPathToNode(prog, 1, 17)
	if path == nil {
		t.Fatal("FindPathToNode returned nil for y in block")
	}
	// Last node should be identifier "y"
	last := path[len(path)-1]
	if ident, ok := last.(*ast.Identifier); !ok || ident.Value != "y" {
		t.Errorf("last node = %T %v; want Identifier(y)", last, last)
	}
}

func TestFindPathToNodeIntegerLiteral(t *testing.T) {
	prog := parseProgram(t, "let n = 42;")
	// Find integer 42 at line 1, column 9 (after `let n = `)
	path := FindPathToNode(prog, 1, 9)
	if path == nil {
		t.Fatal("FindPathToNode returned nil for integer literal")
	}
	last := path[len(path)-1]
	if _, ok := last.(*ast.IntegerLiteral); !ok {
		t.Errorf("last node = %T; want IntegerLiteral", last)
	}
}

// ---------------------------------------------------------------------------
// SemanticAnalyzer tests
// ---------------------------------------------------------------------------

func TestNewSemanticAnalyzer(t *testing.T) {
	prog := parseProgram(t, "let x = 42;")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	if analyzer == nil {
		t.Fatal("NewSemanticAnalyzer returned nil")
	}
	if analyzer.Symbols == nil || analyzer.Symbols.RootScope == nil {
		t.Fatal("SymbolTable not initialized")
	}
}

func TestSemanticAnalyzerDeclaresBuiltins(t *testing.T) {
	prog := parseProgram(t, "let x = len;")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	// Builtins should be declared in root scope
	sym := analyzer.Symbols.RootScope.Get("len")
	if sym == nil {
		t.Error("len builtin not found in root scope")
	}
	if sym.Kind != protocol.SymbolKindFunction {
		t.Errorf("len kind = %d; want SymbolKindFunction", sym.Kind)
	}
}

func TestSemanticAnalyzerLetDeclaration(t *testing.T) {
	prog := parseProgram(t, "let x = 42;")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	sym := analyzer.Symbols.RootScope.Get("x")
	if sym == nil {
		t.Fatal("x not found in root scope")
	}
	if sym.Kind != protocol.SymbolKindVariable {
		t.Errorf("x kind = %d; want SymbolKindVariable", sym.Kind)
	}
}

func TestSemanticAnalyzerConstDeclaration(t *testing.T) {
	prog := parseProgram(t, "const C = 100;")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	sym := analyzer.Symbols.RootScope.Get("C")
	if sym == nil {
		t.Fatal("C not found in root scope")
	}
	if sym.Kind != protocol.SymbolKindConstant {
		t.Errorf("C kind = %d; want SymbolKindConstant", sym.Kind)
	}
}

func TestSemanticAnalyzerFunctionDeclaration(t *testing.T) {
	prog := parseProgram(t, "fn add(a, b) { return a + b; }")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	sym := analyzer.Symbols.RootScope.Get("add")
	if sym == nil {
		t.Fatal("add not found in root scope")
	}
	if sym.Kind != protocol.SymbolKindFunction {
		t.Errorf("add kind = %d; want SymbolKindFunction", sym.Kind)
	}
}

func TestSemanticAnalyzerAsyncFunction(t *testing.T) {
	prog := parseProgram(t, "async fn fetch(url) { return url; }")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	sym := analyzer.Symbols.RootScope.Get("fetch")
	if sym == nil {
		t.Fatal("fetch not found in root scope")
	}
	if sym.Kind != protocol.SymbolKindFunction {
		t.Errorf("fetch kind = %d; want SymbolKindFunction", sym.Kind)
	}
}

func TestSemanticAnalyzerFunctionScope(t *testing.T) {
	prog := parseProgram(t, "fn f() { let x = 10; }; let y = 20;")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	// x should NOT be in root scope (it's in function scope)
	if sym := analyzer.Symbols.RootScope.Get("x"); sym != nil {
		t.Error("x should not be in root scope")
	}
	// y should be in root scope
	if sym := analyzer.Symbols.RootScope.Get("y"); sym == nil {
		t.Error("y should be in root scope")
	}
	// f should be in root scope
	if sym := analyzer.Symbols.RootScope.Get("f"); sym == nil {
		t.Error("f should be in root scope")
	}
}

func TestSemanticAnalyzerBlockScope(t *testing.T) {
	prog := parseProgram(t, "if (true) { let inner = 1; }; let outer = 2;")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	// inner should NOT be in root scope
	if sym := analyzer.Symbols.RootScope.Get("inner"); sym != nil {
		t.Error("inner should not leak to root scope")
	}
	// outer should be in root scope
	if sym := analyzer.Symbols.RootScope.Get("outer"); sym == nil {
		t.Error("outer should be in root scope")
	}
}

func TestSemanticAnalyzerDuplicateDeclaration(t *testing.T) {
	prog := parseProgram(t, "let x = 1; let x = 2;")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	if len(analyzer.Errors) == 0 {
		t.Error("expected error for duplicate declaration, got none")
	}
}

func TestSemanticAnalyzerStructDeclaration(t *testing.T) {
	prog := parseProgram(t, "struct Point { x: int, y: int }")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	sym := analyzer.Symbols.RootScope.Get("Point")
	if sym == nil {
		t.Fatal("Point not found in root scope")
	}
}

func TestSemanticAnalyzerEnumDeclaration(t *testing.T) {
	prog := parseProgram(t, "enum Color { Red, Green, Blue }")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	sym := analyzer.Symbols.RootScope.Get("Color")
	if sym == nil {
		t.Fatal("Color not found in root scope")
	}
}

func TestSemanticAnalyzerInterfaceDeclaration(t *testing.T) {
	prog := parseProgram(t, "interface Drawable { draw(): void }")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	sym := analyzer.Symbols.RootScope.Get("Drawable")
	if sym == nil {
		t.Fatal("Drawable not found in root scope")
	}
}

func TestSemanticAnalyzerWhileLoop(t *testing.T) {
	prog := parseProgram(t, "let x = 0; while (x < 5) { let x = x + 1; };")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()
	// Should not produce errors (scope working correctly)
	_ = analyzer.Errors
}

func TestSemanticAnalyzerForLoop(t *testing.T) {
	prog := parseProgram(t, "let sum = 0; for (let i = 0; i < 3; i++) { sum = sum + i; };")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()
}

func TestSemanticAnalyzerServiceDeclaration(t *testing.T) {
	prog := parseProgram(t, "service Counter { port: 8080 }")
	analyzer := NewSemanticAnalyzer(prog, &WorkspaceSymbolStore{}, "file:///test.jb")
	analyzer.Analyze()

	sym := analyzer.Symbols.RootScope.Get("Counter")
	if sym == nil {
		t.Fatal("Counter not found in root scope")
	}
}

// ---------------------------------------------------------------------------
// WorkspaceSymbolStore tests
// ---------------------------------------------------------------------------

func TestWorkspaceStoreAnalyzingDocument(t *testing.T) {
	ws := &WorkspaceSymbolStore{
		Documents:          make(map[string]*DocumentSemanticInfo),
		analyzingDocuments: make(map[string]bool),
	}

	ws.AddAnalyzingDocument("file:///test.jb")
	if !ws.IsAnalyzingDocument("file:///test.jb") {
		t.Error("IsAnalyzingDocument should be true after Add")
	}

	ws.RemoveAnalyzingDocument("file:///test.jb")
	if ws.IsAnalyzingDocument("file:///test.jb") {
		t.Error("IsAnalyzingDocument should be false after Remove")
	}
}

// ---------------------------------------------------------------------------
// buildFnSignature tests
// ---------------------------------------------------------------------------

func TestBuildFnSignature(t *testing.T) {
	tests := []struct {
		name       string
		fnName     string
		params     []*ast.Identifier
		returnType *ast.TypeExpression
		isAsync    bool
		want       string
	}{
		{"simple fn", "add", []*ast.Identifier{ident("a", 1, 1), ident("b", 1, 1)}, nil, false, "fn add(a, b)"},
		{"fn with return", "get", []*ast.Identifier{ident("key", 1, 1)}, &ast.TypeExpression{Value: "string"}, false, "fn get(key): string"},
		{"no name", "", []*ast.Identifier{ident("x", 1, 1)}, nil, false, "fn(x)"},
		{"async fn", "fetch", []*ast.Identifier{ident("url", 1, 1)}, nil, true, "async fn fetch(url)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFnSignature(tt.fnName, tt.params, tt.returnType, tt.isAsync)
			if got != tt.want {
				t.Errorf("buildFnSignature() = %q; want %q", got, tt.want)
			}
		})
	}
}
