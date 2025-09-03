# <img src="assets/jabline.png" alt="Logo" width="30"> 🚀 Jabline Programming Language

**A modern, powerful, and extensible interpreted programming language with a comprehensive standard library.**

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/user/jabline)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.24.6-00ADD8.svg)](https://golang.org/)

## 🌟 Overview

Jabline is a production-ready interpreted programming language that combines modern syntax with powerful built-in capabilities. Designed for rapid development, system integration, and real-world applications.

### ✨ Key Features

- **🎯 Modern Syntax**: Clean, expressive syntax inspired by JavaScript and Python
- **⚡ Fast Execution**: Optimized interpreter with efficient object system
- **📚 Rich Standard Library**: 60+ built-in functions for I/O, networking, data processing
- **🔧 System Integration**: File operations, HTTP client, environment variables
- **🧪 Built-in Testing**: Complete testing framework with assertions
- **📦 Module System**: Import/export with selective imports
- **🔒 Production Ready**: Error handling, type safety, professional features

## 🚀 Quick Start

### 🔧 Automated Installation (Recommended)

The easiest way to install Jabline with automatic system integration:

```bash
# Clone the repository
git clone --depth 1 https://github.com/Jabline-lang/Jabline && cd Jabline

# Run the automated installer
cd builder && go run .
```

### Hello World

```jabline
echo("Hello, World!");

let name = "Jabline";
echo(`Welcome to ${name} programming!`);
```

### Run Your First Program

```bash
jabline run examples/hello_world.jb
```

## 📖 Language Features

### Variables and Types

```jabline
let name = "Alice";
let age = 30;
let active = true;
let numbers = [1, 2, 3, 4, 5];
let config = {"debug": true, "port": 8080};
```

### Functions and Arrow Functions

```jabline
// Traditional function
fn greet(name) {
    return "Hello, " + name + "!";
}

// Arrow function
let double = x => x * 2;
let add = (a, b) => a + b;
```

### Control Flow

```jabline
// Conditionals
if (age >= 18) {
    echo("Adult");
} else {
    echo("Minor");
}

// Loops
for (let i = 0; i < 5; i++) {
    echo("Count: " + i);
}

// For-each
for (item in collection) {
    echo(item);
}
```

### Error Handling

```jabline
try {
    let data = readFile("config.json");
    let config = parse(data);
} catch (error) {
    echo("Error: " + error);
}
```

## 🛠️ Standard Library

### File System Operations

```jabline
// File operations
let content = readFile("document.txt");
writeFile("output.txt", "Hello World!");
let exists = fileExists("config.json");

// Directory operations
createDir("project");
let files = listDir(".");
let currentDir = getWorkingDir();
```

### HTTP Client

```jabline
// GET request
let response = httpGet("https://api.example.com/users");
if (response["status"] == 200) {
    let users = parse(response["body"]);
    echo("Found " + len(users) + " users");
}

// POST request
let data = stringify({"name": "John", "email": "john@example.com"});
let result = httpPost("https://api.example.com/users", data);
```

### Data Processing

```jabline
import { sum, sort, findIndex } from "arrays";
import { capitalize, isValidEmail } from "strings_minimal";
import { abs, max, pow } from "math";

let numbers = [15, 8, 42, 3, 20];
let total = sum(numbers);           // 88
let sorted = sort(numbers);         // [3, 8, 15, 20, 42]
let power = pow(2, 8);             // 256
```

### JSON Processing

```jabline
import { stringify, parse, isValid } from "data/json";

let data = {"name": "Alice", "age": 30, "skills": ["Go", "JavaScript"]};
let json = stringify(data);
let parsed = parse(json);
let valid = isValid(json);
```

### Testing Framework

```jabline
import { describe, it, assertEqual, assertTrue } from "testing/assert";

describe("Math Operations", function() {
    it("should add numbers correctly", function() {
        assertEqual(2 + 3, 5, "Addition should work");
    });

    it("should handle arrays", function() {
        assertTrue(len([1, 2, 3]) == 3, "Array length should be 3");
    });
});
```

### Date and Time

```jabline
import { createDate, formatDate, isLeapYear } from "time/datetime";

let today = createDate(15, 3, 2024);
echo("Today: " + formatDate(today));          // "15/03/2024"
echo("Is leap year: " + isLeapYear(2024));    // true

let timestamp = now();
let formatted = formatTime(timestamp, "YYYY-MM-DD HH:mm:ss");
```

## 📚 Documentation

- [Standard Library Guide](STDLIB_REFERENCE.md) - Complete API reference
- [Contributing Guidelines](CONTRIBUTING.md) - How to contribute
- [Examples](examples/) - Working code examples

## 🎯 Use Cases

### Web API Client

```jabline
import { stringify, parse } from "data/json";

fn fetchUserData(userId) {
    let url = `https://api.example.com/users/${userId}`;
    let response = httpGet(url);

    if (response["status"] == 200) {
        return parse(response["body"]);
    } else {
        throw "Failed to fetch user data";
    }
}

let user = fetchUserData(123);
echo("User: " + user["name"]);
```

### Configuration Management

```jabline
fn loadConfig() {
    let configFile = "app.json";

    if (!fileExists(configFile)) {
        let defaultConfig = {
            "port": 8080,
            "debug": false,
            "database_url": getEnv("DATABASE_URL") ?? "sqlite://app.db"
        };
        writeFile(configFile, stringify(defaultConfig));
        return defaultConfig;
    }

    let content = readFile(configFile);
    return parse(content);
}

let config = loadConfig();
echo("Starting server on port " + config["port"]);
```

### Data Analysis

```jabline
import { sum, sort, max, min } from "arrays";

fn analyzeData(dataset) {
    return {
        "count": len(dataset),
        "sum": sum(dataset),
        "average": sum(dataset) / len(dataset),
        "max": max(dataset),
        "min": min(dataset),
        "sorted": sort(dataset)
    };
}

let sales = [250, 180, 420, 90, 350];
let analysis = analyzeData(sales);
echo("Analysis: " + stringify(analysis));
```

## 🔧 Built-in Functions

### I/O Operations
- `readFile(filename)` - Read file contents
- `writeFile(filename, content)` - Write to file
- `fileExists(filename)` - Check file existence
- `createDir(dirname)` - Create directory
- `listDir(dirname)` - List directory contents

### Network Operations
- `httpGet(url)` - HTTP GET request
- `httpPost(url, data)` - HTTP POST request

### System Integration
- `getEnv(key)` - Get environment variable
- `setEnv(key, value)` - Set environment variable
- `getWorkingDir()` - Get current directory
- `now()` - Current timestamp
- `sleep(milliseconds)` - Pause execution

### Path Utilities
- `pathJoin(...)` - Join path segments
- `pathBase(path)` - Extract filename
- `pathDir(path)` - Extract directory

## 🌐 Standard Library Modules

- **math** - Mathematical operations and constants
- **strings_minimal** - String processing and validation
- **arrays** - Array manipulation and utilities
- **time/datetime** - Date and time operations
- **data/json** - JSON serialization and parsing
- **data/collections** - Functional programming utilities
- **testing/assert** - Testing framework with assertions
- **crypto/hash** - Hashing and security functions
- **os/env** - Environment and system utilities

## 🎨 Modern Features

### Template Literals
```jabline
let name = "World";
let greeting = `Hello ${name}!`;
let calculation = `Result: ${2 + 3}`;
```

### Nullish Coalescing
```jabline
let user = getEnv("USER") ?? "guest";
let port = config?.port ?? 3000;
```

### Optional Chaining
```jabline
let avatar = user?.profile?.avatar ?? "default.jpg";
```

## 📄 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) file for details.

---

**Jabline v1.0.0** - *From prototype to production-ready programming language*

*Built with ❤️ for modern development*
