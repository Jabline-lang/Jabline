<div align="center">
  <img src="assets/jabline.png" alt="Jabline Logo" width="120">

  # Jabline Programming Language

  *A modern, feature-rich interpreted programming language with advanced closure support*

  [![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)](https://github.com/Jabline-lang/Jabline/releases)
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

Jabline is a modern interpreted programming language designed for rapid development and system integration. It combines familiar syntax with powerful built-in capabilities and **advanced closure support**, making it ideal for functional programming, scripting, data processing, and application development.

### Core Features

- **🔒 Advanced Closures** - Full lexical scoping with variable capture and nested functions
- **📦 Module System** - Complete ES6-style import/export with barrel patterns
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
let name = "Jabline"
const VERSION = "2.0.0"

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

## 🔒 **Advanced Closure System**

Jabline features a complete closure implementation with automatic variable capture:

### Lexical Scoping
```jabline
fn outerFunction(x) {
    let outerVar = x * 2
    
    fn innerFunction() {
        return outerVar + 10  // Captures outerVar automatically
    }
    
    return innerFunction
}

let closure = outerFunction(5)
echo(closure())  // 20
```

### Factory Pattern
```jabline
fn createBankAccount(initialBalance) {
    let balance = initialBalance
    let transactionCount = 0
    
    return {
        "deposit": fn(amount) {
            balance = balance + amount
            transactionCount = transactionCount + 1
            return "Balance: $" + balance
        },
        
        "getBalance": fn() {
            return balance
        }
    }
}

let account = createBankAccount(1000)
echo(account["deposit"](250))  // Balance: $1250
```

### Event Emitters
```jabline
fn createEventEmitter() {
    let listeners = {}
    
    return {
        "on": fn(event, callback) {
            if (listeners[event] == null) {
                listeners[event] = []
            }
            listeners[event] = push(listeners[event], callback)
        },
        
        "emit": fn(event, data) {
            if (listeners[event] != null) {
                for (callback in listeners[event]) {
                    callback(data)
                }
            }
        }
    }
}
```

### Memoization
```jabline
fn memoize(fn) {
    let cache = {}
    
    return fn(arg) {
        let key = str(arg)
        if (cache[key] != null) {
            return cache[key]
        }
        let result = fn(arg)
        cache[key] = result
        return result
    }
}
```

---

## 📦 **Complete Module System**

ES6-style modules with full import/export support:

### Named Exports/Imports
```jabline
// math.jb
export fn add(a, b) { return a + b }
export fn multiply(a, b) { return a * b }
export const PI = 3.14159

// main.jb
import { add, multiply, PI } from "./math"
```

### Default Exports/Imports
```jabline
// calculator.jb
fn calculate(op, a, b) {
    // implementation
}
export default calculate

// main.jb
import calculator from "./calculator"
```

### Mixed and Aliased Imports
```jabline
import utils, { isEmpty as empty, isNumber } from "./utils"
import { multiply as mult } from "./math"
import * as mathLib from "./math"
```

### Barrel Patterns
```jabline
// index.jb
export { add, subtract } from "./math"
export { isEmpty, isNumber } from "./utils"
export { default as validator } from "./validator"
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

Jabline is built with a modular, extensible architecture featuring advanced closure support:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Source      │───▶│     Parser      │───▶│   Evaluator     │
│     Code        │    │      AST        │    │    Runtime      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Lexer       │    │   Syntax Tree   │    │    Closure      │
│    Tokens       │    │   Generation    │    │   Environment   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Components
- **Lexical Analysis** - Token generation and preprocessing
- **Syntax Parsing** - AST construction with error recovery
- **Runtime Evaluation** - Expression evaluation with closure support
- **Object System** - Dynamic typing with closure-aware functions
- **Environment System** - Advanced scoping with variable capture
- **Module System** - Import/export resolution and caching
- **Built-in Functions** - Extensive standard library integration

---

## 💼 **Use Cases**

### Functional Programming
```jabline
// Higher-order functions with closures
let compose = f => g => x => f(g(x))
let addOne = x => x + 1
let double = x => x * 2

let addOneThenDouble = compose(double)(addOne)
echo(addOneThenDouble(5))  // 12

// Partial application
let greet = greeting => name => greeting + ", " + name + "!"
let sayHello = greet("Hello")
echo(sayHello("World"))  // Hello, World!
```

### State Management
```jabline
// Redux-like state management
fn createStore(initialState) {
    let state = initialState
    let listeners = []
    
    return {
        "getState": fn() { return state },
        "dispatch": fn(action) {
            state = reducer(state, action)
            for (listener in listeners) {
                listener(state)
            }
        },
        "subscribe": fn(listener) {
            listeners = push(listeners, listener)
        }
    }
}
```

### Data Processing Pipelines
```jabline
// Functional data transformation
import { filter, map, reduce } from "./functional"

let numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
let result = numbers
    |> filter(x => x % 2 == 0)        // Even numbers
    |> map(x => x * x)                // Square them
    |> reduce((acc, x) => acc + x, 0) // Sum them

echo("Result: " + result)  // 220
```

### Module-based Architecture
```jabline
// api/users.jb
import http from "../http"
import { validate } from "../validators"

export fn createUser(userData) {
    validate(userData)
    return http.post("/users", userData)
}

export fn getUser(id) {
    return http.get("/users/" + id)
}

// main.jb
import { createUser, getUser } from "./api/users"

let newUser = createUser({"name": "Alice", "email": "alice@example.com"})
```

---

## 📊 **Performance & Features**

### Closure Performance
- **Automatic Variable Capture** - Only captures variables actually used
- **Lexical Scoping** - Efficient nested environment resolution  
- **Memory Management** - Optimized closure environment storage
- **Function Inlining** - Smart optimization for simple closures

### Module Performance
- **Module Caching** - Modules loaded once and cached
- **Circular Import Detection** - Prevents infinite import loops
- **Tree Shaking** - Only imports what's actually used
- **Relative Path Resolution** - Efficient file system operations

### General Performance
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

# Test closures
jabline run examples/closures/01_closures_guide.jb
```

---

## 📚 **Documentation**

- **[Language Reference](examples/)** - Complete syntax and feature guide
- **[Closure Examples](examples/closures/)** - Advanced closure patterns and use cases
- **[Module Examples](examples/modules/)** - Import/export patterns and best practices
- **[API Documentation](pkg/)** - Built-in function reference
- **[Contributing Guide](.github/CONTRIBUTING.md)** - Development guidelines

---

## 🆕 **What's New in v2.0.0**

### 🔒 Closures and Advanced Scoping
- **Lexical Scoping** - Variables captured from outer scopes automatically
- **Nested Functions** - Functions inside functions with full closure support
- **Variable Capture** - Smart capture of only necessary variables
- **Arrow Function Closures** - Full closure support for arrow functions
- **Closure Environment** - Advanced environment management system

### 📦 Complete Module System
- **ES6-style Imports/Exports** - `import`, `export`, `export default`
- **Named and Default Exports** - Full flexibility in module design
- **Aliased Imports** - `import { func as alias }` support
- **Namespace Imports** - `import * as name` for full module import
- **Barrel Patterns** - Re-export support for clean APIs
- **Circular Import Protection** - Prevents infinite import loops
- **Module Caching** - Efficient module loading and reuse

### 🎯 Advanced Features
- **Factory Pattern Support** - Create objects with encapsulated state
- **Event Emitter Pattern** - Built-in event system capabilities
- **Memoization Support** - Function result caching with closures
- **State Machines** - Encapsulated state management
- **Pipeline Patterns** - Functional data transformation chains
- **Higher-Order Functions** - Functions that operate on other functions

---

## 📄 **License**

Jabline is released under the [MIT License](LICENSE).

---

<div align="center">

**Jabline v2.0.0** - *Modern Programming with Advanced Closures*

Built with ❤️ using Goolang

[Get Started](https://github.com/Jabline-lang/Jabline#quick-start) • [Documentation](examples/) • [Community](https://github.com/Jabline-lang/Jabline/discussions)

</div>
