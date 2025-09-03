package parser

import (
	"jabline/pkg/ast"
	"jabline/pkg/token"
)

// parseStatement parsea un statement basado en el token actual
func (p *Parser) parseStatement() ast.Statement {
	switch p.curTok.Type {
	case token.SEMICOLON:
		// Empty statement - just return nil
		return nil
	case token.LET:
		return p.parseLetStatement()
	case token.CONST:
		return p.parseConstStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.ECHO:
		return p.parseEchoStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.FUNCTION:
		return p.parseFunctionStatement()
	case token.ASYNC:
		return p.parseAsyncFunctionStatement()
	case token.STRUCT:
		return p.parseStructStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.TRY:
		return p.parseTryStatement()
	case token.THROW:
		return p.parseThrowStatement()
	case token.SWITCH:
		return p.parseSwitchStatement()
	case token.IMPORT:
		return p.parseImportStatement()
	case token.EXPORT:
		return p.parseExportStatement()
	default:
		// Verificar si es una asignación (variable = valor o obj.campo = valor)
		if p.isAssignmentStatement() {
			return p.parseFieldAssignmentStatement()
		}
		// Verificar si es un struct literal
		if p.curTok.Type == token.IDENT && p.peekTok.Type == token.LBRACE {
			return p.parseExpressionStatement()
		}
		return p.parseExpressionStatement()
	}
}

// parseLetStatement parsea una declaración let
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curTok}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement parsea una declaración return
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curTok}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseEchoStatement parsea una declaración echo
func (p *Parser) parseEchoStatement() *ast.EchoStatement {
	stmt := &ast.EchoStatement{Token: p.curTok}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseWhileStatement parsea una declaración while
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curTok}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForStatement parsea una declaración for (tradicional o for-each)
func (p *Parser) parseForStatement() ast.Statement {
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Verificación simple: si el siguiente token es IDENT, verificar si es seguido por IN
	if p.peekTok.Type == token.IDENT {
		// Mirar hacia adelante verificando manualmente los siguientes tokens
		p.nextToken() // mover a IDENT
		if p.peekTok.Type == token.IN {
			// Este es un bucle for-each
			return p.parseForEachStatement()
		}
		// Este es un bucle for tradicional, continuar parseando
	}

	return p.parseTraditionalForStatement()
}

// parseTraditionalForStatement parsea un bucle for tradicional
func (p *Parser) parseTraditionalForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curTok}

	// Parsear init (let i = 0) - podríamos estar ya en el token init
	if p.curTok.Type != token.SEMICOLON {
		if p.curTok.Type != token.LET && p.curTok.Type != token.IDENT {
			p.nextToken()
		}
	} else {
		p.nextToken()
	}
	if p.curTok.Type != token.SEMICOLON {
		stmt.Init = p.parseStatement()
	}

	// Saltar al siguiente punto y coma
	if !p.curTokenIs(token.SEMICOLON) {
		if !p.expectPeek(token.SEMICOLON) {
			return nil
		}
	}

	// Parsear condición (i < 5)
	p.nextToken()
	if p.curTok.Type != token.SEMICOLON {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	// Saltar al siguiente punto y coma
	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	// Parsear actualización (i = i + 1)
	p.nextToken()
	if p.curTok.Type != token.RPAREN {
		if p.curTok.Type == token.IDENT && p.peekTok.Type == token.ASSIGN {
			stmt.Update = p.parseAssignmentStatement()
		} else {
			stmt.Update = p.parseExpressionStatement()
		}
	}

	// Saltar al paréntesis de cierre
	if !p.curTokenIs(token.RPAREN) {
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	// Esperar llave de apertura para el cuerpo
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parsear cuerpo
	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForEachStatement parsea un bucle for-each
func (p *Parser) parseForEachStatement() *ast.ForEachStatement {
	stmt := &ast.ForEachStatement{Token: p.curTok}

	// El cursor ya debería estar en el identificador (variable del bucle)
	stmt.Variable = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	// Esperar 'in'
	if !p.expectPeek(token.IN) {
		return nil
	}

	// Parsear el iterable
	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	// Esperar paréntesis de cierre
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Esperar llave de apertura para el cuerpo
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parsear cuerpo
	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseExpressionStatement parsea una declaración de expresión
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curTok}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseBlockStatement parsea un bloque de declaraciones
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curTok}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseAssignmentStatement parsea una declaración de asignación simple
func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	stmt := &ast.AssignmentStatement{Token: p.curTok}

	stmt.Left = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// parseFieldAssignmentStatement parsea asignación a campos o índices de array
func (p *Parser) parseFieldAssignmentStatement() ast.Statement {
	// Parsear el lado izquierdo (puede ser obj.field o arr[index])
	left := p.parseExpression(LOWEST)

	// Verificar que el siguiente token sea un operador de asignación
	if !p.isAssignmentOperator(p.peekTok.Type) {
		// No es una asignación, devolver como expression statement
		return &ast.ExpressionStatement{Token: p.curTok, Expression: left}
	}

	// Es una asignación
	stmt := &ast.AssignmentStatement{Token: p.curTok}
	stmt.Left = left
	p.nextToken()

	// Si es un operador de asignación compuesto (+=, -=, etc.), crear una expresión infija
	if p.curTok.Type != token.ASSIGN {
		arithmeticOp := p.getArithmeticOperator(p.curTok.Type)
		if arithmeticOp != token.ILLEGAL {
			p.nextToken()
			right := p.parseExpression(LOWEST)

			// Crear expresión infija: left + right
			var opLiteral string
			switch arithmeticOp {
			case token.PLUS:
				opLiteral = "+"
			case token.MINUS:
				opLiteral = "-"
			case token.ASTERISK:
				opLiteral = "*"
			case token.SLASH:
				opLiteral = "/"
			}

			infixExpr := &ast.InfixExpression{
				Token:    token.Token{Type: arithmeticOp, Literal: opLiteral},
				Left:     left,
				Operator: opLiteral,
				Right:    right,
			}
			stmt.Value = infixExpr
		} else {
			p.nextToken()
			stmt.Value = p.parseExpression(LOWEST)
		}
	} else {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	return stmt
}

// parseBreakStatement parsea una declaración break
func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curTok}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseContinueStatement parsea una declaración continue
func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curTok}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseAsyncFunctionStatement parsea statements de funciones async como async fn getData() { ... }
func (p *Parser) parseAsyncFunctionStatement() ast.Statement {
	stmt := &ast.AsyncFunctionStatement{Token: p.curTok}

	if !p.expectPeek(token.FUNCTION) {
		return nil
	}

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

// parseTryStatement parsea una declaración try-catch
func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{Token: p.curTok}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.TryBlock = p.parseBlockStatement()

	if p.peekTokenIs(token.CATCH) {
		p.nextToken() // move to CATCH

		// Check if there's a parameter for the exception
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken() // move to LPAREN
			if !p.expectPeek(token.IDENT) {
				return nil
			}
			stmt.CatchParam = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		stmt.CatchBlock = p.parseBlockStatement()
	}

	return stmt
}

// parseThrowStatement parsea una declaración throw
func (p *Parser) parseThrowStatement() *ast.ThrowStatement {
	stmt := &ast.ThrowStatement{Token: p.curTok}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseImportStatement parsea una declaración import
func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curTok}

	// Check for selective import: import { item1, item2 } from "module"
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken() // move to {

		p.nextToken() // move to first identifier
		for {
			if !p.curTokenIs(token.IDENT) {
				return nil
			}

			stmt.ImportList = append(stmt.ImportList, &ast.Identifier{
				Token: p.curTok,
				Value: p.curTok.Literal,
			})

			if p.peekTokenIs(token.COMMA) {
				p.nextToken() // move to comma
				p.nextToken() // move to next identifier
			} else {
				break
			}
		}

		if !p.expectPeek(token.RBRACE) {
			return nil
		}

		if !p.expectPeek(token.FROM) {
			return nil
		}
	}

	if !p.expectPeek(token.STRING) {
		return nil
	}

	stmt.ModuleName = &ast.StringLiteral{
		Token: p.curTok,
		Value: p.curTok.Literal,
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExportStatement parsea una declaración export
func (p *Parser) parseExportStatement() *ast.ExportStatement {
	stmt := &ast.ExportStatement{Token: p.curTok}

	p.nextToken() // move to the statement being exported

	stmt.Statement = p.parseStatement()

	return stmt
}

// parseConstStatement parsea una declaración const
func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curTok}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseSwitchStatement parsea una declaración switch
func (p *Parser) parseSwitchStatement() *ast.SwitchStatement {
	stmt := &ast.SwitchStatement{Token: p.curTok}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse cases and default
	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()

		if p.curTok.Type == token.CASE {
			caseClause := p.parseCaseClause()
			if caseClause != nil {
				stmt.Cases = append(stmt.Cases, caseClause)
			}
		} else if p.curTok.Type == token.DEFAULT {
			if stmt.DefaultCase != nil {
				p.errors = append(p.errors, "multiple default clauses in switch statement")
				return nil
			}
			stmt.DefaultCase = p.parseDefaultClause()
		} else {
			p.errors = append(p.errors, "expected 'case' or 'default' in switch body")
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return stmt
}

// parseCaseClause parsea una cláusula case
func (p *Parser) parseCaseClause() *ast.CaseClause {
	clause := &ast.CaseClause{Token: p.curTok}

	p.nextToken()
	clause.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	// Parse statements until next case, default, or end of switch
	for !p.peekTokenIs(token.CASE) && !p.peekTokenIs(token.DEFAULT) && !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if stmt := p.parseStatement(); stmt != nil {
			clause.Statements = append(clause.Statements, stmt)
		}
	}

	return clause
}

// parseDefaultClause parsea una cláusula default
func (p *Parser) parseDefaultClause() *ast.DefaultClause {
	clause := &ast.DefaultClause{Token: p.curTok}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	// Parse statements until end of switch
	for !p.peekTokenIs(token.CASE) && !p.peekTokenIs(token.DEFAULT) && !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if stmt := p.parseStatement(); stmt != nil {
			clause.Statements = append(clause.Statements, stmt)
		}
	}

	return clause
}
