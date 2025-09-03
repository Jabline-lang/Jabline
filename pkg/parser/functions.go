package parser

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/token"
)

// parseFunctionStatement parsea una declaración de función
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

// parseFunctionLiteral parsea un literal de función anónima
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

// parseArrowFunction parsea una arrow function como (a, b) => a + b
func (p *Parser) parseArrowFunction() ast.Expression {
	arrowFn := &ast.ArrowFunction{Token: p.curTok}

	// Case: Multiple parameters or zero parameters: (a, b) => a + b or () => 42
	if p.curTok.Type == token.LPAREN {
		// Try to parse parameters
		params := p.parseFunctionParameters()

		// Check if we have an arrow after the parameters
		if !p.peekTokenIs(token.ARROW) {
			// Not an arrow function, reset state and return nil
			// This allows parseGroupedExpression to handle it
			return nil
		}

		// This is indeed an arrow function
		arrowFn.Parameters = params

		// Consume the arrow token
		p.nextToken() // move to '=>'
		p.nextToken() // move to body expression

		arrowFn.Body = p.parseExpression(1) // Use LOWEST precedence value directly
		return arrowFn
	}

	return nil
}

// parseArrowFunctionFromIdent parsea arrow function cuando empezamos con un identificador
func (p *Parser) parseArrowFunctionFromIdent() ast.Expression {
	// Check if this is a single parameter arrow function: x => expr
	if p.peekTok.Type == token.ARROW {
		arrowFn := &ast.ArrowFunction{Token: p.curTok}

		param := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
		arrowFn.Parameters = []*ast.Identifier{param}

		// Consume the arrow token
		p.nextToken() // move to '=>'
		p.nextToken() // move to body expression

		arrowFn.Body = p.parseExpression(1) // Use LOWEST precedence value directly
		return arrowFn
	}

	// Not an arrow function, return regular identifier
	return &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
}

// parseFunctionParameters parsea los parámetros de una función
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

// parseStructStatement parsea una declaración de estructura
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

// parseStructFields parsea los campos de una estructura
func (p *Parser) parseStructFields() map[string]*ast.TypeExpression {
	fields := make(map[string]*ast.TypeExpression)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return fields
	}

	p.nextToken()

	// Parsear primer campo
	if p.curTok.Type != token.IDENT {
		return nil
	}

	name := p.curTok.Literal

	// Debe haber dos puntos seguido del tipo
	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken() // moverse al tipo

	typeExpr := p.parseTypeExpression()
	if typeExpr == nil {
		return nil
	}

	fields[name] = typeExpr

	// Parsear campos restantes
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consumir ','
		p.nextToken() // moverse al nombre del campo

		if p.curTok.Type != token.IDENT {
			break
		}

		fieldName := p.curTok.Literal

		// Debe haber dos puntos seguido del tipo
		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken() // moverse al tipo

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

// parseStructLiteral parsea un literal de estructura
func (p *Parser) parseStructLiteral() ast.Expression {
	lit := &ast.StructLiteral{Token: p.curTok}
	lit.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Fields = p.parseStructLiteralFields()

	return lit
}

// parseStructLiteralFields parsea los campos de un literal de estructura
func (p *Parser) parseStructLiteralFields() map[string]ast.Expression {
	fields := make(map[string]ast.Expression)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return fields
	}

	p.nextToken()

	// Parsear primer campo
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

	// Parsear campos restantes
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consumir ','
		p.nextToken() // moverse al nombre del campo

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

// parseTypeExpression parsea una expresión de tipo
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
