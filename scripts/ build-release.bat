@echo off
REM Build and package script for Jabline (Windows Batch version)
REM Usage: build.bat [debug|release] [platform]
REM This script should be run from the /scripts directory

setlocal enabledelayedexpansion

set "build_mode=release"
set "platform="
set "project_root=.."
set "binary_name=jabline"
set "dist_dir=%project_root%\dist"

REM Parse arguments
:parse_args
if "%~1"=="" goto :end_parse
if "%~1"=="debug" (
    set "build_mode=debug"
    shift
    goto :parse_args
)
if "%~1"=="release" (
    set "build_mode=release"
    shift
    goto :parse_args
)
if "%~1"=="linux" (
    set "platform=linux"
    shift
    goto :parse_args
)
if "%~1"=="darwin" (
    set "platform=darwin"
    shift
    goto :parse_args
)
if "%~1"=="windows" (
    set "platform=windows"
    shift
    goto :parse_args
)
echo Error: Unknown argument '%~1'
echo Usage: build.bat [debug^|release] [linux^|darwin^|windows]
exit /b 1

:end_parse

REM Auto-detect platform if not specified
if "%platform%"=="" (
    set "platform=windows"
)

REM Set platform-specific variables
if "%platform%"=="windows" (
    set "binary_name=jabline.exe"
    set "GOOS=windows"
    set "GOARCH=amd64"
) else if "%platform%"=="darwin" (
    set "GOOS=darwin"
    set "GOARCH=amd64"
) else if "%platform%"=="linux" (
    set "GOOS=linux"
    set "GOARCH=amd64"
)

set "dist_platform_dir=%dist_dir%\%platform%"

REM Validation
if not "%build_mode%"=="debug" if not "%build_mode%"=="release" (
    echo Error: Invalid build mode '%build_mode%'. Use 'debug' or 'release'.
    exit /b 1
)

if not exist "%project_root%\go.mod" (
    echo Error: go.mod not found. Make sure this script is run from the /scripts directory.
    exit /b 1
)

if not exist "%project_root%\main.go" (
    echo Error: main.go not found in project root.
    exit /b 1
)

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo Error: Go is not installed or not in PATH.
    exit /b 1
)

echo ==============================================
echo 🚀 Building Jabline Programming Language
echo ==============================================
echo Build mode: %build_mode%
echo Platform: %platform% (%GOOS%/%GOARCH%)
echo Output: %dist_platform_dir%\%binary_name%
echo ==============================================

REM Create distribution directory
if not exist "%dist_platform_dir%" mkdir "%dist_platform_dir%"

pushd "%project_root%"

REM Set build flags
set "ldflags=-s -w"
if "%build_mode%"=="release" (
    for /f %%i in ('powershell -command "Get-Date -UFormat %%Y-%%m-%%dT%%H:%%M:%%SZ"') do set "build_time=%%i"
    set "ldflags=!ldflags! -X main.version=2.0.0 -X main.buildTime=!build_time!"
)

REM Build the binary
echo 🔧 Building...
if "%build_mode%"=="release" (
    go build -ldflags "%ldflags%" -o "%dist_platform_dir%\%binary_name%" main.go
) else (
    go build -race -o "%dist_platform_dir%\%binary_name%" main.go
)

set "build_success=%errorlevel%"
popd

if not "%build_success%"=="0" (
    echo ❌ Build failed.
    exit /b 1
)

REM Verify binary was created
if not exist "%dist_platform_dir%\%binary_name%" (
    echo ❌ Binary '%binary_name%' not found at %dist_platform_dir%\%binary_name%
    exit /b 1
)

echo ✅ Build completed successfully.

REM Post-build processing for release builds on Windows
if "%build_mode%"=="release" if "%platform%"=="windows" (
    REM Check for UPX compression
    upx --version >nul 2>&1
    if not errorlevel 1 (
        echo 🗜️ Compressing binary with UPX...
        upx --best "%dist_platform_dir%\%binary_name%" >nul 2>&1
        if not errorlevel 1 (
            echo ✅ Binary compressed successfully.
        ) else (
            echo ⚠️ Warning: UPX compression failed.
        )
    )
)

REM Show final binary info
for %%A in ("%dist_platform_dir%\%binary_name%") do set "file_size=%%~zA"
set /a "size_kb=%file_size% / 1024"

echo.
echo ==============================================
echo 🎉 Build Complete!
echo ==============================================
echo Binary: %dist_platform_dir%\%binary_name%
echo Size: %size_kb% KB
echo Platform: %platform% (%GOOS%/%GOARCH%)
echo Mode: %build_mode%

REM Test the binary
echo.
echo 🧪 Testing binary...
"%dist_platform_dir%\%binary_name%" --version >nul 2>&1
if not errorlevel 1 (
    echo ✅ Binary test passed.
) else (
    "%dist_platform_dir%\%binary_name%" version >nul 2>&1
    if not errorlevel 1 (
        echo ✅ Binary test passed.
    ) else (
        echo ⚠️ Warning: Binary test failed or no version command available.
    )
)

echo.
echo Ready to use! Try:
echo   %dist_platform_dir%\%binary_name% run examples\basic\01_variables_operadores.jb
echo ==============================================

endlocal
