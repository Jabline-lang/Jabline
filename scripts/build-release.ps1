# Build and package script for Jabline (PowerShell version)
# Usage: .\build.ps1 [-BuildMode debug|release] [-Platform linux|darwin|windows]
# This script should be run from the /scripts directory

param(
    [ValidateSet("debug", "release")]
    [string]$BuildMode = "release",

    [ValidateSet("linux", "darwin", "windows")]
    [string]$Platform = "",

    [switch]$Help
)

# Show help if requested
if ($Help) {
    Write-Host @"
Jabline Programming Language Build Script

USAGE:
    .\build.ps1 [-BuildMode <mode>] [-Platform <platform>] [-Help]

PARAMETERS:
    -BuildMode    Build mode: debug or release (default: release)
    -Platform     Target platform: linux, darwin, or windows (default: auto-detect)
    -Help         Show this help message

EXAMPLES:
    .\build.ps1                              # Build release for current platform
    .\build.ps1 -BuildMode debug             # Build debug version
    .\build.ps1 -Platform linux              # Build for Linux
    .\build.ps1 -BuildMode release -Platform windows  # Build release for Windows
"@
    exit 0
}

# Configuration
$ProjectRoot = ".."
$BinaryName = "jabline"
$DistDir = "$ProjectRoot\dist"

# Auto-detect platform if not specified
if (-not $Platform) {
    switch ([System.Environment]::OSVersion.Platform) {
        "Win32NT" { $Platform = "windows" }
        "Unix" {
            $unameOutput = uname -s
            switch ($unameOutput) {
                "Linux" { $Platform = "linux" }
                "Darwin" { $Platform = "darwin" }
                default { $Platform = "linux" }
            }
        }
        default { $Platform = "windows" }
    }
}

# Set platform-specific variables
switch ($Platform) {
    "windows" {
        $BinaryName = "jabline.exe"
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
    }
    "darwin" {
        $env:GOOS = "darwin"
        $env:GOARCH = "amd64"
    }
    "linux" {
        $env:GOOS = "linux"
        $env:GOARCH = "amd64"
    }
}

$DistPlatformDir = "$DistDir\$Platform"

# Helper functions
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "🔧 $Message" "Blue"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "✅ $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠️  $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "❌ $Message" "Red"
}

# Validation
if (-not (Test-Path "$ProjectRoot\go.mod")) {
    Write-Error "go.mod not found. Make sure this script is run from the /scripts directory."
    exit 1
}

if (-not (Test-Path "$ProjectRoot\main.go")) {
    Write-Error "main.go not found in project root."
    exit 1
}

# Check if Go is installed
try {
    $goVersion = & go version 2>$null
    if (-not $goVersion) {
        throw "Go not found"
    }
} catch {
    Write-Error "Go is not installed or not in PATH."
    exit 1
}

Write-Host "=============================================="
Write-Host "🚀 Building Jabline Programming Language"
Write-Host "=============================================="
Write-Host "Build mode: $BuildMode"
Write-Host "Platform: $Platform ($($env:GOOS)/$($env:GOARCH))"
Write-Host "Output: $DistPlatformDir\$BinaryName"
Write-Host "=============================================="

# Create distribution directory
if (-not (Test-Path $DistPlatformDir)) {
    New-Item -ItemType Directory -Path $DistPlatformDir -Force | Out-Null
}

# Change to project root
Push-Location $ProjectRoot

try {
    # Set build flags
    $ldflags = "-s -w"
    if ($BuildMode -eq "release") {
        $buildTime = Get-Date -UFormat "%Y-%m-%dT%H:%M:%SZ"
        $ldflags += " -X main.version=2.0.0 -X main.buildTime=$buildTime"
    }

    # Build the binary
    Write-Info "Building..."
    if ($BuildMode -eq "release") {
        & go build -ldflags $ldflags -o "$DistPlatformDir\$BinaryName" main.go
    } else {
        & go build -race -o "$DistPlatformDir\$BinaryName" main.go
    }

    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed."
        exit 1
    }

} finally {
    # Return to original directory
    Pop-Location
}

# Verify binary was created
if (-not (Test-Path "$DistPlatformDir\$BinaryName")) {
    Write-Error "Binary '$BinaryName' not found at $DistPlatformDir\$BinaryName"
    exit 1
}

Write-Success "Build completed successfully."

# Post-build processing for release builds
if ($BuildMode -eq "release") {
    # Check for UPX compression (Windows and Linux)
    if ($Platform -in @("windows", "linux")) {
        try {
            $upxVersion = & upx --version 2>$null
            if ($upxVersion) {
                Write-Info "Compressing binary with UPX..."
                & upx --best "$DistPlatformDir\$BinaryName" 2>$null
                if ($LASTEXITCODE -eq 0) {
                    Write-Success "Binary compressed successfully."
                } else {
                    Write-Warning "UPX compression failed."
                }
            }
        } catch {
            # UPX not available, skip compression
        }
    }
}

# Show final binary info
$binaryInfo = Get-Item "$DistPlatformDir\$BinaryName"
$sizeKB = [math]::Round($binaryInfo.Length / 1KB, 2)
$sizeMB = [math]::Round($binaryInfo.Length / 1MB, 2)

$sizeDisplay = if ($sizeMB -gt 1) { "$sizeMB MB" } else { "$sizeKB KB" }

Write-Host ""
Write-Host "=============================================="
Write-Host "🎉 Build Complete!"
Write-Host "=============================================="
Write-Host "Binary: $DistPlatformDir\$BinaryName"
Write-Host "Size: $sizeDisplay"
Write-Host "Platform: $Platform ($($env:GOOS)/$($env:GOARCH))"
Write-Host "Mode: $BuildMode"

# Test the binary
Write-Host ""
Write-Info "Testing binary..."
try {
    $testOutput = & "$DistPlatformDir\$BinaryName" --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Binary test passed."
    } else {
        $testOutput = & "$DistPlatformDir\$BinaryName" version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Binary test passed."
        } else {
            Write-Warning "Binary test failed or no version command available."
        }
    }
} catch {
    Write-Warning "Binary test failed or no version command available."
}

Write-Host ""
Write-Host "Ready to use! Try:"
if ($Platform -eq "windows") {
    Write-Host "  $DistPlatformDir\$BinaryName run examples\basic\01_variables_operadores.jb"
} else {
    Write-Host "  $DistPlatformDir/$BinaryName run examples/basic/01_variables_operadores.jb"
}
Write-Host "=============================================="

# Clean up environment variables
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
