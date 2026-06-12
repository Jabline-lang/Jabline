package lexer

import (
	"jabline/pkg/token"
	"testing"
)

type testCase struct {
	name     string
	input    string
	expected []token.TokenType
	literals []string
}

func checkTokens(t *testing.T, name string, input string, expected []token.TokenType, literals []string) {
	t.Helper()
	l := New(input)
	for i, expType := range expected {
		tok := l.NextToken()
		if tok.Type != expType {
			t.Errorf("[%s] token[%d] type mismatch: expected=%q, got=%q (literal=%q)", name, i, expType, tok.Type, tok.Literal)
		}
		if i < len(literals) && tok.Literal != literals[i] {
			t.Errorf("[%s] token[%d] literal mismatch: expected=%q, got=%q", name, i, literals[i], tok.Literal)
		}
	}
	// Ensure we consume all input (next token should be EOF)
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("[%s] expected EOF after all tokens, got=%q (literal=%q)", name, tok.Type, tok.Literal)
	}
}

func TestNextTokenSingleChar(t *testing.T) {
	input := "=+-*/%^~(),;{}[]<>.:?"
	expected := []token.TokenType{
		token.ASSIGN, token.PLUS, token.MINUS, token.ASTERISK,
		token.SLASH, token.MOD, token.BIT_XOR, token.BIT_NOT,
		token.LPAREN, token.RPAREN, token.COMMA, token.SEMICOLON,
		token.LBRACE, token.RBRACE, token.LBRACKET, token.RBRACKET,
		token.LT, token.GT, token.DOT, token.COLON, token.QUESTION,
	}
	checkTokens(t, "single-char", input, expected, nil)
}

func TestNextTokenMultiChar(t *testing.T) {
	input := "== != <= >= ++ -- += -= *= /= && || |> << >> ?? ?."
	expected := []token.TokenType{
		token.EQ, token.NOT_EQ, token.LT_EQ, token.GT_EQ,
		token.INCREMENT, token.DECREMENT,
		token.PLUS_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.DIV_ASSIGN,
		token.AND, token.OR, token.PIPE,
		token.SHIFT_LEFT, token.SHIFT_RIGHT,
		token.NULLISH_COALESCING, token.OPTIONAL_CHAINING,
	}
	checkTokens(t, "multi-char", input, expected, nil)
}

func TestNextTokenKeywords(t *testing.T) {
	input := "let const fn return if else for while break continue in switch case default try catch throw retry async await import export from spawn interface enum match meter trace service struct echo null true false as"
	expected := []token.TokenType{
		token.LET, token.CONST, token.FUNCTION, token.RETURN,
		token.IF, token.ELSE, token.FOR, token.WHILE,
		token.BREAK, token.CONTINUE, token.IN,
		token.SWITCH, token.CASE, token.DEFAULT,
		token.TRY, token.CATCH, token.THROW, token.RETRY,
		token.ASYNC, token.AWAIT,
		token.IMPORT, token.EXPORT, token.FROM,
		token.SPAWN, token.INTERFACE,
		token.ENUM, token.MATCH, token.METER, token.TRACE,
		token.SERVICE, token.STRUCT,
		token.ECHO,
		token.NULL, token.TRUE, token.FALSE,
		token.AS,
	}
	checkTokens(t, "keywords", input, expected, nil)
}

func TestNextTokenTypes(t *testing.T) {
	input := "string int int8 int16 int32 int64 uint8 uint16 uint32 uint64 float float32 float64 bool"
	expected := []token.TokenType{
		token.STRING_TYPE, token.INT_TYPE,
		token.INT8_TYPE, token.INT16_TYPE, token.INT32_TYPE, token.INT64_TYPE,
		token.UINT8_TYPE, token.UINT16_TYPE, token.UINT32_TYPE, token.UINT64_TYPE,
		token.FLOAT_TYPE, token.FLOAT32_TYPE, token.FLOAT64_TYPE,
		token.BOOL_TYPE,
	}
	checkTokens(t, "types", input, expected, nil)
}

func TestNextTokenIdentifiers(t *testing.T) {
	input := "foo bar baz _underscore camelCase PascalCase snake_case"
	expected := []token.TokenType{
		token.IDENT, token.IDENT, token.IDENT,
		token.IDENT, token.IDENT, token.IDENT, token.IDENT,
	}
	literals := []string{"foo", "bar", "baz", "_underscore", "camelCase", "PascalCase", "snake_case"}
	checkTokens(t, "identifiers", input, expected, literals)
}

func TestNextTokenNumbers(t *testing.T) {
	input := "42 0 3.14 0.5 100000"
	expected := []token.TokenType{
		token.INT, token.INT, token.FLOAT, token.FLOAT, token.INT,
	}
	literals := []string{"42", "0", "3.14", "0.5", "100000"}
	checkTokens(t, "numbers", input, expected, literals)
}

func TestNextTokenStrings(t *testing.T) {
	input := `"hello" "world" "hello \"world\"" ""`
	expected := []token.TokenType{token.STRING, token.STRING, token.STRING, token.STRING}
	literals := []string{"hello", "world", `hello "world"`, ""}
	checkTokens(t, "strings", input, expected, literals)
}

func TestNextTokenTemplateLiterals(t *testing.T) {
	input := "`hello` `hello ${name}` `${a} + ${b}`"
	expected := []token.TokenType{token.TEMPLATE_LITERAL, token.TEMPLATE_LITERAL, token.TEMPLATE_LITERAL}
	checkTokens(t, "template-literals", input, expected, nil)
}

func TestNextTokenComments(t *testing.T) {
	input := `let x = 1; // this is a comment
let y = 2; /* block comment */
let z = 3;`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN, token.INT, token.SEMICOLON,
		token.LET, token.IDENT, token.ASSIGN, token.INT, token.SEMICOLON,
		token.LET, token.IDENT, token.ASSIGN, token.INT, token.SEMICOLON,
	}
	checkTokens(t, "comments", input, expected, nil)
}

func TestNextTokenFunctionDecl(t *testing.T) {
	input := `fn add(a, b) { return a + b; }`
	expected := []token.TokenType{
		token.FUNCTION, token.IDENT, token.LPAREN,
		token.IDENT, token.COMMA, token.IDENT, token.RPAREN,
		token.LBRACE,
		token.RETURN, token.IDENT, token.PLUS, token.IDENT, token.SEMICOLON,
		token.RBRACE,
	}
	literals := []string{"fn", "add", "(", "a", ",", "b", ")", "{", "return", "a", "+", "b", ";", "}"}
	checkTokens(t, "function", input, expected, literals)
}

func TestNextTokenLetConst(t *testing.T) {
	input := `let name: string = "John";
const age: int = 30;`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.COLON, token.STRING_TYPE, token.ASSIGN, token.STRING, token.SEMICOLON,
		token.CONST, token.IDENT, token.COLON, token.INT_TYPE, token.ASSIGN, token.INT, token.SEMICOLON,
	}
	checkTokens(t, "let-const", input, expected, nil)
}

func TestNextTokenControlFlow(t *testing.T) {
	input := `if (x > 0) { echo("pos"); } else { echo("neg"); }
while (true) { break; }
for (let i = 0; i < 10; i++) { continue; }
for item in items { echo(item); }`
	expected := []token.TokenType{
		token.IF, token.LPAREN, token.IDENT, token.GT, token.INT, token.RPAREN, token.LBRACE,
		token.ECHO, token.LPAREN, token.STRING, token.RPAREN, token.SEMICOLON, token.RBRACE,
		token.ELSE, token.LBRACE,
		token.ECHO, token.LPAREN, token.STRING, token.RPAREN, token.SEMICOLON, token.RBRACE,
		token.WHILE, token.LPAREN, token.TRUE, token.RPAREN, token.LBRACE,
		token.BREAK, token.SEMICOLON, token.RBRACE,
		token.FOR, token.LPAREN, token.LET, token.IDENT, token.ASSIGN, token.INT, token.SEMICOLON,
		token.IDENT, token.LT, token.INT, token.SEMICOLON,
		token.IDENT, token.INCREMENT, token.RPAREN, token.LBRACE,
		token.CONTINUE, token.SEMICOLON, token.RBRACE,
		token.FOR, token.IDENT, token.IN, token.IDENT, token.LBRACE,
		token.ECHO, token.LPAREN, token.IDENT, token.RPAREN, token.SEMICOLON, token.RBRACE,
	}
	checkTokens(t, "control-flow", input, expected, nil)
}

func TestNextTokenSwitch(t *testing.T) {
	input := `switch (x) {
	case 1: echo("one");
	case 2: echo("two");
	default: echo("other");
}`
	expected := []token.TokenType{
		token.SWITCH, token.LPAREN, token.IDENT, token.RPAREN, token.LBRACE,
		token.CASE, token.INT, token.COLON, token.ECHO, token.LPAREN, token.STRING, token.RPAREN, token.SEMICOLON,
		token.CASE, token.INT, token.COLON, token.ECHO, token.LPAREN, token.STRING, token.RPAREN, token.SEMICOLON,
		token.DEFAULT, token.COLON, token.ECHO, token.LPAREN, token.STRING, token.RPAREN, token.SEMICOLON,
		token.RBRACE,
	}
	checkTokens(t, "switch", input, expected, nil)
}

func TestNextTokenTryCatch(t *testing.T) {
	input := `try { throw "err"; } catch (e) { echo(e); }`
	expected := []token.TokenType{
		token.TRY, token.LBRACE, token.THROW, token.STRING, token.SEMICOLON, token.RBRACE,
		token.CATCH, token.LPAREN, token.IDENT, token.RPAREN, token.LBRACE,
		token.ECHO, token.LPAREN, token.IDENT, token.RPAREN, token.SEMICOLON, token.RBRACE,
	}
	checkTokens(t, "try-catch", input, expected, nil)
}

func TestNextTokenStruct(t *testing.T) {
	input := `struct Person {
	name: string;
	age: int;
}`
	expected := []token.TokenType{
		token.STRUCT, token.IDENT, token.LBRACE,
		token.IDENT, token.COLON, token.STRING_TYPE, token.SEMICOLON,
		token.IDENT, token.COLON, token.INT_TYPE, token.SEMICOLON,
		token.RBRACE,
	}
	checkTokens(t, "struct", input, expected, nil)
}

func TestNextTokenEnum(t *testing.T) {
	input := `enum Color { Red, Green, Blue }`
	expected := []token.TokenType{
		token.ENUM, token.IDENT, token.LBRACE,
		token.IDENT, token.COMMA, token.IDENT, token.COMMA, token.IDENT,
		token.RBRACE,
	}
	checkTokens(t, "enum", input, expected, nil)
}

func TestNextTokenService(t *testing.T) {
	input := `service Greeter {
	port: 8080;
	fn hello(req) { return "hi"; }
}`
	expected := []token.TokenType{
		token.SERVICE, token.IDENT, token.LBRACE,
		token.IDENT, token.COLON, token.INT, token.SEMICOLON,
		token.FUNCTION, token.IDENT, token.LPAREN, token.IDENT, token.RPAREN, token.LBRACE,
		token.RETURN, token.STRING, token.SEMICOLON, token.RBRACE,
		token.RBRACE,
	}
	checkTokens(t, "service", input, expected, nil)
}

func TestNextTokenImport(t *testing.T) {
	input := `import "foo";
import { bar } from "baz";
import * as stuff from "lib";
import defaultExport from "mod";`
	expected := []token.TokenType{
		token.IMPORT, token.STRING, token.SEMICOLON,
		token.IMPORT, token.LBRACE, token.IDENT, token.RBRACE, token.FROM, token.STRING, token.SEMICOLON,
		token.IMPORT, token.ASTERISK, token.AS, token.IDENT, token.FROM, token.STRING, token.SEMICOLON,
		token.IMPORT, token.IDENT, token.FROM, token.STRING, token.SEMICOLON,
	}
	checkTokens(t, "import", input, expected, nil)
}

func TestNextTokenArrowFunction(t *testing.T) {
	input := `let add = (a, b) => a + b;`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN,
		token.LPAREN, token.IDENT, token.COMMA, token.IDENT, token.RPAREN,
		token.ARROW,
		token.IDENT, token.PLUS, token.IDENT, token.SEMICOLON,
	}
	checkTokens(t, "arrow", input, expected, nil)
}

func TestNextTokenOptionalChaining(t *testing.T) {
	input := `let x = obj?.prop ?? "default";`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN,
		token.IDENT, token.OPTIONAL_CHAINING, token.IDENT,
		token.NULLISH_COALESCING, token.STRING, token.SEMICOLON,
	}
	checkTokens(t, "optional-chaining", input, expected, nil)
}

func TestNextTokenPipe(t *testing.T) {
	input := `let result = x |> double |> toString;`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN,
		token.IDENT, token.PIPE, token.IDENT, token.PIPE, token.IDENT,
		token.SEMICOLON,
	}
	checkTokens(t, "pipe", input, expected, nil)
}

func TestNextTokenPostfix(t *testing.T) {
	input := `let x = 1; x++; x--;`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN, token.INT, token.SEMICOLON,
		token.IDENT, token.INCREMENT, token.SEMICOLON,
		token.IDENT, token.DECREMENT, token.SEMICOLON,
	}
	checkTokens(t, "postfix", input, expected, nil)
}

func TestNextTokenTernary(t *testing.T) {
	input := `let x = a > b ? a : b;`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN,
		token.IDENT, token.GT, token.IDENT,
		token.QUESTION,
		token.IDENT, token.COLON,
		token.IDENT, token.SEMICOLON,
	}
	checkTokens(t, "ternary", input, expected, nil)
}

func TestNextTokenBitwise(t *testing.T) {
	input := `let x = a & b | c ^ d ~e << 2 >> 1;`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN,
		token.IDENT, token.BIT_AND, token.IDENT,
		token.BIT_OR, token.IDENT,
		token.BIT_XOR, token.IDENT,
		token.BIT_NOT, token.IDENT,
		token.SHIFT_LEFT, token.INT,
		token.SHIFT_RIGHT, token.INT,
		token.SEMICOLON,
	}
	checkTokens(t, "bitwise", input, expected, nil)
}

func TestNextTokenSpawnAwait(t *testing.T) {
	input := `let task = spawn fetchData();
let result = await task;`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN,
		token.SPAWN, token.IDENT, token.LPAREN, token.RPAREN, token.SEMICOLON,
		token.LET, token.IDENT, token.ASSIGN,
		token.AWAIT, token.IDENT, token.SEMICOLON,
	}
	checkTokens(t, "spawn-await", input, expected, nil)
}

func TestNextTokenMatch(t *testing.T) {
	input := `let result = match value {
	1 => "one",
	2 => "two",
	_ => "other"
};`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN,
		token.MATCH, token.IDENT, token.LBRACE,
		token.INT, token.ARROW, token.STRING, token.COMMA,
		token.INT, token.ARROW, token.STRING, token.COMMA,
		token.IDENT, token.ARROW, token.STRING,
		token.RBRACE, token.SEMICOLON,
	}
	checkTokens(t, "match", input, expected, nil)
}

func TestNextTokenMultiLineComment(t *testing.T) {
	input := `let x = 1; /* this is a
multi-line
comment */ let y = 2;`
	expected := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN, token.INT, token.SEMICOLON,
		token.LET, token.IDENT, token.ASSIGN, token.INT, token.SEMICOLON,
	}
	checkTokens(t, "multi-line-comment", input, expected, nil)
}

func TestNextTokenIllegal(t *testing.T) {
	input := "@"
	expected := []token.TokenType{token.ILLEGAL}
	checkTokens(t, "illegal", input, expected, nil)
}

func TestNextTokenPosition(t *testing.T) {
	input := "let x = 1;\nlet y = 2;"
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.LET || tok.Line != 1 || tok.Column != 1 {
		t.Errorf("token[0] position: expected Line=1 Col=1, got Line=%d Col=%d", tok.Line, tok.Column)
	}

	tok = l.NextToken()
	if tok.Type != token.IDENT || tok.Line != 1 || tok.Column != 5 {
		t.Errorf("token[1] position: expected Line=1 Col=5, got Line=%d Col=%d", tok.Line, tok.Column)
	}

	tok = l.NextToken()
	if tok.Line != 1 || tok.Column != 7 {
		t.Errorf("token[2] position: expected Line=1 Col=7, got Line=%d Col=%d", tok.Line, tok.Column)
	}

	tok = l.NextToken()
	if tok.Type != token.INT || tok.Line != 1 || tok.Column != 9 {
		t.Errorf("token[3] position: expected Line=1 Col=9, got Line=%d Col=%d", tok.Line, tok.Column)
	}

	tok = l.NextToken()
	if tok.Line != 1 || tok.Column != 10 {
		t.Errorf("token[4] (;) position: expected Line=1 Col=10, got Line=%d Col=%d", tok.Line, tok.Column)
	}

	// Second line
	tok = l.NextToken()
	if tok.Type != token.LET || tok.Line != 2 || tok.Column != 1 {
		t.Errorf("token[5] position: expected Line=2 Col=1, got Line=%d Col=%d", tok.Line, tok.Column)
	}

	tok = l.NextToken()
	if tok.Type != token.IDENT || tok.Line != 2 || tok.Column != 5 {
		t.Errorf("token[6] position: expected Line=2 Col=5, got Line=%d Col=%d", tok.Line, tok.Column)
	}
}

func TestNextTokenEmptyInput(t *testing.T) {
	input := ""
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("empty input: expected EOF, got=%q", tok.Type)
	}
}

func TestNextTokenWhitespaceOnly(t *testing.T) {
	input := "   \n\t\n  "
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("whitespace only: expected EOF, got=%q", tok.Type)
	}
}

func TestNextTokenBOM(t *testing.T) {
	input := "\xef\xbb\xbflet x = 1;"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.LET {
		t.Errorf("BOM: expected LET, got=%q", tok.Type)
	}
}

func TestNextTokenAssignmentOperators(t *testing.T) {
	input := "x += 1; y -= 2; z *= 3; w /= 4;"
	expected := []token.TokenType{
		token.IDENT, token.PLUS_ASSIGN, token.INT, token.SEMICOLON,
		token.IDENT, token.SUB_ASSIGN, token.INT, token.SEMICOLON,
		token.IDENT, token.MUL_ASSIGN, token.INT, token.SEMICOLON,
		token.IDENT, token.DIV_ASSIGN, token.INT, token.SEMICOLON,
	}
	checkTokens(t, "assignment-ops", input, expected, nil)
}

func TestNextTokenRetryCatch(t *testing.T) {
	input := `retry (3) {
	try { throw "fail"; }
	catch (e) { echo(e); }
}`
	expected := []token.TokenType{
		token.RETRY, token.LPAREN, token.INT, token.RPAREN, token.LBRACE,
		token.TRY, token.LBRACE, token.THROW, token.STRING, token.SEMICOLON, token.RBRACE,
		token.CATCH, token.LPAREN, token.IDENT, token.RPAREN, token.LBRACE,
		token.ECHO, token.LPAREN, token.IDENT, token.RPAREN, token.SEMICOLON, token.RBRACE,
		token.RBRACE,
	}
	checkTokens(t, "retry-catch", input, expected, nil)
}

func TestNextTokenAsyncFunction(t *testing.T) {
	input := `async fn fetch(url) { return await http.get(url); }`
	expected := []token.TokenType{
		token.ASYNC, token.FUNCTION, token.IDENT, token.LPAREN, token.IDENT, token.RPAREN,
		token.LBRACE,
		token.RETURN, token.AWAIT, token.IDENT, token.DOT, token.IDENT, token.LPAREN, token.IDENT, token.RPAREN,
		token.SEMICOLON, token.RBRACE,
	}
	checkTokens(t, "async-function", input, expected, nil)
}

func TestNextTokenInterface(t *testing.T) {
	input := `interface Shape {
	area(): float64;
	perimeter(): float64;
}`
	expected := []token.TokenType{
		token.INTERFACE, token.IDENT, token.LBRACE,
		token.IDENT, token.LPAREN, token.RPAREN, token.COLON, token.FLOAT64_TYPE, token.SEMICOLON,
		token.IDENT, token.LPAREN, token.RPAREN, token.COLON, token.FLOAT64_TYPE, token.SEMICOLON,
		token.RBRACE,
	}
	checkTokens(t, "interface", input, expected, nil)
}

func TestNextTokenMeterTrace(t *testing.T) {
	input := `service S {
	meter("counter").add(1);
	trace("span").end();
}`
	expected := []token.TokenType{
		token.SERVICE, token.IDENT, token.LBRACE,
		token.METER, token.LPAREN, token.STRING, token.RPAREN,
		token.DOT, token.IDENT, token.LPAREN, token.INT, token.RPAREN, token.SEMICOLON,
		token.TRACE, token.LPAREN, token.STRING, token.RPAREN,
		token.DOT, token.IDENT, token.LPAREN, token.RPAREN, token.SEMICOLON,
		token.RBRACE,
	}
	checkTokens(t, "meter-trace", input, expected, nil)
}

func TestNextTokenGenerics(t *testing.T) {
	input := `fn identity<T>(x: T): T { return x; }
let arr: Array<int> = [];`
	expected := []token.TokenType{
		token.FUNCTION, token.IDENT, token.LT, token.IDENT, token.GT,
		token.LPAREN, token.IDENT, token.COLON, token.IDENT, token.RPAREN,
		token.COLON, token.IDENT,
		token.LBRACE, token.RETURN, token.IDENT, token.SEMICOLON, token.RBRACE,
		token.LET, token.IDENT, token.COLON, token.IDENT, token.LT, token.INT_TYPE, token.GT,
		token.ASSIGN, token.LBRACKET, token.RBRACKET, token.SEMICOLON,
	}
	checkTokens(t, "generics", input, expected, nil)
}
