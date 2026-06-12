package jpm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LockEntry represents a single locked dependency.
type LockEntry struct {
	Version  string `json:"version"`
	URL      string `json:"url"`
	Hash     string `json:"hash"`
	Resolved string `json:"resolved"` // resolved version after constraint solving
}

// LockFile represents the project lock file (jb-lock.json).
type LockFile struct {
	Version      int                  `json:"version"` // lock file format version
	GeneratedAt  string               `json:"generated_at"`
	Dependencies map[string]LockEntry `json:"dependencies"`
}

// LoadLockFile reads the lock file from the given directory.
func LoadLockFile(dir string) (*LockFile, error) {
	path := filepath.Join(dir, LockFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}
	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}
	if lock.Dependencies == nil {
		lock.Dependencies = make(map[string]LockEntry)
	}
	return &lock, nil
}

// Save writes the lock file to the given directory.
func (lf *LockFile) Save(dir string) error {
	lf.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	lf.Version = 1
	if lf.Dependencies == nil {
		lf.Dependencies = make(map[string]LockEntry)
	}

	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	path := filepath.Join(dir, LockFileName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}
	return nil
}

// ComputeChecksum computes a SHA-256 checksum for a file.
func ComputeChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// VerifyIntegrity checks that all locked dependencies match their recorded checksums.
// Returns a list of mismatches.
func (lf *LockFile) VerifyIntegrity(libDir string) []string {
	var mismatches []string
	for name, entry := range lf.Dependencies {
		pkgDir := filepath.Join(libDir, name)
		markerFile := filepath.Join(pkgDir, ".jabline-check")
		actualHash, err := ComputeChecksum(markerFile)
		if err != nil {
			mismatches = append(mismatches, fmt.Sprintf("%s: cannot verify (%v)", name, err))
			continue
		}
		if !strings.EqualFold(actualHash, entry.Hash) {
			mismatches = append(mismatches, fmt.Sprintf("%s: hash mismatch (expected %s, got %s)", name, entry.Hash, actualHash))
		}
	}
	return mismatches
}

// UpdateFromProject creates or updates a lock file from a ProjectConfig and resolved versions.
// The resolvedVersions map is package name -> resolved version string.
func UpdateLockFile(proj *ProjectConfig, resolvedVersions map[string]string, libDir string) (*LockFile, error) {
	lock := &LockFile{
		Dependencies: make(map[string]LockEntry),
	}

	addDeps := func(deps map[string]string) {
		for name, urlOrConstraint := range deps {
			resolvedVer, ok := resolvedVersions[name]
			if !ok {
				continue
			}

			entry := LockEntry{
				URL:      urlOrConstraint,
				Version:  resolvedVer,
				Resolved: resolvedVer,
			}

			pkgDir := filepath.Join(libDir, name)
			markerFile := filepath.Join(pkgDir, ".jabline-check")
			if hash, err := ComputeChecksum(markerFile); err == nil {
				entry.Hash = hash
			}

			lock.Dependencies[name] = entry
		}
	}

	addDeps(proj.Dependencies)
	addDeps(proj.DevDependencies)

	return lock, nil
}

// LockFileCompatible checks if the existing lock file is compatible with the current config.
// Returns true if the lock file matches dependencies (same URLs/versions).
func LockFileCompatible(lock *LockFile, proj *ProjectConfig) bool {
	if lock == nil {
		return false
	}
	totalDeps := len(proj.Dependencies) + len(proj.DevDependencies)
	if len(lock.Dependencies) != totalDeps {
		return false
	}
	checkDeps := func(deps map[string]string) bool {
		for name, url := range deps {
			entry, ok := lock.Dependencies[name]
			if !ok {
				return false
			}
			if entry.URL != url {
				return false
			}
		}
		return true
	}
	return checkDeps(proj.Dependencies) && checkDeps(proj.DevDependencies)
}

// SortedDepNames returns dependency names sorted alphabetically.
func (lf *LockFile) SortedDepNames() []string {
	names := make([]string, 0, len(lf.Dependencies))
	for n := range lf.Dependencies {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
