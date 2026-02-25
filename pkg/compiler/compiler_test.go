package compiler

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func TestConstAndLetBinding(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
				let a = 1;
				const b = 2;
			`,
			expectedConstants: []interface{}{int64(1), int64(2)},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 6),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 7),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestEnumCompilation(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `enum Status { Active, Inactive }`,
			expectedConstants: []interface{}{
				&object.Hash{
					Pairs: map[object.HashKey]object.HashPair{
						(&object.String{Value: "Active"}).HashKey(): {
							Key:   &object.String{Value: "Active"},
							Value: &object.Integer{Value: 0},
						},
						(&object.String{Value: "Inactive"}).HashKey(): {
							Key:   &object.String{Value: "Inactive"},
							Value: &object.Integer{Value: 1},
						},
					},
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 6),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}
	}
}

func testInstructions(
	expected []code.Instructions,
	actual code.Instructions,
) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot =%q",
			concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot =%q",
				i, concatted, actual)
		}
	}

	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}
	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}
