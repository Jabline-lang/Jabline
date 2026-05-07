package fmt

import (
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"testing"
)

// formatSource parses source and returns the formatted output.
// It also asserts there are no parse errors.
func formatSource(t *testing.T, src string) string {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	return Format(prog)
}

// assertIdempotent verifies that formatting once and then formatting the result
// produces the same output (i.e. the formatter is stable).
func assertIdempotent(t *testing.T, src string) {
	t.Helper()
	first := formatSource(t, src)
	second := formatSource(t, first)
	if first != second {
		t.Errorf("formatter is NOT idempotent.\nFirst pass:\n%s\n\nSecond pass:\n%s", first, second)
	}
}

// assertFormat verifies the formatted output matches expected.
func assertFormat(t *testing.T, src, expected string) {
	t.Helper()
	got := formatSource(t, src)
	if got != expected {
		t.Errorf("formatted output mismatch.\nSource:\n%s\nExpected:\n%s\nGot:\n%s", src, expected, got)
	}
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestFormatLiterals(t *testing.T) {
	cases := []struct{ src, want string }{
		{`let x = 42;`, `let x = 42;`},
		{`let x = 3.14;`, `let x = 3.14;`},
		{`let x = "hello";`, `let x = "hello";`},
		{`let x = true;`, `let x = true;`},
		{`let x = false;`, `let x = false;`},
		{`let x = null;`, `let x = null;`},
		{"let x = `hello ${name}`;", "let x = `hello ${name}`;"},
	}
	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			assertFormat(t, c.src, c.want)
			assertIdempotent(t, c.src)
		})
	}
}

func TestFormatLetConst(t *testing.T) {
	assertFormat(t, `let a = 1;`, `let a = 1;`)
	assertFormat(t, `const PI = 3.14;`, `const PI = 3.14;`)
	assertIdempotent(t, `let a = 1; const b = 2;`)
}

func TestFormatArrays(t *testing.T) {
	assertFormat(t, `let a = [];`, `let a = [];`)
	assertFormat(t, `let a = [1, 2, 3];`, `let a = [1, 2, 3];`)
	assertIdempotent(t, `let a = [1, 2, 3];`)
	// Nested arrays should be multiline
	assertIdempotent(t, `let a = [[1, 2], [3, 4]];`)
}

func TestFormatHashes(t *testing.T) {
	assertFormat(t, `let h = {};`, `let h = {};`)
	// Hash pairs are sorted deterministically — use pre-sorted keys in idempotent check
	assertIdempotent(t, `let h = {"a": 1, "z": 2};`)
}

func TestFormatInfixAndPrefix(t *testing.T) {
	assertFormat(t, `let x = 1 + 2 * 3;`, `let x = 1 + 2 * 3;`)
	assertFormat(t, `let x = !true;`, `let x = !true;`)
	assertFormat(t, `let x = -5;`, `let x = -5;`)
	assertIdempotent(t, `let x = a && b || c;`)
}

func TestFormatIf(t *testing.T) {
	src := `if (x > 0) { echo("pos"); }`
	assertIdempotent(t, src)

	src2 := `if (x > 0) { echo("pos"); } else { echo("neg"); }`
	assertIdempotent(t, src2)

	// else-if chain
	src3 := `if (x > 0) { echo("pos"); } else if (x < 0) { echo("neg"); } else { echo("zero"); }`
	assertIdempotent(t, src3)
}

func TestFormatWhile(t *testing.T) {
	assertIdempotent(t, `while (i < 10) { i = i + 1; }`)
}

func TestFormatFor(t *testing.T) {
	assertIdempotent(t, `for (let i = 0; i < 10; i = i + 1) { echo(i); }`)
}

func TestFormatForEach(t *testing.T) {
	assertIdempotent(t, `for (item in items) { echo(item); }`)
}

func TestFormatSwitch(t *testing.T) {
	src := `switch (x) { case 1: echo("one"); case 2: echo("two"); default: echo("other"); }`
	assertIdempotent(t, src)
}

func TestFormatMatch(t *testing.T) {
	src := `match (x) { case 1: echo("one"); default: echo("other"); }`
	assertIdempotent(t, src)
}

func TestFormatTryCatch(t *testing.T) {
	assertIdempotent(t, `try { risky(); } catch (e) { echo(e); }`)
	assertIdempotent(t, `try { risky(); } catch { echo("err"); }`)
}

func TestFormatRetry(t *testing.T) {
	assertIdempotent(t, `retry (3) { fetch(); } catch (e) { echo(e); }`)
}

func TestFormatThrow(t *testing.T) {
	assertFormat(t, `throw "error";`, `throw "error";`)
}

func TestFormatBreakContinue(t *testing.T) {
	// The formatter always expands blocks — single-line is not preserved
	assertIdempotent(t, `while (true) { break; }`)
	assertIdempotent(t, `while (true) { if (x) { continue; } }`)
}

func TestFormatReturn(t *testing.T) {
	// Formatter always expands blocks to multiple lines
	assertIdempotent(t, `fn add(a, b) { return a + b; }`)
	assertIdempotent(t, `fn noop() { return; }`)
}

func TestFormatFunctionStatement(t *testing.T) {
	src := `fn greet(name) { echo("Hello", name); }`
	assertIdempotent(t, src)

	// With return type
	src2 := `fn add(a, b): int { return a + b; }`
	assertIdempotent(t, src2)
}

func TestFormatAsyncFunction(t *testing.T) {
	src := `async fn fetch(url) { return await http.get(url); }`
	assertIdempotent(t, src)
}

func TestFormatFunctionLiteral(t *testing.T) {
	src := `let add = fn(a, b) { return a + b; };`
	assertIdempotent(t, src)
}

func TestFormatArrowFunction(t *testing.T) {
	assertIdempotent(t, `let double = x => x * 2;`)
	assertIdempotent(t, `let add = (a, b) => a + b;`)
}

func TestFormatMethodCall(t *testing.T) {
	assertIdempotent(t, `srv.get("/ping", fn(ctx) { return ctx.text("Pong", 200); });`)
}

func TestFormatDotAccess(t *testing.T) {
	assertFormat(t, `let x = a.b;`, `let x = a.b;`)
	assertIdempotent(t, `let x = a.b.c;`)
}

func TestFormatBracketAccess(t *testing.T) {
	assertFormat(t, `let x = arr[0];`, `let x = arr[0];`)
	assertIdempotent(t, `let x = map[key];`)
}

func TestFormatTernary(t *testing.T) {
	assertIdempotent(t, `let x = a > 0 ? "pos" : "neg";`)
}

func TestFormatNullishCoalescing(t *testing.T) {
	assertIdempotent(t, `let x = val ?? "default";`)
}

func TestFormatOptionalChaining(t *testing.T) {
	assertIdempotent(t, `let x = obj?.prop;`)
}

func TestFormatAwaitSpawn(t *testing.T) {
	assertIdempotent(t, `let result = await fetch("http://api.com");`)
	assertIdempotent(t, `spawn worker(data);`)
}

func TestFormatStruct(t *testing.T) {
	src := "struct Point {\n    x: int,\n    y: int\n}"
	assertIdempotent(t, src)
}

func TestFormatInterface(t *testing.T) {
	src := `interface Greetable { greet(name): string }`
	assertIdempotent(t, src)
}

func TestFormatEnum(t *testing.T) {
	src := `enum Status { Active, Inactive, Pending }`
	assertIdempotent(t, src)
}

func TestFormatImport(t *testing.T) {
	assertIdempotent(t, `import { NewRouter, ListenAndServe } from "net/http";`)
	assertIdempotent(t, `import * as http from "net/http";`)
	assertIdempotent(t, `import Router from "net/http";`)
}

func TestFormatExport(t *testing.T) {
	assertIdempotent(t, `export let PI = 3.14;`)
	assertIdempotent(t, `export fn greet(name) { echo(name); }`)
	assertIdempotent(t, `export { foo, bar };`)
}

func TestFormatMeterTrace(t *testing.T) {
	assertIdempotent(t, `meter requests_total++;`)
	assertIdempotent(t, `trace db_query { query(); }`)
}

func TestFormatEcho(t *testing.T) {
	assertFormat(t, `echo("hello", "world");`, `echo("hello", "world");`)
	assertIdempotent(t, `echo(1, 2, 3);`)
}

func TestFormatAssignment(t *testing.T) {
	assertFormat(t, `x = 42;`, `x = 42;`)
	assertIdempotent(t, `obj.name = "Alice";`)
}

func TestFormatComplexProgram(t *testing.T) {
	src := `
import { NewRouter, ListenAndServe } from "net/http";

fn handlePing(ctx) {
    return ctx.text("Pong", 200);
}

let srv = NewRouter();
srv.get("/ping", handlePing);
ListenAndServe(8080, srv);
`
	assertIdempotent(t, src)
}

func TestFormatFibonacci(t *testing.T) {
	src := `
fn fib(n) {
    if (n <= 1) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

echo(fib(10));
`
	assertIdempotent(t, src)
}

func TestFormatFizzBuzz(t *testing.T) {
	src := `
let i = 1;
while (i <= 100) {
    if (i % 15 == 0) {
        echo("FizzBuzz");
    } else if (i % 3 == 0) {
        echo("Fizz");
    } else if (i % 5 == 0) {
        echo("Buzz");
    } else {
        echo(i);
    }
    i = i + 1;
}
`
	assertIdempotent(t, src)
}
