package jpm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	LibDir      = "lib"
	ChecksumExt = ".sha256"
)

// DownloadDependency downloads a remote dependency into the lib directory.
// If a version is specified (not empty), it checks out that version after cloning.
// Returns the package name and a checksum of the downloaded content.
func DownloadDependency(url, version string) (string, string, error) {
	if _, err := os.Stat(LibDir); os.IsNotExist(err) {
		err = os.Mkdir(LibDir, 0755)
		if err != nil {
			return "", "", fmt.Errorf("failed to create %s directory: %w", LibDir, err)
		}
	}

	pkgName := derivePackageName(url)
	destPath := filepath.Join(LibDir, pkgName)

	if _, err := os.Stat(destPath); err == nil {
		os.RemoveAll(destPath)
	}

	fmt.Printf("Downloading %s to %s...\n", url, destPath)

	args := []string{"clone", "--depth", "1"}
	if version != "" && !strings.HasPrefix(version, "^") && !strings.HasPrefix(version, "~") && !strings.ContainsAny(version, "><= ") {
		args = append(args, "--branch", version)
	}
	args = append(args, url, destPath)

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("git clone failed: %s\nOutput: %s", err, string(output))
	}

	if version != "" {
		if err := checkoutVersion(destPath, version); err != nil {
			fmt.Printf("Warning: could not checkout version %s: %s\n", version, err)
		}
	}

	checksum, err := computeDirChecksum(destPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to compute checksum: %w", err)
	}

	if err := saveChecksum(destPath, checksum); err != nil {
		fmt.Printf("Warning: could not save checksum: %s\n", err)
	}

	return pkgName, checksum, nil
}

// DownloadDependencyLegacy downloads without checksum/version (backward compat).
func DownloadDependencyLegacy(url string) (string, error) {
	name, _, err := DownloadDependency(url, "")
	return name, err
}

func DerivePackageName(url string) string {
	name := filepath.Base(url)
	name = strings.TrimSuffix(name, ".git")
	return name
}

func derivePackageName(url string) string {
	return DerivePackageName(url)
}

func checkoutVersion(dir, version string) error {
	cmd := exec.Command("git", "-C", dir, "checkout", version)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %s\nOutput: %s", err, string(output))
	}
	return nil
}

func computeDirChecksum(dir string) (string, error) {
	hash := sha256.New()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, ".git") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash.Write([]byte(rel))
		hash.Write([]byte{0})
		hash.Write(data)
		hash.Write([]byte{0})
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func saveChecksum(dir, checksum string) error {
	path := filepath.Join(dir, ".jabline-check")
	return os.WriteFile(path, []byte(checksum), 0644)
}

// VerifyPackageIntegrity checks if a downloaded package matches its expected checksum.
func VerifyPackageIntegrity(libDir, pkgName, expectedChecksum string) error {
	markerFile := filepath.Join(libDir, pkgName, ".jabline-check")
	data, err := os.ReadFile(markerFile)
	if err != nil {
		return fmt.Errorf("cannot read checksum marker for %s: %w", pkgName, err)
	}
	actual := strings.TrimSpace(string(data))
	if !strings.EqualFold(actual, expectedChecksum) {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", pkgName, expectedChecksum, actual)
	}
	return nil
}

// RemovePackage deletes a package from the lib directory.
func RemovePackage(name string) error {
	destPath := filepath.Join(LibDir, name)
	if _, err := os.Stat(destPath); err == nil {
		return os.RemoveAll(destPath)
	}
	return nil
}

// ListInstalledPackages returns the names of all installed packages.
func ListInstalledPackages() ([]string, error) {
	entries, err := os.ReadDir(LibDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
