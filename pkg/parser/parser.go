package parser

import (
	"jabline/pkg/ast"
	"jabline/pkg/lexer"
	"jabline/pkg/token"
)

type Parser struct {
	l        *lexer.Lexer
	curTok   token.Token
	peekTok  token.Token
	peekTok2 token.Token

	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	p.registerPrefixFunctions()

	p.registerInfixFunctions()

	p.nextToken()
	p.nextToken()
	p.nextToken()
	return p
}


func (p *Parser) registerPrefixFunctions() {
	p.registerPrefix(token.IDENT, p.parseArrowFunctionFromIdent)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.TEMPLATE_LITERAL, p.parseTemplateLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.NULL, p.parseNull)
	p.registerPrefix(token.LPAREN, p.parseGroupedOrArrowFunction)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.ASYNC, p.parseAsyncFunctionLiteral)
	p.registerPrefix(token.AWAIT, p.parseAwaitExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashOrStructLiteral)
}

func (p *Parser) registerInfixFunctions() {
	p.registerInfix(token.NULLISH_COALESCING, p.parseNullishCoalescingExpression)
	p.registerInfix(token.OPTIONAL_CHAINING, p.parseOptionalChainingExpression)
	p.registerInfix(token.QUESTION, p.parseTernaryExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.GT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.DOT, p.parseIndexExpression)
	p.registerInfix(token.LBRACKET, p.parseArrayIndexExpression)
	p.registerInfix(token.INCREMENT, p.parsePostfixExpression)
	p.registerInfix(token.DECREMENT, p.parsePostfixExpression)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	p.peekTok = p.peekTok2
	p.peekTok2 = p.l.NextToken()
}

func (p *Parser) peekToken2Is(t token.TokenType) bool {
	return p.peekTok2.Type == t
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}
	prog.Statements = []ast.Statement{}

	for p.curTok.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
		p.nextToken()
	}

	return prog
}
