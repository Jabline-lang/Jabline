package ast

import (
	"testing"

	"jabline/pkg/token"
)

func assertString(t *testing.T, node Node, expected string) {
	t.Helper()
	if got := node.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}

func tokenOf(typ token.TokenType, lit string) token.Token {
	return token.Token{Type: typ, Literal: lit, Line: 1, Column: 1}
}

func ident(name string) *Identifier {
	return &Identifier{Token: tokenOf(token.IDENT, name), Value: name}
}

func idents(names ...string) []*Identifier {
	out := make([]*Identifier, len(names))
	for i, n := range names {
		out[i] = ident(n)
	}
	return out
}

func TestProgramString(t *testing.T) {
	prog := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: tokenOf(token.LET, "let"),
				Name:  ident("x"),
				Value: &IntegerLiteral{Token: tokenOf(token.INT, "5"), Value: 5},
			},
		},
	}
	assertString(t, prog, "x = 5;")
}

func TestLetStatement(t *testing.T) {
	node := &LetStatement{
		Token: tokenOf(token.LET, "let"),
		Name:  ident("x"),
		Value: &IntegerLiteral{Token: tokenOf(token.INT, "5"), Value: 5},
	}
	assertString(t, node, "x = 5;")

	node.Value = nil
	assertString(t, node, "x = ;")

	node.Type = &TypeExpression{Token: tokenOf(token.STRING_TYPE, "string"), Value: "string"}
	node.Value = &StringLiteral{Token: tokenOf(token.STRING, `"hello"`), Value: "hello"}
	assertString(t, node, "x: string = \"hello\";")
}

func TestConstStatement(t *testing.T) {
	node := &ConstStatement{
		Token: tokenOf(token.CONST, "const"),
		Name:  ident("PI"),
		Value: &FloatLiteral{Token: tokenOf(token.FLOAT, "3.14"), Value: 3.14},
	}
	assertString(t, node, "const PI = 3.14;")
}

func TestEchoStatement(t *testing.T) {
	node := &EchoStatement{
		Token:  tokenOf(token.ECHO, "echo"),
		Values: []Expression{ident("x"), &StringLiteral{Token: tokenOf(token.STRING, `"hi"`), Value: "hi"}},
	}
	assertString(t, node, `echo(x, "hi");`)
}

func TestBlockStatement(t *testing.T) {
	node := &BlockStatement{
		Statements: []Statement{
			&ExpressionStatement{Expression: ident("a")},
			&ExpressionStatement{Expression: ident("b")},
		},
	}
	assertString(t, node, "ab")
}

func TestReturnStatement(t *testing.T) {
	node := &ReturnStatement{
		Token:       tokenOf(token.RETURN, "return"),
		ReturnValue: &IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1},
	}
	assertString(t, node, "return 1;")

	node.ReturnValue = nil
	assertString(t, node, "return;")
}

func TestAssignmentStatement(t *testing.T) {
	node := &AssignmentStatement{
		Token: tokenOf(token.ASSIGN, "="),
		Left:  ident("x"),
		Value: &IntegerLiteral{Token: tokenOf(token.INT, "2"), Value: 2},
	}
	assertString(t, node, "x = 2;")
}

func TestBreakContinue(t *testing.T) {
	assertString(t, &BreakStatement{Token: tokenOf(token.BREAK, "break")}, "break;")
	assertString(t, &ContinueStatement{Token: tokenOf(token.CONTINUE, "continue")}, "continue;")
}

func TestIntegerLiteral(t *testing.T) {
	node := &IntegerLiteral{Token: tokenOf(token.INT, "42"), Value: 42}
	assertString(t, node, "42")
	if node.TokenLiteral() != "42" {
		t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), "42")
	}
}

func TestFloatLiteral(t *testing.T) {
	node := &FloatLiteral{Token: tokenOf(token.FLOAT, "3.14"), Value: 3.14}
	assertString(t, node, "3.14")
}

func TestStringLiteral(t *testing.T) {
	node := &StringLiteral{Token: tokenOf(token.STRING, `"hello"`), Value: "hello"}
	assertString(t, node, `"hello"`)
}

func TestBoolean(t *testing.T) {
	assertString(t, &Boolean{Token: tokenOf(token.TRUE, "true"), Value: true}, "true")
	assertString(t, &Boolean{Token: tokenOf(token.FALSE, "false"), Value: false}, "false")
}

func TestNull(t *testing.T) {
	node := &Null{Token: tokenOf(token.NULL, "null")}
	assertString(t, node, "null")
}

func TestArrayLiteral(t *testing.T) {
	node := &ArrayLiteral{
		Token:    tokenOf(token.LBRACKET, "["),
		Elements: []Expression{&IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1}, &IntegerLiteral{Token: tokenOf(token.INT, "2"), Value: 2}},
	}
	assertString(t, node, "[1, 2]")
}

func TestHashLiteral(t *testing.T) {
	node := &HashLiteral{
		Token: tokenOf(token.LBRACE, "{"),
		Pairs: map[Expression]Expression{
			&StringLiteral{Token: tokenOf(token.STRING, `"a"`), Value: "a"}: &IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1},
		},
	}
	out := node.String()
	if out != `{"a": 1}` {
		t.Errorf("String() = %q, want %q", out, `{"a": 1}`)

	}
}

func TestIdentifier(t *testing.T) {
	node := ident("foo")
	assertString(t, node, "foo")
	node.Type = &TypeExpression{Token: tokenOf(token.STRING_TYPE, "string"), Value: "string"}
	assertString(t, node, "foo: string")
}

func TestPrefixExpression(t *testing.T) {
	node := &PrefixExpression{
		Token:    tokenOf(token.BANG, "!"),
		Operator: "!",
		Right:    &Boolean{Token: tokenOf(token.TRUE, "true"), Value: true},
	}
	assertString(t, node, "(!true)")
}

func TestInfixExpression(t *testing.T) {
	node := &InfixExpression{
		Token:    tokenOf(token.PLUS, "+"),
		Left:     &IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1},
		Operator: "+",
		Right:    &IntegerLiteral{Token: tokenOf(token.INT, "2"), Value: 2},
	}
	assertString(t, node, "(1 + 2)")
}

func TestPostfixExpression(t *testing.T) {
	node := &PostfixExpression{
		Token:    tokenOf(token.INCREMENT, "++"),
		Left:     ident("x"),
		Operator: "++",
	}
	assertString(t, node, "(x++)")
}

func TestIfExpression(t *testing.T) {
	node := &IfExpression{
		Token:       tokenOf(token.IF, "if"),
		Condition:   &Boolean{Token: tokenOf(token.TRUE, "true"), Value: true},
		Consequence: &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}},
	}
	assertString(t, node, "if (true) x")
}

func TestIfElseExpression(t *testing.T) {
	node := &IfExpression{
		Token:       tokenOf(token.IF, "if"),
		Condition:   &Boolean{Token: tokenOf(token.TRUE, "true"), Value: true},
		Consequence: &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}},
		Alternative: &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("y")}}},
	}
	assertString(t, node, "if (true) x else y")
}

func TestFunctionLiteral(t *testing.T) {
	node := &FunctionLiteral{
		Token:      tokenOf(token.FUNCTION, "fn"),
		Parameters: idents("a", "b"),
		Body:       &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("a")}}},
	}
	assertString(t, node, "fn(a, b) a")

	node.TypeParameters = idents("T")
	assertString(t, node, "fn<T>(a, b) a")
}

func TestCallExpression(t *testing.T) {
	node := &CallExpression{
		Token:     tokenOf(token.LPAREN, "("),
		Function:  ident("foo"),
		Arguments: []Expression{&IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1}},
	}
	assertString(t, node, "foo(1)")
}

func TestArrowFunction(t *testing.T) {
	node := &ArrowFunction{
		Token:      tokenOf(token.ARROW, "=>"),
		Parameters: idents("x"),
		Body:       &InfixExpression{Left: ident("x"), Operator: "+", Right: &IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1}},
	}
	assertString(t, node, "x => (x + 1)")

	node2 := &ArrowFunction{
		Token:      tokenOf(token.ARROW, "=>"),
		Parameters: idents("a", "b"),
		Body:       ident("a"),
	}
	assertString(t, node2, "(a, b) => a")
}

func TestWhileStatement(t *testing.T) {
	node := &WhileStatement{
		Token:     tokenOf(token.WHILE, "while"),
		Condition: &Boolean{Token: tokenOf(token.TRUE, "true"), Value: true},
		Body:      &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}},
	}
	assertString(t, node, "while true x")
}

func TestForStatement(t *testing.T) {
	node := &ForStatement{
		Token:     tokenOf(token.FOR, "for"),
		Init:      &LetStatement{Token: tokenOf(token.LET, "let"), Name: ident("i"), Value: &IntegerLiteral{Token: tokenOf(token.INT, "0"), Value: 0}},
		Condition: &InfixExpression{Left: ident("i"), Operator: "<", Right: &IntegerLiteral{Token: tokenOf(token.INT, "10"), Value: 10}},
		Update:    &ExpressionStatement{Expression: &PostfixExpression{Left: ident("i"), Operator: "++"}},
		Body:      &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}},
	}
	out := node.String()
	// LetStatement.String() includes trailing ";", plus ForStatement adds ";" separators
	if out != "for (i = 0;; (i < 10); (i++)) x" {
		t.Errorf("unexpected ForStatement string: %q", out)
	}
}

func TestForEachStatement(t *testing.T) {
	node := &ForEachStatement{
		Token:    tokenOf(token.FOR, "for"),
		Variable: ident("v"),
		Iterable: ident("arr"),
		Body:     &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}},
	}
	assertString(t, node, "for (v in arr) x")
}

func TestTryCatchStatement(t *testing.T) {
	node := &TryStatement{
		Token:      tokenOf(token.TRY, "try"),
		TryBlock:   &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}},
		CatchParam: ident("e"),
		CatchBlock: &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("y")}}},
	}
	assertString(t, node, "try x catch(e) y")
}

func TestThrowStatement(t *testing.T) {
	node := &ThrowStatement{
		Token: tokenOf(token.THROW, "throw"),
		Value: &StringLiteral{Token: tokenOf(token.STRING, `"err"`), Value: "err"},
	}
	assertString(t, node, `throw "err"`)
}

func TestSwitchStatement(t *testing.T) {
	node := &SwitchStatement{
		Token:      tokenOf(token.SWITCH, "switch"),
		Expression: ident("x"),
		Cases: []*CaseClause{
			{Token: tokenOf(token.CASE, "case"), Value: &IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1}, Statements: []Statement{&ExpressionStatement{Expression: ident("a")}}},
		},
		DefaultCase: &DefaultClause{Token: tokenOf(token.DEFAULT, "default"), Statements: []Statement{&ExpressionStatement{Expression: ident("b")}}},
	}
	assertString(t, node, "switch (x) {case 1:adefault:b}")
}

func TestArrayIndexExpression(t *testing.T) {
	node := &ArrayIndexExpression{
		Token: tokenOf(token.LBRACKET, "["),
		Left:  ident("arr"),
		Index: &IntegerLiteral{Token: tokenOf(token.INT, "0"), Value: 0},
	}
	assertString(t, node, "(arr[0])")
}

func TestIndexExpression(t *testing.T) {
	node := &IndexExpression{
		Token: tokenOf(token.DOT, "."),
		Left:  ident("obj"),
		Index: ident("field"),
	}
	assertString(t, node, "(obj.field)")
}

func TestTernaryExpression(t *testing.T) {
	node := &TernaryExpression{
		Token:      tokenOf(token.QUESTION, "?"),
		Condition:  &Boolean{Token: tokenOf(token.TRUE, "true"), Value: true},
		TrueValue:  &IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1},
		FalseValue: &IntegerLiteral{Token: tokenOf(token.INT, "2"), Value: 2},
	}
	assertString(t, node, "(true ? 1 : 2)")
}

func TestFunctionStatement(t *testing.T) {
	node := &FunctionStatement{
		Token:          tokenOf(token.FUNCTION, "fn"),
		Name:           ident("add"),
		TypeParameters: idents("T"),
		Parameters:     idents("a", "b"),
		ReturnType:     &TypeExpression{Token: tokenOf(token.INT_TYPE, "int"), Value: "int"},
		Body:           &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("a")}}},
	}
	assertString(t, node, "fn add<T>(a, b): int a")
}

func TestStructStatement(t *testing.T) {
	node := &StructStatement{
		Token: tokenOf(token.STRUCT, "struct"),
		Name:  ident("Point"),
		Fields: map[string]*TypeExpression{
			"x": {Token: tokenOf(token.INT_TYPE, "int"), Value: "int"},
			"y": {Token: tokenOf(token.INT_TYPE, "int"), Value: "int"},
		},
	}
	out := node.String()
	if out != "struct Point { x: int, y: int }" && out != "struct Point { y: int, x: int }" {
		t.Errorf("unexpected StructStatement string: %q", out)
	}
}

func TestImportStatement(t *testing.T) {
	tests := []struct {
		name string
		node *ImportStatement
		want string
	}{
		{"side_effect", &ImportStatement{Token: tokenOf(token.IMPORT, "import"), ImportType: IMPORT_SIDE_EFFECT, ModuleName: &StringLiteral{Token: tokenOf(token.STRING, `"./lib"`), Value: "./lib"}}, `import "./lib"`},
		{"default", &ImportStatement{Token: tokenOf(token.IMPORT, "import"), ImportType: IMPORT_DEFAULT, DefaultImport: ident("foo"), ModuleName: &StringLiteral{Token: tokenOf(token.STRING, `"./lib"`), Value: "./lib"}}, `import foo from "./lib"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertString(t, tt.node, tt.want)
		})
	}
}

func TestTemplateLiteral(t *testing.T) {
	node := &TemplateLiteral{
		Token:       tokenOf(token.TEMPLATE_LITERAL, "`"),
		Parts:       []string{"hello ", " world"},
		Expressions: []Expression{ident("name")},
	}
	assertString(t, node, "`hello ${name} world`")
}

func TestEnumStatement(t *testing.T) {
	node := &EnumStatement{
		Token:  tokenOf(token.ENUM, "enum"),
		Name:   ident("Color"),
		Values: idents("Red", "Green", "Blue"),
	}
	assertString(t, node, "enum Color { Red, Green, Blue }")
}

func TestServiceStatement(t *testing.T) {
	node := &ServiceStatement{
		Token:  tokenOf(token.SERVICE, "service"),
		Name:   ident("MyService"),
		Fields: map[string]Expression{"port": &IntegerLiteral{Token: tokenOf(token.INT, "8080"), Value: 8080}},
		Methods: []*FunctionStatement{
			{Token: tokenOf(token.FUNCTION, "fn"), Name: ident("handle"), Parameters: idents("req"), Body: &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}}},
		},
	}
	out := node.String()
	if out != "service MyService {\n  port: 8080\nfn handle(req) x\n}" {
		t.Errorf("unexpected ServiceStatement: %q", out)
	}
}

func TestTypeExpression(t *testing.T) {
	node := &TypeExpression{Token: tokenOf(token.INT_TYPE, "int"), Value: "int"}
	assertString(t, node, "int")

	node2 := &TypeExpression{
		Token:     tokenOf(token.IDENT, "Array"),
		Value:     "Array",
		Arguments: []*TypeExpression{{Token: tokenOf(token.INT_TYPE, "int"), Value: "int"}},
	}
	assertString(t, node2, "Array<int>")
}

func TestSpawnExpression(t *testing.T) {
	node := &SpawnExpression{
		Token: tokenOf(token.SPAWN, "spawn"),
		Call: &CallExpression{
			Token:     tokenOf(token.LPAREN, "("),
			Function:  ident("foo"),
			Arguments: []Expression{},
		},
	}
	assertString(t, node, "spawn foo()")
}

func TestAwaitExpression(t *testing.T) {
	node := &AwaitExpression{
		Token: tokenOf(token.AWAIT, "await"),
		Value: ident("prom"),
	}
	assertString(t, node, "await prom")
}

func TestInstantiatedExpression(t *testing.T) {
	node := &InstantiatedExpression{
		Token:         tokenOf(token.IDENT, "Array"),
		Left:          ident("Array"),
		TypeArguments: []*TypeExpression{{Token: tokenOf(token.INT_TYPE, "int"), Value: "int"}},
	}
	assertString(t, node, "Array<int>")
}

func TestMatchStatement(t *testing.T) {
	node := &MatchStatement{
		Token:      tokenOf(token.MATCH, "match"),
		Expression: ident("x"),
		Cases: []*MatchCase{
			{Token: tokenOf(token.CASE, "case"), Pattern: &IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1}, Statements: []Statement{&ExpressionStatement{Expression: ident("a")}}},
		},
	}
	assertString(t, node, "match (x) {case 1:a}")
}

func TestMeterTraceStatements(t *testing.T) {
	meter := &MeterStatement{
		Token:    tokenOf(token.METER, "meter"),
		Name:     ident("metric_a"),
		Operator: "++",
	}
	assertString(t, meter, "meter metric_a++;")

	trace := &TraceStatement{
		Token: tokenOf(token.TRACE, "trace"),
		Name:  ident("span_1"),
		Body:  &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}},
	}
	assertString(t, trace, "trace span_1 x")
}

func TestInterfaceStatement(t *testing.T) {
	node := &InterfaceStatement{
		Token: tokenOf(token.INTERFACE, "interface"),
		Name:  ident("Stringer"),
		Methods: map[string]*FunctionSignature{
			"String": {Token: tokenOf(token.FUNCTION, "fn"), Name: "String", Parameters: idents(), ReturnType: &TypeExpression{Token: tokenOf(token.STRING_TYPE, "string"), Value: "string"}},
		},
	}
	out := node.String()
	if out != "interface Stringer { String(): string }" {
		t.Errorf("unexpected InterfaceStatement: %q", out)
	}
}

func TestRetryStatement(t *testing.T) {
	node := &RetryStatement{
		Token:      tokenOf(token.RETRY, "retry"),
		Attempts:   &IntegerLiteral{Token: tokenOf(token.INT, "3"), Value: 3},
		RetryBlock: &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}},
		CatchParam: ident("e"),
		CatchBlock: &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("y")}}},
	}
	assertString(t, node, "retry (3) x catch(e) y")
}

func TestStructLiteral(t *testing.T) {
	node := &StructLiteral{
		Token:  tokenOf(token.LBRACE, "{"),
		Name:   ident("Point"),
		Fields: map[string]Expression{"x": &IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1}},
	}
	out := node.String()
	if out != "Point { x: 1 }" {
		t.Errorf("unexpected StructLiteral: %q", out)
	}
}

func TestProgramTokenLiteral(t *testing.T) {
	prog := &Program{}
	if prog.TokenLiteral() != "" {
		t.Errorf("empty program TokenLiteral() = %q, want empty", prog.TokenLiteral())
	}

	prog.Statements = []Statement{
		&LetStatement{Token: tokenOf(token.LET, "let"), Name: ident("x"), Value: &IntegerLiteral{Token: tokenOf(token.INT, "5"), Value: 5}},
	}
	if prog.TokenLiteral() != "let" {
		t.Errorf("program TokenLiteral() = %q, want %q", prog.TokenLiteral(), "let")
	}
}

func TestGetToken(t *testing.T) {
	tok := tokenOf(token.LET, "let")
	node := &LetStatement{Token: tok}
	if node.GetToken() != tok {
		t.Errorf("GetToken() returned different token")
	}
}

func TestAsyncFunctionStatement(t *testing.T) {
	node := &AsyncFunctionStatement{
		Token:          tokenOf(token.ASYNC, "async"),
		Name:           ident("fetchData"),
		TypeParameters: idents("T"),
		Parameters:     idents("url"),
		ReturnType:     &TypeExpression{Token: tokenOf(token.STRING_TYPE, "string"), Value: "string"},
		Body:           &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("x")}}},
	}
	assertString(t, node, "async fn fetchData[T](url): string x")
}

func TestAsyncFunctionLiteral(t *testing.T) {
	node := &AsyncFunctionLiteral{
		Token:          tokenOf(token.ASYNC, "async"),
		TypeParameters: idents("T"),
		Parameters:     idents("a"),
		ReturnType:     &TypeExpression{Token: tokenOf(token.INT_TYPE, "int"), Value: "int"},
		Body:           &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: ident("a")}}},
	}
	assertString(t, node, "async[T](a): int a")
}

func TestExportStatement(t *testing.T) {
	node := &ExportStatement{
		Token:      tokenOf(token.EXPORT, "export"),
		ExportType: EXPORT_DECLARATION,
		Statement:  &LetStatement{Token: tokenOf(token.LET, "let"), Name: ident("x"), Value: &IntegerLiteral{Token: tokenOf(token.INT, "1"), Value: 1}},
	}
	assertString(t, node, "export x = 1;")
}

func TestReExportStatement(t *testing.T) {
	node := &ReExportStatement{
		Token:      tokenOf(token.EXPORT, "export"),
		ModuleName: &StringLiteral{Token: tokenOf(token.STRING, `"./lib"`), Value: "./lib"},
	}
	assertString(t, node, `export * from "./lib"`)
}

func TestNullishCoalescingExpression(t *testing.T) {
	node := &NullishCoalescingExpression{
		Token: tokenOf(token.NULLISH_COALESCING, "??"),
		Left:  ident("x"),
		Right: &IntegerLiteral{Token: tokenOf(token.INT, "0"), Value: 0},
	}
	assertString(t, node, "(x ?? 0)")
}

func TestOptionalChainingExpression(t *testing.T) {
	node := &OptionalChainingExpression{
		Token: tokenOf(token.OPTIONAL_CHAINING, "?."),
		Left:  ident("obj"),
		Right: ident("field"),
	}
	assertString(t, node, "(obj?.field)")
}
