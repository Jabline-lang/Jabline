<div align="center">
  <img src="assets/jabline.png" alt="Jabline Logo" width="120">

  # Jabline Programming Language

  *A modern, feature-rich interpreted programming language*

  [![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/Jabline-lang/Jabline/releases)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  [![Go Version](https://img.shields.io/badge/go-1.24.6-00ADD8.svg)](https://golang.org/)
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
```

---

## 📝 **Language Overview**

Jabline is a modern interpreted programming language designed for rapid development and system integration. It combines familiar syntax with powerful built-in capabilities, making it ideal for scripting, data processing, and application development.

### Core Features

- **🔧 Rich Built-in Library** - JSON, mathematics, regex, HTTP, file operations
- **🐛 Advanced Error Handling** - Colored output, stack traces, intelligent suggestions
- **📊 Native JSON Support** - Built-in parsing and serialization
- **🧮 Scientific Computing** - Complete mathematical function library
- **🌐 HTTP Client** - Integrated web request capabilities
- **📁 File System Operations** - Comprehensive file and directory management
- **⚡ Performance** - Optimized Go-based interpreter

### Syntax Highlights

```jabline
// Variables and constants
let name = "Jabline";
const VERSION = "1.0.0";

// Functions
fn calculate(x, y) {
    return sqrt(pow(x, 2) + pow(y, 2));
}

// Data structures
let user = {"name": "Alice", "age": 30};
let numbers = [1, 2, 3, 4, 5];

// Control flow
if (user["age"] >= 18) {
    echo("Adult user");
} else {
    echo("Minor user");
}

// JSON operations
let json = stringify(user);
let parsed = parse(json);

// Mathematical operations
let result = abs(-15) + sqrt(25) + pow(2, 8);
```

---

## 🔧 **Built-in Capabilities**

### Data Operations
- **JSON Processing** - `stringify()`, `parse()` for seamless data conversion
- **String Manipulation** - Case conversion, splitting, joining, pattern matching
- **Array Operations** - Push, pop, sorting, filtering, transformation
- **Hash Operations** - Key-value manipulation, merging, iteration

### Mathematical Computing
- **Basic Math** - `abs()`, `sqrt()`, `pow()`, `min()`, `max()`
- **Trigonometry** - `sin()`, `cos()`, `tan()`, `asin()`, `acos()`, `atan()`
- **Logarithms** - `log()`, `log10()`, `log2()`, `exp()`
- **Utilities** - `floor()`, `ceil()`, `round()`, `factorial()`, `random()`
- **Constants** - `PI()`, `E()` for mathematical precision

### Pattern Matching
- **Email Validation** - `isEmail()` for RFC-compliant email checking
- **URL Validation** - `isURL()` for web address verification  
- **Phone Validation** - `isPhone()` for telephone number formats
- **Custom Patterns** - `test()`, `match()`, `replace()` for regex operations

### System Integration
- **File Operations** - Read, write, create, delete files and directories
- **HTTP Requests** - GET and POST operations with automatic response parsing
- **Environment** - Access and modify environment variables
- **Input/Output** - Console interaction and formatted output

### Development Tools
- **Debug Output** - `debug()` with colored, formatted messages
- **Assertions** - `assert()` with visual feedback for testing
- **Execution Tracing** - `trace()` for code flow analysis
- **Stack Inspection** - `stackTrace()` for error diagnosis

---

## 🏗️ **Architecture**

Jabline is built with a modular, extensible architecture:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Source      │───▶│     Parser      │───▶│   Evaluator     │
│     Code        │    │      AST        │    │    Runtime      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Lexer       │    │   Syntax Tree   │    │    Built-ins    │
│    Tokens       │    │   Generation    │    │   Functions     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Components
- **Lexical Analysis** - Token generation and preprocessing
- **Syntax Parsing** - AST construction with error recovery
- **Runtime Evaluation** - Expression evaluation and execution
- **Object System** - Dynamic typing with primitive and composite types
- **Built-in Functions** - Extensive standard library integration

---

## 💼 **Use Cases**

### Data Processing
```jabline
// Process JSON data from APIs
let response = httpGet("https://api.example.com/users");
let users = parse(response["body"]);

for (user in users) {
    if (isEmail(user["email"])) {
        echo("Valid user: " + user["name"]);
    }
}
```

### Mathematical Computing
```jabline
// Scientific calculations
fn calculateDistance(x1, y1, x2, y2) {
    let dx = x2 - x1;
    let dy = y2 - y1;
    return sqrt(pow(dx, 2) + pow(dy, 2));
}

let distance = calculateDistance(0, 0, 3, 4); // 5.0
```

### System Automation
```jabline
// File processing with validation
let files = listDir("./data");
for (file in files) {
    if (fileExists(file) && endsWith(file, ".json")) {
        let content = readFile(file);
        let data = parse(content);
        debug("Processed file:", file);
    }
}
```

---

## 📊 **Performance & Reliability**

- **Optimized Runtime** - Built on Go for high performance and memory efficiency
- **Error Recovery** - Graceful handling of syntax and runtime errors
- **Memory Management** - Automatic garbage collection with minimal overhead
- **Cross Platform** - Runs on Windows, macOS, and Linux
- **Production Ready** - Comprehensive error reporting and debugging tools

---

## 🤝 **Contributing**

We welcome contributions to make Jabline better:

1. **Fork** the repository
2. **Create** a feature branch
3. **Add** your improvements with tests
4. **Submit** a pull request

### Development Setup
```bash
# Install Go 1.24.6 or later
# Clone and build
go build -o jabline main.go

# Run tests
go test ./...
```

---

## 📚 **Documentation**

- **[Language Reference](examples/)** - Complete syntax and feature guide
- **[API Documentation](pkg/)** - Built-in function reference
- **[Examples](examples/)** - Code samples and tutorials
- **[Contributing Guide](.github/CONTRIBUTING.md)** - Development guidelines

---

## 📄 **License**

Jabline is released under the [MIT License](LICENSE).

---

<div align="center">

**Jabline v1.0.0** - *Modern Programming Made Simple*

Built with ❤️ using Goolang

[Get Started](https://github.com/Jabline-lang/Jabline#quick-start) • [Documentation](examples/) • [Community](https://github.com/Jabline-lang/Jabline/discussions)

</div>
