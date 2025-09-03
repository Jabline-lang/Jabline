package parser

import (
	"fmt"
	"strconv"

	"jabline/pkg/ast"
	"jabline/pkg/lexer"
	"jabline/pkg/token"
)

// parseExpression parsea una expresión con la precedencia dada
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curTok.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curTok.Type)
		return nil
	}

	var leftExp ast.Expression

	// Verificar si esto es un struct literal
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

// parseIdentifier parsea un identificador
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
}

// parseIntegerLiteral parsea un literal entero
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curTok}

	value, err := strconv.ParseInt(p.curTok.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curTok.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral parsea un literal de punto flotante
func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curTok}

	value, err := strconv.ParseFloat(p.curTok.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curTok.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parsea un literal de cadena
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curTok, Value: p.curTok.Literal}
}

// parseTemplateLiteral parsea un template literal como `Hello ${name}`
func (p *Parser) parseTemplateLiteral() ast.Expression {
	lit := &ast.TemplateLiteral{Token: p.curTok}

	// Parsear el contenido del template literal
	content := p.curTok.Literal
	lit.Parts, lit.Expressions = p.parseTemplateContent(content)

	return lit
}

// parseTemplateContent parsea el contenido de un template literal y extrae partes de texto y expresiones
func (p *Parser) parseTemplateContent(content string) ([]string, []ast.Expression) {
	parts := []string{}
	expressions := []ast.Expression{}

	currentPart := ""
	i := 0

	for i < len(content) {
		if i < len(content)-1 && content[i] == '$' && content[i+1] == '{' {
			// Encontramos inicio de interpolación ${
			parts = append(parts, currentPart)
			currentPart = ""
			i += 2 // saltar ${

			// Encontrar el cierre }
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
				// Extraer la expresión
				exprContent := content[exprStart:i]

				// Crear un mini-lexer y parser para la expresión
				exprLexer := lexer.New(exprContent)
				exprParser := New(exprLexer)
				expr := exprParser.parseExpression(1) // LOWEST precedence

				if expr != nil {
					expressions = append(expressions, expr)
				} else {
					// Si no se puede parsear, tratar como string literal
					expressions = append(expressions, &ast.StringLiteral{
						Token: token.Token{Type: token.STRING, Literal: exprContent},
						Value: exprContent,
					})
				}
				i++ // saltar el }
			} else {
				// No se encontró cierre, tratar como texto normal
				currentPart += "${"
			}
		} else {
			currentPart += string(content[i])
			i++
		}
	}

	// Agregar la última parte
	parts = append(parts, currentPart)

	return parts, expressions
}

// parseBoolean parsea un literal booleano
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curTok, Value: p.curTokenIs(token.TRUE)}
}

// parsePrefixExpression parsea una expresión prefija como -x o !x
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curTok,
		Operator: p.curTok.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression parsea una expresión infija como x + y
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

// parseGroupedExpression parsea una expresión agrupada como (x + y)
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseGroupedOrArrowFunction parsea expresión agrupada o arrow function
func (p *Parser) parseGroupedOrArrowFunction() ast.Expression {
	// For now, just parse as grouped expression to avoid conflicts
	// Arrow functions with parentheses are parsed separately when they start with (
	return p.parseGroupedExpression()
}

// parseIfExpression parsea una expresión if
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

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// parseArrayLiteral parsea literales de array como [1, 2, 3]
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curTok}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// parseArrayIndexExpression parsea acceso por índice como arr[0]
func (p *Parser) parseArrayIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.ArrayIndexExpression{Token: p.curTok, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

// parseIndexExpression parsea acceso a campo como obj.field
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curTok, Left: left}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	exp.Index = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	return exp
}

// parseCallExpression parsea una llamada a función
func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curTok, Function: fn}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// parseExpressionList parsea una lista de expresiones separadas por comas
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

// parsePostfixExpression parsea expresiones postfijas como x++ o x--
func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	return &ast.PostfixExpression{
		Token:    p.curTok,
		Left:     left,
		Operator: p.curTok.Literal,
	}
}

// parseHashOrStructLiteral parsea tanto hash literals como struct literals
func (p *Parser) parseHashOrStructLiteral() ast.Expression {
	// Si el token actual es un identificador seguido de {, es un struct literal
	if p.curTok.Type == token.IDENT {
		return p.parseStructLiteral()
	}
	// De lo contrario, es un hash literal
	return p.parseHashLiteral()
}

// parseHashLiteral parsea un literal hash como {"key": "value"}
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curTok}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return hash
	}

	p.nextToken()

	// Parsear el primer par clave-valor
	key := p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)
	hash.Pairs[key] = value

	// Parsear el resto de pares
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consumir ','
		p.nextToken() // moverse a la clave

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

// parseTernaryExpression parsea expresiones ternarias como condition ? trueValue : falseValue
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

// parseNullishCoalescingExpression parsea expresiones de coalescencia nula como a ?? b
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

// parseOptionalChainingExpression parsea expresiones de encadenamiento opcional como obj?.prop
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

// parseNull parsea un literal null
func (p *Parser) parseNull() ast.Expression {
	return &ast.Null{Token: p.curTok}
}

// parseAsyncFunctionLiteral parsea literales de funciones async como async fn(x, y) { ... }
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

// parseAwaitExpression parsea expresiones await como await somePromise
func (p *Parser) parseAwaitExpression() ast.Expression {
	expression := &ast.AwaitExpression{Token: p.curTok}

	p.nextToken()

	expression.Value = p.parseExpression(PREFIX)

	return expression
}
