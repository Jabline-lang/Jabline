package parser

import (
	"fmt"
	"jabline/pkg/token"
)

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curTok.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekTok.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("line %d, column %d: expected next token to be %s, got %s instead",
		p.peekTok.Line, p.peekTok.Column, t, p.peekTok.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("line %d, column %d: no prefix parse function for %s found",
		p.curTok.Line, p.curTok.Column, t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf("line %d, column %d: ", p.curTok.Line, p.curTok.Column)
	msg += fmt.Sprintf(format, args...)
	p.errors = append(p.errors, msg)
}

func (p *Parser) isAssignmentStatement() bool {
	if p.curTok.Type != token.IDENT {
		return false
	}

	if p.peekTok.Type == token.DOT {
		return true
	}

	if p.peekTok.Type == token.LBRACKET {
		return true
	}

	if p.isAssignmentOperator(p.peekTok.Type) {
		return true
	}

	return false
}
