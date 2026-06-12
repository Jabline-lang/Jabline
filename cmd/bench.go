package cmd

import (
	"fmt"
	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"jabline/pkg/vm"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:   "bench [file.jb]",
	Short: "Run benchmarks for a Jabline file",
	Long: `Runs the given .jb file repeatedly and reports execution time.
If no file is specified, runs built-in benchmarks.

Examples:
  jabline bench myfile.jb
  jabline bench            # Run built-in benchmarks
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			benchFile(args[0])
		} else {
			runBuiltinBenchmarks()
		}
	},
}

func init() {
	rootCmd.AddCommand(benchCmd)
}

func benchFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %s\n", path, err)
		os.Exit(1)
	}
	input := string(data)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors in %s:\n", path)
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
		os.Exit(1)
	}

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Fprintf(os.Stderr, "Compile error: %s\n", err)
		os.Exit(1)
	}

	bc := comp.Bytecode()
	n := benchIterations
	fmt.Printf("Benchmark: %s\n", path)
	fmt.Printf("Iterations: %d\n", n)

	var total time.Duration
	fastest := time.Duration(1<<63 - 1)
	slowest := time.Duration(0)

	for i := 0; i < n; i++ {
		machine := vm.New(bc.Instructions, bc.Constants, path)
		start := time.Now()
		if err := machine.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Run error: %s\n", err)
			os.Exit(1)
		}
		duration := time.Since(start)
		total += duration
		if duration < fastest {
			fastest = duration
		}
		if duration > slowest {
			slowest = duration
		}
	}

	avg := total / time.Duration(n)
	fmt.Printf("Total: %v\n", total)
	fmt.Printf("Avg:   %v\n", avg)
	fmt.Printf("Fast:  %v\n", fastest)
	fmt.Printf("Slow:  %v\n", slowest)
	fmt.Printf("Ops/s: %.0f\n", float64(n)/total.Seconds())
}

var benchIterations = 100

func runBuiltinBenchmarks() {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"Fibonacci(20)", `let fib = fn(n) { if (n < 2) { return n }; return fib(n - 1) + fib(n - 2) }; fib(20)`},
		{"Loop 10k", `let s = 0; for (let i = 0; i < 10000; i++) { s = s + i }; s`},
		{"TCO Factorial(1000)", `let f = fn(n, a) { if (n == 0) { return a }; return f(n - 1, n * a) }; f(1000, 1)`},
		{"Array ops", `let a = []; for (let i = 0; i < 1000; i++) { a = push(a, i) }; len(a)`},
		{"Hash ops", `let h = {}; for (let i = 0; i < 1000; i++) { h[to_str(i)] = i }; h["500"]`},
		{"Closure call", `let mk = fn(x) { return fn(y) { return x + y } }; let a = mk(5); let r = 0; for (let i = 0; i < 1000; i++) { r = a(i) }; r`},
	}

	fmt.Println("Jabline Built-in Benchmarks")
	fmt.Println(string(repeatChar('=', 60)))
	fmt.Printf("%-25s %12s %12s %12s\n", "Benchmark", "Avg", "Fast", "Ops/s")
	fmt.Println(string(repeatChar('-', 60)))

	for _, bm := range benchmarks {
		l := lexer.New(bm.input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			fmt.Fprintf(os.Stderr, "  %s: parse error, skipping\n", bm.name)
			continue
		}
		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: compile error, skipping\n", bm.name)
			continue
		}

		n := 50
		var total time.Duration
		fastest := time.Duration(1<<63 - 1)

		bc := comp.Bytecode()
		for i := 0; i < n; i++ {
			machine := vm.New(bc.Instructions, bc.Constants, "bench")
			start := time.Now()
			if err := machine.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "  %s: run error: %s\n", bm.name, err)
				continue
			}
			d := time.Since(start)
			total += d
			if d < fastest {
				fastest = d
			}
		}

		avg := total / time.Duration(n)
		fmt.Printf("%-25s %12v %12v %12.0f\n", bm.name, avg, fastest, float64(n)/total.Seconds())
	}
	fmt.Println(string(repeatChar('=', 60)))
}

func repeatChar(c rune, n int) string {
	s := make([]rune, n)
	for i := range s {
		s[i] = c
	}
	return string(s)
}
