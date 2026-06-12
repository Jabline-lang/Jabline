package typechecker

import (
	"testing"

	"jabline/pkg/ast"
	"jabline/pkg/token"
)

func mkToken(tokType token.TokenType, lit string, line, col int) token.Token {
	return token.Token{Type: tokType, Literal: lit, Line: line, Column: col}
}

func intLit(v int64) *ast.IntegerLiteral {
	return &ast.IntegerLiteral{Token: mkToken(token.INT, "0", 1, 1), Value: v}
}

func floatLit(v float64) *ast.FloatLiteral {
	return &ast.FloatLiteral{Token: mkToken(token.FLOAT, "0.0", 1, 1), Value: v}
}

func strLit(v string) *ast.StringLiteral {
	return &ast.StringLiteral{Token: mkToken(token.STRING, v, 1, 1), Value: v}
}

func boolLit(v bool) *ast.Boolean {
	return &ast.Boolean{Token: mkToken(token.TRUE, "true", 1, 1), Value: v}
}

func nullExpr() *ast.Null {
	return &ast.Null{Token: mkToken(token.NULL, "null", 1, 1)}
}

func ident(name string) *ast.Identifier {
	return &ast.Identifier{Token: mkToken(token.IDENT, name, 1, 1), Value: name}
}

func typeExpr(value string) *ast.TypeExpression {
	return &ast.TypeExpression{Token: mkToken(token.IDENT, value, 1, 1), Value: value}
}

func blockStmt(stmts ...ast.Statement) *ast.BlockStatement {
	return &ast.BlockStatement{
		Token:      mkToken(token.LBRACE, "{", 1, 1),
		Statements: stmts,
	}
}

func exprStmt(expr ast.Expression) *ast.ExpressionStatement {
	return &ast.ExpressionStatement{
		Token:      mkToken(token.ILLEGAL, "", 1, 1),
		Expression: expr,
	}
}

// ---------------------------------------------------------------------------
// Environment tests
// ---------------------------------------------------------------------------

func TestNewEnvironment(t *testing.T) {
	env := NewEnvironment()
	if env == nil {
		t.Fatal("NewEnvironment returned nil")
	}
	if env.store == nil {
		t.Error("store should be initialized")
	}
	if env.outer != nil {
		t.Error("root env should have nil outer")
	}
	if env.expectedReturn != TypeAny {
		t.Errorf("expectedReturn should be TypeAny, got %s", env.expectedReturn)
	}
	if env.loopDepth != 0 {
		t.Errorf("loopDepth should be 0, got %d", env.loopDepth)
	}
}

func TestEnvironmentSetAndGet(t *testing.T) {
	env := NewEnvironment()
	env.Set("x", TypeInt)
	env.Set("y", TypeString)

	if typ, ok := env.Get("x"); !ok || typ != TypeInt {
		t.Errorf("Get(x) = %s, %v; want int, true", typ, ok)
	}
	if typ, ok := env.Get("y"); !ok || typ != TypeString {
		t.Errorf("Get(y) = %s, %v; want string, true", typ, ok)
	}
	if _, ok := env.Get("z"); ok {
		t.Error("Get(z) should be false for undefined variable")
	}
}

func TestEnvironmentOverride(t *testing.T) {
	env := NewEnvironment()
	env.Set("x", TypeInt)
	env.Set("x", TypeString)

	if typ, ok := env.Get("x"); !ok || typ != TypeString {
		t.Errorf("Get(x) = %s, %v; want string, true", typ, ok)
	}
}

func TestEnclosedEnvironment(t *testing.T) {
	outer := NewEnvironment()
	outer.Set("x", TypeInt)
	outer.Set("y", TypeString)

	inner := NewEnclosedEnvironment(outer)
	inner.Set("z", TypeBool)

	// Inner can see all three
	if typ, ok := inner.Get("x"); !ok || typ != TypeInt {
		t.Errorf("inner.Get(x) = %s, %v; want int, true", typ, ok)
	}
	if typ, ok := inner.Get("y"); !ok || typ != TypeString {
		t.Errorf("inner.Get(y) = %s, %v; want string, true", typ, ok)
	}
	if typ, ok := inner.Get("z"); !ok || typ != TypeBool {
		t.Errorf("inner.Get(z) = %s, %v; want bool, true", typ, ok)
	}

	// Outer cannot see inner's variable
	if _, ok := outer.Get("z"); ok {
		t.Error("outer.Get(z) should be false (shadowing)")
	}

	// Inner shadows outer
	inner.Set("x", TypeFloat)
	if typ, ok := inner.Get("x"); !ok || typ != TypeFloat {
		t.Errorf("shadowed inner.Get(x) = %s, %v; want float, true", typ, ok)
	}
	if typ, ok := outer.Get("x"); !ok || typ != TypeInt {
		t.Errorf("outer.Get(x) after shadow = %s, %v; want int, true", typ, ok)
	}
}

func TestEnclosedEnvironmentInheritsExpectedReturn(t *testing.T) {
	outer := NewEnvironment()
	outer.expectedReturn = TypeInt

	inner := NewEnclosedEnvironment(outer)
	if inner.expectedReturn != TypeInt {
		t.Errorf("inner.expectedReturn = %s; want int", inner.expectedReturn)
	}

	inner.expectedReturn = TypeString
	if outer.expectedReturn != TypeInt {
		t.Errorf("outer.expectedReturn changed to %s; want int", outer.expectedReturn)
	}
}

func TestEnclosedEnvironmentInheritsLoopDepth(t *testing.T) {
	outer := NewEnvironment()
	outer.loopDepth = 2

	inner := NewEnclosedEnvironment(outer)
	if inner.loopDepth != 2 {
		t.Errorf("inner.loopDepth = %d; want 2", inner.loopDepth)
	}

	inner.loopDepth++
	if outer.loopDepth != 2 {
		t.Errorf("outer.loopDepth changed to %d; want 2", outer.loopDepth)
	}
}

// ---------------------------------------------------------------------------
// ParseASTType tests
// ---------------------------------------------------------------------------

func TestParseASTType(t *testing.T) {
	tests := []struct {
		input    string
		expected TypeType
	}{
		{"int", TypeInt},
		{"int8", TypeInt8},
		{"int16", TypeInt16},
		{"int32", TypeInt32},
		{"int64", TypeInt64},
		{"uint8", TypeUint8},
		{"uint16", TypeUint16},
		{"uint32", TypeUint32},
		{"uint64", TypeUint64},
		{"float", TypeFloat},
		{"float32", TypeFloat32},
		{"float64", TypeFloat64},
		{"string", TypeString},
		{"bool", TypeBool},
		{"void", TypeVoid},
		{"any", TypeAny},
		{"Array", TypeArray},
		{"Hash", TypeHash},
		{"Channel", TypeChannel},
		{"null", TypeNull},
		{"MyCustomType", TypeType("MyCustomType")},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseASTType(typeExpr(tt.input), nil)
			if result != tt.expected {
				t.Errorf("ParseASTType(%q, nil) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseASTTypeNil(t *testing.T) {
	result := ParseASTType(nil, nil)
	if result != TypeAny {
		t.Errorf("ParseASTType(nil, nil) = %s; want any", result)
	}
}

// ---------------------------------------------------------------------------
// Helper: run type checker and check errors
// ---------------------------------------------------------------------------

func checkProgram(t *testing.T, stmts []ast.Statement) *Checker {
	t.Helper()
	checker := New()
	program := &ast.Program{Statements: stmts}
	checker.Check(program)
	return checker
}

func assertNoErrors(t *testing.T, checker *Checker) {
	t.Helper()
	if len(checker.errors) > 0 {
		t.Errorf("Expected no errors, got:\n")
		for _, e := range checker.errors {
			t.Errorf("  %s", e)
		}
	}
}

func assertErrors(t *testing.T, checker *Checker, expected int) {
	t.Helper()
	if len(checker.errors) != expected {
		t.Errorf("Expected %d errors, got %d:\n", expected, len(checker.errors))
		for _, e := range checker.errors {
			t.Errorf("  %s", e)
		}
	}
}

// ---------------------------------------------------------------------------
// Expression type checking tests
// ---------------------------------------------------------------------------

func TestCheckIntegerLiteral(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{exprStmt(intLit(42))})
	if len(checker.errors) > 0 {
		t.Errorf("Expected no errors for integer literal, got: %v", checker.errors)
	}
}

func TestCheckFloatLiteral(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{exprStmt(floatLit(3.14))})
	assertNoErrors(t, checker)
}

func TestCheckStringLiteral(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{exprStmt(strLit("hello"))})
	assertNoErrors(t, checker)
}

func TestCheckBooleanLiteral(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{exprStmt(boolLit(true))})
	assertNoErrors(t, checker)
}

func TestCheckNullLiteral(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{exprStmt(nullExpr())})
	assertNoErrors(t, checker)
}

func TestCheckIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		setup   []ast.Statement
		checker func(*testing.T, *Checker)
	}{
		{
			"defined variable returns its type",
			[]ast.Statement{
				&ast.LetStatement{Token: mkToken(token.LET, "let", 1, 1), Name: ident("x"), Value: intLit(42)},
				exprStmt(ident("x")),
			},
			func(t *testing.T, c *Checker) { assertNoErrors(t, c) },
		},
		{
			"undefined variable treated as any",
			[]ast.Statement{
				exprStmt(ident("undefinedVar")),
			},
			func(t *testing.T, c *Checker) { assertNoErrors(t, c) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			tt.checker(t, checker)
		})
	}
}

func TestCheckInfixArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"int + int", &ast.InfixExpression{Left: intLit(1), Operator: "+", Right: intLit(2)}, 0},
		{"int - int", &ast.InfixExpression{Left: intLit(1), Operator: "-", Right: intLit(2)}, 0},
		{"int * int", &ast.InfixExpression{Left: intLit(1), Operator: "*", Right: intLit(2)}, 0},
		{"int / int", &ast.InfixExpression{Left: intLit(1), Operator: "/", Right: intLit(2)}, 0},
		{"float + float", &ast.InfixExpression{Left: floatLit(1.0), Operator: "+", Right: floatLit(2.0)}, 0},
		{"int + float", &ast.InfixExpression{Left: intLit(1), Operator: "+", Right: floatLit(2.0)}, 0},
		{"float + int", &ast.InfixExpression{Left: floatLit(1.0), Operator: "+", Right: intLit(2)}, 0},
		{"string + string", &ast.InfixExpression{Left: strLit("a"), Operator: "+", Right: strLit("b")}, 0},
		{"string + int (error)", &ast.InfixExpression{Left: strLit("a"), Operator: "+", Right: intLit(1)}, 1},
		{"int + string (error)", &ast.InfixExpression{Left: intLit(1), Operator: "+", Right: strLit("b")}, 1},
		{"string - string (error)", &ast.InfixExpression{Left: strLit("a"), Operator: "-", Right: strLit("b")}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckInfixMod(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"int % int", &ast.InfixExpression{Left: intLit(10), Operator: "%", Right: intLit(3)}, 0},
		{"float % float (error)", &ast.InfixExpression{Left: floatLit(10.0), Operator: "%", Right: floatLit(3.0)}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckInfixEquality(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"int == int", &ast.InfixExpression{Left: intLit(1), Operator: "==", Right: intLit(2)}, 0},
		{"int != float", &ast.InfixExpression{Left: intLit(1), Operator: "!=", Right: floatLit(2.0)}, 0},
		{"string == string", &ast.InfixExpression{Left: strLit("a"), Operator: "==", Right: strLit("b")}, 0},
		{"string == int (error)", &ast.InfixExpression{Left: strLit("a"), Operator: "==", Right: intLit(1)}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckInfixComparison(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"int < int", &ast.InfixExpression{Left: intLit(1), Operator: "<", Right: intLit(2)}, 0},
		{"int > float", &ast.InfixExpression{Left: intLit(1), Operator: ">", Right: floatLit(2.0)}, 0},
		{"string < string (no error, same type)", &ast.InfixExpression{Left: strLit("a"), Operator: "<", Right: strLit("b")}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckInfixLogical(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"bool && bool", &ast.InfixExpression{Left: boolLit(true), Operator: "&&", Right: boolLit(false)}, 0},
		{"bool || bool", &ast.InfixExpression{Left: boolLit(true), Operator: "||", Right: boolLit(false)}, 0},
		{"int && int (error)", &ast.InfixExpression{Left: intLit(1), Operator: "&&", Right: intLit(0)}, 2},
		{"bool || int (error)", &ast.InfixExpression{Left: boolLit(true), Operator: "||", Right: intLit(1)}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckInfixBitwise(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"int & int", &ast.InfixExpression{Left: intLit(1), Operator: "&", Right: intLit(3)}, 0},
		{"int | int", &ast.InfixExpression{Left: intLit(1), Operator: "|", Right: intLit(3)}, 0},
		{"int ^ int", &ast.InfixExpression{Left: intLit(1), Operator: "^", Right: intLit(3)}, 0},
		{"int << int", &ast.InfixExpression{Left: intLit(1), Operator: "<<", Right: intLit(3)}, 0},
		{"int >> int", &ast.InfixExpression{Left: intLit(1), Operator: ">>", Right: intLit(3)}, 0},
		{"float & float (error)", &ast.InfixExpression{Left: floatLit(1.0), Operator: "&", Right: floatLit(3.0)}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckPrefix(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"!bool", &ast.PrefixExpression{Operator: "!", Right: boolLit(true)}, 0},
		{"-int", &ast.PrefixExpression{Operator: "-", Right: intLit(5)}, 0},
		{"-float", &ast.PrefixExpression{Operator: "-", Right: floatLit(3.14)}, 0},
		{"~int", &ast.PrefixExpression{Operator: "~", Right: intLit(7)}, 0},
		{"!int (error)", &ast.PrefixExpression{Operator: "!", Right: intLit(1)}, 1},
		{"-string (error)", &ast.PrefixExpression{Operator: "-", Right: strLit("a")}, 1},
		{"~bool (error)", &ast.PrefixExpression{Operator: "~", Right: boolLit(true)}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckPostfix(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"x++ on int variable",
			[]ast.Statement{
				&ast.LetStatement{Token: mkToken(token.LET, "let", 1, 1), Name: ident("x"), Value: intLit(0)},
				exprStmt(&ast.PostfixExpression{Left: ident("x"), Operator: "++"}),
			},
			0,
		},
		{
			"x-- on float variable",
			[]ast.Statement{
				&ast.LetStatement{Token: mkToken(token.LET, "let", 1, 1), Name: ident("f"), Value: floatLit(0.0)},
				exprStmt(&ast.PostfixExpression{Left: ident("f"), Operator: "--"}),
			},
			0,
		},
		{
			"undefined variable in postfix (error)",
			[]ast.Statement{
				exprStmt(&ast.PostfixExpression{Left: ident("nope"), Operator: "++"}),
			},
			1,
		},
		{
			"string postfix (error)",
			[]ast.Statement{
				&ast.LetStatement{Token: mkToken(token.LET, "let", 1, 1), Name: ident("s"), Value: strLit("hi")},
				exprStmt(&ast.PostfixExpression{Left: ident("s"), Operator: "++"}),
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckIfExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"if bool { block }", &ast.IfExpression{
			Condition:   boolLit(true),
			Consequence: blockStmt(exprStmt(intLit(1))),
		}, 0},
		{"if int { block } (error)", &ast.IfExpression{
			Condition:   intLit(1),
			Consequence: blockStmt(exprStmt(intLit(1))),
		}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckArrayIndex(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"array[int]", &ast.ArrayIndexExpression{
			Left:  &ast.ArrayLiteral{Elements: []ast.Expression{intLit(1)}},
			Index: intLit(0),
		}, 0},
		{"string[int]", &ast.ArrayIndexExpression{
			Left:  strLit("hello"),
			Index: intLit(0),
		}, 0},
		{"array[string] (error)", &ast.ArrayIndexExpression{
			Left:  &ast.ArrayLiteral{Elements: []ast.Expression{intLit(1)}},
			Index: strLit("zero"),
		}, 1},
		{"int[int] (error)", &ast.ArrayIndexExpression{
			Left:  intLit(42),
			Index: intLit(0),
		}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckTernary(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"bool ? int : int", &ast.TernaryExpression{
			Condition: boolLit(true), TrueValue: intLit(1), FalseValue: intLit(2),
		}, 0},
		{"bool ? int : string (error)", &ast.TernaryExpression{
			Condition: boolLit(true), TrueValue: intLit(1), FalseValue: strLit("two"),
		}, 1},
		{"int ? int : int (error)", &ast.TernaryExpression{
			Condition: intLit(1), TrueValue: intLit(1), FalseValue: intLit(2),
		}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckNullishCoalescing(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{"null ?? int", &ast.NullishCoalescingExpression{Left: nullExpr(), Right: intLit(42)}, 0},
		{"int ?? int", &ast.NullishCoalescingExpression{Left: intLit(1), Right: intLit(2)}, 0},
		{"bool ?? string (error)", &ast.NullishCoalescingExpression{Left: boolLit(true), Right: strLit("x")}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckCallExpression(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		exprStmt(&ast.CallExpression{
			Function:  ident("f"),
			Arguments: []ast.Expression{intLit(1), strLit("a")},
		}),
	})
	assertNoErrors(t, checker)
}

func TestCheckFunctionLiteral(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		exprStmt(&ast.FunctionLiteral{
			Parameters: []*ast.Identifier{ident("x")},
			Body:       blockStmt(exprStmt(intLit(42))),
		}),
	})
	assertNoErrors(t, checker)
}

func TestCheckArrowFunction(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantErrs int
	}{
		{
			"arrow with no return type annotation",
			&ast.ArrowFunction{
				Parameters: []*ast.Identifier{ident("x")},
				Body:       intLit(1),
			},
			0,
		},
		{
			"arrow with matching return type",
			&ast.ArrowFunction{
				Parameters: []*ast.Identifier{ident("x")},
				ReturnType: typeExpr("int"),
				Body:       intLit(1),
			},
			0,
		},
		{
			"arrow with mismatched return type (error)",
			&ast.ArrowFunction{
				Parameters: []*ast.Identifier{ident("x")},
				ReturnType: typeExpr("int"),
				Body:       strLit("wrong"),
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, []ast.Statement{exprStmt(tt.expr)})
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckStructLiteral(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"defined struct type",
			[]ast.Statement{
				&ast.StructStatement{
					Token:  mkToken(token.STRUCT, "struct", 1, 1),
					Name:   ident("Point"),
					Fields: map[string]*ast.TypeExpression{"x": typeExpr("int")},
				},
				exprStmt(&ast.StructLiteral{
					Name:   ident("Point"),
					Fields: map[string]ast.Expression{"x": intLit(0)},
				}),
			},
			0,
		},
		{
			"undefined struct type",
			[]ast.Statement{
				exprStmt(&ast.StructLiteral{
					Name:   ident("NonExistent"),
					Fields: map[string]ast.Expression{},
				}),
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

// ---------------------------------------------------------------------------
// Statement type checking tests
// ---------------------------------------------------------------------------

func TestCheckLetStatement(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"let x = 5",
			[]ast.Statement{
				&ast.LetStatement{Token: mkToken(token.LET, "let", 1, 1), Name: ident("x"), Value: intLit(5)},
			},
			0,
		},
		{
			"let x: int = 5",
			[]ast.Statement{
				&ast.LetStatement{Token: mkToken(token.LET, "let", 1, 1), Name: ident("x"), Type: typeExpr("int"), Value: intLit(5)},
			},
			0,
		},
		{
			"let x: string = 5 (error)",
			[]ast.Statement{
				&ast.LetStatement{Token: mkToken(token.LET, "let", 1, 1), Name: ident("x"), Type: typeExpr("string"), Value: intLit(5)},
			},
			1,
		},
		{
			"let x: int = float (ok, numeric)",
			[]ast.Statement{
				&ast.LetStatement{Token: mkToken(token.LET, "let", 1, 1), Name: ident("x"), Type: typeExpr("int"), Value: floatLit(5.0)},
			},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckLetStatementStoresType(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.LetStatement{Token: mkToken(token.LET, "let", 1, 1), Name: ident("x"), Type: typeExpr("bool"), Value: boolLit(true)},
		exprStmt(&ast.PrefixExpression{Operator: "!", Right: ident("x")}),
	})
	assertNoErrors(t, checker)
}

func TestCheckConstStatement(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"const x = 5",
			[]ast.Statement{
				&ast.ConstStatement{Token: mkToken(token.CONST, "const", 1, 1), Name: ident("x"), Value: intLit(5)},
			},
			0,
		},
		{
			"const x: string = 5 (error)",
			[]ast.Statement{
				&ast.ConstStatement{Token: mkToken(token.CONST, "const", 1, 1), Name: ident("x"), Type: typeExpr("string"), Value: intLit(5)},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckReturnStatement(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"return int from int function",
			[]ast.Statement{
				&ast.FunctionStatement{
					Token:      mkToken(token.FUNCTION, "fn", 1, 1),
					Name:       ident("f"),
					ReturnType: typeExpr("int"),
					Body: blockStmt(
						&ast.ReturnStatement{Token: mkToken(token.RETURN, "return", 2, 1), ReturnValue: intLit(42)},
					),
				},
			},
			0,
		},
		{
			"return string from int function (error)",
			[]ast.Statement{
				&ast.FunctionStatement{
					Token:      mkToken(token.FUNCTION, "fn", 1, 1),
					Name:       ident("f"),
					ReturnType: typeExpr("int"),
					Body: blockStmt(
						&ast.ReturnStatement{Token: mkToken(token.RETURN, "return", 2, 1), ReturnValue: strLit("hi")},
					),
				},
			},
			1,
		},
		{
			"return void (no value) from void function",
			[]ast.Statement{
				&ast.FunctionStatement{
					Token:      mkToken(token.FUNCTION, "fn", 1, 1),
					Name:       ident("f"),
					ReturnType: typeExpr("void"),
					Body: blockStmt(
						&ast.ReturnStatement{Token: mkToken(token.RETURN, "return", 2, 1)},
					),
				},
			},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckWhileCondition(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"while bool",
			[]ast.Statement{
				&ast.WhileStatement{
					Condition: boolLit(true),
					Body:      blockStmt(),
				},
			},
			0,
		},
		{
			"while int (error)",
			[]ast.Statement{
				&ast.WhileStatement{
					Condition: intLit(1),
					Body:      blockStmt(),
				},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckBreakContinueInLoops(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"break inside while",
			[]ast.Statement{
				&ast.WhileStatement{
					Condition: boolLit(true),
					Body: blockStmt(
						&ast.BreakStatement{Token: mkToken(token.BREAK, "break", 2, 1)},
					),
				},
			},
			0,
		},
		{
			"continue inside for",
			[]ast.Statement{
				&ast.ForStatement{
					Init:      &ast.LetStatement{Name: ident("i"), Value: intLit(0)},
					Condition: boolLit(true),
					Body: blockStmt(
						&ast.ContinueStatement{Token: mkToken(token.CONTINUE, "continue", 2, 1)},
					),
				},
			},
			0,
		},
		{
			"break outside loop (error)",
			[]ast.Statement{
				&ast.BreakStatement{Token: mkToken(token.BREAK, "break", 1, 1)},
			},
			1,
		},
		{
			"continue outside loop (error)",
			[]ast.Statement{
				&ast.ContinueStatement{Token: mkToken(token.CONTINUE, "continue", 1, 1)},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckForCondition(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"for with bool condition",
			[]ast.Statement{
				&ast.ForStatement{
					Condition: boolLit(true),
					Body:      blockStmt(),
				},
			},
			0,
		},
		{
			"for with int condition (error)",
			[]ast.Statement{
				&ast.ForStatement{
					Condition: intLit(1),
					Body:      blockStmt(),
				},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckForEach(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"foreach with array",
			[]ast.Statement{
				&ast.ForEachStatement{
					Variable: ident("v"),
					Iterable: &ast.ArrayLiteral{Elements: []ast.Expression{intLit(1)}},
					Body:     blockStmt(),
				},
			},
			0,
		},
		{
			"foreach with string",
			[]ast.Statement{
				&ast.ForEachStatement{
					Variable: ident("c"),
					Iterable: strLit("hello"),
					Body:     blockStmt(),
				},
			},
			0,
		},
		{
			"foreach with int (error)",
			[]ast.Statement{
				&ast.ForEachStatement{
					Variable: ident("x"),
					Iterable: intLit(42),
					Body:     blockStmt(),
				},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckAssignment(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"assign matching type",
			[]ast.Statement{
				&ast.LetStatement{Name: ident("x"), Value: intLit(0)},
				&ast.AssignmentStatement{Left: ident("x"), Value: intLit(42)},
			},
			0,
		},
		{
			"assign mismatched type (error)",
			[]ast.Statement{
				&ast.LetStatement{Name: ident("x"), Type: typeExpr("int"), Value: intLit(0)},
				&ast.AssignmentStatement{Left: ident("x"), Value: strLit("hi")},
			},
			1,
		},
		{
			"assign to undefined (error)",
			[]ast.Statement{
				&ast.AssignmentStatement{Left: ident("nope"), Value: intLit(1)},
			},
			1,
		},
		{
			"index assignment",
			[]ast.Statement{
				&ast.AssignmentStatement{
					Left:  &ast.IndexExpression{Left: ident("obj"), Index: strLit("key")},
					Value: intLit(1),
				},
			},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckStructStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.StructStatement{
			Token:  mkToken(token.STRUCT, "struct", 1, 1),
			Name:   ident("Point"),
			Fields: map[string]*ast.TypeExpression{"x": typeExpr("int"), "y": typeExpr("float")},
		},
	})
	assertNoErrors(t, checker)
	if typ, ok := checker.env.Get("Point"); !ok || typ != TypeType("Point") {
		t.Errorf("Point type = %s; want Point", typ)
	}
}

func TestCheckEnumStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.EnumStatement{
			Token:  mkToken(token.ENUM, "enum", 1, 1),
			Name:   ident("Color"),
			Values: []*ast.Identifier{ident("Red"), ident("Green"), ident("Blue")},
		},
	})
	assertNoErrors(t, checker)
	if typ, ok := checker.env.Get("Color"); !ok || typ != TypeHash {
		t.Errorf("Color type = %s; want Hash", typ)
	}
}

func TestCheckInterfaceStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.InterfaceStatement{
			Token: mkToken(token.INTERFACE, "interface", 1, 1),
			Name:  ident("Drawable"),
			Methods: map[string]*ast.FunctionSignature{
				"draw": {Name: "draw", ReturnType: typeExpr("void")},
			},
		},
	})
	assertNoErrors(t, checker)
	if typ, ok := checker.env.Get("Drawable"); !ok || typ != TypeType("Drawable") {
		t.Errorf("Drawable type = %s; want Drawable", typ)
	}
}

func TestCheckServiceStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.ServiceStatement{
			Token: mkToken(token.SERVICE, "service", 1, 1),
			Name:  ident("MyService"),
			Fields: map[string]ast.Expression{
				"port": intLit(8080),
			},
			Methods: []*ast.FunctionStatement{
				{
					Token:      mkToken(token.FUNCTION, "fn", 2, 1),
					Name:       ident("handle"),
					ReturnType: typeExpr("int"),
					Body:       blockStmt(&ast.ReturnStatement{ReturnValue: intLit(0)}),
				},
			},
		},
	})
	assertNoErrors(t, checker)
	if typ, ok := checker.env.Get("MyService"); !ok || typ != TypeType("MyService") {
		t.Errorf("MyService type = %s; want MyService", typ)
	}
}

func TestCheckTryStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.TryStatement{
			TryBlock:   blockStmt(exprStmt(intLit(1))),
			CatchBlock: blockStmt(exprStmt(ident("e"))),
			CatchParam: ident("e"),
		},
	})
	assertNoErrors(t, checker)
}

func TestCheckRetryStatement(t *testing.T) {
	tests := []struct {
		name     string
		setup    []ast.Statement
		wantErrs int
	}{
		{
			"retry with int attempts",
			[]ast.Statement{
				&ast.RetryStatement{
					Attempts:   intLit(3),
					RetryBlock: blockStmt(exprStmt(intLit(1))),
				},
			},
			0,
		},
		{
			"retry with string attempts (error)",
			[]ast.Statement{
				&ast.RetryStatement{
					Attempts:   strLit("three"),
					RetryBlock: blockStmt(exprStmt(intLit(1))),
				},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := checkProgram(t, tt.setup)
			assertErrors(t, checker, tt.wantErrs)
		})
	}
}

func TestCheckThrowStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.ThrowStatement{Token: mkToken(token.THROW, "throw", 1, 1), Value: strLit("error")},
	})
	assertNoErrors(t, checker)
}

func TestCheckSwitchStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.SwitchStatement{
			Expression: ident("x"),
			Cases: []*ast.CaseClause{
				{Value: intLit(1), Statements: []ast.Statement{exprStmt(intLit(10))}},
			},
		},
	})
	assertNoErrors(t, checker)
}

func TestCheckMatchStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.MatchStatement{
			Expression: intLit(1),
			Cases: []*ast.MatchCase{
				{Pattern: intLit(1), Statements: []ast.Statement{exprStmt(strLit("one"))}},
			},
		},
	})
	assertNoErrors(t, checker)
}

func TestCheckEchoStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.EchoStatement{
			Token:  mkToken(token.ECHO, "echo", 1, 1),
			Values: []ast.Expression{intLit(1), strLit("hello"), boolLit(true)},
		},
	})
	assertNoErrors(t, checker)
}

func TestCheckImportStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.ImportStatement{
			ModuleName: strLit("math"),
		},
	})
	assertNoErrors(t, checker)
}

func TestCheckExportStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.ExportStatement{
			Statement: &ast.LetStatement{Name: ident("x"), Value: intLit(5)},
		},
	})
	assertNoErrors(t, checker)
}

func TestCheckMeterStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.MeterStatement{
			Token:    mkToken(token.METER, "meter", 1, 1),
			Name:     strLit("requests_total"),
			Operator: "++",
		},
	})
	assertNoErrors(t, checker)
}

func TestCheckTraceStatement(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.TraceStatement{
			Token: mkToken(token.TRACE, "trace", 1, 1),
			Name:  strLit("db_query"),
			Body:  blockStmt(exprStmt(intLit(1))),
		},
	})
	assertNoErrors(t, checker)
}

func TestCheckSpawnExpression(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		exprStmt(&ast.SpawnExpression{
			Call: &ast.CallExpression{
				Function:  ident("f"),
				Arguments: []ast.Expression{intLit(1)},
			},
		}),
	})
	assertNoErrors(t, checker)
}

func TestCheckAwaitExpression(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		exprStmt(&ast.AwaitExpression{
			Value: ident("future"),
		}),
	})
	assertNoErrors(t, checker)
}

func TestCheckOptionalChaining(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		exprStmt(&ast.OptionalChainingExpression{
			Left:  ident("obj"),
			Right: ident("prop"),
		}),
	})
	assertNoErrors(t, checker)
}

func TestCheckTemplateLiteral(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		exprStmt(&ast.TemplateLiteral{
			Parts:       []string{"hello ", " world"},
			Expressions: []ast.Expression{ident("name")},
		}),
	})
	assertNoErrors(t, checker)
}

// ---------------------------------------------------------------------------
// Integration tests: valid programs
// ---------------------------------------------------------------------------

func TestCheckerFibonacciProgram(t *testing.T) {
	checker := checkProgram(t, []ast.Statement{
		&ast.FunctionStatement{
			Token:      mkToken(token.FUNCTION, "fn", 1, 1),
			Name:       ident("fib"),
			ReturnType: typeExpr("int"),
			Parameters: []*ast.Identifier{ident("n")},
			Body: blockStmt(
				exprStmt(&ast.IfExpression{
					Condition:   &ast.InfixExpression{Left: ident("n"), Operator: "<", Right: intLit(2)},
					Consequence: blockStmt(&ast.ReturnStatement{ReturnValue: ident("n")}),
					Alternative: blockStmt(
						&ast.ReturnStatement{
							ReturnValue: &ast.InfixExpression{
								Left:  &ast.CallExpression{Function: ident("fib"), Arguments: []ast.Expression{&ast.InfixExpression{Left: ident("n"), Operator: "-", Right: intLit(1)}}},
								Operator: "+",
								Right: &ast.CallExpression{Function: ident("fib"), Arguments: []ast.Expression{&ast.InfixExpression{Left: ident("n"), Operator: "-", Right: intLit(2)}}},
							},
						},
					),
				}),
			),
		},
	})
	assertNoErrors(t, checker)
}
