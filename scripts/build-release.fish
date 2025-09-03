#!/usr/bin/env fish
# Build and package script for Jabline (Fish Shell version)
# Usage: ./build.fish [debug|release] [platform]
# This script should be run from the /scripts directory

set -e

set build_mode "release"
set platform ""
set project_root ".."
set binary_name "jabline"
set dist_dir "$project_root/dist"

# Colors for output
set RED '\033[0;31m'
set GREEN '\033[0;32m'
set YELLOW '\033[1;33m'
set BLUE '\033[0;34m'
set NC '\033[0m' # No Color

# Helper functions
function log_info
    echo -e "$BLUE🔧 $argv[1]$NC"
end

function log_success
    echo -e "$GREEN✅ $argv[1]$NC"
end

function log_warning
    echo -e "$YELLOW⚠️  $argv[1]$NC"
end

function log_error
    echo -e "$RED❌ $argv[1]$NC"
end

# Parse arguments
for arg in $argv
    switch $arg
        case debug release
            set build_mode $arg
        case linux darwin windows
            set platform $arg
        case '*'
            log_error "Unknown argument '$arg'"
            echo "Usage: ./build.fish [debug|release] [linux|darwin|windows]"
            exit 1
    end
end

# Auto-detect platform if not specified
if test -z "$platform"
    switch (uname -s)
        case Linux
            set platform linux
        case Darwin
            set platform darwin
        case 'CYGWIN*' 'MINGW*' 'MSYS*'
            set platform windows
        case '*'
            set platform linux
    end
end

# Set platform-specific variables
switch $platform
    case windows
        set binary_name "jabline.exe"
        set -x GOOS windows
        set -x GOARCH amd64
    case darwin
        set -x GOOS darwin
        set -x GOARCH amd64
    case linux
        set -x GOOS linux
        set -x GOARCH amd64
end

set dist_platform_dir "$dist_dir/$platform"

# Validation
if test "$build_mode" != "debug" -a "$build_mode" != "release"
    log_error "Invalid build mode '$build_mode'. Use 'debug' or 'release'."
    exit 1
end

if not test -f "$project_root/go.mod"
    log_error "go.mod not found. Make sure this script is run from the /scripts directory."
    exit 1
end

if not test -f "$project_root/main.go"
    log_error "main.go not found in project root."
    exit 1
end

# Check if Go is installed
if not command -v go &> /dev/null
    log_error "Go is not installed or not in PATH."
    exit 1
end

echo "=============================================="
echo "🚀 Building Jabline Programming Language"
echo "=============================================="
echo "Build mode: $build_mode"
echo "Platform: $platform ($GOOS/$GOARCH)"
echo "Output: $dist_platform_dir/$binary_name"
echo "=============================================="

# Create distribution directory
mkdir -p "$dist_platform_dir"

pushd "$project_root"

# Set build flags
set ldflags "-s -w"
if test "$build_mode" = "release"
    set ldflags "$ldflags -X main.version=2.0.0 -X main.buildTime="(date -u +%Y-%m-%dT%H:%M:%SZ)
end

# Build the binary
log_info "Building..."
if test "$build_mode" = "release"
    go build -ldflags "$ldflags" -o "$dist_platform_dir/$binary_name" main.go
else
    go build -race -o "$dist_platform_dir/$binary_name" main.go
end

set build_success $status
popd

if test $build_success -ne 0
    log_error "Build failed."
    exit 1
end

# Verify binary was created
if not test -f "$dist_platform_dir/$binary_name"
    log_error "Binary '$binary_name' not found at $dist_platform_dir/$binary_name"
    exit 1
end

log_success "Build completed successfully."

# Post-build processing for release builds on Linux
if test "$build_mode" = "release" -a "$platform" = "linux"
    # Strip binary if available
    if command -v strip &> /dev/null
        log_info "Stripping binary..."
        if strip "$dist_platform_dir/$binary_name"
            log_success "Binary stripped successfully."
        else
            log_warning "Failed to strip binary."
        end
    end

    # Compress with UPX if available
    if command -v upx &> /dev/null
        log_info "Compressing binary with UPX..."
        if upx --best "$dist_platform_dir/$binary_name"
            log_success "Binary compressed successfully."
        else
            log_warning "UPX compression failed."
        end
    end
end

# Show final binary info
set final_size (du -h "$dist_platform_dir/$binary_name" | cut -f1)
echo ""
echo "=============================================="
echo "🎉 Build Complete!"
echo "=============================================="
echo "Binary: $dist_platform_dir/$binary_name"
echo "Size: $final_size"
echo "Platform: $platform ($GOOS/$GOARCH)"
echo "Mode: $build_mode"

# Test the binary
echo ""
log_info "Testing binary..."
if "$dist_platform_dir/$binary_name" --version &> /dev/null; or "$dist_platform_dir/$binary_name" version &> /dev/null
    log_success "Binary test passed."
else
    log_warning "Binary test failed or no version command available."
end

echo ""
echo "Ready to use! Try:"
echo "  $dist_platform_dir/$binary_name run examples/basic/01_variables_operadores.jb"
echo "=============================================="
