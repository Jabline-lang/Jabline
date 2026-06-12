<div align="center">
  <img src="assets/jabline.png" alt="Jabline Logo" width="120">

  # Jabline

  A compiled, cloud-native programming language with a custom bytecode virtual machine.

  [![Build](https://github.com/Jabline-lang/Jabline/actions/workflows/ci.yml/badge.svg)](https://github.com/Jabline-lang/Jabline/actions/workflows/ci.yml)
  [![Version](https://img.shields.io/badge/version-v0.6.0-blue.svg)](https://github.com/Jabline-lang/Jabline/releases)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)
  [![Tests](https://img.shields.io/badge/tests-516%20passing-brightgreen.svg)](https://github.com/Jabline-lang/Jabline/actions)
  [![Docs](https://img.shields.io/badge/docs-388%20examples-blueviolet.svg)](docs/)
  [![Registry](https://img.shields.io/badge/registry-7%20packages-orange.svg)](registry/)

</div>

---

## Quick Start

```bash
# Install
git clone https://github.com/Jabline-lang/Jabline.git
cd Jabline && go build -o jabline .

# Hello world
echo 'echo("Hello, World!")' > hello.jb
./jabline run hello.jb

# REPL
./jabline repl

# Try more examples
./jabline run examples/hello.jb
```

---

## Performance

| Benchmark | Jabline v0.6.0 | Go (native) | Ratio |
|-----------|----------------|-------------|-------|
| Fibonacci(20) | ~0.8ms | ~0.02ms | 40x |
| Loop 10k iterations | ~0.3ms | ~0.005ms | 60x |
| String concat 100x | ~0.05ms | ~0.001ms | 50x |
| Hash ops 1k | ~0.4ms | ~0.01ms | 40x |
| Concurrent spawn | ~0.9ms | ~0.1ms | 9x |

*Jabline is a bytecode-interpreted VM. Performance is competitive with Python, Ruby, and Lua for most workloads.*

---

## Features

- **Full OOP**: structs, methods, interfaces, enums, generics `<T>`
- **Concurrency**: `spawn`, channels, `select`, `async`/`await`
- **Error handling**: `try`/`catch`/`finally`, `panic`/`recover`, `throw`
- **Expressions**: pipe `|>`, optional chaining `?.`, nullish coalescing `??`, template literals `` `hello ${name}` ``
- **Pattern matching**: `match` expressions, `switch`/`case`, `enum` variants
- **Built-in toolchain**: LSP server, formatter, test runner, package manager, debugger (DAP)
- **FFI**: Call C libraries directly from Jabline
- **Sandbox**: Execute untrusted code with configurable permission levels
- **Telemetry**: `meter` and `trace` for observability
- **35 stdlib modules**: HTTP server, SQLite, JSON, YAML, CSV, crypto, regex, datetime, websocket, SSH, Redis, and more

---

## Quick Tour

```javascript
// Variables and types
let name: string = "Jabline";
const VERSION: int = 6;
let pi = 3.14159;

// Functions and closures
fn counter() {
    let n = 0;
    return fn() { n = n + 1; return n; };
}
let c = counter();
echo(c()); // 1
echo(c()); // 2

// Structs with methods
struct User { name: string, age: int }
fn (u User) greet() { echo("Hi, I'm " + u.name); }
let alice = User{name: "Alice", age: 30};
alice.greet();

// Generics
struct Box[T] { value: T }
fn identity<T>(x: T): T { return x; }

// Concurrency
let ch = make_chan();
spawn fn() { send(ch, "done"); }();
echo(recv(ch));

// Error handling
try {
    throw "something went wrong";
} catch (err) {
    echo("Caught: " + err);
}

// Pattern matching
match (value) {
    type string: echo("text");
    type int: echo("number");
    else: echo("other");
}

// Pipe operator
let result = [1, 2, 3]
    |> fn(arr) { let s = 0; for (x in arr) { s = s + x }; return s; }
    |> fn(s) { s * 2 };
echo(result); // 12
```

---

## CLI

```
Usage:  jabline <command> [options]

Run:      jabline run <file.jb>            Run a script
          jabline run -e "<code>"           Inline execution
          jabline run <file> --hot          Hot-reload on changes
          jabline run <file> --bytecode     Dump bytecode
          jabline run <file> --ast          Print AST

Build:    jabline build <file.jb>           Compile to native binary
          jabline build --standalone        Alias for build

Test:     jabline test                      Run *_test.jb files
          jabline run <test_file>           Run a specific test

Format:   jabline fmt <path>                Format source files
          jabline fmt --check               Check formatting (CI)
          jabline fmt --lint                Lint style issues
          jabline fmt --watch               Auto-format on save

Debug:    jabline debug <file.jb>           Start DAP debugger

Packages: jabline init [name]               Create a project
          jabline get <pkg>                 Add a dependency
          jabline install                   Install all deps
          jabline search <query>            Search registry
          jabline publish                   Publish a package

Tools:    jabline lsp                       Start LSP server
          jabline repl                      Interactive REPL
          jabline --version                 Print version
```

---

## Documentation

Comprehensive documentation with 388 runnable examples is available in [`docs/`](docs/):

| Category | Files | What you'll learn |
|----------|-------|-------------------|
| [Getting Started](docs/getting-started/) | 11 | Installation, hello world, CLI basics |
| [Basics](docs/basics/) | 30 | Variables, types, operators, control flow |
| [Functions](docs/functions/) | 20 | Closures, recursion, pipe, defer |
| [Collections](docs/collections/) | 25 | Arrays, hashes, sets, operations |
| [Structs](docs/structs/) | 15 | Definitions, methods, composition |
| [Interfaces](docs/interfaces/) | 10 | Duck typing, polymorphism |
| [Enums](docs/enums/) | 10 | C-style enums, matching |
| [Generics](docs/generics/) | 10 | Generic functions and types |
| [Error Handling](docs/error-handling/) | 15 | Try/catch, panic/recover |
| [Concurrency](docs/concurrency/) | 15 | Spawn, channels, select, async |
| [Modules](docs/modules/) | 12 | Import/export, JPM packages |
| [Stdlib](docs/stdlib/) | 149 | All standard library modules |
| [Advanced](docs/advanced/) | 15 | FFI, services, telemetry, sandbox |
| [Tooling](docs/tooling/) | 15 | LSP, formatter, debugger, REPL |
| [Testing](docs/testing/) | 15 | Assertions, describe/it, mocking |
| [Patterns](docs/patterns/) | 20 | Singleton, factory, observer, actor |

Each example is **max 50 lines** and runs with `jabline run docs/<category>/<file>.jb`.

---

## Examples

| Example | Description |
|---------|-------------|
| [hello.jb](examples/hello.jb) | Hello World |
| [fibonacci.jb](examples/fibonacci.jb) | Fibonacci with recursion |
| [http_server.jb](examples/http_server.jb) | HTTP server with SQLite |
| [concurrency.jb](examples/concurrency.jb) | Channels and spawn |
| [generics.jb](examples/generics.jb) | Generic functions and types |
| [structs.jb](examples/structs.jb) | Structs and methods |
| [error_handling.jb](examples/error_handling.jb) | Try/catch patterns |
| [task-api](examples/task-api/) | Full CRUD HTTP + SQLite app |

---

## Package Registry

| Package | Description | Version |
|---------|-------------|---------|
| [validation](https://github.com/Jabline-lang/jabline-validation) | Email, URL, IP, UUID, schema validation | 0.1.0 |
| [collections](https://github.com/Jabline-lang/jabline-collections) | Stack, Queue, Option, Result, Pair | 0.1.0 |
| [http](https://github.com/Jabline-lang/jabline-http) | HTTP client/server utilities | 0.1.0 |
| [crypto](https://github.com/Jabline-lang/jabline-crypto) | Hash, base64, hex, HMAC | 0.1.0 |
| [strings](https://github.com/Jabline-lang/jabline-strings) | Case conversion, split/join, trim | 0.1.0 |
| [datetime](https://github.com/Jabline-lang/jabline-datetime) | Date/time formatting, parsing | 0.1.0 |
| [testing](https://github.com/Jabline-lang/jabline-testing) | Assertions, describe/it, test runner | 0.1.0 |

---

## VS Code Extension

The [Jabline VS Code extension](vscode-jabline/) provides:
- Syntax highlighting
- Autocompletion and hover docs
- Error diagnostics
- Code formatting on save
- Debugging via DAP

Install from the `.vsix` file in [`vscode-jabline/`](vscode-jabline/).

---

## Architecture

```
.jb source  ──►  Lexer  ──►  Parser  ──►  Compiler  ──►  JBVM
                 tokens        AST           bytecode      execution

Optimizer: constant folding, dead code elimination, peephole optimization
Runtime:   pooled frames, generational stack, inline caching for methods
Tooling:   LSP, formatter, DAP debugger, JPM package manager, test runner
```

---

## Project Structure

```
├── cmd/           CLI commands (13 commands)
├── pkg/
│   ├── compiler/  Bytecode compiler + optimizer
│   ├── vm/        Stack-based virtual machine (72 opcodes)
│   ├── stdlib/    35 native modules + 31 embedded .jb modules
│   ├── lsp/       Language Server Protocol (15 handlers)
│   ├── fmt/       Code formatter + linter
│   ├── dap/       Debug Adapter Protocol
│   └── jpm/       Package manager with semver resolution
├── docs/          388 runnable examples (16 categories)
├── examples/      20+ example programs
├── packages/      7 JPM packages
├── registry/      Package registry index
└── internal/      Builder, embedded modules
```

---

## Building from Source

**Prerequisites:** Go 1.21+

```bash
git clone https://github.com/Jabline-lang/Jabline.git
cd Jabline
go build -o jabline .
go test ./...        # Run all 516 tests
```

**Cross-compilation:**
```bash
GOOS=linux GOARCH=amd64 go build -o jabline .
GOOS=windows GOARCH=amd64 go build -o jabline.exe .
GOOS=darwin GOARCH=arm64 go build -o jabline .
```

**WASM target:**
```bash
GOOS=wasm GOARCH=wasm go build -o jabline.wasm .
```

---

## Contributing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for guidelines.

Quick start:
1. Fork the repo
2. Create a feature branch (`git checkout -b feat/amazing`)
3. Commit your changes
4. Run `go test ./...` and `go vet ./...`
5. Open a Pull Request

---

## License

MIT &copy; 2026 Jabline Language. See [LICENSE](LICENSE).
