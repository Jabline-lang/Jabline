# Jabline for VS Code

Jabline language support for Visual Studio Code — syntax highlighting, code completion, debugging, and more.

## Features

- ✅ **Syntax highlighting** for `.jb` files (TextMate grammar)
- ✅ **Code snippets** for all Jabline keywords and constructs
- ✅ **Language Server** integration via `jabline lsp`:
  - Diagnostics (parse errors, semantic errors)
  - Code completion with keyword snippets + scope symbols
  - Hover information with type/signature display
  - Go to definition, find references
  - Rename symbol
  - Signature help
  - Document formatting
  - Document symbols (outline)
  - Code actions (quick fixes)
  - Folding ranges
  - Semantic tokens (syntax highlighting)
- ✅ **Run file** command (Ctrl+Alt+R)
- ✅ **Build file** command
- ✅ **Debug configuration** support

## Requirements

- Jabline compiler (`jabline`) must be installed and available in PATH.
- Extension requires VS Code 1.82.0+.

## Extension Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `jabline.executablePath` | `jabline` | Path to the jabline executable |
| `jabline.sandboxLevel` | `none` | Default sandbox level |
| `jabline.lsp.enabled` | `true` | Enable LSP |
| `jabline.lsp.trace.server` | `off` | Trace LSP communication |

## Commands

- `Jabline: Run File` (Ctrl+Alt+R) — Run the active `.jb` file
- `Jabline: Build File` — Build the active `.jb` file to an executable

## Known Issues

First public release.
