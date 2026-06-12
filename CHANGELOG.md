# Changelog

## v0.6.0 (2026-06-11)

### Language Features
- **Optional chaining**: `a?.b`, `a?.()`, `a?.[0]` — short-circuit on null
- **Bytecode optimizer**: `OptimizeBytecode()` with constant folding, redundant instruction elimination, and pattern replacement (off by default)
- **WASM/WASI target**: `ffi_wasm.go` stub with build constraints for `GOOS=wasm`
- **Pipe operator**: `|>` — value piped as last argument to function
- **Template literals**: `` `hello ${name}` `` — string interpolation
- **Enum enhancements**: C-style integer enums with full pattern matching support
- **Defer statement**: deferred execution on scope exit
- **Select statement**: multiplex on channel operations
- **Match expression**: pattern matching with type, literal, and condition arms

### Standard Library (18 native + 31 embedded = 49 modules)
- **New native modules**: `_websocket`, `_tls`, `_compress`, `_sync`, `_xml`, `_config`, `_health`, `_log`, `_metrics`, `_parallel`, `_context`, `_resilience`, `_http_advanced`, `_image`
- **Embedded `.jb` modules** (31): `collections`, `datetime`, `crypto`, `strings`, `testing`, `validation`, `http/cookie`, `http/router`, `io/stream`, `io/csv`, `net/url`, `net/mime`, `data/xml`, `data/json/pointer`, `crypto/hash`, `crypto/cipher`, `crypto/encoding`, `math/stat`, `math/vector`, `time/zone`, `time/format`, `os/env`, `os/fs`, `sync/async`, `sync/future`, `text/template`, `text/format`, `types`, `sys/info`, `log/format`, `testing/mock`

### Tooling
- **DAP debugger** (`jabline debug`): breakpoints, evaluate, stack frames, named variables, step in/out, pause, exceptionInfo
- **Formatter**: CRLF on Windows, LF on Unix — automatic line-ending detection
- **LSP**: 15 handlers — completion, hover, diagnostics, go-to-definition, references, signature help, document symbols, workspace symbols, formatting, rename, code actions, folding, selection range, semantic tokens, document links
- **JPM packages**: `collections`, `http`, `crypto`, `strings`, `datetime`, `testing` — each with `jabline.toml` + `main.jb` + registry entry
- **Hot reload** (`--hot`): automatic re-run on file changes via `fsnotify`

### Documentation
- **388 self-contained `.jb` examples** across 16 categories in `docs/`
- Categories: getting-started, basics, functions, collections, structs, interfaces, enums, generics, error-handling, concurrency, modules, stdlib (149 files), advanced, tooling, testing, patterns
- Each example max 50 lines, runs with `jabline run docs/<category>/<file>.jb`

### Example Projects
- `examples/hello.jb` — Hello World
- `examples/fibonacci.jb` — Recursive fibonacci
- `examples/http_server.jb` — HTTP server
- `examples/concurrency.jb` — Channels and spawn
- `examples/generics.jb` — Generic functions
- `examples/structs.jb` — Structs and methods
- `examples/error_handling.jb` — Try/catch
- `examples/cli-todo/` — CLI todo manager
- `examples/chat-server/` — WebSocket chat room
- `examples/file-processor/` — Directory watcher

### Build & CI
- Integration tests: build tag removed, now run with `go test ./...`
- `lib/` populated: 32 embedded module files copied for distribution
- VSCode cleanup: migrated to `vscode-jabline/` (old `editor/vscode/` removed)
- `go.mod` updated with all dependency requirements

### Fixes
- Global mutation lost in closure calls (pre-existing — tracked as known limitation)
- Various LSP crash-on-malformed-input fixes
- Formatter edge cases with nested blocks
- Bytecode dump formatting for long constants

---

## v0.5.0 (2026-03-15)

### Language
- Generics: `<T>` syntax, constraints, generic structs and methods
- Error handling: `try`/`catch`/`finally`, `throw`, `panic`/`recover`
- Structs: methods, composition, embedding
- Interfaces: duck typing, interface composition, type assertions
- Modules: `import`/`export` with relative and module resolution
- Concurrency: `spawn`, channels, `async`/`await`
- Defer statement (initial implementation)

### Standard Library
- Core modules: `math`, `strings`, `json`, `time`, `regex`, `io`, `os`
- Network: `http` (server + client), `db` (SQLite), `redis`, `ssh`, `smtp`
- Data: `csv`, `yaml`, `encoding` (base64, hex), `template`
- Security: `jwt` (sign, verify, custom claims)
- FFI: C library loading on Windows/Unix

### Tooling
- LSP server with completion, diagnostics, hover
- Code formatter with `fmt` command
- Test runner: `assert`, `describe`/`it`, `_test.jb` convention
- Package manager: `init`, `get`, `install`, `search`, `publish`
- REPL with multiline editing
- `jabline build` with sandbox flag
- VSCode extension with syntax highlighting

### Infrastructure
- 516 unit tests across all packages
- CI workflow with Go test matrix
- Package registry hosted on GitHub Issues
- JPM package scaffolding

---

## v0.4.0 (2025-11-20)

### Language
- Struct type definitions and instantiation
- Method declarations with receiver syntax
- `while` and `for` loops
- Array and hash literal syntax
- Closures and anonymous functions

### Virtual Machine
- 45 opcodes with stack-based execution
- Closure support with free variable capture
- Array/hash operations
- Builtin function dispatch

### Tooling
- Basic CLI with `run`, `repl`, `fmt` commands
- Initial LSP scaffolding
- Simple formatter
