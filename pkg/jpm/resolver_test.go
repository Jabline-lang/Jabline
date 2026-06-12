package jpm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolverBasic(t *testing.T) {
	config := &ProjectConfig{
		Project: ProjectMetadata{
			Name:    "test",
			Version: "0.1.0",
		},
		Dependencies: map[string]string{
			"mylib": "https://github.com/user/mylib.git",
		},
	}

	r := NewResolver(config)
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}

	if len(r.Resolved) != 0 {
		t.Error("expected empty resolved map initially")
	}

	if r.Config != config {
		t.Error("expected config to match")
	}
}

func TestResolverEmpty(t *testing.T) {
	config := &ProjectConfig{
		Project: ProjectMetadata{
			Name: "test",
		},
		Dependencies: map[string]string{},
	}

	r := NewResolver(config)
	if err := r.ResolveAll(); err != nil {
		t.Fatal(err)
	}

	if len(r.Graph) != 0 {
		t.Error("expected empty graph for no dependencies")
	}
}

func TestResolverCircular(t *testing.T) {
	config := &ProjectConfig{
		Project: ProjectMetadata{
			Name: "test",
		},
		Dependencies: map[string]string{
			"a": "https://github.com/user/a.git",
		},
	}

	r := NewResolver(config)

	nodeA := &DependencyNode{Name: "a", Constraint: "https://github.com/user/a.git", URL: "https://github.com/user/a.git"}
	nodeB := &DependencyNode{Name: "b", Constraint: "https://github.com/user/b.git", URL: "https://github.com/user/b.git"}
	nodeA.Deps = append(nodeA.Deps, nodeB)
	nodeB.Deps = append(nodeB.Deps, nodeA)

	r.Visiting["a"] = true
	err := r.resolveNode(nodeA)
	if err == nil {
		t.Error("expected circular dependency error")
	}
}

func TestDerivePackageName(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/user/mylib.git", "mylib"},
		{"https://github.com/user/mylib", "mylib"},
		{"git@github.com:user/repo.git", "repo"},
	}
	for _, tt := range tests {
		got := DerivePackageName(tt.url)
		if got != tt.want {
			t.Errorf("DerivePackageName(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestFindLatest(t *testing.T) {
	versions := VersionsFromStrings([]string{"1.0.0", "2.0.0", "1.5.0", "0.9.0"})
	latest := FindLatest(versions)
	if latest.String() != "2.0.0" {
		t.Errorf("FindLatest = %s, want 2.0.0", latest.String())
	}

	empty := FindLatest([]Version{})
	if empty.String() != "0.0.0" {
		t.Errorf("FindLatest(empty) = %s, want 0.0.0", empty.String())
	}
}

func TestResolverDependencyTree(t *testing.T) {
	config := &ProjectConfig{
		Project: ProjectMetadata{
			Name: "test",
		},
		Dependencies: map[string]string{
			"mylib": "https://github.com/user/mylib.git",
		},
	}

	r := NewResolver(config)
	r.Graph["mylib"] = &DependencyNode{
		Name:     "mylib",
		URL:      "https://github.com/user/mylib.git",
		Resolved: &Version{Major: 1, Minor: 0, Patch: 0},
	}

	tree := r.DependencyTree()
	if tree == "" {
		t.Error("expected non-empty dependency tree")
	}
}

func TestComputeDirChecksum(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "sub", "bar.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	checksum, err := computeDirChecksum(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(checksum) != 64 {
		t.Errorf("expected 64-char hex, got %d", len(checksum))
	}

	checksum2, err := computeDirChecksum(dir)
	if err != nil {
		t.Fatal(err)
	}
	if checksum != checksum2 {
		t.Error("same directory should produce same checksum")
	}
}

func TestVerifyPackageIntegrity(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "mypkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	checksum, err := computeDirChecksum(pkgDir)
	if err != nil {
		t.Fatal(err)
	}

	if err := saveChecksum(pkgDir, checksum); err != nil {
		t.Fatal(err)
	}

	if err := VerifyPackageIntegrity(dir, "mypkg", checksum); err != nil {
		t.Errorf("expected integrity pass: %s", err)
	}

	if err := VerifyPackageIntegrity(dir, "mypkg", "badchecksum123456789012345678901234567890123456789012345678901234567890"); err == nil {
		t.Error("expected integrity failure with bad checksum")
	}
}

func TestListInstalledPackages(t *testing.T) {
	testDir := t.TempDir()
	origDir := LibDir
	LibDir = testDir
	defer func() { LibDir = origDir }()

	pkgs, err := ListInstalledPackages()
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 0 {
		t.Error("expected no packages in empty dir")
	}

	if err := os.MkdirAll(filepath.Join(testDir, "pkg1"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(testDir, "pkg2"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(testDir, ".hidden"), 0755); err != nil {
		t.Fatal(err)
	}

	pkgs, err = ListInstalledPackages()
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 2 {
		t.Errorf("expected 2 packages, got %d", len(pkgs))
	}
}

func TestResolvedURL(t *testing.T) {
	node := &DependencyNode{
		Name: "test",
		URL:  "https://github.com/user/test.git",
	}
	if url := ResolvedURL(node); url != "https://github.com/user/test.git" {
		t.Errorf("ResolvedURL = %q, want %q", url, "https://github.com/user/test.git")
	}

	node.URL = ""
	if url := ResolvedURL(node); url != "" {
		t.Errorf("ResolvedURL with empty URL = %q, want empty", url)
	}

	if url := ResolvedURL(nil); url != "" {
		t.Errorf("ResolvedURL(nil) = %q, want empty", url)
	}
}
