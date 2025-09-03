package token

const (
	// Keywords
	ECHO        = "ECHO"        // palabra reservada
	LET         = "LET"         // declaración de variables
	CONST       = "CONST"       // declaración de constantes
	NULL        = "NULL"        // valor null
	IF          = "IF"          // condicional if
	ELSE        = "ELSE"        // condicional else
	FOR         = "FOR"         // bucle for
	WHILE       = "WHILE"       // bucle while
	FUNCTION    = "FUNCTION"    // declaración de función
	RETURN      = "RETURN"      // retorno de función
	STRUCT      = "STRUCT"      // declaración de estructura
	BREAK       = "BREAK"       // break statement
	CONTINUE    = "CONTINUE"    // continue statement
	IN          = "IN"          // for-in statement
	SWITCH      = "SWITCH"      // switch statement
	CASE        = "CASE"        // case statement
	DEFAULT     = "DEFAULT"     // default statement
	TRY         = "TRY"         // try statement
	CATCH       = "CATCH"       // catch statement
	THROW       = "THROW"       // throw statement
	ASYNC       = "ASYNC"       // async keyword
	AWAIT       = "AWAIT"       // await keyword
	IMPORT      = "IMPORT"      // import statement
	EXPORT      = "EXPORT"      // export statement
	FROM        = "FROM"        // from keyword
	STRING_TYPE = "STRING_TYPE" // string type
	INT_TYPE    = "INT_TYPE"    // int type
	FLOAT_TYPE  = "FLOAT_TYPE"  // float type
	BOOL_TYPE   = "BOOL_TYPE"   // bool type
)

// keywords mapea las palabras reservadas a sus tipos de token
var keywords = map[string]TokenType{
	"echo":     ECHO,
	"let":      LET,
	"const":    CONST,
	"null":     NULL,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"while":    WHILE,
	"function": FUNCTION,
	"fn":       FUNCTION, // alias para function
	"return":   RETURN,
	"struct":   STRUCT,
	"break":    BREAK,
	"continue": CONTINUE,
	"in":       IN,
	"switch":   SWITCH,
	"case":     CASE,
	"default":  DEFAULT,
	"try":      TRY,
	"catch":    CATCH,
	"throw":    THROW,
	"async":    ASYNC,
	"await":    AWAIT,
	"import":   IMPORT,
	"export":   EXPORT,
	"from":     FROM,
	"string":   STRING_TYPE,
	"int":      INT_TYPE,
	"float":    FLOAT_TYPE,
	"bool":     BOOL_TYPE,
}

// LookupIdent verifica si un identificador es una palabra reservada
// Si es una keyword devuelve el TokenType correspondiente, sino devuelve IDENT
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
