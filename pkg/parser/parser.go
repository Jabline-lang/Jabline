package parser

import (
	"jabline/pkg/ast"
	"jabline/pkg/lexer"
	"jabline/pkg/token"
)

// Parser representa el analizador sintáctico
type Parser struct {
	l       *lexer.Lexer
	curTok  token.Token
	peekTok token.Token

	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New crea una nueva instancia del parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Inicializar mapas de funciones de parsing
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	// Registrar funciones de parsing para expresiones prefix
	p.registerPrefixFunctions()

	// Registrar funciones de parsing para expresiones infix
	p.registerInfixFunctions()

	// Lee dos tokens para llenar curTok y peekTok
	p.nextToken()
	p.nextToken()
	return p
}

// registerPrefixFunctions registra todas las funciones de parsing prefix
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

// registerInfixFunctions registra todas las funciones de parsing infix
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

// registerPrefix registra una función de parsing prefix
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registra una función de parsing infix
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// nextToken avanza el parser al siguiente token
func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	p.peekTok = p.l.NextToken()
}

// Errors devuelve la lista de errores encontrados durante el parsing
func (p *Parser) Errors() []string {
	return p.errors
}

// ParseProgram parsea un programa completo y devuelve el AST
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
