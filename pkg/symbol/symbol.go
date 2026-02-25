package symbol

import "jabline/pkg/object"

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
	TypeScope     SymbolScope = "TYPE"
)

type Symbol struct {
	Name       string
	Scope      SymbolScope
	Index      int
	IsExported bool
	IsConst    bool          // true when declared with 'const'
	Value      object.Object // stores the actual object for constants/structs
	DataType   string        // stores the type name for static validation
}

type SymbolTable struct {
	Outer          *SymbolTable
	store          map[string]Symbol
	numDefinitions int
	FreeSymbols    []Symbol
}

func (s *SymbolTable) MarkExported(name string) {
	obj, ok := s.store[name]
	if ok {
		obj.IsExported = true
		s.store[name] = obj
	}
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store:       make(map[string]Symbol),
		FreeSymbols: []Symbol{},
	}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (s *SymbolTable) Define(name string) Symbol {
	return s.DefineWithType(name, "")
}

func (s *SymbolTable) DefineWithType(name string, dataType string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions, DataType: dataType}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineType(name string) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: TypeScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) DefineConst(name string) Symbol {
	return s.DefineConstWithType(name, "")
}

func (s *SymbolTable) DefineConstWithType(name string, dataType string) Symbol {
	symbol := s.DefineWithType(name, dataType)
	symbol.IsConst = true
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) IsConstant(name string) bool {
	obj, ok := s.store[name]
	if ok {
		return obj.IsConst
	}
	if s.Outer != nil {
		return s.Outer.IsConstant(name)
	}
	return false
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}

		if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
			return obj, ok
		}

		free := s.defineFree(obj)
		return free, true
	}
	return obj, ok
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: FunctionScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)
	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1, Scope: FreeScope}
	s.store[original.Name] = symbol
	return symbol
}

func (s *SymbolTable) GetStore() map[string]Symbol {
	return s.store
}

func (s *SymbolTable) NumDefinitions() int {
	return s.numDefinitions
}
