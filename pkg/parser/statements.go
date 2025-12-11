package parser

import (
	"jabline/pkg/ast"
	"jabline/pkg/token"
)

func (p *Parser) parseStatement() ast.Statement {
	switch p.curTok.Type {
	case token.SEMICOLON:
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
		if p.peekTokenIs(token.LPAREN) {
			return p.parseExpressionStatement()
		}
		return p.parseFunctionStatement()
	case token.ASYNC:
		if p.peekTokenIs(token.FUNCTION) {
			if p.peekToken2Is(token.LPAREN) {
				return p.parseExpressionStatement()
			}
			return p.parseAsyncFunctionStatement()
		}
		return p.parseExpressionStatement()

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
		if p.isAssignmentStatement() {
			return p.parseFieldAssignmentStatement()
		}
		if p.curTok.Type == token.IDENT && p.peekTok.Type == token.LBRACE {
			return p.parseExpressionStatement()
		}
		return p.parseExpressionStatement()
	}
}

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

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curTok}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

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

func (p *Parser) parseForStatement() ast.Statement {
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	if p.peekTok.Type == token.IDENT {
		p.nextToken()
		if p.peekTok.Type == token.IN {

			return p.parseForEachStatement()
		}

	}

	return p.parseTraditionalForStatement()
}

func (p *Parser) parseTraditionalForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curTok}

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

	if !p.curTokenIs(token.SEMICOLON) {
		if !p.expectPeek(token.SEMICOLON) {
			return nil
		}
	}

	p.nextToken()
	if p.curTok.Type != token.SEMICOLON {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	p.nextToken()
	if p.curTok.Type != token.RPAREN {
		if p.curTok.Type == token.IDENT && p.peekTok.Type == token.ASSIGN {
			stmt.Update = p.parseAssignmentStatement()
		} else {
			stmt.Update = p.parseExpressionStatement()
		}
	}

	if !p.curTokenIs(token.RPAREN) {
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseForEachStatement() *ast.ForEachStatement {
	stmt := &ast.ForEachStatement{Token: p.curTok}

	stmt.Variable = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.IN) {
		return nil
	}

	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curTok}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

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

	if len(block.Statements) == 0 {
		p.errors = append(p.errors, "block statement must contain at least one statement")
	}

	if p.curTokenIs(token.EOF) && !p.curTokenIs(token.RBRACE) {
		p.errors = append(p.errors, "unclosed block: expected '}' before end of file")
		return nil
	}

	return block
}

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

func (p *Parser) parseFieldAssignmentStatement() ast.Statement {

	left := p.parseExpression(LOWEST)

	if !p.isAssignmentOperator(p.peekTok.Type) {

		return &ast.ExpressionStatement{Token: p.curTok, Expression: left}
	}

	stmt := &ast.AssignmentStatement{Token: p.curTok}
	stmt.Left = left
	p.nextToken()

	if p.curTok.Type != token.ASSIGN {
		arithmeticOp := p.getArithmeticOperator(p.curTok.Type)
		if arithmeticOp != token.ILLEGAL {
			p.nextToken()
			right := p.parseExpression(LOWEST)

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

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curTok}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curTok}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

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

func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{Token: p.curTok}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.TryBlock = p.parseBlockStatement()

	if p.peekTokenIs(token.CATCH) {
		p.nextToken()

		if p.peekTokenIs(token.LPAREN) {
			p.nextToken()
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

func (p *Parser) parseThrowStatement() *ast.ThrowStatement {
	stmt := &ast.ThrowStatement{Token: p.curTok}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curTok}

	if p.peekTokenIs(token.LBRACE) {
		stmt.ImportType = ast.IMPORT_NAMED
		p.nextToken()

		p.nextToken()
		for {
			if !p.curTokenIs(token.IDENT) {
				p.addError("expected identifier in import list, got %s", p.curTok.Literal)
				return nil
			}

			importItem := &ast.ImportItem{
				Name: &ast.Identifier{
					Token: p.curTok,
					Value: p.curTok.Literal,
				},
			}

			if p.peekTokenIs(token.IDENT) && p.peekTok.Literal == "as" {
				p.nextToken()
				if !p.expectPeek(token.IDENT) {
					return nil
				}
				importItem.Alias = &ast.Identifier{
					Token: p.curTok,
					Value: p.curTok.Literal,
				}
			}

			stmt.NamedImports = append(stmt.NamedImports, importItem)

			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
				p.nextToken()
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

		if !p.expectPeek(token.STRING) {
			return nil
		}

		stmt.ModuleName = &ast.StringLiteral{
			Token: p.curTok,
			Value: p.curTok.Literal,
		}
	} else if p.peekTokenIs(token.ASTERISK) {

		stmt.ImportType = ast.IMPORT_NAMESPACE
		p.nextToken()

		if !p.expectPeek(token.AS) {
			p.addError("expected 'as' after '*' in import statement")
			return nil
		}

		if !p.expectPeek(token.IDENT) {
			return nil
		}

		stmt.NamespaceAlias = &ast.Identifier{
			Token: p.curTok,
			Value: p.curTok.Literal,
		}

		if !p.expectPeek(token.FROM) {
			return nil
		}

		if !p.expectPeek(token.STRING) {
			return nil
		}

		stmt.ModuleName = &ast.StringLiteral{
			Token: p.curTok,
			Value: p.curTok.Literal,
		}
	} else if p.peekTokenIs(token.IDENT) {

		p.nextToken()

		defaultImport := &ast.Identifier{
			Token: p.curTok,
			Value: p.curTok.Literal,
		}

		if p.peekTokenIs(token.COMMA) {
			stmt.ImportType = ast.IMPORT_MIXED
			p.nextToken()

			if !p.expectPeek(token.LBRACE) {
				return nil
			}

			p.nextToken()
			for {
				if !p.curTokenIs(token.IDENT) {
					p.addError("expected identifier in named import list")
					return nil
				}

				importItem := &ast.ImportItem{
					Name: &ast.Identifier{
						Token: p.curTok,
						Value: p.curTok.Literal,
					},
				}

				if p.peekTokenIs(token.IDENT) && p.peekTok.Literal == "as" {
					p.nextToken()
					if !p.expectPeek(token.IDENT) {
						return nil
					}
					importItem.Alias = &ast.Identifier{
						Token: p.curTok,
						Value: p.curTok.Literal,
					}
				}

				stmt.NamedImports = append(stmt.NamedImports, importItem)

				if p.peekTokenIs(token.COMMA) {
					p.nextToken()
					p.nextToken()
				} else {
					break
				}
			}

			if !p.expectPeek(token.RBRACE) {
				return nil
			}
		} else {
			stmt.ImportType = ast.IMPORT_DEFAULT
		}

		stmt.DefaultImport = defaultImport

		if !p.expectPeek(token.FROM) {
			return nil
		}

		if !p.expectPeek(token.STRING) {
			return nil
		}

		stmt.ModuleName = &ast.StringLiteral{
			Token: p.curTok,
			Value: p.curTok.Literal,
		}
	} else if p.peekTokenIs(token.STRING) {

		stmt.ImportType = ast.IMPORT_SIDE_EFFECT
		p.nextToken()

		stmt.ModuleName = &ast.StringLiteral{
			Token: p.curTok,
			Value: p.curTok.Literal,
		}
	} else {
		p.addError("invalid import syntax")
		return nil
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExportStatement() *ast.ExportStatement {
	stmt := &ast.ExportStatement{Token: p.curTok}

	if p.peekTokenIs(token.LBRACE) {

		p.nextToken()

		p.nextToken()
		for {
			if !p.curTokenIs(token.IDENT) {
				p.addError("expected identifier in export list")
				return nil
			}

			exportItem := &ast.ExportItem{
				Name: &ast.Identifier{
					Token: p.curTok,
					Value: p.curTok.Literal,
				},
			}

			if p.peekTokenIs(token.IDENT) && p.peekTok.Literal == "as" {
				p.nextToken()
				if !p.expectPeek(token.IDENT) {
					return nil
				}
				exportItem.Alias = &ast.Identifier{
					Token: p.curTok,
					Value: p.curTok.Literal,
				}
			}

			stmt.ExportList = append(stmt.ExportList, exportItem)

			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
				p.nextToken()
			} else {
				break
			}
		}

		if !p.expectPeek(token.RBRACE) {
			return nil
		}

		if p.peekTokenIs(token.FROM) {
			stmt.ExportType = ast.EXPORT_NAMED_FROM
			p.nextToken()

			if !p.expectPeek(token.STRING) {
				return nil
			}

			stmt.ModuleName = &ast.StringLiteral{
				Token: p.curTok,
				Value: p.curTok.Literal,
			}
		} else {
			stmt.ExportType = ast.EXPORT_LIST
		}
	} else if p.peekTokenIs(token.ASTERISK) {

		p.nextToken()

		if p.peekTokenIs(token.IDENT) && p.peekTok.Literal == "as" {
			stmt.ExportType = ast.EXPORT_ALL_AS
			p.nextToken()
			if !p.expectPeek(token.IDENT) {
				return nil
			}
			stmt.NamespaceAlias = &ast.Identifier{
				Token: p.curTok,
				Value: p.curTok.Literal,
			}
		} else {
			stmt.ExportType = ast.EXPORT_ALL
		}

		if !p.expectPeek(token.FROM) {
			return nil
		}

		if !p.expectPeek(token.STRING) {
			return nil
		}

		stmt.ModuleName = &ast.StringLiteral{
			Token: p.curTok,
			Value: p.curTok.Literal,
		}
	} else if p.peekTokenIs(token.DEFAULT) {

		stmt.ExportType = ast.EXPORT_DEFAULT
		stmt.IsDefault = true
		p.nextToken()
		p.nextToken()

		stmt.Statement = p.parseStatement()
	} else {

		stmt.ExportType = ast.EXPORT_DECLARATION
		p.nextToken()
		stmt.Statement = p.parseStatement()
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

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

func (p *Parser) parseCaseClause() *ast.CaseClause {
	clause := &ast.CaseClause{Token: p.curTok}

	p.nextToken()
	clause.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	for !p.peekTokenIs(token.CASE) && !p.peekTokenIs(token.DEFAULT) && !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if stmt := p.parseStatement(); stmt != nil {
			clause.Statements = append(clause.Statements, stmt)
		}
	}

	return clause
}

func (p *Parser) parseDefaultClause() *ast.DefaultClause {
	clause := &ast.DefaultClause{Token: p.curTok}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	for !p.peekTokenIs(token.CASE) && !p.peekTokenIs(token.DEFAULT) && !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if stmt := p.parseStatement(); stmt != nil {
			clause.Statements = append(clause.Statements, stmt)
		}
	}

	return clause
}
