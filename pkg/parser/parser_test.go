package parser

import (
	"jabline/pkg/ast"
	"jabline/pkg/lexer"
	"testing"
)

func parseProgram(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	return prog
}

func checkStatements(t *testing.T, prog *ast.Program, expectedCount int) {
	t.Helper()
	if len(prog.Statements) != expectedCount {
		t.Fatalf("expected %d statements, got %d", expectedCount, len(prog.Statements))
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`let x = 5;`, "x = 5;"},
		{`let y = true;`, "y = true;"},
		{`let z = "hello";`, `z = "hello";`},
		{`let foo: int = 42;`, "foo: int = 42;"},
		{`let bar: string = "hi";`, `bar: string = "hi";`},
		{`let arr: Array<int> = [];`, "arr: Array<int> = [];"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		ls := prog.Statements[0].(*ast.LetStatement)
		if ls.String() != tt.expected {
			t.Errorf("LetStatement.String() = %q, want %q", ls.String(), tt.expected)
		}
	}
}

func TestConstStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`const x = 5;`, "const x = 5;"},
		{`const y = true;`, "const y = true;"},
		{`const z: float64 = 3.14;`, "const z: float64 = 3.14;"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		cs := prog.Statements[0].(*ast.ConstStatement)
		if cs.String() != tt.expected {
			t.Errorf("ConstStatement.String() = %q, want %q", cs.String(), tt.expected)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`return 5;`, "return 5;"},
		{`return x;`, "return x;"},
		{`return;`, "return;"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		rs := prog.Statements[0].(*ast.ReturnStatement)
		if rs.String() != tt.expected {
			t.Errorf("ReturnStatement.String() = %q, want %q", rs.String(), tt.expected)
		}
	}
}

func TestEchoStatements(t *testing.T) {
	input := `echo("hello");`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	es := prog.Statements[0].(*ast.EchoStatement)
	if es.TokenLiteral() != "echo" {
		t.Errorf("EchoStatement.TokenLiteral() = %q, want 'echo'", es.TokenLiteral())
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := `foobar;`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected *ast.Identifier, got %T", stmt.Expression)
	}
	if ident.Value != "foobar" {
		t.Errorf("ident.Value = %q, want %q", ident.Value, "foobar")
	}
	if ident.String() != "foobar" {
		t.Errorf("ident.String() = %q, want %q", ident.String(), "foobar")
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := `42;`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	lit, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected *ast.IntegerLiteral, got %T", stmt.Expression)
	}
	if lit.Value != 42 {
		t.Errorf("lit.Value = %d, want %d", lit.Value, 42)
	}
	if lit.String() != "42" {
		t.Errorf("lit.String() = %q, want %q", lit.String(), "42")
	}
}

func TestFloatLiteralExpression(t *testing.T) {
	input := `3.14;`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	lit, ok := stmt.Expression.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("expected *ast.FloatLiteral, got %T", stmt.Expression)
	}
	if lit.Value != 3.14 {
		t.Errorf("lit.Value = %f, want %f", lit.Value, 3.14)
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	lit, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected *ast.StringLiteral, got %T", stmt.Expression)
	}
	if lit.Value != "hello world" {
		t.Errorf("lit.Value = %q, want %q", lit.Value, "hello world")
	}
	if lit.String() != `"hello world"` {
		t.Errorf("lit.String() = %q, want %q", lit.String(), `"hello world"`)
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`true;`, true},
		{`false;`, false},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		b, ok := stmt.Expression.(*ast.Boolean)
		if !ok {
			t.Fatalf("expected *ast.Boolean, got %T", stmt.Expression)
		}
		if b.Value != tt.expected {
			t.Errorf("b.Value = %t, want %t", b.Value, tt.expected)
		}
	}
}

func TestNullExpression(t *testing.T) {
	input := `null;`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.Null)
	if !ok {
		t.Fatalf("expected *ast.Null, got %T", stmt.Expression)
	}
}

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`!true;`, "(!true)"},
		{`-42;`, "(-42)"},
		{`~x;`, "(~x)"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		pe, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("expected *ast.PrefixExpression, got %T", stmt.Expression)
		}
		if pe.String() != tt.expected {
			t.Errorf("PrefixExpression.String() = %q, want %q", pe.String(), tt.expected)
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`5 + 5;`, "(5 + 5)"},
		{`5 - 5;`, "(5 - 5)"},
		{`5 * 5;`, "(5 * 5)"},
		{`5 / 5;`, "(5 / 5)"},
		{`5 % 5;`, "(5 % 5)"},
		{`5 == 5;`, "(5 == 5)"},
		{`5 != 5;`, "(5 != 5)"},
		{`5 < 5;`, "(5 < 5)"},
		{`5 > 5;`, "(5 > 5)"},
		{`5 <= 5;`, "(5 <= 5)"},
		{`5 >= 5;`, "(5 >= 5)"},
		{`true && false;`, "(true && false)"},
		{`true || false;`, "(true || false)"},
		{`a & b;`, "(a & b)"},
		{`a | b;`, "(a | b)"},
		{`a ^ b;`, "(a ^ b)"},
		{`a << 2;`, "(a << 2)"},
		{`a >> 2;`, "(a >> 2)"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		ie, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("expected *ast.InfixExpression, got %T", stmt.Expression)
		}
		if ie.String() != tt.expected {
			t.Errorf("InfixExpression.String() = %q, want %q", ie.String(), tt.expected)
		}
	}
}

func TestPostfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`x++;`, "(x++)"},
		{`x--;`, "(x--)"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		pe, ok := stmt.Expression.(*ast.PostfixExpression)
		if !ok {
			t.Fatalf("expected *ast.PostfixExpression, got %T", stmt.Expression)
		}
		if pe.String() != tt.expected {
			t.Errorf("PostfixExpression.String() = %q, want %q", pe.String(), tt.expected)
		}
	}
}

func TestTernaryExpression(t *testing.T) {
	input := `a > b ? a : b;`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	te, ok := stmt.Expression.(*ast.TernaryExpression)
	if !ok {
		t.Fatalf("expected *ast.TernaryExpression, got %T", stmt.Expression)
	}
	if te.String() != "((a > b) ? a : b)" {
		t.Errorf("TernaryExpression.String() = %q, want %q", te.String(), "((a > b) ? a : b)")
	}
}

func TestNullishCoalescingExpression(t *testing.T) {
	input := `x ?? "default";`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	nce, ok := stmt.Expression.(*ast.NullishCoalescingExpression)
	if !ok {
		t.Fatalf("expected *ast.NullishCoalescingExpression, got %T", stmt.Expression)
	}
	if nce.String() != `(x ?? "default")` {
		t.Errorf("NullishCoalescingExpression.String() = %q, want %q", nce.String(), `(x ?? "default")`)
	}
}

func TestOptionalChainingExpression(t *testing.T) {
	input := `obj?.prop;`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	oce, ok := stmt.Expression.(*ast.OptionalChainingExpression)
	if !ok {
		t.Fatalf("expected *ast.OptionalChainingExpression, got %T", stmt.Expression)
	}
	if oce.String() != "(obj?.prop)" {
		t.Errorf("OptionalChainingExpression.String() = %q, want %q", oce.String(), "(obj?.prop)")
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`-a * b;`, "((-a) * b)"},
		{`!-a;`, "(!(-a))"},
		{`a + b + c;`, "((a + b) + c)"},
		{`a + b - c;`, "((a + b) - c)"},
		{`a * b * c;`, "((a * b) * c)"},
		{`a * b / c;`, "((a * b) / c)"},
		{`a + b / c;`, "(a + (b / c))"},
		{`a + b * c + d;`, "((a + (b * c)) + d)"},
		{`a * b + c / d;`, "((a * b) + (c / d))"},
		{`3 + 4 * 5 == 3 * 1 + 4 * 5;`, "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{`a > b && c < d;`, "((a > b) && (c < d))"},
		{`a || b && c;`, "(a || (b && c))"},
		{`a | b & c;`, "(a | (b & c))"},
		{`a ?? b ?? c;`, "((a ?? b) ?? c)"},
		{`1 + 2 << 3;`, "(1 + (2 << 3))"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		got := stmt.Expression.String()
		if got != tt.expected {
			t.Errorf("precedence: input=%q, got=%q, want=%q", tt.input, got, tt.expected)
		}
	}
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`if (x > 0) { x; }`, "if ((x > 0)) x"},
		{`if (x > 0) { x; } else { -x; }`, "if ((x > 0)) x else (-x)"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		ie, ok := stmt.Expression.(*ast.IfExpression)
		if !ok {
			t.Fatalf("expected *ast.IfExpression, got %T", stmt.Expression)
		}
		if ie.String() != tt.expected {
			t.Errorf("IfExpression.String() = %q, want %q", ie.String(), tt.expected)
		}
	}
}

func TestIfElseIfExpression(t *testing.T) {
	input := `if (x > 0) { "pos"; } else if (x < 0) { "neg"; } else { "zero"; }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	ie, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected *ast.IfExpression, got %T", stmt.Expression)
	}
	got := ie.String()
	expected := `if ((x > 0)) "pos" else if ((x < 0)) "neg" else "zero"`
	if got != expected {
		t.Errorf("IfExpression.String() = %q, want %q", got, expected)
	}
}

func TestFunctionLiteral(t *testing.T) {
	input := `let f = fn(x, y) { return x + y; };`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	ls := prog.Statements[0].(*ast.LetStatement)
	fn, ok := ls.Value.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected *ast.FunctionLiteral, got %T", ls.Value)
	}
	if len(fn.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
	}
	if fn.Parameters[0].String() != "x" {
		t.Errorf("param[0] = %q, want %q", fn.Parameters[0].String(), "x")
	}
	if fn.Parameters[1].String() != "y" {
		t.Errorf("param[1] = %q, want %q", fn.Parameters[1].String(), "y")
	}
}

func TestFunctionStatement(t *testing.T) {
	input := `fn add(a, b) { return a + b; }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	fs, ok := prog.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("expected *ast.FunctionStatement, got %T", prog.Statements[0])
	}
	if fs.Name.Value != "add" {
		t.Errorf("fs.Name = %q, want %q", fs.Name.Value, "add")
	}
	if len(fs.Parameters) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fs.Parameters))
	}
}

func TestFunctionWithReceiver(t *testing.T) {
	input := `fn (p Person) greet() { return "hi"; }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	fs := prog.Statements[0].(*ast.FunctionStatement)
	if fs.ReceiverName == nil || fs.ReceiverType == nil {
		t.Fatalf("expected receiver on function")
	}
	if fs.ReceiverName.Value != "p" {
		t.Errorf("receiver name = %q, want %q", fs.ReceiverName.Value, "p")
	}
	if fs.ReceiverType.Value != "Person" {
		t.Errorf("receiver type = %q, want %q", fs.ReceiverType.Value, "Person")
	}
}

func TestCallExpression(t *testing.T) {
	input := `add(1, 2 * 3);`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	ce, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", stmt.Expression)
	}
	if len(ce.Arguments) != 2 {
		t.Fatalf("expected 2 args, got %d", len(ce.Arguments))
	}
}

func TestCallExpressionPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`add(a, b);`, "add(a, b)"},
		{`add(a + b);`, "add((a + b))"},
		{`a + add(b * c);`, "(a + add((b * c)))"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		got := stmt.Expression.String()
		if got != tt.expected {
			t.Errorf("call precedence: input=%q, got=%q, want=%q", tt.input, got, tt.expected)
		}
	}
}

func TestWhileStatement(t *testing.T) {
	input := `while (x > 0) { echo(x); }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	ws := prog.Statements[0].(*ast.WhileStatement)
	if ws.Condition.String() != "(x > 0)" {
		t.Errorf("while condition = %q, want %q", ws.Condition.String(), "(x > 0)")
	}
}

func TestForStatement(t *testing.T) {
	input := `for (let i = 0; i < 10; i++) { echo(i); }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	fs := prog.Statements[0].(*ast.ForStatement)
	if fs.Init == nil || fs.Condition == nil || fs.Update == nil {
		t.Fatalf("expected init, condition, and update in for-statement")
	}
}

func TestForEachStatement(t *testing.T) {
	input := `for (item in items) { echo(item); }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	fes := prog.Statements[0].(*ast.ForEachStatement)
	if fes.Variable.Value != "item" {
		t.Errorf("foreach variable = %q, want %q", fes.Variable.Value, "item")
	}
}

func TestBreakContinue(t *testing.T) {
	tests := []string{
		`break;`,
		`continue;`,
	}
	for _, input := range tests {
		prog := parseProgram(t, input)
		checkStatements(t, prog, 1)
	}
}

func TestArrayLiteral(t *testing.T) {
	input := `[1, 2, 3];`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	arr, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expected *ast.ArrayLiteral, got %T", stmt.Expression)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestHashLiteral(t *testing.T) {
	input := `{ "a": 1, "b": 2 };`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("expected *ast.HashLiteral, got %T", stmt.Expression)
	}
	if len(hash.Pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(hash.Pairs))
	}
}

func TestIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`obj.prop;`, `(obj."prop")`},
		{`arr[0];`, "(arr[0])"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		got := stmt.Expression.String()
		if got != tt.expected {
			t.Errorf("index: input=%q, got=%q, want=%q", tt.input, got, tt.expected)
		}
	}
}

func TestSwitchStatement(t *testing.T) {
	input := `switch (x) { case 1: echo("one"); default: echo("other"); }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	ss := prog.Statements[0].(*ast.SwitchStatement)
	if ss.Expression.String() != "x" {
		t.Errorf("switch expr = %q, want %q", ss.Expression.String(), "x")
	}
	if len(ss.Cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(ss.Cases))
	}
	if ss.DefaultCase == nil {
		t.Fatal("expected default case")
	}
}

func TestMatchStatement(t *testing.T) {
	input := `match (value) { case 1: "one" case _: "other" }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	ms := prog.Statements[0].(*ast.MatchStatement)
	if ms.Expression.String() != "value" {
		t.Errorf("match expr = %q, want %q", ms.Expression.String(), "value")
	}
	if len(ms.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(ms.Cases))
	}
}

func TestTryCatchThrow(t *testing.T) {
	tests := []string{
		`try { throw "err"; } catch (e) { echo(e); }`,
		`throw "error";`,
	}
	for _, input := range tests {
		prog := parseProgram(t, input)
		checkStatements(t, prog, 1)
	}
}

func TestRetryStatement(t *testing.T) {
	input := `retry (3) { try { throw "fail"; } catch (e) { echo(e); } }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	rs := prog.Statements[0].(*ast.RetryStatement)
	if rs.Attempts.String() != "3" {
		t.Errorf("retry attempts = %q, want %q", rs.Attempts.String(), "3")
	}
}

func TestAssignmentStatement(t *testing.T) {
	input := `x = 42;`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	as := prog.Statements[0].(*ast.AssignmentStatement)
	if as.Left.String() != "x" {
		t.Errorf("assignment left = %q, want %q", as.Left.String(), "x")
	}
	if as.Value.String() != "42" {
		t.Errorf("assignment value = %q, want %q", as.Value.String(), "42")
	}
}

func TestStructStatement(t *testing.T) {
	input := `struct Person { name: string, age: int }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	ss := prog.Statements[0].(*ast.StructStatement)
	if ss.Name.Value != "Person" {
		t.Errorf("struct name = %q, want %q", ss.Name.Value, "Person")
	}
	if len(ss.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(ss.Fields))
	}
}

func TestStructLiteral(t *testing.T) {
	input := `Person { name: "Alice", age: 30 };`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	sl, ok := stmt.Expression.(*ast.StructLiteral)
	if !ok {
		t.Fatalf("expected *ast.StructLiteral, got %T", stmt.Expression)
	}
	if sl.Name.String() != "Person" {
		t.Errorf("struct literal name = %q, want %q", sl.Name.String(), "Person")
	}
	if len(sl.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(sl.Fields))
	}
}

func TestEnumStatement(t *testing.T) {
	input := `enum Color { Red, Green, Blue }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	es := prog.Statements[0].(*ast.EnumStatement)
	if es.Name.Value != "Color" {
		t.Errorf("enum name = %q, want %q", es.Name.Value, "Color")
	}
	if len(es.Values) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(es.Values))
	}
}

func TestInterfaceStatement(t *testing.T) {
	input := `interface Shape { area(): float64; perimeter(): float64; }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	is := prog.Statements[0].(*ast.InterfaceStatement)
	if is.Name.Value != "Shape" {
		t.Errorf("interface name = %q, want %q", is.Name.Value, "Shape")
	}
	if len(is.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(is.Methods))
	}
}

func TestServiceStatement(t *testing.T) {
	input := `service Greeter { port: 8080; fn hello(req) { return "hi"; } }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	ss := prog.Statements[0].(*ast.ServiceStatement)
	if ss.Name.Value != "Greeter" {
		t.Errorf("service name = %q, want %q", ss.Name.Value, "Greeter")
	}
}

func TestImportStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`import "foo";`, `import "foo"`},
		{`import { bar } from "baz";`, `import { bar } from "baz"`},
		{`import * as stuff from "lib";`, `import * as stuff from "lib"`},
		{`import defaultExport from "mod";`, `import defaultExport from "mod"`},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		is := prog.Statements[0].(*ast.ImportStatement)
		if is.String() != tt.expected {
			t.Errorf("ImportStatement.String() = %q, want %q", is.String(), tt.expected)
		}
	}
}

func TestExportStatement(t *testing.T) {
	input := `export fn hello() { return "hi"; }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	es := prog.Statements[0].(*ast.ExportStatement)
	if es.ExportType != ast.EXPORT_DECLARATION {
		t.Errorf("expected EXPORT_DECLARATION, got %v", es.ExportType)
	}
}

func TestArrowFunction(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`x => x + 1;`, "x => (x + 1)"},
		{`(a, b) => a + b;`, "(a, b) => (a + b)"},
	}
	for _, tt := range tests {
		prog := parseProgram(t, tt.input)
		checkStatements(t, prog, 1)
		stmt := prog.Statements[0].(*ast.ExpressionStatement)
		af, ok := stmt.Expression.(*ast.ArrowFunction)
		if !ok {
			t.Fatalf("expected *ast.ArrowFunction, got %T", stmt.Expression)
		}
		if af.String() != tt.expected {
			t.Errorf("ArrowFunction.String() = %q, want %q", af.String(), tt.expected)
		}
	}
}

func TestAsyncFunction(t *testing.T) {
	input := `async fn fetch(url) { return await http.get(url); }`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	afs := prog.Statements[0].(*ast.AsyncFunctionStatement)
	if afs.Name.Value != "fetch" {
		t.Errorf("async fn name = %q, want %q", afs.Name.Value, "fetch")
	}
}

func TestSpawnAwait(t *testing.T) {
	tests := []string{
		`let task = spawn fetchData();`,
		`let result = await task;`,
	}
	for _, input := range tests {
		prog := parseProgram(t, input)
		checkStatements(t, prog, 1)
	}
}

func TestTemplateLiteral(t *testing.T) {
	input := "`hello ${name}`;"
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	tl, ok := stmt.Expression.(*ast.TemplateLiteral)
	if !ok {
		t.Fatalf("expected *ast.TemplateLiteral, got %T", stmt.Expression)
	}
	expected := "`hello ${name}`"
	if tl.String() != expected {
		t.Errorf("TemplateLiteral.String() = %q, want %q", tl.String(), expected)
	}
}

func TestPipeExpression(t *testing.T) {
	input := `x |> double |> toString;`
	prog := parseProgram(t, input)
	checkStatements(t, prog, 1)
	stmt := prog.Statements[0].(*ast.ExpressionStatement)
	ce, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression (pipe desugars to call), got %T", stmt.Expression)
	}
	// x |> double |> toString desugars to toString(double(x))
	if ce.Function.String() != "toString" {
		t.Errorf("outer call function = %q, want %q", ce.Function.String(), "toString")
	}
	if len(ce.Arguments) != 1 {
		t.Fatalf("expected 1 arg to outer call, got %d", len(ce.Arguments))
	}
	innerCall, ok := ce.Arguments[0].(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected inner *ast.CallExpression, got %T", ce.Arguments[0])
	}
	if innerCall.Function.String() != "double" {
		t.Errorf("inner call function = %q, want %q", innerCall.Function.String(), "double")
	}
	if len(innerCall.Arguments) != 1 {
		t.Fatalf("expected 1 arg to inner call, got %d", len(innerCall.Arguments))
	}
	if innerCall.Arguments[0].String() != "x" {
		t.Errorf("inner call arg = %q, want %q", innerCall.Arguments[0].String(), "x")
	}
}

func TestMeterTrace(t *testing.T) {
	tests := []string{
		`meter "counter";`,
		`trace "span" { echo("work"); }`,
	}
	for _, input := range tests {
		prog := parseProgram(t, input)
		checkStatements(t, prog, 1)
	}
}

func TestEmptyProgram(t *testing.T) {
	input := ``
	prog := parseProgram(t, input)
	if len(prog.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(prog.Statements))
	}
}

func TestMultipleStatements(t *testing.T) {
	input := `let x = 1; let y = 2; fn add(a, b) { return a + b; }`
	prog := parseProgram(t, input)
	if len(prog.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(prog.Statements))
	}
	_, ok1 := prog.Statements[0].(*ast.LetStatement)
	_, ok2 := prog.Statements[1].(*ast.LetStatement)
	_, ok3 := prog.Statements[2].(*ast.FunctionStatement)
	if !ok1 || !ok2 || !ok3 {
		t.Errorf("unexpected statement types in multi-statement program")
	}
}

func TestParseErrors(t *testing.T) {
	tests := []string{
		`let x = ;`,
		`fn ( { }`,
		`if (true) { } else`,
	}
	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()
		if len(p.Errors()) == 0 {
			t.Errorf("expected parse errors for input: %q", input)
		}
	}
}

func TestTypeCastExpression(t *testing.T) {
	tests := []string{
		`int64(x);`,
		`float64(x);`,
		`string(x);`,
	}
	for _, input := range tests {
		prog := parseProgram(t, input)
		checkStatements(t, prog, 1)
	}
}

func TestGenericsSyntax(t *testing.T) {
	tests := []string{
		`fn identity<T>(x: T): T { return x; }`,
	}
	for _, input := range tests {
		prog := parseProgram(t, input)
		checkStatements(t, prog, 1)
	}
}

func TestCommentsAreSkipped(t *testing.T) {
	input := `// this is a comment
let x = 1; /* block comment */
let y = 2;`
	prog := parseProgram(t, input)
	if len(prog.Statements) != 2 {
		t.Fatalf("expected 2 statements (comments skipped), got %d", len(prog.Statements))
	}
}
