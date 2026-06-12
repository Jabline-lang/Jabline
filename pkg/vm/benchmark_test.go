package vm

import (
	"jabline/pkg/code"
	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
	"testing"
)

// Test to satisfy Go testing framework
func TestBenchmarkSuite(t *testing.T) {}

func benchmarkJabline(input string, b *testing.B) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		b.Fatalf("parse errors: %v", p.Errors())
	}

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		b.Fatalf("compile error: %v", err)
	}

	bc := comp.Bytecode()
	machine := New(bc.Instructions, bc.Constants, "bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		machine = New(bc.Instructions, bc.Constants, "bench")
		if err := machine.Run(); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}

func BenchmarkFibonacci(b *testing.B) {
	input := `
let fib = fn(n) {
	if (n < 2) { return n }
	return fib(n - 1) + fib(n - 2)
}
fib(20)
`
	benchmarkJabline(input, b)
}

func BenchmarkTCOFactorial(b *testing.B) {
	input := `
let fact = fn(n, acc) {
	if (n == 0) { return acc }
	return fact(n - 1, n * acc)
}
fact(1000, 1)
`
	benchmarkJabline(input, b)
}

func BenchmarkLoop(b *testing.B) {
	input := `
let sum = 0
for (let i = 0; i < 10000; i++) {
	sum = sum + i
}
sum
`
	benchmarkJabline(input, b)
}

func BenchmarkArrayMap(b *testing.B) {
	input := `
let arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
let result = []
for (let i = 0; i < len(arr); i++) {
	result = push(result, arr[i] * 2)
}
result
`
	benchmarkJabline(input, b)
}

func BenchmarkStringConcat(b *testing.B) {
	input := `
let s = ""
for (let i = 0; i < 100; i++) {
	s = s + "hello"
}
s
`
	benchmarkJabline(input, b)
}

func BenchmarkRecursiveFib(b *testing.B) {
	input := `
let fib = fn(n) {
	if (n <= 1) { return n }
	return fib(n - 1) + fib(n - 2)
}
fib(20)
`
	benchmarkJabline(input, b)
}

func BenchmarkClosure(b *testing.B) {
	input := `
let makeAdder = fn(x) {
	return fn(y) { return x + y }
}
let add5 = makeAdder(5)
let r = 0
for (let i = 0; i < 1000; i++) {
	r = add5(i)
}
r
`
	benchmarkJabline(input, b)
}

// Benchmark against equivalent Go operations

func BenchmarkGoSimpleLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < 10000; j++ {
			sum += j
		}
		_ = sum
	}
}

func BenchmarkGoFib20(b *testing.B) {
	var fib func(n int) int
	fib = func(n int) int {
		if n < 2 {
			return n
		}
		return fib(n-1) + fib(n-2)
	}
	for i := 0; i < b.N; i++ {
		fib(20)
	}
}

// Parallel benchmark
func BenchmarkGoParallelSum(b *testing.B) {
	data := make([]int, 10000)
	for i := range data {
		data[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for _, v := range data {
			sum += v
		}
		_ = sum
	}
}

// Bytecode execution benchmark
func BenchmarkBytecodeExec(b *testing.B) {
	instr := code.Instructions{
		byte(code.OpConstant), 0, 0,
		byte(code.OpPop),
		byte(code.OpTrue),
		byte(code.OpPop),
		byte(code.OpConstant), 1, 0,
		byte(code.OpPop),
		byte(code.OpConstant), 2, 0,
		byte(code.OpPop),
		byte(code.OpNull),
	}
	constants := []object.Object{
		&object.Integer{Value: 42},
		&object.String{Value: "hello"},
		&object.Array{Elements: []object.Object{
			&object.Integer{Value: 1},
			&object.Integer{Value: 2},
			&object.Integer{Value: 3},
		}},
	}

	machine := New(instr, constants, "bench")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		machine = New(instr, constants, "bench")
		if err := machine.Run(); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}

// Benchmark hash map operations
func BenchmarkHashOps(b *testing.B) {
	input := `
let h = {"a": 1, "b": 2, "c": 3}
let keys = ["a", "b", "c"]
let sum = 0
for (let i = 0; i < 1000; i++) {
	sum = sum + h[keys[i % 3]]
}
sum
`
	benchmarkJabline(input, b)
}

func BenchmarkConcurrentSpawn(b *testing.B) {
	input := `
let ch = make_chan()
let fn = fn() { send(ch, 42) }
spawn(fn)
recv(ch)
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		b.Fatalf("parse errors: %v", p.Errors())
	}

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		b.Fatalf("compile error: %v", err)
	}

	bc := comp.Bytecode()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		machine := New(bc.Instructions, bc.Constants, "bench")
		if err := machine.Run(); err != nil {
			b.Fatalf("run error: %v", err)
		}
	}
}


