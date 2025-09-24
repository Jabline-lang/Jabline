package token

// TokenType representa el tipo de un token
type TokenType string

// Token representa un token individual con su tipo y valor literal
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}
