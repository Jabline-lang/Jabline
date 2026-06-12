package jpm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLockFileRoundTrip(t *testing.T) {
	dir := t.TempDir()

	lock := &LockFile{
		Dependencies: map[string]LockEntry{
			"mylib": {
				URL:      "https://github.com/user/mylib.git",
				Version:  "^1.0.0",
				Resolved: "1.2.3",
				Hash:     "abc123",
			},
		},
	}

	if err := lock.Save(dir); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadLockFile(dir)
	if err != nil {
		t.Fatal(err)
	}

	entry, ok := loaded.Dependencies["mylib"]
	if !ok {
		t.Fatal("expected dependency 'mylib'")
	}
	if entry.URL != "https://github.com/user/mylib.git" {
		t.Errorf("URL = %q, want %q", entry.URL, "https://github.com/user/mylib.git")
	}
	if entry.Resolved != "1.2.3" {
		t.Errorf("Resolved = %q, want %q", entry.Resolved, "1.2.3")
	}
}

func TestLockFileCompatible(t *testing.T) {
	lock := &LockFile{
		Dependencies: map[string]LockEntry{
			"lib1": {URL: "https://example.com/lib1.git", Resolved: "1.0.0"},
			"lib2": {URL: "https://example.com/lib2.git", Resolved: "2.0.0"},
		},
	}
	proj := &ProjectConfig{
		Dependencies: map[string]string{
			"lib1": "https://example.com/lib1.git",
			"lib2": "https://example.com/lib2.git",
		},
	}
	if !LockFileCompatible(lock, proj) {
		t.Error("expected compatible")
	}

	// Different URL
	proj.Dependencies["lib1"] = "https://other.com/lib1.git"
	if LockFileCompatible(lock, proj) {
		t.Error("expected incompatible with different URL")
	}
}

func TestComputeChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	hash, err := ComputeChecksum(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) != 64 {
		t.Errorf("expected 64-char hex hash, got %d chars", len(hash))
	}

	// Same content = same hash
	path2 := filepath.Join(dir, "test2.txt")
	if err := os.WriteFile(path2, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	hash2, _ := ComputeChecksum(path2)
	if hash != hash2 {
		t.Error("same content should produce same hash")
	}

	// Different content = different hash
	path3 := filepath.Join(dir, "test3.txt")
	if err := os.WriteFile(path3, []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}
	hash3, _ := ComputeChecksum(path3)
	if hash == hash3 {
		t.Error("different content should produce different hash")
	}
}

func TestLockFileNotFound(t *testing.T) {
	dir := t.TempDir()
	lock, err := LoadLockFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if lock != nil {
		t.Error("expected nil for missing lock file")
	}
}
