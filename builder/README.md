# 🔨 Jabline Builder

The automated installation system for the Jabline Programming Language.

## 🚀 Quick Start

The fastest way to install Jabline:

```bash
# Clone the repository
git clone --depth 1 https://github.com/Jabline-lang/Jabline && cd Jabline

# Run the automated builder
cd builder && go run .
```

## 📖 What This Does

The builder automatically:

- ✅ **Detects your operating system** (Linux, macOS, Windows)
- ✅ **Checks prerequisites** (Go, Git, internet connection)
- ✅ **Clones the repository** with shallow clone for speed
- ✅ **Builds optimized binary** with release flags (`-ldflags "-s -w"`)
- ✅ **Installs to appropriate location** based on your system and permissions
- ✅ **Handles permissions** automatically (uses sudo when needed on Unix)
- ✅ **Verifies installation** to ensure everything works
- ✅ **Provides beautiful output** with colors and progress indicators

## 🖥️ Installation Paths

### Linux
- **Root/Sudo**: `/usr/local/bin/jabline`
- **User**: `~/.local/bin/jabline` (creates directory if needed)

### macOS
- **Root/Sudo**: `/usr/local/bin/jabline`
- **User**: `~/bin/jabline` (creates directory if needed)

### Windows
- **Administrator**: `C:\Program Files\Jabline\jabline.exe`
- **User**: `%USERPROFILE%\AppData\Local\Programs\Jabline\jabline.exe`

## 🔧 Prerequisites

Before running the builder, ensure you have:

- **Go 1.21 or higher** - Download from [golang.org](https://golang.org/dl/)
- **Git** - For cloning the repository
- **Internet connection** - To download the source code

The builder will check these automatically and provide helpful error messages if anything is missing.

## 🛠️ Advanced Usage

### Custom Installation Path

You can override the default installation path by modifying the source or using environment variables (implementation dependent).

### Building from Local Source

If you already have the source code:

```bash
cd path/to/jabline/builder
go run .
```

### Development Build

For development purposes, you can modify the build flags in `main.go`:

```go
// Change from:
cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", i.BinaryName, "main.go")

// To (for debug builds):
cmd := exec.Command("go", "build", "-race", "-o", i.BinaryName, "main.go")
```

## 🎨 Features

### Beautiful Output
- Color-coded messages for different types of information
- Progress indicators for each installation step
- Clear success/error messaging
- System information display

### Smart Detection
- Automatic OS and architecture detection
- Permission level detection
- PATH configuration recommendations

### Error Handling
- Comprehensive prerequisite checking
- Graceful failure with helpful error messages
- Automatic cleanup of temporary files
- Detailed error reporting

### Cross-Platform Support
- Works on Linux, macOS, and Windows
- Handles platform-specific installation paths
- Manages permissions appropriately for each platform

## 🔍 Troubleshooting

### Common Issues

#### "Go is not installed or not in PATH"
**Solution:** Install Go from https://golang.org/dl/ and ensure it's in your PATH.

#### "Git is not installed or not in PATH"
**Solution:** Install Git and ensure it's accessible from the command line.

#### "No internet connection"
**Solution:** Check your internet connection and firewall settings.

#### Permission Denied
**Solution:** The installer will automatically use `sudo` on Unix systems when needed. On Windows, run as Administrator if installing system-wide.

#### Binary exists but cannot be executed
**Solution:** The installation directory may not be in your PATH. The installer will provide instructions for adding it.

### Debugging

To see more detailed output, you can modify the source to add debug logging or run with verbose Go output:

```bash
go run -x .  # Shows Go build commands
```

## 🏗️ Architecture

The builder is structured as follows:

```
builder/
├── main.go          # Main installer logic
├── go.mod           # Go module definition
└── README.md        # This file
```

### Key Components

- **Installer struct**: Holds system information and configuration
- **System detection**: Automatically determines OS, architecture, and install paths
- **Prerequisites checking**: Validates Go, Git, and internet connectivity
- **Repository cloning**: Downloads source with shallow clone for efficiency
- **Binary building**: Compiles optimized release binary
- **Installation**: Copies binary to appropriate system location
- **Verification**: Tests the installed binary

## 🤝 Contributing

To improve the builder:

1. Fork the repository
2. Make your changes to `builder/main.go`
3. Test on multiple platforms if possible
4. Submit a pull request

### Development Guidelines

- Maintain cross-platform compatibility
- Provide clear error messages
- Use appropriate colors for terminal output
- Handle edge cases gracefully
- Add comments for complex logic

## 📝 License

This builder is part of the Jabline Programming Language project and follows the same license terms.

## 🔗 Related

- Main project: [Jabline Programming Language](../../README.md)
- Installation guide: [INSTALL.md](../../INSTALL.md)
- Alternative installers: 
  - Unix shell script: [install.sh](../../install.sh)
  - Windows batch: [install.bat](../../install.bat)
  - PowerShell: [install.ps1](../../install.ps1)