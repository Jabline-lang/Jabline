package object

// ObjectType representa el tipo de un objeto en el lenguaje
type ObjectType string

const (
	// Primitive types
	INTEGER_OBJ = "INTEGER"
	FLOAT_OBJ   = "FLOAT"
	BOOLEAN_OBJ = "BOOLEAN"
	STRING_OBJ  = "STRING"
	NULL_OBJ    = "NULL"

	// Control flow types
	ERROR_OBJ     = "ERROR"
	RETURN_OBJ    = "RETURN"
	BREAK_OBJ     = "BREAK"
	CONTINUE_OBJ  = "CONTINUE"
	EXCEPTION_OBJ = "EXCEPTION"

	// Complex types
	FUNCTION_OBJ       = "FUNCTION"
	ARROW_FUNCTION_OBJ = "ARROW_FUNCTION"
	ASYNC_FUNCTION_OBJ = "ASYNC_FUNCTION"
	BUILTIN_OBJ        = "BUILTIN"
	STRUCT_OBJ         = "STRUCT"
	INSTANCE_OBJ       = "INSTANCE"
	ARRAY_OBJ          = "ARRAY"
	HASH_OBJ           = "HASH"
	PROMISE_OBJ        = "PROMISE"
)

// Object interface que deben implementar todos los objetos
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Hashable interface para objetos que pueden ser usados como claves en hash maps
type Hashable interface {
	HashKey() HashKey
}
