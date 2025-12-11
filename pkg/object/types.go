package object

type ObjectType string

const (
	INTEGER_OBJ = "INTEGER"
	FLOAT_OBJ   = "FLOAT"
	BOOLEAN_OBJ = "BOOLEAN"
	STRING_OBJ  = "STRING"
	NULL_OBJ    = "NULL"

	ERROR_OBJ     = "ERROR"
	RETURN_OBJ    = "RETURN"
	BREAK_OBJ     = "BREAK"
	CONTINUE_OBJ  = "CONTINUE"
	EXCEPTION_OBJ = "EXCEPTION"

	FUNCTION_OBJ       = "FUNCTION"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION_OBJ"
	CLOSURE_OBJ           = "CLOSURE"
	ARROW_FUNCTION_OBJ    = "ARROW_FUNCTION"

	ASYNC_FUNCTION_OBJ = "ASYNC_FUNCTION"
	BUILTIN_OBJ        = "BUILTIN"
	STRUCT_OBJ         = "STRUCT"
	INSTANCE_OBJ       = "INSTANCE"
	ARRAY_OBJ          = "ARRAY"
	HASH_OBJ           = "HASH"
	PROMISE_OBJ        = "PROMISE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
}
