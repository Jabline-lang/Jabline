# 🚀 Jabline Language Builder

Automated installation tool for the Jabline programming language that detects your operating system and installs the binary system-wide.

## 📋 Prerequisites

- Go 1.21 or higher
- Git (for cloning the repository)
- Administrative privileges may be required for system-wide installation

## 🛠️ Usage

### Quick Installation

```bash
# Clone the repository
git clone --depth 1 https://github.com/Jabline-lang/Jabline && cd Jabline

# Run the automated builder
cd builder && go run .
```

### What it does

The builder automatically:

1. **🔍 Detects your operating system** (Windows, macOS, Linux)
2. **🏗️ Builds the Jabline binary** using the appropriate build script
3. **📦 Installs the binary** to the system PATH
4. **✅ Verifies the installation** 

## 🖥️ Platform Support

| Platform | Script Used | Installation Path | Notes |
|----------|-------------|-------------------|--------|
| **Windows** | `build-release.bat` | `%USERPROFILE%\AppData\Local\Programs\Jabline` | May need to add to PATH manually |
| **macOS** | `build-release.sh` | `/usr/local/bin` | May require `sudo` password |
| **Linux** | `build-release.sh` | `/usr/local/bin` | May require `sudo` password |

## 🚨 Troubleshooting

### Permission Issues (macOS/Linux)

If you get permission errors, the builder will automatically attempt to use `sudo`:

```bash
⚠️ No write permissions to /usr/local/bin, attempting with sudo...
[sudo] password for user: 
✅ Binary installed to: /usr/local/bin/jabline (with sudo)
```

### Go Not Found

```
❌ Error: Environment validation failed: Go is not installed or not in PATH
```

**Solution:** Install Go from https://golang.org/dl/

### Build Script Not Found

```
❌ Error: Build failed: build script not found: ../scripts/build-release.sh
```

**Solution:** Make sure you're running from the `builder` directory in a complete Jabline repository.

## 🧪 Testing Installation

After installation, test that Jabline is working:

```bash
# Check version
jabline --version

# Run a sample program
jabline run examples/basic/01_variables_operadores.jb

# Get help
jabline --help
```

## 📁 Directory Structure

```
Jabline/
├── builder/
│   ├── main.go          # This automated installer
│   ├── go.mod          # Module definition
│   └── README.md       # This file
├── scripts/
│   ├── build-release.bat   # Windows build script
│   ├── build-release.sh    # Unix build script
│   ├── build-release.ps1   # PowerShell build script
│   └── build-release.fish  # Fish shell build script
├── dist/               # Built binaries (created during build)
│   ├── windows/
│   ├── linux/
│   └── darwin/
└── ...                # Other project files
```

## 🎯 Features

- **🤖 Automatic OS detection** - No need to specify your platform
- **🔧 Smart permission handling** - Automatically uses `sudo` when needed
- **🎨 Beautiful terminal output** - Colored progress indicators
- **⚡ Fast installation** - Optimized build process
- **🛡️ Error handling** - Clear error messages and troubleshooting tips
- **✅ Installation verification** - Confirms everything works correctly

## 🔧 Advanced Usage

### Building for a Specific Platform

The builder automatically detects your platform, but the underlying build scripts support cross-compilation:

```bash
# From the scripts directory
./build-release.sh release linux
./build-release.sh release darwin
./build-release.sh release windows
```

### Development Mode

For development builds with debug symbols:

```bash
# From the scripts directory
./build-release.sh debug
```

## 🤝 Contributing

If you want to improve the builder:

1. Make your changes to `main.go`
2. Test on different platforms
3. Update this README if needed
4. Submit a pull request

## 📝 License

This builder tool is part of the Jabline project and follows the same license terms.

---

**Happy coding with Jabline! 🎉**