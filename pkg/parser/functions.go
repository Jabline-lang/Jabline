package parser

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/token"
)

func (p *Parser) parseFunctionStatement() ast.Statement {
	stmt := &ast.FunctionStatement{Token: p.curTok}

	// Check for Method Receiver: fn (receiver Type) name
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken() // Move to (

		// Expect receiver name
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.ReceiverName = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

		// Expect receiver type
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.ReceiverType = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken()
		stmt.TypeParameters = p.parseTypeParameters()
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	// Optional return type: `fn name(params): int`
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume COLON
		p.nextToken() // move to type token
		stmt.ReturnType = p.parseTypeExpression()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curTok}

	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken()
		lit.TypeParameters = p.parseTypeParameters()
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	// Optional return type: `fn(params): int { ... }`
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume COLON
		p.nextToken() // move to type token
		lit.ReturnType = p.parseTypeExpression()
	}

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

	// Optional type annotation: `param: int`
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume COLON
		p.nextToken() // move to type token
		ident.Type = p.parseTypeExpression()
	}

	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume COMMA
		p.nextToken() // move to next parameter

		ident := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

		// Optional type annotation: `, param: int`
		if p.peekTokenIs(token.COLON) {
			p.nextToken() // consume COLON
			p.nextToken() // move to type token
			ident.Type = p.parseTypeExpression()
		}

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

	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken()
		stmt.TypeParameters = p.parseTypeParameters()
	}

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

func (p *Parser) parseStructLiteralInfix(left ast.Expression) ast.Expression {
	lit := &ast.StructLiteral{Token: p.curTok, Name: left}

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
	case token.INT8_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "int8"}
	case token.INT16_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "int16"}
	case token.INT32_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "int32"}
	case token.INT64_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "int64"}
	case token.UINT8_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "uint8"}
	case token.UINT16_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "uint16"}
	case token.UINT32_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "uint32"}
	case token.UINT64_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "uint64"}
	case token.FLOAT_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "float"}
	case token.FLOAT32_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "float32"}
	case token.FLOAT64_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "float64"}
	case token.BOOL_TYPE:
		return &ast.TypeExpression{Token: p.curTok, Value: "bool"}
	case token.IDENT:
		te := &ast.TypeExpression{Token: p.curTok, Value: p.curTok.Literal}
		if p.peekTokenIs(token.LBRACKET) {
			p.nextToken() // move to [
			te.Arguments = p.parseTypeArguments()
		}
		return te
	default:
		p.errors = append(p.errors, fmt.Sprintf("expected type, got %s", p.curTok.Type))
		return nil
	}
}

func (p *Parser) parseTypeArguments() []*ast.TypeExpression {
	args := []*ast.TypeExpression{}

	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseTypeExpression())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseTypeExpression())
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return args
}

func (p *Parser) parseTypeParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()
	ident := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return identifiers
}
