package parser

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/token"
)

func (p *Parser) parseFunctionStatement() ast.Statement {
	stmt := &ast.FunctionStatement{Token: p.curTok}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curTok}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseArrowFunction() ast.Expression {
	arrowFn := &ast.ArrowFunction{Token: p.curTok}

	if p.curTok.Type == token.LPAREN {

		params := p.parseFunctionParameters()

		if !p.peekTokenIs(token.ARROW) {
			return nil
		}

		arrowFn.Parameters = params

		p.nextToken()
		p.nextToken()

		arrowFn.Body = p.parseExpression(1)
		return arrowFn
	}

	return nil
}

func (p *Parser) parseArrowFunctionFromIdent() ast.Expression {

	if p.peekTok.Type == token.ARROW {
		arrowFn := &ast.ArrowFunction{Token: p.curTok}

		param := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
		arrowFn.Parameters = []*ast.Identifier{param}

		p.nextToken()
		p.nextToken()

		arrowFn.Body = p.parseExpression(1)
		return arrowFn
	}

	return &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseStructStatement() ast.Statement {
	stmt := &ast.StructStatement{Token: p.curTok}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Fields = p.parseStructFields()

	return stmt
}

func (p *Parser) parseStructFields() map[string]*ast.TypeExpression {
	fields := make(map[string]*ast.TypeExpression)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return fields
	}

	p.nextToken()

	if p.curTok.Type != token.IDENT {
		return nil
	}

	name := p.curTok.Literal

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()

	typeExpr := p.parseTypeExpression()
	if typeExpr == nil {
		return nil
	}

	fields[name] = typeExpr

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		if p.curTok.Type != token.IDENT {
			break
		}

		fieldName := p.curTok.Literal

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()

		fieldType := p.parseTypeExpression()
		if fieldType == nil {
			return nil
		}

		fields[fieldName] = fieldType
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return fields
}

func (p *Parser) parseStructLiteral() ast.Expression {
	lit := &ast.StructLiteral{Token: p.curTok}
	lit.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Fields = p.parseStructLiteralFields()

	return lit
}

func (p *Parser) parseStructLiteralFields() map[string]ast.Expression {
	fields := make(map[string]ast.Expression)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return fields
	}

	p.nextToken()

	if p.curTok.Type != token.IDENT {
		return nil
	}

	name := p.curTok.Literal

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)
	fields[name] = value

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		if p.curTok.Type != token.IDENT {
			break
		}

		fieldName := p.curTok.Literal

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		fieldValue := p.parseExpression(LOWEST)
		fields[fieldName] = fieldValue
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return fields
}

func (p *Parser) parseTypeExpression() *ast.TypeExpression {
	switch p.curTok.Type {
	case token.STRING_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "string"}
	case token.INT_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "int"}
	case token.FLOAT_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "float"}
	case token.BOOL_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "bool"}
	default:
		p.errors = append(p.errors, fmt.Sprintf("expected type, got %s", p.curTok.Type))
		return nil
	}
}
