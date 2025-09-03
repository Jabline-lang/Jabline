package token

const (
	// Special tokens
	ILLEGAL = "ILLEGAL" // token/carácter desconocido
	EOF     = "EOF"     // fin de archivo

	// Identifiers + literals
	IDENT            = "IDENT"            // variables, ej: x, foo
	INT              = "INT"              // números enteros, ej: 123
	FLOAT            = "FLOAT"            // números decimales, ej: 3.14
	STRING           = "STRING"           // cadenas, ej: "hello"
	TEMPLATE_LITERAL = "TEMPLATE_LITERAL" // template literals, ej: `Hello ${name}`
	TRUE             = "TRUE"             // booleano true
	FALSE            = "FALSE"            // booleano false
)
