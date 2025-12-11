<div align="center">
  <img src="assets/jabline.png" alt="Jabline Logo" width="120">

  # Jabline Programming Language

  *The Jabline programming language is a compiled (JBVM), modern, cloud-oriented programming language with concurrency and parallelism.*

  [![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/Jabline-lang/Jabline/releases)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  [![Go Version](https://img.shields.io/badge/go-1.25.3-00ADD8.svg)](https://golang.org/)
</div>

---

## 🚀 **Quick Start**

### Installation

```bash
# Clone the repository
git clone https://github.com/Jabline-lang/Jabline.git
cd Jabline

# Build using the automated builder
cd builder && go run .

# Run your first program
jabline run hello.jb
```

### Basic Usage

```bash
# Execute a .jb file
jabline run program.jb

# Show version
jabline --version

# Show help
jabline --help

# Activation Language Server Protocol
jabline lsp

# Compile .jb files
jabline build program.jb -o program && ./program
```

---

## 📝 Language Overview

Jabline is a modern compiled programming language, designed for rapid development and cloud integration, with a familiar syntax and powerful built-in features.

### Core Features

- **📠 Read-Evaluate-Print-Loop** - Functional and easy-to-use REPL
- **💻 Language Server Protocol** - Complete language server protocol system for your code editor
- **🔒 Advanced Closures** - Full lexical scoping with variable capture and nested functions
- **📦 Module System** - Complete ES6-style import/export with barrel patterns
- **🔧 Rich Built-in Library** - JSON, mathematics, regex, HTTP, file operations
- **🐛 Advanced Error Handling** - Colored output, stack traces, intelligent suggestions
- **📊 Native JSON Support** - Built-in parsing and serialization
- **🧮 Scientific Computing** - Complete mathematical function library
- **🌐 HTTP Client** - Integrated web request capabilities
- **📁 File System Operations** - Comprehensive file and directory management
- **⚡ Performance** - Optimized Go-based compiled

### Syntax Highlights

```jabline
// Variables and constants
let name = "Jabline"
const VERSION = "0.1.0"

// Functions with closures
fn createCounter(start) {
    let count = start

    return fn() {
        count = count + 1
        return count
    }
}

// Arrow functions with closures
let createMultiplier = factor => x => x * factor
let double = createMultiplier(2)

// Module imports/exports
import { calculate, PI } from "./math"
import utils, { isEmpty } from "./utils"
export default myFunction

// Data structures
let user = {"name": "Alice", "age": 30}
let numbers = [1, 2, 3, 4, 5]

// Control flow
if (user["age"] >= 18) {
    echo("Adult user")
} else {
    echo("Minor user")
}
```

---


## 📚 **Documentation**

- **[Language Reference](https://jabline-doc.choqlitodev.xyz/)** - Complete syntax and feature guide
- **[API Documentation](pkg/)** - Built-in function reference
- **[Contributing Guide](.github/CONTRIBUTING.md)** - Development guidelines

## 📄 **License**

Jabline is released under the [MIT License](LICENSE).

---

<div align="center">

**Jabline v0.0.1** - *Modern Programming*

Built with ❤️ using Goolang

[Get Started](https://github.com/Jabline-lang/Jabline#quick-start) • [Documentation](https://jabline-doc.choqlitodev.xyz) • [Community](https://discord.gg/invite/4FN7pA8RWm)

</div>
