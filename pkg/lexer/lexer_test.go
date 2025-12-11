package lexer

import (
	"jabline/pkg/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5;
let ten = 10;

let add = fn(x, y) {
  x + y;
};

let result = add(five, ten);
!-*5;
5 < 10 > 5;

if (ten > five) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
10 <= 10;
10 >= 9;
10 += 5;
10 -= 5;
10 *= 5;
10 /= 5;
10 ?? 0;
null?.property;
"foobar";
"foo bar";
` + "`" + `template literal` + "`" + `;
` + "`" + `template ${expression} literal` + "`" + `;
[1, 2];
{"foo": "bar"};
1.5;
// This is a comment
/* This is a
multi-line comment */
enum Status { Active, Inactive }
struct User { name: string, age: int }
for (x in y) {}
async fn myAsync() {}
await myPromise;
try {} catch(e) {}
throw "error";
switch(x) { case 1: break; default: continue; }
import { a, b as c } from "module";
export let a = 1;
() => 1;
(a, b) => a + b;
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "ten"},
		{token.GT, ">"},
		{token.IDENT, "five"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.LT_EQ, "<="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.GT_EQ, ">="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.PLUS_ASSIGN, "+="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.SUB_ASSIGN, "-="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.MUL_ASSIGN, "*="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.DIV_ASSIGN, "/="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NULLISH_COALESCING, "??"},
		{token.INT, "0"},
		{token.SEMICOLON, ";"},
		{token.NULL, "null"},
		{token.OPTIONAL_CHAINING, "?."},
		{token.IDENT, "property"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foobar"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foo bar"},
		{token.SEMICOLON, ";"},
		{token.TEMPLATE_LITERAL, "template literal"},
		{token.SEMICOLON, ";"},
		{token.TEMPLATE_LITERAL, "template ${expression} literal"},
		{token.SEMICOLON, ";"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.FLOAT, "1.5"},
		{token.SEMICOLON, ";"},
		{token.ENUM, "enum"},
		{token.IDENT, "Status"},
		{token.LBRACE, "{"},
		{token.IDENT, "Active"},
		{token.COMMA, ","},
		{token.IDENT, "Inactive"},
		{token.RBRACE, "}"},
		{token.STRUCT, "struct"},
		{token.IDENT, "User"},
		{token.LBRACE, "{"},
		{token.IDENT, "name"},
		{token.COLON, ":"},
		{token.STRING_TYPE, "string"},
		{token.COMMA, ","},
		{token.IDENT, "age"},
		{token.COLON, ":"},
		{token.INT_TYPE, "int"},
		{token.RBRACE, "}"},
		{token.FOR, "for"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.IN, "in"},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.ASYNC, "async"},
		{token.FUNCTION, "fn"},
		{token.IDENT, "myAsync"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.AWAIT, "await"},
		{token.IDENT, "myPromise"},
		{token.SEMICOLON, ";"},
		{token.TRY, "try"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.CATCH, "catch"},
		{token.LPAREN, "("},
		{token.IDENT, "e"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.THROW, "throw"},
		{token.STRING, "error"},
		{token.SEMICOLON, ";"},
		{token.SWITCH, "switch"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.CASE, "case"},
		{token.INT, "1"},
		{token.COLON, ":"},
		{token.BREAK, "break"},
		{token.SEMICOLON, ";"},
		{token.DEFAULT, "default"},
		{token.COLON, ":"},
		{token.CONTINUE, "continue"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.IMPORT, "import"},
		{token.LBRACE, "{"},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.AS, "as"},
		{token.IDENT, "c"},
		{token.RBRACE, "}"},
		{token.FROM, "from"},
		{token.STRING, "module"},
		{token.SEMICOLON, ";"},
		{token.EXPORT, "export"},
		{token.LET, "let"},
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.ARROW, "=>"},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.RPAREN, ")"},
		{token.ARROW, "=>"},
		{token.IDENT, "a"},
		{token.PLUS, "+"},
		{token.IDENT, "b"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}