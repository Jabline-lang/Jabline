package token

const (
	// Assignment operators
	ASSIGN      = "="
	PLUS_ASSIGN = "+="
	SUB_ASSIGN  = "-="
	MUL_ASSIGN  = "*="
	DIV_ASSIGN  = "/="

	// Arithmetic operators
	PLUS      = "+"
	MINUS     = "-"
	ASTERISK  = "*"
	SLASH     = "/"
	MOD       = "%"
	INCREMENT = "++"
	DECREMENT = "--"

	// Logical operators
	BANG = "!"
	AND  = "&&"
	OR   = "||"

	// Comparison operators
	LT     = "<"
	GT     = ">"
	LT_EQ  = "<="
	GT_EQ  = ">="
	EQ     = "=="
	NOT_EQ = "!="

	// Ternary operator
	QUESTION = "?"

	// Arrow function operator
	ARROW = "=>"

	// Modern operators
	NULLISH_COALESCING = "??"
	OPTIONAL_CHAINING  = "?."
)
