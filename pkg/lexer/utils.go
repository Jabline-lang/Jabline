package lexer

import "jabline/pkg/token"

// newToken crea un nuevo token con el tipo y carácter dados
// Es una función de utilidad para crear tokens de un solo carácter
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
