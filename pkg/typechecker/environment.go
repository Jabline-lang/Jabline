package typechecker

import "jabline/pkg/ast"

// TypeType represents a type in our type system
type TypeType string

const (
	TypeInt      TypeType = "int"
	TypeInt8     TypeType = "int8"
	TypeInt16    TypeType = "int16"
	TypeInt32    TypeType = "int32"
	TypeInt64    TypeType = "int64"
	TypeUint8    TypeType = "uint8"
	TypeUint16   TypeType = "uint16"
	TypeUint32   TypeType = "uint32"
	TypeUint64   TypeType = "uint64"
	TypeFloat    TypeType = "float"
	TypeFloat32  TypeType = "float32"
	TypeFloat64  TypeType = "float64"
	TypeString   TypeType = "string"
	TypeBool     TypeType = "bool"
	TypeVoid     TypeType = "void"
	TypeAny      TypeType = "any"
	TypeError    TypeType = "error"
	TypeArray    TypeType = "Array"
	TypeHash     TypeType = "Hash"
	TypeFunction TypeType = "Function"
	TypeChannel  TypeType = "Channel"
	TypeNull     TypeType = "null"
)

// Environment holds the type information for variables in a specific scope.
type Environment struct {
	store          map[string]TypeType
	consts         map[string]bool // Track const-declared variables
	outer          *Environment
	expectedReturn TypeType // Set when inside a function body
	loopDepth      int      // Track loop nesting for break/continue validation
	typeAliases    map[string]TypeType // User-defined type aliases
}

// NewEnvironment creates a new root type environment.
func NewEnvironment() *Environment {
	return &Environment{
		store:          make(map[string]TypeType),
		consts:         make(map[string]bool),
		outer:          nil,
		expectedReturn: TypeAny,
		loopDepth:      0,
		typeAliases:    make(map[string]TypeType),
	}
}

// NewEnclosedEnvironment creates a new environment scoped within an outer environment.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	env.expectedReturn = outer.expectedReturn
	env.loopDepth = outer.loopDepth
	return env
}

// IsConst returns whether a variable was declared with const.
func (e *Environment) IsConst(name string) bool {
	if val, ok := e.consts[name]; ok {
		return val
	}
	if e.outer != nil {
		return e.outer.IsConst(name)
	}
	return false
}

// MarkConst marks a variable as const-declared.
func (e *Environment) MarkConst(name string) {
	e.consts[name] = true
}

// ExistsInCurrentScope returns true if the variable is defined directly in this scope (not outer).
func (e *Environment) ExistsInCurrentScope(name string) bool {
	_, ok := e.store[name]
	return ok
}

// Get retrieves the type of a variable by name.
func (e *Environment) Get(name string) (TypeType, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set defines or updates the type of a variable in the current scope.
func (e *Environment) Set(name string, val TypeType) TypeType {
	e.store[name] = val
	return val
}

// DefineAlias registers a type alias (e.g., `type MyInt = int`).
func (e *Environment) DefineAlias(name string, target TypeType) {
	e.typeAliases[name] = target
}

// ResolveAlias resolves a type alias to its underlying type. Returns the type and true if found.
func (e *Environment) ResolveAlias(name string) (TypeType, bool) {
	t, ok := e.typeAliases[name]
	if ok {
		// Allow transitive aliases (A -> B -> int)
		if inner, ok := e.typeAliases[string(t)]; ok {
			return inner, true
		}
	}
	return t, ok
}

// ParseASTType converts an ast.TypeExpression into an internal TypeType.
// When called on a Checker, it also checks user-defined type aliases.
// If env is nil, only built-in types are recognized.
func ParseASTType(typeExpr *ast.TypeExpression, env *Environment) TypeType {
	if typeExpr == nil {
		return TypeAny
	}

	switch typeExpr.Value {
	case "int":
		return TypeInt
	case "int8":
		return TypeInt8
	case "int16":
		return TypeInt16
	case "int32":
		return TypeInt32
	case "int64":
		return TypeInt64
	case "uint8":
		return TypeUint8
	case "uint16":
		return TypeUint16
	case "uint32":
		return TypeUint32
	case "uint64":
		return TypeUint64
	case "float":
		return TypeFloat
	case "float32":
		return TypeFloat32
	case "float64":
		return TypeFloat64
	case "string":
		return TypeString
	case "bool":
		return TypeBool
	case "void":
		return TypeVoid
	case "Array":
		return TypeArray
	case "Hash":
		return TypeHash
	case "Channel":
		return TypeChannel
	case "null":
		return TypeNull
	default:
		if env != nil {
			if resolved, ok := env.ResolveAlias(typeExpr.Value); ok {
				return resolved
			}
		}
		return TypeType(typeExpr.Value)
	}
}
