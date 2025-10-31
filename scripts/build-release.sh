#!/bin/bash
# Build and package script for Jabline
# Usage: ./build.sh [debug|release] [platform]
# This script should be run from the /scripts directory

set -e

build_mode="release"
platform=""
project_root=".."
binary_name="jabline"
dist_dir="$project_root/dist"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        debug|release)
            build_mode="$1"
            shift
            ;;
        linux|darwin|windows)
            platform="$1"
            shift
            ;;
        *)
            echo "Error: Unknown argument '$1'"
            echo "Usage: ./build.sh [debug|release] [linux|darwin|windows]"
            exit 1
            ;;
    esac
done

# Auto-detect platform if not specified
if [ -z "$platform" ]; then
    case "$(uname -s)" in
        Linux*)     platform=linux;;
        Darwin*)    platform=darwin;;
        CYGWIN*|MINGW*|MSYS*) platform=windows;;
        *)          platform=linux;;
    esac
fi

# Set platform-specific variables
case "$platform" in
    windows)
        binary_name="jabline.exe"
        GOOS=windows
        GOARCH=amd64
        ;;
    darwin)
        GOOS=darwin
        GOARCH=amd64
        ;;
    linux)
        GOOS=linux
        GOARCH=amd64
        ;;
esac

dist_platform_dir="$dist_dir/$platform"

# Validation
if [[ "$build_mode" != "debug" && "$build_mode" != "release" ]]; then
    echo "Error: Invalid build mode '$build_mode'. Use 'debug' or 'release'."
    exit 1
fi

if [ ! -f "$project_root/go.mod" ]; then
    echo "Error: go.mod not found. Make sure this script is run from the /scripts directory."
    exit 1
fi

if [ ! -f "$project_root/main.go" ]; then
    echo "Error: main.go not found in project root."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH."
    exit 1
fi

echo "=============================================="
echo "🚀 Building Jabline Programming Language"
echo "=============================================="
echo "Build mode: $build_mode"
echo "Platform: $platform ($GOOS/$GOARCH)"
echo "Output: $dist_platform_dir/$binary_name"
echo "=============================================="

# Create distribution directory
mkdir -p "$dist_platform_dir"

pushd "$project_root" > /dev/null

# Set build flags
ldflags="-s -w"
if [ "$build_mode" = "release" ]; then
    ldflags="$ldflags -X main.version=2.0.0 -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
fi

# Build the binary
echo "Building..."
export GOOS GOARCH
if [ "$build_mode" = "release" ]; then
    go build -ldflags "$ldflags" -o "$dist_platform_dir/$binary_name" main.go
else
    go build -race -o "$dist_platform_dir/$binary_name" main.go
fi

build_success=$?
popd > /dev/null

if [ $build_success -ne 0 ]; then
    echo "❌ Build failed."
    exit 1
fi

# Verify binary was created
if [ ! -f "$dist_platform_dir/$binary_name" ]; then
    echo "❌ Binary '$binary_name' not found at $dist_platform_dir/$binary_name"
    exit 1
fi

echo "✅ Build completed successfully."

# Post-build processing for release builds on Linux
if [ "$build_mode" = "release" ] && [ "$platform" = "linux" ]; then
    # Strip binary if available
    if command -v strip &> /dev/null; then
        echo "🔧 Stripping binary..."
        if strip "$dist_platform_dir/$binary_name"; then
            echo "✅ Binary stripped successfully."
        else
            echo "⚠️  Warning: Failed to strip binary."
        fi
    fi

    # Compress with UPX if available
    if command -v upx &> /dev/null; then
        echo "🗜️  Compressing binary with UPX..."
        if upx --best "$dist_platform_dir/$binary_name"; then
            echo "✅ Binary compressed successfully."
        else
            echo "⚠️  Warning: UPX compression failed."
        fi
    fi
fi

# Show final binary info
final_size=$(du -h "$dist_platform_dir/$binary_name" | cut -f1)
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
echo "🧪 Testing binary..."
if "$dist_platform_dir/$binary_name" --version &> /dev/null || "$dist_platform_dir/$binary_name" version &> /dev/null; then
    echo "✅ Binary test passed."
else
    echo "⚠️  Warning: Binary test failed or no version command available."
fi

echo ""
echo "Ready to use! Try:"
echo "  $dist_platform_dir/$binary_name run examples/basic/01_variables_operadores.jb"
echo "=============================================="
