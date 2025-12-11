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

	var leftExp ast.Expression

	if p.curTok.Type == token.IDENT && p.peekTok.Type == token.LBRACE {
		leftExp = p.parseStructLiteral()
	} else {
		leftExp = prefix()
	}

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

	expressions := p.parseExpressionList(token.RPAREN)

	if !p.peekTokenIs(token.ARROW) {

		if len(expressions) == 0 {
			p.errors = append(p.errors, "Cannot use '()' as an expression.")
			return nil
		}
		if len(expressions) > 1 {
			p.errors = append(p.errors, "Cannot use comma-separated expressions in a grouped expression.")
			return nil
		}
		return expressions[0]
	}

	p.nextToken()
	p.nextToken()

	params := []*ast.Identifier{}
	for _, exp := range expressions {
		ident, ok := exp.(*ast.Identifier)
		if !ok {
			p.errors = append(p.errors, fmt.Sprintf("Expected identifier in arrow function parameter list, but got %T", exp))
			return nil
		}
		params = append(params, ident)
	}

	arrow := &ast.ArrowFunction{
		Token:      lParenToken,
		Parameters: params,
	}

	arrow.Body = p.parseExpression(LOWEST)

	return arrow
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
	exp := &ast.ArrayIndexExpression{Token: p.curTok, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curTok, Left: left}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	exp.Index = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

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

func (p *Parser) parseHashOrStructLiteral() ast.Expression {
	if p.curTok.Type == token.IDENT {
		return p.parseStructLiteral()
	}

	return p.parseHashLiteral()
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curTok}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return hash
	}

	p.nextToken()

	key := p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)
	hash.Pairs[key] = value

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		key := p.parseExpression(LOWEST)

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
