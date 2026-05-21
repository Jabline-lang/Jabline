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
	outer          *Environment
	expectedReturn TypeType // Set when inside a function body
	loopDepth      int      // Track loop nesting for break/continue validation
}

// NewEnvironment creates a new root type environment.
func NewEnvironment() *Environment {
	return &Environment{
		store:          make(map[string]TypeType),
		outer:          nil,
		expectedReturn: TypeAny,
		loopDepth:      0,
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

// ParseASTType converts an ast.TypeExpression into an internal TypeType.
func ParseASTType(typeExpr *ast.TypeExpression) TypeType {
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
		return TypeType(typeExpr.Value)
	}
}
