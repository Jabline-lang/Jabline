package parser

import (
	"fmt"
	"strconv"

	"jabline/pkg/ast"
	"jabline/pkg/lexer"
	"jabline/pkg/token"
)

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curTok.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curTok.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekTok.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		newLeft := infix(leftExp)
		if newLeft == nil {
			return nil
		}
		leftExp = newLeft
	}

	return leftExp
}

func (p *Parser) parseSpawnExpression() ast.Expression {
	exp := &ast.SpawnExpression{Token: p.curTok}

	p.nextToken() // move past 'spawn'

	callExp := p.parseExpression(LOWEST)
	call, ok := callExp.(*ast.CallExpression)
	if !ok {
		p.errors = append(p.errors, fmt.Sprintf("expected function call after spawn, got %T", callExp))
		return nil
	}

	exp.Call = call
	return exp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curTok}

	value, err := strconv.ParseInt(p.curTok.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d, column %d: could not parse %q as integer", p.curTok.Line, p.curTok.Column, p.curTok.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curTok}

	value, err := strconv.ParseFloat(p.curTok.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d, column %d: could not parse %q as float", p.curTok.Line, p.curTok.Column, p.curTok.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curTok, Value: p.curTok.Literal}
}

func (p *Parser) parseTemplateLiteral() ast.Expression {
	lit := &ast.TemplateLiteral{Token: p.curTok}

	content := p.curTok.Literal
	lit.Parts, lit.Expressions = p.parseTemplateContent(content)

	return lit
}

func (p *Parser) parseTemplateContent(content string) ([]string, []ast.Expression) {
	parts := []string{}
	expressions := []ast.Expression{}

	currentPart := ""
	i := 0

	for i < len(content) {
		if i < len(content)-1 && content[i] == '$' && content[i+1] == '{' {

			parts = append(parts, currentPart)
			currentPart = ""
			i += 2

			braceCount := 1
			exprStart := i

			for i < len(content) && braceCount > 0 {
				if content[i] == '{' {
					braceCount++
				} else if content[i] == '}' {
					braceCount--
				}
				if braceCount > 0 {
					i++
				}
			}

			if braceCount == 0 {
				exprContent := content[exprStart:i]

				exprLexer := lexer.New(exprContent)
				exprParser := New(exprLexer)
				expr := exprParser.parseExpression(1)

				if expr != nil {
					expressions = append(expressions, expr)
				} else {
					expressions = append(expressions, &ast.StringLiteral{
						Token: token.Token{Type: token.STRING, Literal: exprContent},
						Value: exprContent,
					})
				}
				i++
			} else {
				currentPart += "${"
			}
		} else {
			currentPart += string(content[i])
			i++
		}
	}

	parts = append(parts, currentPart)

	return parts, expressions
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curTok, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curTok,
		Operator: p.curTok.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curTok,
		Left:     left,
		Operator: p.curTok.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseGroupedOrArrowFunction() ast.Expression {
	lParenToken := p.curTok

	// Peak ahead to see if it's an arrow function or empty param list
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken() // curTok = RPAREN
		// Must be an arrow function: () => ... OR (): type => ...
		var returnType *ast.TypeExpression
		if p.peekTokenIs(token.COLON) {
			p.nextToken() // curTok = COLON
			p.nextToken() // curTok = type
			returnType = p.parseTypeExpression()
		}

		if !p.expectPeek(token.ARROW) {
			return nil
		}

		p.nextToken() // move to body
		arrow := &ast.ArrowFunction{
			Token:      lParenToken,
			Parameters: []*ast.Identifier{},
			ReturnType: returnType,
		}
		arrow.Body = p.parseExpression(LOWEST)
		return arrow
	}

	// Try to parse as expressions first.
	// But wait, if it has types like (x: int), parseExpressionList(token.RPAREN) will fail.
	// So we need to be smarter.

	// Let's use parseFunctionParameters if we see a COLON after the first identifier or if we have multiple params.
	// For simplicity, let's try to parse it as parameters IF we see ARROW or COLON after the list.

	// Actually, the easiest is to allow COLON in a custom list parser.

	p.nextToken() // Move to first element

	// We'll collect both expressions and identifiers
	exprs := []ast.Expression{}
	idents := []*ast.Identifier{}
	isArrow := false

	for {
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			return nil
		}

		// If the expression is an identifier and followed by a COLON, it's definitely a typed parameter
		if ident, ok := expr.(*ast.Identifier); ok && p.peekTokenIs(token.COLON) {
			isArrow = true
			p.nextToken() // COLON
			p.nextToken() // type
			ident.Type = p.parseTypeExpression()
			idents = append(idents, ident)
		} else if ident, ok := expr.(*ast.Identifier); ok {
			idents = append(idents, ident)
			exprs = append(exprs, expr)
		} else {
			exprs = append(exprs, expr)
		}

		if !p.peekTokenIs(token.COMMA) {
			break
		}
		p.nextToken() // COMMA
		p.nextToken() // next element
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// After RPAREN, check for COLON or ARROW
	var returnType *ast.TypeExpression
	if p.peekTokenIs(token.COLON) {
		isArrow = true
		p.nextToken() // COLON
		p.nextToken() // type
		returnType = p.parseTypeExpression()
	}

	if p.peekTokenIs(token.ARROW) || isArrow {
		if !p.expectPeek(token.ARROW) {
			return nil
		}
		p.nextToken() // body

		arrow := &ast.ArrowFunction{
			Token:      lParenToken,
			Parameters: idents,
			ReturnType: returnType,
		}
		arrow.Body = p.parseExpression(LOWEST)
		return arrow
	}

	// Not an arrow function, must be a grouped expression
	if len(exprs) != 1 {
		p.addError("unexpected comma in expression")
		return nil
	}
	return exprs[0]
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curTok}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if p.peekTokenIs(token.IF) {
			p.nextToken()
			ifToken := p.curTok

			block := &ast.BlockStatement{Token: ifToken}

			ifExpression := p.parseIfExpression()
			if ifExpression == nil {
				return nil
			}

			elseIfStatement := &ast.ExpressionStatement{
				Token:      ifToken,
				Expression: ifExpression,
			}

			block.Statements = append(block.Statements, elseIfStatement)
			expression.Alternative = block
		} else {
			if !p.expectPeek(token.LBRACE) {
				return nil
			}
			expression.Alternative = p.parseBlockStatement()
		}
	}
	return expression
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curTok}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseArrayIndexExpression(left ast.Expression) ast.Expression {
	// Si el token actual es [, el peek podría ser un tipo
	if p.isTypeStart(p.peekTok.Type) {
		exp := &ast.InstantiatedExpression{Token: p.curTok, Left: left}
		// No avanzamos aquí, parseTypeArguments se encarga de saltar el [
		exp.TypeArguments = p.parseTypeArguments()
		return exp
	}

	exp := &ast.ArrayIndexExpression{Token: p.curTok, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) isTypeStart(t token.TokenType) bool {
	switch t {
	case token.STRING_TYPE, token.INT_TYPE, token.INT8_TYPE, token.INT16_TYPE,
		token.INT32_TYPE, token.INT64_TYPE, token.UINT8_TYPE, token.UINT16_TYPE,
		token.UINT32_TYPE, token.UINT64_TYPE, token.FLOAT_TYPE, token.FLOAT32_TYPE,
		token.FLOAT64_TYPE, token.BOOL_TYPE, token.IDENT:
		return true
	}
	return false
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curTok, Left: left}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// Treat dot notation property access as a string literal index
	// obj.prop becomes equivalent to obj["prop"]
	exp.Index = &ast.StringLiteral{Token: p.curTok, Value: p.curTok.Literal}

	return exp
}

func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curTok, Function: fn}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return args
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	return &ast.PostfixExpression{
		Token:    p.curTok,
		Left:     left,
		Operator: p.curTok.Literal,
	}
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curTok}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return hash
	}

	p.nextToken()

	var key ast.Expression
	if p.curTok.Type == token.IDENT && p.peekTokenIs(token.COLON) {
		key = &ast.StringLiteral{Token: p.curTok, Value: p.curTok.Literal}
	} else {
		key = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)
	hash.Pairs[key] = value

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		if p.curTok.Type == token.IDENT && p.peekTokenIs(token.COLON) {
			key = &ast.StringLiteral{Token: p.curTok, Value: p.curTok.Literal}
		} else {
			key = p.parseExpression(LOWEST)
		}

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

func (p *Parser) parseTernaryExpression(left ast.Expression) ast.Expression {
	expression := &ast.TernaryExpression{
		Token:     p.curTok,
		Condition: left,
	}

	p.nextToken()
	expression.TrueValue = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	expression.FalseValue = p.parseExpression(TERNARY - 1)

	return expression
}

func (p *Parser) parseNullishCoalescingExpression(left ast.Expression) ast.Expression {
	expression := &ast.NullishCoalescingExpression{
		Token: p.curTok,
		Left:  left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseOptionalChainingExpression(left ast.Expression) ast.Expression {
	expression := &ast.OptionalChainingExpression{
		Token: p.curTok,
		Left:  left,
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	expression.Right = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	return expression
}

func (p *Parser) parseNull() ast.Expression {
	return &ast.Null{Token: p.curTok}
}

func (p *Parser) parseAsyncFunctionLiteral() ast.Expression {
	lit := &ast.AsyncFunctionLiteral{Token: p.curTok}

	if !p.expectPeek(token.FUNCTION) {
		return nil
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	// Optional return type: `async fn(params): int { ... }`
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

func (p *Parser) parseAwaitExpression() ast.Expression {
	expression := &ast.AwaitExpression{Token: p.curTok}

	p.nextToken()

	expression.Value = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseTypeCastExpression() ast.Expression {
	tok := p.curTok // Current token is INT8_TYPE, UINT16_TYPE, etc.

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken() // Advance to the expression inside parentheses
	argExp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Create a CallExpression that will call the builtin function (e.g., "int8")
	// The function part of the CallExpression will be an Identifier representing the type name.
	callExp := &ast.CallExpression{
		Token:     tok,                                             // The type token itself (e.g., INT8_TYPE)
		Function:  &ast.Identifier{Token: tok, Value: tok.Literal}, // Identifier "int8"
		Arguments: []ast.Expression{argExp},
	}

	return callExp
}
