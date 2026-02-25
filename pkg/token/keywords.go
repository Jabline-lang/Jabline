package token

const (
	ECHO         = "ECHO"
	LET          = "LET"
	CONST        = "CONST"
	NULL         = "NULL"
	IF           = "IF"
	ELSE         = "ELSE"
	FOR          = "FOR"
	WHILE        = "WHILE"
	FUNCTION     = "FUNCTION"
	RETURN       = "RETURN"
	STRUCT       = "STRUCT"
	BREAK        = "BREAK"
	CONTINUE     = "CONTINUE"
	IN           = "IN"
	SWITCH       = "SWITCH"
	CASE         = "CASE"
	DEFAULT      = "DEFAULT"
	TRY          = "TRY"
	CATCH        = "CATCH"
	THROW        = "THROW"
	ASYNC        = "ASYNC"
	AWAIT        = "AWAIT"
	IMPORT       = "IMPORT"
	EXPORT       = "EXPORT"
	FROM         = "FROM"
	STRING_TYPE  = "STRING_TYPE"
	INT_TYPE     = "INT_TYPE"
	INT8_TYPE    = "INT8_TYPE"
	INT16_TYPE   = "INT16_TYPE"
	INT32_TYPE   = "INT32_TYPE"
	INT64_TYPE   = "INT64_TYPE"
	UINT8_TYPE   = "UINT8_TYPE"
	UINT16_TYPE  = "UINT16_TYPE"
	UINT32_TYPE  = "UINT32_TYPE"
	UINT64_TYPE  = "UINT64_TYPE"
	FLOAT_TYPE   = "FLOAT_TYPE"
	FLOAT32_TYPE = "FLOAT32_TYPE"
	FLOAT64_TYPE = "FLOAT64_TYPE"
	BOOL_TYPE    = "BOOL_TYPE"
	TRUE         = "TRUE"
	FALSE        = "FALSE"
	ENUM         = "ENUM"
	AS           = "AS"
	RETRY        = "RETRY"
	SERVICE      = "SERVICE"
	SPAWN        = "SPAWN"
)

var keywords = map[string]TokenType{
	"echo":     ECHO,
	"let":      LET,
	"const":    CONST,
	"service":  SERVICE,
	"null":     NULL,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"while":    WHILE,
	"fn":       FUNCTION,
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
	"retry":    RETRY,
	"async":    ASYNC,
	"await":    AWAIT,
	"import":   IMPORT,
	"export":   EXPORT,
	"from":     FROM,
	"spawn":    SPAWN,
	"string":   STRING_TYPE,
	"int":      INT_TYPE,
	"int8":     INT8_TYPE,
	"int16":    INT16_TYPE,
	"int32":    INT32_TYPE,
	"int64":    INT64_TYPE,
	"uint8":    UINT8_TYPE,
	"uint16":   UINT16_TYPE,
	"uint32":   UINT32_TYPE,
	"uint64":   UINT64_TYPE,
	"float":    FLOAT_TYPE,
	"float32":  FLOAT32_TYPE,
	"float64":  FLOAT64_TYPE,
	"bool":     BOOL_TYPE,
	"enum":     ENUM,
	"as":       AS,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
