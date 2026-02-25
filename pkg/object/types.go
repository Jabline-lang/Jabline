package object

type ObjectType string

const (
	INTEGER_OBJ = "INTEGER"
	FLOAT_OBJ   = "FLOAT"

	INT8_OBJ  = "INT8"
	INT16_OBJ = "INT16"
	INT32_OBJ = "INT32"
	INT64_OBJ = "INT64"

	UINT8_OBJ  = "UINT8"
	UINT16_OBJ = "UINT16"
	UINT32_OBJ = "UINT32"
	UINT64_OBJ = "UINT64"

	FLOAT32_OBJ = "FLOAT32"
	FLOAT64_OBJ = "FLOAT64"

	BOOLEAN_OBJ = "BOOLEAN"
	STRING_OBJ  = "STRING"
	NULL_OBJ    = "NULL"

	ERROR_OBJ     = "ERROR"
	RETURN_OBJ    = "RETURN"
	BREAK_OBJ     = "BREAK"
	CONTINUE_OBJ  = "CONTINUE"
	EXCEPTION_OBJ = "EXCEPTION"

	FUNCTION_OBJ          = "FUNCTION"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION_OBJ"
	CLOSURE_OBJ           = "CLOSURE"
	ARROW_FUNCTION_OBJ    = "ARROW_FUNCTION"
	BOUND_METHOD_OBJ      = "BOUND_METHOD"
	SERVICE_OBJ           = "SERVICE"

	ASYNC_FUNCTION_OBJ        = "ASYNC_FUNCTION"
	BUILTIN_OBJ               = "BUILTIN"
	STRUCT_OBJ                = "STRUCT"
	INSTANTIATED_STRUCT_OBJ   = "INSTANTIATED_STRUCT"
	INSTANTIATED_FUNCTION_OBJ = "INSTANTIATED_FUNCTION"
	INSTANCE_OBJ              = "INSTANCE"
	ARRAY_OBJ                 = "ARRAY"
	HASH_OBJ                  = "HASH"
	PROMISE_OBJ               = "PROMISE"
	CHANNEL_OBJ               = "CHANNEL"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
}

// VMExecutor is a function hook to execute a closure in a new VM/Context.
// Used to break dependency cycles between stdlib and vm.
type VMExecutor func(closure Object, args []Object) Object
