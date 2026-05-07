package jpm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	LibDir = "lib"
)

// DownloadDependency downloads a remote dependency into the lib directory.
func DownloadDependency(url string) (string, error) {
	// Create lib directory if it doesn't exist
	if _, err := os.Stat(LibDir); os.IsNotExist(err) {
		err = os.Mkdir(LibDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create %s directory: %w", LibDir, err)
		}
	}

	// Derive package name from URL
	pkgName := filepath.Base(url)
	if len(pkgName) > 4 && pkgName[len(pkgName)-4:] == ".git" {
		pkgName = pkgName[:len(pkgName)-4]
	}

	destPath := filepath.Join(LibDir, pkgName)

	// If it already exists, remove it (for now, simple sync)
	if _, err := os.Stat(destPath); err == nil {
		os.RemoveAll(destPath)
	}

	fmt.Printf("Downloading %s to %s...\n", url, destPath)

	// Use git clone to download the dependency
	cmd := exec.Command("git", "clone", "--depth", "1", url, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git clone failed: %s\nOutput: %s", err, string(output))
	}

	return pkgName, nil
}
