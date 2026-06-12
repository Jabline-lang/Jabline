package lexer

import (
	"jabline/pkg/token"
	"strings"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}

	// Detect and skip BOMs
	if strings.HasPrefix(l.input, "\xef\xbb\xbf") { // UTF-8
		l.readPosition = 3
	} else if strings.HasPrefix(l.input, "\xff\xfe") || strings.HasPrefix(l.input, "\xfe\xff") { // UTF-16
		l.readPosition = 2
	}

	l.readChar()
	return l
}

func (l *Lexer) newToken(tokenType token.TokenType, literal string) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column,
	}
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	var tok token.Token

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.EQ, string(ch)+string(l.ch))
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.ARROW, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.ASSIGN, string(l.ch))
		}
	case '+':
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.INCREMENT, string(ch)+string(l.ch))
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.PLUS_ASSIGN, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.PLUS, string(l.ch))
		}
	case '-':
		if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.DECREMENT, string(ch)+string(l.ch))
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.SUB_ASSIGN, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.MINUS, string(l.ch))
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.NOT_EQ, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.BANG, string(l.ch))
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.MUL_ASSIGN, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.ASTERISK, string(l.ch))
		}
	case '/':
		if l.peekChar() == '/' {
			l.skipComment()
			return l.NextToken()
		} else if l.peekChar() == '*' {
			l.skipMultiLineComment()
			return l.NextToken()
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.DIV_ASSIGN, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.SLASH, string(l.ch))
		}
	case '%':
		tok = l.newToken(token.MOD, string(l.ch))
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.AND, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.BIT_AND, string(l.ch))
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.OR, string(ch)+string(l.ch))
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.PIPE, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.BIT_OR, string(l.ch))
		}
	case '^':
		tok = l.newToken(token.BIT_XOR, string(l.ch))
	case '~':
		tok = l.newToken(token.BIT_NOT, string(l.ch))
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.LT_EQ, string(ch)+string(l.ch))
		} else if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.SHIFT_LEFT, string(ch)+string(l.ch))
		} else if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.ARROW_LEFT, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.LT, string(l.ch))
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.GT_EQ, string(ch)+string(l.ch))
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.SHIFT_RIGHT, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.GT, string(l.ch))
		}
	case '(':
		tok = l.newToken(token.LPAREN, string(l.ch))
	case ')':
		tok = l.newToken(token.RPAREN, string(l.ch))
	case '{':
		tok = l.newToken(token.LBRACE, string(l.ch))
	case '}':
		tok = l.newToken(token.RBRACE, string(l.ch))
	case '[':
		tok = l.newToken(token.LBRACKET, string(l.ch))
	case ']':
		tok = l.newToken(token.RBRACKET, string(l.ch))
	case ',':
		tok = l.newToken(token.COMMA, string(l.ch))
	case ';':
		tok = l.newToken(token.SEMICOLON, string(l.ch))
	case ':':
		tok = l.newToken(token.COLON, string(l.ch))
	case '.':
		if l.peekChar() == '.' {
			// Look ahead two characters to see if this is ...
			nextNext := byte(0)
			if l.readPosition+1 < len(l.input) {
				nextNext = l.input[l.readPosition+1]
			}
			if nextNext == '.' {
				// Three dots: ... → ELLIPSIS
				tok = l.newToken(token.ELLIPSIS, "...")
				l.readChar() // consume second dot
				l.readChar() // consume third dot
				break
			}
		}
		tok = l.newToken(token.DOT, string(l.ch))
	case '?':
		if l.peekChar() == '?' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.NULLISH_COALESCING, string(ch)+string(l.ch))
		} else if l.peekChar() == '.' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.OPTIONAL_CHAINING, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(token.QUESTION, string(l.ch))
		}
	case '"':
		line := l.line
		column := l.column
		lit := l.readString()
		tok = token.Token{Type: token.STRING, Literal: lit, Line: line, Column: column}
	case '`':
		line := l.line
		column := l.column
		lit := l.readTemplateLiteral()
		tok = token.Token{Type: token.TEMPLATE_LITERAL, Literal: lit, Line: line, Column: column}
	case 0:
		return l.newToken(token.EOF, "")
	default:
		if isLetter(l.ch) {
			line := l.line
			column := l.column
			lit := l.readIdentifier()
			return token.Token{Type: token.LookupIdent(lit), Literal: lit, Line: line, Column: column}
		} else if isDigit(l.ch) {
			line := l.line
			column := l.column
			numLiteral := l.readNumber()
			tokenType := token.TokenType(token.INT)
			if strings.Contains(numLiteral, ".") {
				tokenType = token.FLOAT
			}
			return token.Token{Type: tokenType, Literal: numLiteral, Line: line, Column: column}
		} else {
			tok = l.newToken(token.ILLEGAL, string(l.ch))
		}
	}

	l.readChar()
	return tok
}
