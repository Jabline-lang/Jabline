# Contributing to Jabline

Thank you for considering contributing to Jabline! We welcome contributions of all kinds: bug fixes, features, documentation, tests, and examples.

## Code of Conduct

This project adheres to the [Contributor Covenant](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/Jabline.git`
3. Create a branch: `git checkout -b feat/your-feature`
4. Make your changes
5. Run tests: `go test ./...`
6. Run vet: `go vet ./...`
7. Commit and push, then open a PR

## Development Guidelines

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use 4-space indentation in Jabline source files
- Keep functions focused and small
- Add comments for non-obvious logic

### Jabline Language Changes

If you're modifying the language itself:

1. **Lexer** (`pkg/lexer/`): Add tokens for new syntax
2. **Parser** (`pkg/parser/`): Add parsing rules with tests
3. **Compiler** (`pkg/compiler/`): Generate bytecode from the new AST nodes
4. **VM** (`pkg/vm/`): Implement opcode execution
5. **Tests**: Add tests at every layer (lexer, parser, compiler, VM, integration)

### Adding Standard Library Modules

1. Create a Go file in `pkg/stdlib/` (for native functions)
2. Register it in `pkg/stdlib/modules.go`
3. Create a `.jb` wrapper in `internal/embedded/modules/`
4. Add examples in `docs/stdlib/<module>/`

### Testing

- All new code must have tests
- Run `go test ./...` before submitting
- Integration tests are in `tests/` and run with `go test ./tests/`
- Jabline test files (`*_test.jb`) run with `jabline test`

### Pull Request Process

1. Ensure all tests pass and the build is clean
2. Update documentation if needed
3. Add a changelog entry in `CHANGELOG.md`
4. The PR will be reviewed by a maintainer

## Project Structure

```
cmd/           CLI commands (cobra)
pkg/
  compiler/    Bytecode compiler
  vm/          Virtual machine (72 opcodes)
  stdlib/      35 native modules
  lsp/         Language Server Protocol
  fmt/         Code formatter + linter
  dap/         Debug Adapter Protocol
  jpm/         Package manager
  lexer/       Tokenizer
  parser/      Pratt parser
  ast/         AST definitions
  code/        Bytecode definitions
  object/      Runtime objects
  token/       Token types
  symbol/      Symbol table
  sandbox/     Security policies
docs/          388 runnable examples
examples/      20+ example programs
packages/      7 JPM packages
```

## Questions?

Open a [Discussion](https://github.com/Jabline-lang/Jabline/discussions) or ask in the [issues](https://github.com/Jabline-lang/Jabline/issues).
