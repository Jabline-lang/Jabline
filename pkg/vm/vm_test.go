package vm

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
	"testing"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
	}

	return nil
}

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := comp.Bytecode()
		vm := New(bytecode.Instructions, bytecode.Constants, "test.jb")
		err = vm.Run()
		if err != nil {
			switch expected := tt.expected.(type) {
			case error:
				if err.Error() != expected.Error() {
					t.Fatalf("wrong error returned. \nwant=%q\ngot=%q", expected, err)
				}
				continue // Error matched, test pass
			default:
				t.Fatalf("vm error: %s", err)
			}
		}

		if _, ok := tt.expected.(error); ok {
			t.Fatalf("expected error %q but got none", tt.expected)
		}

		stackElem := vm.LastPoppedStackElem()

		switch expected := tt.expected.(type) {
		case int:
			err := testIntegerObject(int64(expected), stackElem)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	}
}

func TestTypeCheckingInVM(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
				let a: int = 15
				a
			`,
			expected: 15,
		},
		{
			input: `
				let a: string = 15
				a
			`,
			// It should fail on type check
			expected: fmt.Errorf("runtime error: type error: expected type string, got INTEGER"),
		},
	}

	runVmTests(t, tests)
}
