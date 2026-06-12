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
	case token.DEFER:
		return p.parseDeferStatement()
	case token.DO:
		return p.parseDoWhileStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.FUNCTION:
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
	case token.INTERFACE:
		return p.parseInterfaceStatement()
	case token.BREAK:

		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.TRY:
		return p.parseTryStatement()
	case token.RETRY:
		return p.parseRetryStatement()
	case token.SERVICE:
		return p.parseServiceStatement()
	case token.THROW:
		return p.parseThrowStatement()
	case token.SWITCH:
		return p.parseSwitchStatement()
	case token.IMPORT:
		return p.parseImportStatement()
	case token.EXPORT:
		return p.parseExportStatement()
	case token.ENUM:
		return p.parseEnumStatement()
	case token.MATCH:
		return p.parseMatchStatement()
	case token.SELECT:
		return p.parseSelectStatement()
	case token.METER:
		return p.parseMeterStatement()
	case token.TRACE:
		return p.parseTraceStatement()
	case token.ALIAS:
		return p.parseTypeAliasStatement()
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

func (p *Parser) parseTypeAliasStatement() *ast.TypeAliasStatement {
	stmt := &ast.TypeAliasStatement{Token: p.curTok}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Type = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curTok}

	// Check for destructuring: let [a, b] = ... or let {x, y} = ...
	if p.peekTokenIs(token.LBRACKET) || p.peekTokenIs(token.LBRACE) {
		isHash := p.peekTokenIs(token.LBRACE)
		p.nextToken() // consume LBRACKET or LBRACE
		stmt.Destructure = p.parseDestructuringPattern(isHash)

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

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	// Optional type annotation: `let x: int = 5;`
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume COLON
		p.nextToken() // move to type token
		stmt.Type = p.parseTypeExpression()
	}

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

func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curTok}

	// Check for destructuring: const [a, b] = ... or const {x, y} = ...
	if p.peekTokenIs(token.LBRACKET) || p.peekTokenIs(token.LBRACE) {
		isHash := p.peekTokenIs(token.LBRACE)
		p.nextToken() // consume LBRACKET or LBRACE
		stmt.Destructure = p.parseDestructuringPattern(isHash)

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

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	// Optional type annotation: `const x: int = 5;`
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // consume COLON
		p.nextToken() // move to type token
		stmt.Type = p.parseTypeExpression()
	}

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

	if p.curTokenIs(token.SEMICOLON) {
		return stmt
	}

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

	stmt.Values = p.parseExpressionList(token.RPAREN)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseDeferStatement() *ast.DeferStatement {
	stmt := &ast.DeferStatement{Token: p.curTok}

	p.nextToken()
	stmt.Call = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseDoWhileStatement() *ast.DoWhileStatement {
	stmt := &ast.DoWhileStatement{Token: p.curTok}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	if !p.expectPeek(token.WHILE) {
		return nil
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

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

	// Optional return type: `async fn name(params): int`
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

func (p *Parser) parseServiceStatement() *ast.ServiceStatement {
	stmt := &ast.ServiceStatement{Token: p.curTok}
	stmt.Fields = make(map[string]ast.Expression)
	stmt.Methods = []*ast.FunctionStatement{}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()

		// Method: fn name() {}
		if p.curTok.Type == token.FUNCTION {
			fnStmt := p.parseFunctionStatement()
			if fnNode, ok := fnStmt.(*ast.FunctionStatement); ok {
				// Implicitly bind to this service? Compiler handles name mangling.
				stmt.Methods = append(stmt.Methods, fnNode)
			}
			continue
		}

		// Field: name: value (Config)
		if p.curTok.Type == token.IDENT {
			key := p.curTok.Literal
			if !p.expectPeek(token.COLON) {
				return nil
			}
			p.nextToken() // Skip colon

			val := p.parseExpression(LOWEST)
			stmt.Fields[key] = val

			if p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
			// Comma support for fields
			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
			}
			continue
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return stmt
}

func (p *Parser) parseRetryStatement() *ast.RetryStatement {
	stmt := &ast.RetryStatement{Token: p.curTok}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Attempts = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.RetryBlock = p.parseBlockStatement()

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

			// Parse optional type annotation: catch(e: TypeName)
			if p.peekTokenIs(token.COLON) {
				p.nextToken() // consume ':'
				p.nextToken() // move to type name
				stmt.CatchType = p.parseTypeExpression()
			}

			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		stmt.CatchBlock = p.parseBlockStatement()
	}

	// Parse optional finally block
	if p.peekTokenIs(token.FINALLY) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		stmt.Finally = p.parseBlockStatement()
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
		p.nextToken()
		stmt.ModuleName = &ast.StringLiteral{
			Token: p.curTok,
			Value: p.curTok.Literal,
		}

		if p.peekTokenIs(token.AS) {
			stmt.ImportType = ast.IMPORT_ALIAS
			p.nextToken()
			if !p.expectPeek(token.IDENT) {
				return nil
			}
			stmt.NamespaceAlias = &ast.Identifier{
				Token: p.curTok,
				Value: p.curTok.Literal,
			}
		} else {
			stmt.ImportType = ast.IMPORT_SIDE_EFFECT
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

// parseEnumStatement parses: enum Name { Variant1, Variant2, ... }
func (p *Parser) parseEnumStatement() *ast.EnumStatement {
	stmt := &ast.EnumStatement{Token: p.curTok}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Values = []*ast.Identifier{}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if !p.curTokenIs(token.IDENT) {
			p.addError("expected identifier in enum body, got %s", p.curTok.Literal)
			return nil
		}
		variant := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
		stmt.Values = append(stmt.Values, variant)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // consume comma
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
func (p *Parser) parseMatchStatement() *ast.MatchStatement {
	stmt := &ast.MatchStatement{Token: p.curTok}

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
			matchCase := p.parseMatchCase()
			if matchCase != nil {
				stmt.Cases = append(stmt.Cases, matchCase)
			}
		} else if p.curTok.Type == token.DEFAULT {
			matchCase := p.parseMatchCase() // Reusing the same function as it handles default
			if matchCase != nil {
				stmt.Cases = append(stmt.Cases, matchCase)
			}
		} else {
			p.addError("expected 'case' or 'default' in match body, got %s", p.curTok.Literal)
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return stmt
}

func (p *Parser) parseMatchCase() *ast.MatchCase {
	clause := &ast.MatchCase{Token: p.curTok}

	if p.curTokenIs(token.DEFAULT) {
		clause.IsDefault = true
	} else {
		p.nextToken()
		clause.Pattern = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	// Support optional braced block body: case X: { ... }
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken() // consume '{'
		block := p.parseBlockStatement()
		if block != nil {
			clause.Statements = block.Statements
		}
	} else {
		// Unbraced: collect statements until next case/default/closing brace
		for !p.peekTokenIs(token.CASE) && !p.peekTokenIs(token.DEFAULT) && !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
			p.nextToken()
			if stmt := p.parseStatement(); stmt != nil {
				clause.Statements = append(clause.Statements, stmt)
			}
		}
	}

	return clause
}
func (p *Parser) parseSelectStatement() *ast.SelectStatement {
	stmt := &ast.SelectStatement{Token: p.curTok}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if p.curTokenIs(token.CASE) {
			c := p.parseSelectCase()
			if c != nil {
				stmt.Cases = append(stmt.Cases, c)
			}
		} else if p.curTokenIs(token.DEFAULT) {
			stmt.DefaultCase = p.parseSelectDefault()
		} else {
			p.addError("expected 'case' or 'default' in select body, got %s", p.curTok.Literal)
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return stmt
}

func (p *Parser) parseSelectCase() *ast.SelectCase {
	c := &ast.SelectCase{Token: p.curTok}
	p.nextToken()

	if p.curTokenIs(token.ARROW_LEFT) {
		// case <-ch:  (recv without binding)
		p.nextToken() // consume <-
		c.IsSend = false
		c.Channel = p.parseExpression(LOWEST)
	} else if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ARROW_LEFT) {
		// case ch <- expr:  (send) — parse channel ident directly
		ch := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
		p.nextToken() // consume ident
		p.nextToken() // consume <-
		c.IsSend = true
		c.Channel = ch
		c.Value = p.parseExpression(LOWEST)
	} else {
		p.addError("expected '<-' or 'channel <-' in select case, got %s", p.curTok.Literal)
		return nil
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		block := p.parseBlockStatement()
		if block != nil {
			c.Statements = block.Statements
		}
	} else {
		for !p.peekTokenIs(token.CASE) && !p.peekTokenIs(token.DEFAULT) && !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
			p.nextToken()
			if stmt := p.parseStatement(); stmt != nil {
				c.Statements = append(c.Statements, stmt)
			}
		}
	}
	return c
}

func (p *Parser) parseSelectDefault() *ast.DefaultClause {
	clause := &ast.DefaultClause{Token: p.curTok}
	if !p.expectPeek(token.COLON) {
		return nil
	}
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		block := p.parseBlockStatement()
		if block != nil {
			clause.Statements = block.Statements
		}
	} else {
		for !p.peekTokenIs(token.CASE) && !p.peekTokenIs(token.DEFAULT) && !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
			p.nextToken()
			if stmt := p.parseStatement(); stmt != nil {
				clause.Statements = append(clause.Statements, stmt)
			}
		}
	}
	return clause
}

func (p *Parser) parseMeterStatement() *ast.MeterStatement {
	stmt := &ast.MeterStatement{Token: p.curTok}

	p.nextToken()
	stmt.Name = p.parseExpression(POSTFIX)

	if p.peekTokenIs(token.INCREMENT) {
		p.nextToken()
		stmt.Operator = "++"
	} else if p.peekTokenIs(token.DECREMENT) {
		p.nextToken()
		stmt.Operator = "--"
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseTraceStatement() *ast.TraceStatement {
	stmt := &ast.TraceStatement{Token: p.curTok}

	p.nextToken()
	stmt.Name = p.parseExpression(POSTFIX)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseDestructuringPattern(isHash bool) *ast.DestructuringPattern {
	dp := &ast.DestructuringPattern{
		Token:  p.curTok,
		IsHash: isHash,
	}

	if isHash {
		// Parse {x, y, z: alias}
		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return dp
		}
		p.nextToken()

		for {
			if p.curTokenIs(token.RBRACE) {
				break
			}

			field := ast.DestructuringField{}

			if p.curTokenIs(token.IDENT) {
				field.Key = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
				field.Value = field.Key

				// Check for alias: {orig: newName}
				if p.peekTokenIs(token.COLON) {
					p.nextToken() // consume COLON
					if !p.expectPeek(token.IDENT) {
						return nil
					}
					field.Value = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
				}
			}

			dp.Fields = append(dp.Fields, field)

			if !p.peekTokenIs(token.COMMA) {
				break
			}
			p.nextToken() // consume COMMA
			p.nextToken() // move to next field
		}

		if !p.expectPeek(token.RBRACE) {
			return nil
		}
	} else {
		// Parse [a, b, ...rest]
		if p.peekTokenIs(token.RBRACKET) {
			p.nextToken()
			return dp
		}
		p.nextToken()

		for {
			if p.curTokenIs(token.RBRACKET) {
				break
			}

			field := ast.DestructuringField{}

			if p.curTokenIs(token.ELLIPSIS) {
				field.Rest = true
				if !p.expectPeek(token.IDENT) {
					return nil
				}
				field.Value = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
			} else if p.curTokenIs(token.IDENT) {
				field.Value = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
			}

			dp.Fields = append(dp.Fields, field)

			if field.Rest || !p.peekTokenIs(token.COMMA) {
				break
			}
			p.nextToken() // consume COMMA
			p.nextToken() // move to next field
		}

		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
	}

	return dp
}
