package parser

import "jabline/pkg/token"

// Precedencias de operadores
const (
	_ int = iota
	LOWEST
	NULLISH_COALESCING // ??
	TERNARY            // ? :
	OR                 // ||
	AND                // &&
	EQUALS             // ==
	LESSGREATER        // > or <
	SUM                // +
	PRODUCT            // *
	PREFIX             // -X or !X
	CALL               // myFunction(X)
	INDEX              // obj.field
	OPTIONAL_CHAINING  // ?.
	POSTFIX            // x++ x--
)

// precedences mapea los tipos de token a sus precedencias
var precedences = map[token.TokenType]int{
	token.NULLISH_COALESCING: NULLISH_COALESCING,
	token.QUESTION:           TERNARY,
	token.OR:                 OR,
	token.AND:                AND,
	token.EQ:                 EQUALS,
	token.NOT_EQ:             EQUALS,
	token.LT:                 LESSGREATER,
	token.GT:                 LESSGREATER,
	token.LT_EQ:              LESSGREATER,
	token.GT_EQ:              LESSGREATER,
	token.PLUS:               SUM,
	token.MINUS:              SUM,
	token.SLASH:              PRODUCT,
	token.ASTERISK:           PRODUCT,
	token.MOD:                PRODUCT,
	token.LPAREN:             CALL,              // for function calls
	token.DOT:                INDEX,             // for field access
	token.LBRACKET:           INDEX,             // for array index access
	token.OPTIONAL_CHAINING:  OPTIONAL_CHAINING, // for optional chaining ?.
	token.INCREMENT:          POSTFIX,           // postfix ++
	token.DECREMENT:          POSTFIX,           // postfix --
}

// peekPrecedence devuelve la precedencia del siguiente token
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekTok.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence devuelve la precedencia del token actual
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curTok.Type]; ok {
		return p
	}
	return LOWEST
}

// isAssignmentOperator verifica si el token es un operador de asignación
func (p *Parser) isAssignmentOperator(tok token.TokenType) bool {
	return tok == token.ASSIGN ||
		tok == token.PLUS_ASSIGN ||
		tok == token.SUB_ASSIGN ||
		tok == token.MUL_ASSIGN ||
		tok == token.DIV_ASSIGN
}

// getArithmeticOperator obtiene el operador aritmético correspondiente a un operador de asignación
func (p *Parser) getArithmeticOperator(tok token.TokenType) token.TokenType {
	switch tok {
	case token.PLUS_ASSIGN:
		return token.PLUS
	case token.SUB_ASSIGN:
		return token.MINUS
	case token.MUL_ASSIGN:
		return token.ASTERISK
	case token.DIV_ASSIGN:
		return token.SLASH
	default:
		return token.ILLEGAL
	}
}
