package vm

import (
	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
	"testing"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func parseCompileAndRun(t *testing.T, input string) *VM {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	c := compiler.New()
	err := c.Compile(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	bc := c.Bytecode()
	vmInst := New(bc.Instructions, bc.Constants, "test")
	err = vmInst.Run()
	if err != nil {
		t.Fatalf("VM error: %v", err)
	}
	return vmInst
}

func runVMTests(t *testing.T, tests []vmTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			vmInst := parseCompileAndRun(t, tt.input)
			result := vmInst.LastPoppedStackElem()
			checkExpected(t, result, tt.expected)
		})
	}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{`1 + 2;`, 3},
		{`5 - 3;`, 2},
		{`2 * 3;`, 6},
		{`10 / 2;`, 5},
		{`7 % 3;`, 1},
		{`1 + 2 * 3;`, 7},
		{`(1 + 2) * 3;`, 9},
		{`-5;`, -5},
	}
	runVMTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`true;`, true},
		{`false;`, false},
		{`!true;`, false},
		{`!false;`, true},
		{`1 == 1;`, true},
		{`1 != 1;`, false},
		{`1 < 2;`, true},
		{`1 > 2;`, false},
		{`true && true;`, true},
		{`true && false;`, false},
		{`true || false;`, true},
	}
	runVMTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"hello";`, "hello"},
		{`"hello" + " " + "world";`, "hello world"},
	}
	runVMTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{`[];`, []int{}},
		{`[1, 2, 3];`, []int{1, 2, 3}},
	}
	runVMTests(t, tests)
}

func TestLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{`let x = 5; x;`, 5},
		{`let x = 5; let y = x; y;`, 5},
		{`let x = 5; let y = 10; let z = x + y; z;`, 15},
	}
	runVMTests(t, tests)
}

func TestConstStatements(t *testing.T) {
	tests := []vmTestCase{
		{`const x = 42; x;`, 42},
		{`const x = 10; const y = x * 2; y;`, 20},
	}
	runVMTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{`if (1 == 1) { 10; } else { 20; };`, 10},
		{`if (1 != 1) { 10; } else { 20; };`, 20},
		{`if (1 != 1) { 10; };`, nil},
		{`if (1 < 2) { 10; } else { 20; };`, 10},
	}
	runVMTests(t, tests)
}

func TestWhileLoop(t *testing.T) {
	tests := []vmTestCase{
		{`let x = 0; while (x < 3) { x = x + 1; }; x;`, 3},
	}
	runVMTests(t, tests)
}

func TestForLoop(t *testing.T) {
	tests := []vmTestCase{
		{`let sum = 0; for (let i = 0; i < 3; i = i + 1) { sum = sum + i; }; sum;`, 3},
	}
	runVMTests(t, tests)
}

func TestBreakContinue(t *testing.T) {
	tests := []vmTestCase{
		{`let x = 0; while (x < 5) { if (x > 2) { break; }; x = x + 1; }; x;`, 3},
	}
	runVMTests(t, tests)
}

func TestFunctionCalls(t *testing.T) {
	tests := []vmTestCase{
		{`let f = fn() { return 42; }; f();`, 42},
		{`let f = fn(x) { return x + 1; }; f(41);`, 42},
		{`let f = fn(a, b) { return a + b; }; f(10, 20);`, 30},
		{`let f = fn() { 42; }; f();`, 42},
		{`fn add(a, b) { return a + b; }; add(3, 4);`, 7},
	}
	runVMTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{`let add = fn(a, b) { return a + b; }; add(1, 2);`, 3},
		{`let maker = fn(x) { return fn() { return x; }; }; let f = maker(42); f();`, 42},
	}
	runVMTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`fn fib(n) { if (n < 2) { return n; }; return fib(n - 1) + fib(n - 2); }; fib(10);`, 55},
	}
	runVMTests(t, tests)
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`len("hello");`, 5},
		{`len([1, 2, 3]);`, 3},
		{`first([10, 20, 30]);`, 10},
		{`last([10, 20, 30]);`, 30},
	}
	runVMTests(t, tests)
}

func TestNullValues(t *testing.T) {
	tests := []vmTestCase{
		{`null;`, nil},
		{`let x = null; x;`, nil},
	}
	runVMTests(t, tests)
}

func TestAssignment(t *testing.T) {
	tests := []vmTestCase{
		{`let x = 5; x = 10; x;`, 10},
	}
	runVMTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`let arr = [1, 2, 3]; arr[0];`, 1},
		{`let arr = [1, 2, 3]; arr[1];`, 2},
		{`let arr = [1, 2, 3]; arr[2];`, 3},
	}
	runVMTests(t, tests)
}

func TestPostfixExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`let x = 5; x++; x;`, 6},
		{`let x = 5; x--; x;`, 4},
	}
	runVMTests(t, tests)
}

func TestTernaryExpression(t *testing.T) {
	tests := []vmTestCase{
		{`true ? 42 : 0;`, 42},
		{`false ? 42 : 0;`, 0},
	}
	runVMTests(t, tests)
}

func TestNullishCoalescing(t *testing.T) {
	tests := []vmTestCase{
		{`null ?? 42;`, 42},
		{`false ?? 42;`, false},
	}
	runVMTests(t, tests)
}

func TestGlobalScope(t *testing.T) {
	tests := []vmTestCase{
		{`let x = 1; fn f() { return x; }; x;`, 1},
	}
	runVMTests(t, tests)
}

// ─── Helpers ───────────────────────────────────────────────────────

func checkExpected(t *testing.T, got object.Object, expected interface{}) {
	t.Helper()
	switch e := expected.(type) {
	case int:
		checkIntegerObject(t, got, int64(e))
	case int64:
		checkIntegerObject(t, got, e)
	case string:
		checkStringObject(t, got, e)
	case bool:
		checkBooleanObject(t, got, e)
	case nil:
		checkNullObject(t, got)
	case []int:
		arr, ok := got.(*object.Array)
		if !ok {
			t.Fatalf("expected *object.Array, got %T", got)
		}
		if len(arr.Elements) != len(e) {
			t.Fatalf("array length mismatch: expected %d, got %d", len(e), len(arr.Elements))
		}
		for i, elem := range arr.Elements {
			checkIntegerObject(t, elem, int64(e[i]))
		}
	default:
		t.Fatalf("unexpected expected type: %T", expected)
	}
}

func checkIntegerObject(t *testing.T, obj object.Object, expected int64) {
	t.Helper()
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Fatalf("expected *object.Integer, got %T (value=%+v)", obj, obj)
	}
	if result.Value != expected {
		t.Errorf("integer value mismatch: expected %d, got %d", expected, result.Value)
	}
}

func checkStringObject(t *testing.T, obj object.Object, expected string) {
	t.Helper()
	result, ok := obj.(*object.String)
	if !ok {
		t.Fatalf("expected *object.String, got %T", obj)
	}
	if result.Value != expected {
		t.Errorf("string value mismatch: expected %q, got %q", expected, result.Value)
	}
}

func checkBooleanObject(t *testing.T, obj object.Object, expected bool) {
	t.Helper()
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Fatalf("expected *object.Boolean, got %T", obj)
	}
	if result.Value != expected {
		t.Errorf("boolean value mismatch: expected %t, got %t", expected, result.Value)
	}
}

func checkNullObject(t *testing.T, obj object.Object) {
	t.Helper()
	if obj != Null {
		t.Errorf("expected null, got %T (%+v)", obj, obj)
	}
}
