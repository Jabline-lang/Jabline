package parser

import (
	"jabline/pkg/ast"
	"jabline/pkg/lexer"
	"testing"
)

func TestArrowFunctionParsing(t *testing.T) {
	tests := []struct {
		input              string
		expectedParamCount int
		expectedBody       string
	}{
		{"() => 1", 0, "1"},
		{"(a) => a", 1, "a"},
		{"(a, b) => a + b", 2, "(a + b)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		arrow, ok := stmt.Expression.(*ast.ArrowFunction)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.ArrowFunction. got=%T", stmt.Expression)
		}

		if len(arrow.Parameters) != tt.expectedParamCount {
			t.Fatalf("function has wrong number of parameters. want=%d, got=%d",
				tt.expectedParamCount, len(arrow.Parameters))
		}

		if arrow.Body.String() != tt.expectedBody {
			t.Errorf("body is not %q. got=%q", tt.expectedBody, arrow.Body.String())
		}
	}
}

func TestGroupedExpressionParsing(t *testing.T) {
	input := "(5 + 5)"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Expression.String() != "(5 + 5)" {
		t.Errorf("expression is not %q. got=%q", "(5 + 5)", stmt.Expression.String())
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
