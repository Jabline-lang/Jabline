# Change Log

## [0.2.0] - 2026-05-22

### Added

- Full LSP code intelligence (24 handlers): hover, completion, signature help, go to definition, references, rename, code actions, folding, semantic tokens, workspace symbols
- Code Lens (reference counts on functions/structs/interfaces/enums/services)
- Call Hierarchy (prepare/incoming/outgoing calls)
- Inlay Hints (type hints on variables, parameter name hints on call arguments)
- Document formatting via `jabline lsp` (TextDocumentFormatting)
- 22 code snippets for all Jabline constructs
- Fixed PowerShell compatibility for Run/Build commands (no quotes around executable)
- `.gitattributes` for proper GitHub language detection

## [0.1.0] - 2026-05-22

### Added

- Initial release
- Syntax highlighting for `.jb` files
- Language Server Protocol integration via `jabline lsp`
- Code snippets for all Jabline keywords and constructs
- Run and Build commands for Jabline files
- Debug configuration support
