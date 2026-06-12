package token

import "testing"

func TestTokenTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		typ  TokenType
		want string
	}{
		{"ILLEGAL", ILLEGAL, "ILLEGAL"},
		{"EOF", EOF, "EOF"},
		{"IDENT", IDENT, "IDENT"},
		{"INT", INT, "INT"},
		{"FLOAT", FLOAT, "FLOAT"},
		{"STRING", STRING, "STRING"},
		{"TEMPLATE_LITERAL", TEMPLATE_LITERAL, "TEMPLATE_LITERAL"},
		{"ECHO", ECHO, "ECHO"},
		{"LET", LET, "LET"},
		{"CONST", CONST, "CONST"},
		{"NULL", NULL, "NULL"},
		{"IF", IF, "IF"},
		{"ELSE", ELSE, "ELSE"},
		{"FOR", FOR, "FOR"},
		{"WHILE", WHILE, "WHILE"},
		{"FUNCTION", FUNCTION, "FUNCTION"},
		{"RETURN", RETURN, "RETURN"},
		{"STRUCT", STRUCT, "STRUCT"},
		{"BREAK", BREAK, "BREAK"},
		{"CONTINUE", CONTINUE, "CONTINUE"},
		{"IN", IN, "IN"},
		{"SWITCH", SWITCH, "SWITCH"},
		{"CASE", CASE, "CASE"},
		{"DEFAULT", DEFAULT, "DEFAULT"},
		{"TRY", TRY, "TRY"},
		{"CATCH", CATCH, "CATCH"},
		{"THROW", THROW, "THROW"},
		{"ASYNC", ASYNC, "ASYNC"},
		{"AWAIT", AWAIT, "AWAIT"},
		{"IMPORT", IMPORT, "IMPORT"},
		{"EXPORT", EXPORT, "EXPORT"},
		{"FROM", FROM, "FROM"},
		{"STRING_TYPE", STRING_TYPE, "STRING_TYPE"},
		{"INT_TYPE", INT_TYPE, "INT_TYPE"},
		{"BOOL_TYPE", BOOL_TYPE, "BOOL_TYPE"},
		{"TRUE", TRUE, "TRUE"},
		{"FALSE", FALSE, "FALSE"},
		{"ENUM", ENUM, "ENUM"},
		{"AS", AS, "AS"},
		{"RETRY", RETRY, "RETRY"},
		{"SERVICE", SERVICE, "SERVICE"},
		{"SPAWN", SPAWN, "SPAWN"},
		{"INTERFACE", INTERFACE, "INTERFACE"},
		{"MATCH", MATCH, "MATCH"},
		{"METER", METER, "METER"},
		{"TRACE", TRACE, "TRACE"},
		{"ASSIGN", ASSIGN, "="},
		{"PLUS", PLUS, "+"},
		{"MINUS", MINUS, "-"},
		{"ASTERISK", ASTERISK, "*"},
		{"SLASH", SLASH, "/"},
		{"BANG", BANG, "!"},
		{"EQ", EQ, "=="},
		{"NOT_EQ", NOT_EQ, "!="},
		{"LT", LT, "<"},
		{"GT", GT, ">"},
		{"LPAREN", LPAREN, "("},
		{"RPAREN", RPAREN, ")"},
		{"LBRACE", LBRACE, "{"},
		{"RBRACE", RBRACE, "}"},
		{"LBRACKET", LBRACKET, "["},
		{"RBRACKET", RBRACKET, "]"},
		{"COMMA", COMMA, ","},
		{"SEMICOLON", SEMICOLON, ";"},
		{"DOT", DOT, "."},
		{"COLON", COLON, ":"},
		{"AND", AND, "&&"},
		{"OR", OR, "||"},
		{"ARROW", ARROW, "=>"},
		{"PIPE", PIPE, "|>"},
		{"NULLISH_COALESCING", NULLISH_COALESCING, "??"},
		{"OPTIONAL_CHAINING", OPTIONAL_CHAINING, "?."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.typ) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, string(tt.typ), tt.want)
			}
		})
	}
}

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		input string
		want  TokenType
	}{
		{"fn", FUNCTION},
		{"let", LET},
		{"const", CONST},
		{"if", IF},
		{"else", ELSE},
		{"for", FOR},
		{"while", WHILE},
		{"return", RETURN},
		{"struct", STRUCT},
		{"true", TRUE},
		{"false", FALSE},
		{"null", NULL},
		{"break", BREAK},
		{"continue", CONTINUE},
		{"in", IN},
		{"switch", SWITCH},
		{"case", CASE},
		{"default", DEFAULT},
		{"try", TRY},
		{"catch", CATCH},
		{"throw", THROW},
		{"retry", RETRY},
		{"async", ASYNC},
		{"await", AWAIT},
		{"import", IMPORT},
		{"export", EXPORT},
		{"from", FROM},
		{"spawn", SPAWN},
		{"interface", INTERFACE},
		{"string", STRING_TYPE},
		{"int", INT_TYPE},
		{"int8", INT8_TYPE},
		{"int16", INT16_TYPE},
		{"int32", INT32_TYPE},
		{"int64", INT64_TYPE},
		{"uint8", UINT8_TYPE},
		{"uint16", UINT16_TYPE},
		{"uint32", UINT32_TYPE},
		{"uint64", UINT64_TYPE},
		{"float", FLOAT_TYPE},
		{"float32", FLOAT32_TYPE},
		{"float64", FLOAT64_TYPE},
		{"bool", BOOL_TYPE},
		{"enum", ENUM},
		{"as", AS},
		{"match", MATCH},
		{"meter", METER},
		{"trace", TRACE},
		{"service", SERVICE},
		{"echo", ECHO},
		{"myVariable", IDENT},
		{"x", IDENT},
		{"_foo", IDENT},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := LookupIdent(tt.input); got != tt.want {
				t.Errorf("LookupIdent(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTokenStruct(t *testing.T) {
	tok := Token{Type: IDENT, Literal: "foo", Line: 5, Column: 3}
	if tok.Type != IDENT {
		t.Errorf("Token.Type = %v, want IDENT", tok.Type)
	}
	if tok.Literal != "foo" {
		t.Errorf("Token.Literal = %q, want %q", tok.Literal, "foo")
	}
	if tok.Line != 5 {
		t.Errorf("Token.Line = %d, want 5", tok.Line)
	}
	if tok.Column != 3 {
		t.Errorf("Token.Column = %d, want 3", tok.Column)
	}
}
