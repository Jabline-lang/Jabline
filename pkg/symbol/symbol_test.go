package symbol

import (
	"testing"
)

func TestNewSymbolTable(t *testing.T) {
	st := NewSymbolTable()
	if st == nil {
		t.Fatal("NewSymbolTable() returned nil")
	}
	if st.Outer != nil {
		t.Errorf("expected nil Outer")
	}
	if st.NumDefinitions() != 0 {
		t.Errorf("expected 0 definitions, got %d", st.NumDefinitions())
	}
	if len(st.FreeSymbols) != 0 {
		t.Errorf("expected 0 free symbols, got %d", len(st.FreeSymbols))
	}
}

func TestDefine(t *testing.T) {
	st := NewSymbolTable()
	sym := st.Define("a")
	if sym.Name != "a" {
		t.Errorf("Name = %q, want %q", sym.Name, "a")
	}
	if sym.Scope != GlobalScope {
		t.Errorf("Scope = %v, want %v", sym.Scope, GlobalScope)
	}
	if sym.Index != 0 {
		t.Errorf("Index = %d, want 0", sym.Index)
	}
	if sym.IsConst {
		t.Errorf("expected IsConst = false")
	}

	sym2 := st.Define("b")
	if sym2.Index != 1 {
		t.Errorf("Index = %d, want 1", sym2.Index)
	}
}

func TestDefineLocalScope(t *testing.T) {
	outer := NewSymbolTable()
	inner := NewEnclosedSymbolTable(outer)
	inner.IsFunctionScope = true

	sym := inner.Define("localVar")
	if sym.Scope != LocalScope {
		t.Errorf("Scope = %v, want %v", sym.Scope, LocalScope)
	}
}

func TestDefineWithType(t *testing.T) {
	st := NewSymbolTable()
	sym := st.DefineWithType("x", "int")
	if sym.DataType != "int" {
		t.Errorf("DataType = %q, want %q", sym.DataType, "int")
	}
}

func TestDefineConst(t *testing.T) {
	st := NewSymbolTable()
	sym := st.DefineConst("PI")
	if !sym.IsConst {
		t.Errorf("expected IsConst = true")
	}
	if sym.Scope != GlobalScope {
		t.Errorf("Scope = %v, want %v", sym.Scope, GlobalScope)
	}
}

func TestDefineConstWithType(t *testing.T) {
	st := NewSymbolTable()
	sym := st.DefineConstWithType("PI", "float")
	if !sym.IsConst {
		t.Errorf("expected IsConst = true")
	}
	if sym.DataType != "float" {
		t.Errorf("DataType = %q, want %q", sym.DataType, "float")
	}
}

func TestIsConstant(t *testing.T) {
	st := NewSymbolTable()
	st.Define("x")
	st.DefineConst("PI")
	if st.IsConstant("x") {
		t.Errorf("expected x to not be constant")
	}
	if !st.IsConstant("PI") {
		t.Errorf("expected PI to be constant")
	}
	if st.IsConstant("nonexistent") {
		t.Errorf("expected nonexistent to not be constant")
	}
}

func TestResolveGlobal(t *testing.T) {
	st := NewSymbolTable()
	st.Define("a")
	sym, ok := st.Resolve("a")
	if !ok {
		t.Fatal("expected to resolve 'a'")
	}
	if sym.Scope != GlobalScope {
		t.Errorf("Scope = %v, want %v", sym.Scope, GlobalScope)
	}
}

func TestResolveLocal(t *testing.T) {
	outer := NewSymbolTable()
	outer.IsFunctionScope = true
	inner := NewEnclosedSymbolTable(outer)
	inner.IsFunctionScope = true

	inner.Define("localVar")
	sym, ok := inner.Resolve("localVar")
	if !ok {
		t.Fatal("expected to resolve 'localVar'")
	}
	if sym.Scope != LocalScope {
		t.Errorf("Scope = %v, want %v", sym.Scope, LocalScope)
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	outer := NewEnclosedSymbolTable(global)
	outer.IsFunctionScope = true
	outer.Define("b")

	inner := NewEnclosedSymbolTable(outer)
	inner.IsFunctionScope = true
	inner.Define("c")

	obj, ok := inner.Resolve("a")
	if !ok {
		t.Fatal("expected to resolve 'a' from inner scope")
	}
	if obj.Scope != GlobalScope {
		t.Errorf("'a' Scope = %v, want %v", obj.Scope, GlobalScope)
	}

	obj, ok = inner.Resolve("b")
	if !ok {
		t.Fatal("expected to resolve 'b' from inner scope")
	}
	if obj.Scope != FreeScope {
		t.Errorf("'b' Scope = %v, want %v", obj.Scope, FreeScope)
	}
}

func TestResolveNotFound(t *testing.T) {
	st := NewSymbolTable()
	_, ok := st.Resolve("nonexistent")
	if ok {
		t.Errorf("expected not to resolve nonexistent symbol")
	}
}

func TestDefineBuiltin(t *testing.T) {
	st := NewSymbolTable()
	sym := st.DefineBuiltin(0, "len")
	if sym.Scope != BuiltinScope {
		t.Errorf("Scope = %v, want %v", sym.Scope, BuiltinScope)
	}
	if sym.Index != 0 {
		t.Errorf("Index = %d, want 0", sym.Index)
	}

	resolved, ok := st.Resolve("len")
	if !ok {
		t.Fatal("expected to resolve 'len'")
	}
	if resolved.Scope != BuiltinScope {
		t.Errorf("Scope = %v, want %v", resolved.Scope, BuiltinScope)
	}
}

func TestDefineBuiltinInNestedScope(t *testing.T) {
	global := NewSymbolTable()
	global.DefineBuiltin(0, "len")
	inner := NewEnclosedSymbolTable(global)
	inner.IsFunctionScope = true

	resolved, ok := inner.Resolve("len")
	if !ok {
		t.Fatal("expected to resolve 'len' from inner scope")
	}
	if resolved.Scope != BuiltinScope {
		t.Errorf("Scope = %v, want %v", resolved.Scope, BuiltinScope)
	}
}

func TestDefineFunctionName(t *testing.T) {
	st := NewSymbolTable()
	sym := st.DefineFunctionName("myFunc")
	if sym.Scope != FunctionScope {
		t.Errorf("Scope = %v, want %v", sym.Scope, FunctionScope)
	}
	if sym.Index != 0 {
		t.Errorf("Index = %d, want 0", sym.Index)
	}

	resolved, ok := st.Resolve("myFunc")
	if !ok {
		t.Fatal("expected to resolve 'myFunc'")
	}
	if resolved.Scope != FunctionScope {
		t.Errorf("Scope = %v, want %v", resolved.Scope, FunctionScope)
	}
}

func TestFreeSymbols(t *testing.T) {
	outer := NewSymbolTable()
	outer.IsFunctionScope = true
	outer.Define("a")
	outer.Define("b")

	inner := NewEnclosedSymbolTable(outer)
	inner.IsFunctionScope = true
	inner.Define("c")

	_, ok := inner.Resolve("a")
	if !ok {
		t.Fatal("expected to resolve 'a'")
	}
	_, ok = inner.Resolve("b")
	if !ok {
		t.Fatal("expected to resolve 'b'")
	}
	if len(inner.FreeSymbols) != 2 {
		t.Errorf("expected 2 free symbols, got %d", len(inner.FreeSymbols))
	}

	if inner.FreeSymbols[0].Name != "a" {
		t.Errorf("FreeSymbols[0].Name = %q, want %q", inner.FreeSymbols[0].Name, "a")
	}
}

func TestDefineType(t *testing.T) {
	st := NewSymbolTable()
	sym := st.DefineType("MyStruct")
	if sym.Scope != TypeScope {
		t.Errorf("Scope = %v, want %v", sym.Scope, TypeScope)
	}
	if sym.Index != 0 {
		t.Errorf("Index = %d, want 0", sym.Index)
	}
}

func TestMarkExported(t *testing.T) {
	st := NewSymbolTable()
	st.Define("x")
	st.MarkExported("x")

	resolved, ok := st.Resolve("x")
	if !ok {
		t.Fatal("expected to resolve 'x'")
	}
	if !resolved.IsExported {
		t.Errorf("expected x to be exported")
	}
}

func TestGetStore(t *testing.T) {
	st := NewSymbolTable()
	st.Define("a")
	store := st.GetStore()
	if len(store) != 1 {
		t.Errorf("expected store size 1, got %d", len(store))
	}
	if _, ok := store["a"]; !ok {
		t.Errorf("expected 'a' in store")
	}
}

func TestIsGlobalScope(t *testing.T) {
	global := NewSymbolTable()
	if !global.IsGlobalScope() {
		t.Errorf("expected root table to be global scope")
	}

	inner := NewEnclosedSymbolTable(global)
	if !inner.IsGlobalScope() {
		t.Errorf("expected enclosed table with global outer to BE global scope")
	}

	inner.IsFunctionScope = true
	if inner.IsGlobalScope() {
		t.Errorf("expected function scope to NOT be global scope")
	}
}

func TestNestedScopeResolution(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	first := NewEnclosedSymbolTable(global)
	first.IsFunctionScope = true
	first.Define("c")

	second := NewEnclosedSymbolTable(first)
	second.IsFunctionScope = true
	second.Define("d")

	tests := []struct {
		name     string
		scope    SymbolScope
		free     bool
	}{
		{"a", GlobalScope, false},
		{"b", GlobalScope, false},
		{"c", FreeScope, true},
		{"d", LocalScope, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sym, ok := second.Resolve(tt.name)
			if !ok {
				t.Fatalf("failed to resolve %q", tt.name)
			}
			if sym.Scope != tt.scope {
				t.Errorf("Scope = %v, want %v", sym.Scope, tt.scope)
			}
		})
	}
	if len(second.FreeSymbols) != 1 {
		t.Errorf("expected 1 free symbol, got %d", len(second.FreeSymbols))
	}
}

func TestIsConstantNested(t *testing.T) {
	global := NewSymbolTable()
	global.DefineConst("PI")
	global.Define("x")

	inner := NewEnclosedSymbolTable(global)
	if !inner.IsConstant("PI") {
		t.Errorf("expected PI to be constant in nested scope")
	}
	if inner.IsConstant("x") {
		t.Errorf("expected x to not be constant")
	}
}
