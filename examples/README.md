# 📚 Jabline Examples

This directory contains comprehensive examples demonstrating Jabline's features and standard library capabilities.

## 🚀 Getting Started

### Quick Start
```bash
# Basic language features
./jabline run examples/basic/01_variables_operadores.jb

# Standard library I/O functions
./jabline run examples/stdlib_io_demo.jb

# Complete standard library demonstration
./jabline run examples/stdlib_complete_demo.jb
```

## 📁 Directory Structure

### `/basic/` - Core Language Features
Examples of fundamental Jabline language constructs:
- Variables and operators
- Control flow (if/else, loops)
- Functions and arrow functions
- Data structures (arrays, hash maps, structs)
- Error handling
- Modern operators (??, ?.)

### `/modules/` - Module System Examples
Demonstrations of Jabline's module system:
- Basic module usage
- Import/export patterns
- Selective imports
- Module composition

### `/advanced/` - Advanced Features
Complex examples showcasing:
- Template literals
- Advanced function patterns
- Complex data processing
- Real-world application patterns

### `/modern/` - Modern Language Features
Examples of cutting-edge Jabline features:
- Arrow functions
- Template literals with interpolation
- Modern operators
- Async patterns (if available)

## 🎯 Featured Examples

### Standard Library Demonstrations

#### `stdlib_io_demo.jb` - I/O Built-ins Test
Comprehensive test of all I/O built-in functions:
- File system operations
- HTTP requests
- Environment variables
- Path utilities
- Time functions

**Run:** `./jabline run examples/stdlib_io_demo.jb`

#### `stdlib_complete_demo.jb` - Complete Integration
Real-world application demonstrating:
- Mathematical operations
- String processing
- Array manipulation
- File operations
- Data analysis
- User management simulation

**Run:** `./jabline run examples/stdlib_complete_demo.jb`

#### `stdlib_quickstart.jb` - Quick Overview
Fast introduction to standard library features:
- DateTime operations
- JSON processing  
- Testing framework
- Built-in function usage

**Run:** `./jabline run examples/stdlib_quickstart.jb`

### Module System Examples

#### Basic Module Usage
```bash
./jabline run examples/modules/01_basic_modules.jb
```
Demonstrates importing and using the core modules:
- Math operations
- String processing
- Array manipulation

## 🧪 Testing Examples

All examples include error handling and demonstrate best practices:

```jabline
// Error handling pattern
try {
    let result = riskyOperation();
    echo("Success: " + result);
} catch (error) {
    echo("Error: " + error);
}
```

## 💡 Usage Patterns

### Configuration Management
```jabline
import { stringify, parse } from "data/json";

fn loadConfig() {
    if (!fileExists("config.json")) {
        let defaultConfig = {"port": 8080, "debug": false};
        writeFile("config.json", stringify(defaultConfig));
        return defaultConfig;
    }
    return parse(readFile("config.json"));
}
```

### Data Processing Pipeline
```jabline
import { sum, sort, max } from "arrays";
import { capitalize } from "strings_minimal";

let processed = data
    .map(item => processItem(item))
    .filter(item => item.valid)
    .sort();
```

### HTTP API Integration
```jabline
fn fetchUserData(id) {
    let response = httpGet(`https://api.example.com/users/${id}`);
    if (response.status == 200) {
        return parse(response.body);
    }
    throw "Failed to fetch user data";
}
```

## 📊 Example Categories

| Category | Count | Description |
|----------|-------|-------------|
| Basic | 20+ | Core language features |
| Modules | 5+ | Module system usage |
| Advanced | 10+ | Complex patterns |
| Modern | 8+ | Latest language features |
| Standard Library | 3 | Built-in function demos |

## 🎓 Learning Path

### Beginner
1. Start with `basic/01_variables_operadores.jb`
2. Progress through basic examples
3. Try `stdlib_quickstart.jb`

### Intermediate
1. Explore module examples
2. Run `stdlib_complete_demo.jb`
3. Study advanced patterns

### Advanced
1. Review modern language features
2. Build your own applications
3. Contribute new examples

## 🔧 Requirements

- Jabline interpreter (build with `go build -o jabline main.go`)
- Go 1.19+ (for building from source)
- Internet connection (for HTTP examples)

## 🤝 Contributing

To add new examples:

1. Place files in appropriate subdirectory
2. Include comprehensive comments
3. Demonstrate error handling
4. Add entry to this README
5. Test thoroughly

## 📖 Documentation

- [Language Reference](../README.md)
- [Standard Library Reference](../STDLIB_REFERENCE.md)
- [Contributing Guidelines](../CONTRIBUTING.md)

---

**Happy coding with Jabline! 🚀**

*All examples are tested and ready to run*