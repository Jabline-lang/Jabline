package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	REPO_URL    = "https://github.com/Jabline-lang/Jabline"
	BINARY_NAME = "jabline"
)

const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorReset  = "\033[0m"
)

type Installer struct {
	OS             string
	Arch           string
	InstallPath    string
	BinaryName     string
	TempDir        string
	IsWindows      bool
	NeedsElevation bool
}

func main() {
	installer := &Installer{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		BinaryName: BINARY_NAME,
		IsWindows:  runtime.GOOS == "windows",
	}

	if installer.IsWindows {
		installer.BinaryName += ".exe"
	}

	printBanner()

	fmt.Printf("%s🔍 Detecting system...%s\n", ColorBlue, ColorReset)
	installer.detectSystem()

	fmt.Printf("%s📋 System Information:%s\n", ColorCyan, ColorReset)
	fmt.Printf("   OS: %s\n", installer.OS)
	fmt.Printf("   Architecture: %s\n", installer.Arch)
	fmt.Printf("   Install Path: %s\n", installer.InstallPath)
	fmt.Printf("   Binary Name: %s\n", installer.BinaryName)

	if installer.NeedsElevation {
		fmt.Printf("%s⚠️  This installation requires administrator privileges%s\n", ColorYellow, ColorReset)
	}

	fmt.Print("\nContinue with installation? (y/N): ")
	if !askForConfirmation() {
		fmt.Printf("%s❌ Installation cancelled%s\n", ColorRed, ColorReset)
		return
	}

	// Step 1: Check prerequisites
	fmt.Printf("\n%s🔧 Checking prerequisites...%s\n", ColorBlue, ColorReset)
	if err := installer.checkPrerequisites(); err != nil {
		fmt.Printf("%s❌ Prerequisites check failed: %v%s\n", ColorRed, err, ColorReset)
		return
	}

	// Step 2: Setup temporary directory
	fmt.Printf("%s📁 Setting up temporary directory...%s\n", ColorBlue, ColorReset)
	if err := installer.setupTempDir(); err != nil {
		fmt.Printf("%s❌ Failed to setup temp directory: %v%s\n", ColorRed, err, ColorReset)
		return
	}
	defer installer.cleanup()

	// Step 3: Clone repository
	fmt.Printf("%s📥 Cloning repository...%s\n", ColorBlue, ColorReset)
	if err := installer.cloneRepo(); err != nil {
		fmt.Printf("%s❌ Failed to clone repository: %v%s\n", ColorRed, err, ColorReset)
		return
	}

	// Step 4: Build binary
	fmt.Printf("%s🔨 Building binary...%s\n", ColorBlue, ColorReset)
	if err := installer.buildBinary(); err != nil {
		fmt.Printf("%s❌ Failed to build binary: %v%s\n", ColorRed, err, ColorReset)
		return
	}

	// Step 5: Install binary
	fmt.Printf("%s📦 Installing binary...%s\n", ColorBlue, ColorReset)
	if err := installer.installBinary(); err != nil {
		fmt.Printf("%s❌ Failed to install binary: %v%s\n", ColorRed, err, ColorReset)
		return
	}

	// Step 6: Verify installation
	fmt.Printf("%s✅ Verifying installation...%s\n", ColorBlue, ColorReset)
	if err := installer.verifyInstallation(); err != nil {
		fmt.Printf("%s⚠️  Installation completed but verification failed: %v%s\n", ColorYellow, err, ColorReset)
		fmt.Printf("%sYou may need to add %s to your PATH manually%s\n", ColorYellow, installer.InstallPath, ColorReset)
	}

	printSuccess(installer)
}

func printBanner() {
	fmt.Printf("%s", ColorPurple)
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    🚀 JABLINE INSTALLER 🚀                   ║")
	fmt.Println("║                                                               ║")
	fmt.Println("║        Automated installer for Jabline Programming Language  ║")
	fmt.Println("║                     Version 1.0.0                            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Printf("%s\n", ColorReset)
}

func printSuccess(installer *Installer) {
	fmt.Printf("\n%s", ColorGreen)
	fmt.Println("🎉 ═══════════════════════════════════════════════════════════")
	fmt.Println("🎉  INSTALLATION COMPLETED SUCCESSFULLY!")
	fmt.Println("🎉 ═══════════════════════════════════════════════════════════")
	fmt.Printf("%s", ColorReset)

	fmt.Printf("\n%s📍 Binary installed to: %s%s%s\n", ColorCyan, ColorWhite, installer.InstallPath, ColorReset)

	fmt.Printf("\n%s🚀 Quick Start:%s\n", ColorGreen, ColorReset)
	fmt.Printf("   %s --version    # Check installation\n", installer.BinaryName)
	fmt.Printf("   %s --help       # Get help\n", installer.BinaryName)
	fmt.Printf("   %s run file.jb  # Run a Jabline program\n", installer.BinaryName)

	fmt.Printf("\n%s💡 Next Steps:%s\n", ColorYellow, ColorReset)
	fmt.Println("   • Visit the Doc Oficial for more information")
	fmt.Println("   • Visit the GitHub repository for more information")

	fmt.Printf("\n%sHappy coding with Jabline! 💻✨%s\n\n", ColorGreen, ColorReset)
}

func (i *Installer) detectSystem() {
	switch i.OS {
	case "linux":
		if isRoot() {
			i.InstallPath = "/usr/local/bin/" + i.BinaryName
		} else {
			homeDir, _ := os.UserHomeDir()
			i.InstallPath = filepath.Join(homeDir, ".local", "bin", i.BinaryName)
			i.NeedsElevation = false
		}
	case "darwin":
		if isRoot() {
			i.InstallPath = "/usr/local/bin/" + i.BinaryName
		} else {
			homeDir, _ := os.UserHomeDir()
			i.InstallPath = filepath.Join(homeDir, "bin", i.BinaryName)
			i.NeedsElevation = false
		}
	case "windows":
		if isAdmin() {
			i.InstallPath = filepath.Join("C:", "Program Files", "Jabline", i.BinaryName)
		} else {
			homeDir, _ := os.UserHomeDir()
			i.InstallPath = filepath.Join(homeDir, "AppData", "Local", "Programs", "Jabline", i.BinaryName)
			i.NeedsElevation = false
		}
	default:
		homeDir, _ := os.UserHomeDir()
		i.InstallPath = filepath.Join(homeDir, "bin", i.BinaryName)
	}
}

func (i *Installer) checkPrerequisites() error {

	fmt.Print("   Checking Go installation... ")
	if err := exec.Command("go", "version").Run(); err != nil {
		fmt.Printf("%s❌%s\n", ColorRed, ColorReset)
		return fmt.Errorf("Go is not installed or not in PATH. Please install Go 1.21+ from https://golang.org/dl/")
	}
	fmt.Printf("%s✅%s\n", ColorGreen, ColorReset)

	fmt.Print("   Checking Git installation... ")
	if err := exec.Command("git", "--version").Run(); err != nil {
		fmt.Printf("%s❌%s\n", ColorRed, ColorReset)
		return fmt.Errorf("Git is not installed or not in PATH. Please install Git first")
	}
	fmt.Printf("%s✅%s\n", ColorGreen, ColorReset)

	fmt.Print("   Checking internet connectivity... ")
	if err := checkInternetConnection(); err != nil {
		fmt.Printf("%s❌%s\n", ColorRed, ColorReset)
		return fmt.Errorf("no internet connection: %v", err)
	}
	fmt.Printf("%s✅%s\n", ColorGreen, ColorReset)

	return nil
}

func (i *Installer) setupTempDir() error {
	tempDir, err := os.MkdirTemp("", "jabline-install-")
	if err != nil {
		return err
	}
	i.TempDir = tempDir
	fmt.Printf("   Temporary directory: %s\n", tempDir)
	return nil
}

func (i *Installer) cloneRepo() error {
	repoDir := filepath.Join(i.TempDir, "Jabline")

	cmd := exec.Command("git", "clone", "--depth", "1", REPO_URL, repoDir)
	cmd.Dir = i.TempDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %v\nOutput: %s", err, output)
	}

	fmt.Printf("   Repository cloned to: %s\n", repoDir)
	return nil
}

func (i *Installer) buildBinary() error {
	repoDir := filepath.Join(i.TempDir, "Jabline")

	ldflags := "-s -w"
	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", i.BinaryName, "main.go")
	cmd.Dir = repoDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v\nOutput: %s", err, output)
	}

	binaryPath := filepath.Join(repoDir, i.BinaryName)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary was not created at %s", binaryPath)
	}

	fmt.Printf("   Binary built successfully: %s\n", binaryPath)
	return nil
}

func (i *Installer) installBinary() error {
	sourceFile := filepath.Join(i.TempDir, "Jabline", i.BinaryName)

	destDir := filepath.Dir(i.InstallPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		if !i.IsWindows && i.NeedsElevation {
			return i.installWithElevation(sourceFile)
		}
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	if err := copyFile(sourceFile, i.InstallPath); err != nil {
		if !i.IsWindows && i.NeedsElevation {
			return i.installWithElevation(sourceFile)
		}
		return fmt.Errorf("failed to copy binary: %v", err)
	}

	if !i.IsWindows {
		if err := os.Chmod(i.InstallPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %v", err)
		}
	}

	return nil
}

func (i *Installer) installWithElevation(sourceFile string) error {
	fmt.Printf("%s🔑 Installation requires administrator privileges%s\n", ColorYellow, ColorReset)

	destDir := filepath.Dir(i.InstallPath)

	var cmd *exec.Cmd
	if i.IsWindows {

		return fmt.Errorf("please run this installer as Administrator")
	} else {

		fmt.Print("   Creating directory with sudo... ")
		cmd = exec.Command("sudo", "mkdir", "-p", destDir)
		if err := cmd.Run(); err != nil {
			fmt.Printf("%s❌%s\n", ColorRed, ColorReset)
			return fmt.Errorf("failed to create directory with sudo: %v", err)
		}
		fmt.Printf("%s✅%s\n", ColorGreen, ColorReset)

		fmt.Print("   Copying binary with sudo... ")
		cmd = exec.Command("sudo", "cp", sourceFile, i.InstallPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("%s❌%s\n", ColorRed, ColorReset)
			return fmt.Errorf("failed to copy binary with sudo: %v", err)
		}
		fmt.Printf("%s✅%s\n", ColorGreen, ColorReset)

		fmt.Print("   Setting permissions with sudo... ")
		cmd = exec.Command("sudo", "chmod", "755", i.InstallPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("%s❌%s\n", ColorRed, ColorReset)
			return fmt.Errorf("failed to set permissions with sudo: %v", err)
		}
		fmt.Printf("%s✅%s\n", ColorGreen, ColorReset)
	}

	return nil
}

func (i *Installer) verifyInstallation() error {

	if _, err := os.Stat(i.InstallPath); os.IsNotExist(err) {
		return fmt.Errorf("binary not found at %s", i.InstallPath)
	}

	cmd := exec.Command(i.InstallPath, "--version")
	output, err := cmd.Output()
	if err != nil {

		cmd = exec.Command(strings.TrimSuffix(i.BinaryName, ".exe"), "--version")
		output, err = cmd.Output()
		if err != nil {
			return fmt.Errorf("binary exists but cannot be executed")
		}
	}

	fmt.Printf("   Version check: %s", strings.TrimSpace(string(output)))
	return nil
}

func (i *Installer) cleanup() {
	if i.TempDir != "" {
		os.RemoveAll(i.TempDir)
	}
}

func askForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func isRoot() bool {
	return os.Getuid() == 0
}

func isAdmin() bool {

	return false
}

func checkInternetConnection() error {
	client := &http.Client{Timeout: 10 * time.Second}
	_, err := client.Get("https://github.com")
	return err
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
