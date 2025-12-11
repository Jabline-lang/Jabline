package token

const (
	ECHO        = "ECHO"
	LET         = "LET"
	CONST       = "CONST"
	NULL        = "NULL"
	IF          = "IF"
	ELSE        = "ELSE"
	FOR         = "FOR"
	WHILE       = "WHILE"
	FUNCTION    = "FUNCTION"
	RETURN      = "RETURN"
	STRUCT      = "STRUCT"
	BREAK       = "BREAK"
	CONTINUE    = "CONTINUE"
	IN          = "IN"
	SWITCH      = "SWITCH"
	CASE        = "CASE"
	DEFAULT     = "DEFAULT"
	TRY         = "TRY"
	CATCH       = "CATCH"
	THROW       = "THROW"
	ASYNC       = "ASYNC"
	AWAIT       = "AWAIT"
	IMPORT      = "IMPORT"
	EXPORT      = "EXPORT"
	FROM        = "FROM"
	STRING_TYPE = "STRING_TYPE"
	INT_TYPE    = "INT_TYPE"
	FLOAT_TYPE  = "FLOAT_TYPE"
	BOOL_TYPE   = "BOOL_TYPE"
	TRUE        = "TRUE"
	FALSE       = "FALSE"
	ENUM        = "ENUM"
	AS          = "AS"
)

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
	"async":    ASYNC,
	"await":    AWAIT,
	"import":   IMPORT,
	"export":   EXPORT,
	"from":     FROM,
	"string":   STRING_TYPE,
	"int":      INT_TYPE,
	"float":    FLOAT_TYPE,
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
