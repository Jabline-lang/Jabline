package jpm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// RegistryURL is the default URL for the Jabline package registry index.
	// This points to a JSON file hosted on GitHub that maps package names to git URLs.
	RegistryURL = "https://raw.githubusercontent.com/Jabline-lang/registry/main/index.json"

	// CacheDir is the local directory for caching registry data.
	CacheDir = ".jb_cache"

	// CacheFile is the filename for the cached registry index.
	CacheFile = "registry.json"

	// CacheTTL is how long the cached registry is valid before re-fetching.
	CacheTTL = 1 * time.Hour
)

// RegistryEntry represents a single package in the registry.
type RegistryEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Version     string `json:"version"`
	Author      string `json:"author"`
}

// RegistryIndex is the full registry: a map of package name → entry.
type RegistryIndex map[string]RegistryEntry

// FetchRegistry downloads the registry index from the remote URL,
// using a local cache to avoid unnecessary network requests.
func FetchRegistry() (RegistryIndex, error) {
	cachePath := filepath.Join(CacheDir, CacheFile)

	// Check if cache exists and is fresh
	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < CacheTTL {
			return loadCachedRegistry(cachePath)
		}
	}

	// Fetch from remote
	fmt.Println("Fetching package registry...")
	resp, err := http.Get(RegistryURL)
	if err != nil {
		// If network fails, try to use stale cache
		if cached, cacheErr := loadCachedRegistry(cachePath); cacheErr == nil {
			fmt.Println("Warning: Using cached registry (network unavailable)")
			return cached, nil
		}
		return nil, fmt.Errorf("failed to fetch registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// If registry doesn't exist yet (404), return empty registry
		if resp.StatusCode == 404 {
			fmt.Println("Registry not found at remote URL. Using empty registry.")
			return make(RegistryIndex), nil
		}
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry response: %w", err)
	}

	// Parse the registry
	var index RegistryIndex
	if err := json.Unmarshal(body, &index); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	// Cache the result
	if err := cacheRegistry(cachePath, body); err != nil {
		fmt.Printf("Warning: Could not cache registry: %s\n", err)
	}

	return index, nil
}

// ResolvePackage looks up a package name in the registry and returns its git URL.
// If the name looks like a URL (contains "://" or ".git"), it's returned as-is.
func ResolvePackage(name string) (string, error) {
	// If it already looks like a URL, return as-is
	if isURL(name) {
		return name, nil
	}

	registry, err := FetchRegistry()
	if err != nil {
		return "", fmt.Errorf("could not resolve package '%s': %w", name, err)
	}

	entry, ok := registry[name]
	if !ok {
		return "", fmt.Errorf("package '%s' not found in registry. Use a full git URL instead", name)
	}

	return entry.URL, nil
}

// ListPackages returns all available packages in the registry.
func ListPackages() (RegistryIndex, error) {
	return FetchRegistry()
}

// isURL checks if a string looks like a URL.
func isURL(s string) bool {
	return len(s) > 8 && (s[:8] == "https://" || s[:7] == "http://" || s[:6] == "git://")
}

// loadCachedRegistry reads the registry from the local cache file.
func loadCachedRegistry(path string) (RegistryIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var index RegistryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	return index, nil
}

// cacheRegistry writes registry data to the local cache.
func cacheRegistry(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
