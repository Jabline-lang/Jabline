package parser

import (
	"fmt"
	"jabline/pkg/token"
)

// curTokenIs verifica si el token actual es del tipo especificado
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curTok.Type == t
}

// peekTokenIs verifica si el siguiente token es del tipo especificado
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekTok.Type == t
}

// expectPeek verifica si el siguiente token es del tipo esperado y avanza si es correcto
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// peekError añade un error cuando el siguiente token no es el esperado
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("line %d, column %d: expected next token to be %s, got %s instead",
		p.peekTok.Line, p.peekTok.Column, t, p.peekTok.Type)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError añade un error cuando no se encuentra función de parsing prefix
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("line %d, column %d: no prefix parse function for %s found",
		p.curTok.Line, p.curTok.Column, t)
	p.errors = append(p.errors, msg)
}

// isAssignmentStatement verifica si el statement actual es una asignación
func (p *Parser) isAssignmentStatement() bool {
	if p.curTok.Type != token.IDENT {
		return false
	}

	// Verificar si es asignación a campo (obj.field = ...)
	if p.peekTok.Type == token.DOT {
		return true
	}

	// Verificar si es asignación a índice de array (arr[i] = ...)
	if p.peekTok.Type == token.LBRACKET {
		return true
	}

	// Verificar si es asignación simple (var = ...)
	if p.isAssignmentOperator(p.peekTok.Type) {
		return true
	}

	return false
}
