package compiler

import (
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
	"testing"
)

type compilerTestCase struct {
	input    string
	constant interface{}
}

func parseAndCompile(t *testing.T, input string) *Compiler {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	c := New()
	err := c.Compile(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return c
}

func TestIntegerConstants(t *testing.T) {
	tests := []compilerTestCase{
		{input: `1;`, constant: int64(1)},
		{input: `42;`, constant: int64(42)},
		{input: `1 + 2;`, constant: int64(3)},
	}
	for _, tt := range tests {
		c := parseAndCompile(t, tt.input)
		bc := c.Bytecode()
		if len(bc.Constants) == 0 {
			t.Fatalf("expected at least 1 constant, got 0")
		}
		constant := bc.Constants[0]
		integer, ok := constant.(*object.Integer)
		if !ok {
			t.Fatalf("expected *object.Integer, got %T", constant)
		}
		if integer.Value != tt.constant.(int64) {
			t.Errorf("constant value: expected %d, got %d", tt.constant.(int64), integer.Value)
		}
	}
}

func TestStringConstants(t *testing.T) {
	c := parseAndCompile(t, `"hello";`)
	bc := c.Bytecode()
	if len(bc.Constants) < 1 {
		t.Fatalf("expected at least 1 constant, got %d", len(bc.Constants))
	}
	str, ok := bc.Constants[0].(*object.String)
	if !ok {
		t.Fatalf("expected *object.String, got %T", bc.Constants[0])
	}
	if str.Value != "hello" {
		t.Errorf("string constant: expected %q, got %q", "hello", str.Value)
	}
}

func TestSourceMap(t *testing.T) {
	c := parseAndCompile(t, `let x = 1;`)
	bc := c.Bytecode()
	if bc.SourceMap == nil {
		t.Fatal("expected SourceMap to be non-nil")
	}
}

func TestSymbolTable(t *testing.T) {
	c := parseAndCompile(t, `let x = 5;`)
	st := c.GetSymbolTable()
	if st == nil {
		t.Fatal("expected symbol table to be non-nil")
	}
}

func TestLetConstCompilation(t *testing.T) {
	tests := []string{
		`let x = 5;`,
		`const x = 10;`,
		`let x = 5; let y = x;`,
	}
	for _, input := range tests {
		c := parseAndCompile(t, input)
		if len(c.Bytecode().Instructions) == 0 {
			t.Errorf("expected non-empty instructions for: %s", input)
		}
	}
}

func TestFunctionCompilation(t *testing.T) {
	tests := []string{
		`fn add(a, b) { return a + b; };`,
	}
	for _, input := range tests {
		c := parseAndCompile(t, input)
		bc := c.Bytecode()
		// CompiledFunctions may not be the first constant due to constant folding.
		// Check that at least one constant is a CompiledFunction.
		found := false
		for _, constant := range bc.Constants {
			if _, ok := constant.(*object.CompiledFunction); ok {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected at least one *object.CompiledFunction in constants for: %s", input)
		}
	}
}

func TestConditionalCompilation(t *testing.T) {
	tests := []string{
		// These use boolean literals, which the compiler folds away at compile-time,
		// and the optimizer may eliminate the remaining dead OpConstant+OpPop pairs.
		// Use non-constant conditions to ensure real bytecode is emitted.
		`let x = 1; if (x) { 1; } else { 2; };`,
		`let x = 0; if (x) { 1; };`,
		`if (1 < 2) { 1; } else { 2; };`,
	}
	for _, input := range tests {
		c := parseAndCompile(t, input)
		bc := c.Bytecode()
		if len(bc.Instructions) == 0 {
			t.Errorf("expected non-empty instructions for: %s", input)
		}
	}
}

func TestLoopCompilation(t *testing.T) {
	tests := []string{
		`while (true) { break; };`,
		`while (true) { continue; };`,
	}
	for _, input := range tests {
		c := parseAndCompile(t, input)
		bc := c.Bytecode()
		if len(bc.Instructions) == 0 {
			t.Errorf("expected non-empty instructions for: %s", input)
		}
	}
}

func TestCompileErrors(t *testing.T) {
	l := lexer.New(`break;`)
	p := parser.New(l)
	prog := p.ParseProgram()
	c := New()
	err := c.Compile(prog)
	if err == nil {
		t.Errorf("expected compile error for break outside loop")
	}
}

func TestEmptyProgram(t *testing.T) {
	c := parseAndCompile(t, ``)
	bc := c.Bytecode()
	if len(bc.Instructions) != 0 {
		t.Errorf("expected empty instructions, got %d", len(bc.Instructions))
	}
}

func TestPostfixCompilation(t *testing.T) {
	tests := []string{
		`let x = 1; x++;`,
		`let x = 1; x--;`,
	}
	for _, input := range tests {
		c := parseAndCompile(t, input)
		bc := c.Bytecode()
		if len(bc.Instructions) == 0 {
			t.Errorf("expected non-empty instructions for: %s", input)
		}
	}
}

func TestArrayCompilation(t *testing.T) {
	tests := []string{
		`[];`,
		`[1, 2, 3];`,
	}
	for _, input := range tests {
		c := parseAndCompile(t, input)
		bc := c.Bytecode()
		if len(bc.Instructions) == 0 {
			t.Errorf("expected non-empty instructions for: %s", input)
		}
	}
}

func TestHashCompilation(t *testing.T) {
	c := parseAndCompile(t, `{};`)
	bc := c.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions for hash literal")
	}
}

func TestIndexCompilation(t *testing.T) {
	c := parseAndCompile(t, `[1, 2][0];`)
	bc := c.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestExportCompilation(t *testing.T) {
	c := parseAndCompile(t, `export fn hello() { return "hi"; }`)
	bc := c.Bytecode()
	if bc.Exports == nil {
		t.Fatal("expected exports map to be non-nil")
	}
}

func TestFunctionWithReturnType(t *testing.T) {
	input := `fn add(a: int, b: int): int { return a + b; };`
	c := parseAndCompile(t, input)
	bc := c.Bytecode()
	if len(bc.Constants) == 0 {
		t.Fatal("expected function constant")
	}
}
