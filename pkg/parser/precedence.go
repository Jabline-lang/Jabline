package parser

import "jabline/pkg/token"

const (
	_ int = iota
	LOWEST
	NULLISH_COALESCING
	TERNARY
	CHANNEL_SEND
	OR
	AND
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
	OPTIONAL_CHAINING
	POSTFIX
)

var precedences = map[token.TokenType]int{
	token.ARROW_LEFT:         CHANNEL_SEND,
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
	token.BIT_OR:             SUM,
	token.BIT_XOR:            SUM,
	token.SLASH:              PRODUCT,
	token.ASTERISK:           PRODUCT,
	token.MOD:                PRODUCT,
	token.BIT_AND:            PRODUCT,
	token.SHIFT_LEFT:         PRODUCT,
	token.SHIFT_RIGHT:        PRODUCT,
	token.LPAREN:             CALL,
	token.DOT:                INDEX,
	token.LBRACKET:           INDEX,
	token.OPTIONAL_CHAINING:  OPTIONAL_CHAINING,
	token.INCREMENT:          POSTFIX,
	token.DECREMENT:          POSTFIX,
	token.LBRACE:             CALL,
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekTok.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curTok.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) isAssignmentOperator(tok token.TokenType) bool {
	return tok == token.ASSIGN ||
		tok == token.PLUS_ASSIGN ||
		tok == token.SUB_ASSIGN ||
		tok == token.MUL_ASSIGN ||
		tok == token.DIV_ASSIGN
}

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
