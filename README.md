<div align="center">
  <img src="assets/jabline.png" alt="Jabline Logo" width="120">

  # Jabline

  A compiled, cloud-native programming language with a custom bytecode virtual machine.

  [![Version](https://img.shields.io/badge/version-v0.6.0-blue.svg)](https://github.com/Jabline-lang/Jabline/releases)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  [![Built with](https://img.shields.io/badge/built%20with-Go-00ADD8.svg)](https://golang.org/)

</div>

---

## Overview

Jabline is a general-purpose programming language that compiles to bytecode and runs on its own stack-based virtual machine (JBVM). It combines the syntax familiarity of JavaScript/TypeScript with Go-style concurrency and a built-in toolchain.

**What makes it different:**
- Compiles to bytecode instead of interpreting an AST — faster execution than tree-walking interpreters.
- Native concurrency via `spawn` and channels (CSP model).
- Ships with an LSP server, code formatter, test runner, and package manager out of the box.
- Generics support with `<T>` syntax.

---

## Installation

Requires [Go 1.21+](https://golang.org/dl/).

```bash
git clone https://github.com/Jabline-lang/Jabline.git
cd Jabline
go build -o jabline .
```

On Windows:
```powershell
go build -o jabline.exe .
```

Optionally, move the binary to your system PATH.

---

## CLI

```
jabline run <file.jb>           Run a Jabline script
jabline run -e "<code>"         Execute inline code
jabline run <file> --hot        Run with hot-reloading
jabline run <file> --bytecode   Dump compiled bytecode
jabline run <file> --ast        Print the AST

jabline build <file.jb>         Compile to binary
jabline test                    Run *_test.jb files
jabline fmt <path>              Format source code
jabline debug <file.jb>         Start the debugger

jabline init [name]             Create a new project
jabline get <package>           Add a dependency
jabline install                 Install all dependencies
jabline search [query]          Search the package registry
jabline publish                 Publish a package to the registry

jabline lsp                     Start the Language Server
jabline repl                    Start interactive mode
jabline --version               Print version
```

---

## Language Syntax

### Variables

```javascript
let name = "Jabline";
const PI = 3.14159;

// Type annotations are optional
let count: int = 0;
let label: string = "hello";
```

### Functions

```javascript
fn add(a, b) {
    return a + b;
}

// With type annotations
fn greet(name: string): string {
    return "Hello, " + name;
}

// Arrow functions
let double = (x) => x * 2;

// Closures
fn counter() {
    let n = 0;
    return fn() {
        n = n + 1;
        return n;
    };
}
```

### Generics

```javascript
fn identity<T>(value: T): T {
    return value;
}

struct Pair<T, U> {
    first: T,
    second: U
}

let items = Array<string>();
```

### Structs and Methods

```javascript
struct User {
    name: string,
    age: int
}

fn (u User) greet() {
    echo("Hi, I'm " + u.name);
}

let user = User { name: "Alice", age: 30 };
user.greet();
```

### Control Flow

```javascript
// If / else
if (x > 10) {
    echo("big");
} else if (x > 5) {
    echo("medium");
} else {
    echo("small");
}

// For loop
for (let i = 0; i < 10; i++) {
    echo(i);
}

// For-in
for item in items {
    echo(item);
}

// While
while (running) {
    process();
}

// Switch
switch (value) {
    case 1: echo("one");
    case 2: echo("two");
    default: echo("other");
}

// Try / catch
try {
    riskyOperation();
} catch (err) {
    echo("Error: " + err);
}
```

### Concurrency

```javascript
// Spawn a concurrent task
let ch = make_chan();

spawn fn() {
    let result = heavyComputation();
    send(ch, result);
}();

let value = recv(ch);

// Async / await
async fn fetchData() {
    let response = await httpGet("/api/data");
    return response;
}
```

### Modules

```javascript
// Standard library imports
import * as http from "std/net/http";
import * as json from "std/encoding/json";
import * as crypto from "std/crypto";

// Named imports
import { parse, stringify } from "std/encoding/json";

// Local imports
import * as utils from "./utils";
```

### Other Features

```javascript
// Pipe operator
let result = data |> transform |> format;

// Optional chaining
let city = user?.address?.city;

// Nullish coalescing
let name = user.name ?? "Anonymous";

// Template literals
let msg = `Hello ${name}, you have ${count} items`;

// Ternary
let label = x > 0 ? "positive" : "negative";
```

---

## Type System

Jabline supports the following types:

| Type | Description |
|------|-------------|
| `int` | Default integer (64-bit) |
| `int8`, `int16`, `int32`, `int64` | Signed integers |
| `uint8`, `uint16`, `uint32`, `uint64` | Unsigned integers |
| `float32`, `float64` | Floating point |
| `string` | Text |
| `bool` | `true` / `false` |
| `null` | Null value |
| `Array` | Ordered collection |
| `Hash` | Key-value map |

Type annotations are optional. When provided, they serve as documentation and enable LSP features.

---

## Standard Library

| Module | Description |
|--------|-------------|
| `std/net/http` | HTTP server, router, request handling |
| `std/db` | SQLite database operations |
| `std/encoding/json` | JSON parse and stringify |
| `std/encoding/base64` | Base64 encoding/decoding |
| `std/encoding/hex` | Hex encoding/decoding |
| `std/crypto` | SHA-256, MD5, AES, random bytes, UUID |
| `std/time/datetime` | Date and time operations |
| `std/sys/io` | File system read/write |
| `std/data/collections` | Stack, Queue, and other data structures |
| `std/math` | Math constants and functions |
| `std/strings` | String manipulation utilities |

---

## Package Manager (JPM)

Jabline has a built-in package manager that uses a central registry hosted on GitHub.

**Using packages:**
```bash
jabline search router        # Find packages
jabline get http-router      # Install by name
jabline get https://github.com/user/repo.git  # Install by URL
jabline install              # Install all from jabline.toml
```

**Publishing packages:**
```bash
jabline init my-library      # Create a project
# ... write your code ...
jabline publish              # Submit to the registry
```

The `publish` command reads your `jabline.toml`, detects your git remote, and opens a GitHub issue on the [Jabline Registry](https://github.com/Jabline-lang/registry) for review.

Dependencies are stored in `lib/` and declared in `jabline.toml`:
```toml
[project]
name = "my-app"
version = "0.1.0"
description = "My Jabline application"

[dependencies]
http-router = "https://github.com/Jabline-lang/http-router.git"
```

---

## Editor Support

Jabline includes a VS Code extension with full LSP integration:

- Syntax highlighting
- Autocompletion
- Hover documentation
- Error diagnostics
- Code formatting on save
- Go-to-definition

The extension is located in `editor/vscode/`. It automatically detects `jabline.exe` in the workspace root.

---

## Architecture

```
Source Code (.jb)
       │
       ▼
    Lexer        → Tokenizes source into tokens
       │
       ▼
    Parser       → Builds Abstract Syntax Tree (Pratt parser)
       │
       ▼
   Compiler      → Generates bytecode with constant folding
       │
       ▼
     JBVM        → Stack-based virtual machine executes bytecode
```

The compiler performs:
- **Constant folding**: `10 * 3600` compiles to `OpConstant(36000)`.
- **Dead code elimination**: Unreachable branches are pruned.
- **Symbol resolution**: Variables are resolved at compile time.

---

## Project Structure

```
├── cmd/           CLI commands (run, build, fmt, test, lsp, get, publish...)
├── pkg/
│   ├── lexer/     Tokenizer
│   ├── parser/    Pratt parser and AST builder
│   ├── ast/       AST node definitions
│   ├── compiler/  Bytecode compiler
│   ├── vm/        Virtual machine
│   ├── object/    Runtime object system
│   ├── stdlib/    Built-in functions
│   ├── lsp/       Language Server Protocol implementation
│   ├── fmt/       Code formatter
│   ├── jpm/       Package manager and registry
│   ├── token/     Token types and keywords
│   ├── code/      Bytecode instruction set
│   └── symbol/    Symbol table
├── editor/        VS Code extension
├── examples/      Example programs
├── main.go        Entry point
└── go.mod
```

---

## License

Jabline is released under the [MIT License](LICENSE).
