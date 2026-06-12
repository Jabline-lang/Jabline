package jpm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// DependencyNode represents a single node in the dependency graph.
type DependencyNode struct {
	Name       string
	Constraint string
	URL        string
	Resolved   *Version
	Deps       []*DependencyNode
}

// Resolver handles dependency graph resolution with transitive dependencies.
type Resolver struct {
	Config    *ProjectConfig
	LibDir    string
	RegIndex  RegistryIndex
	Resolved  map[string]*Version
	Graph     map[string]*DependencyNode
	Visiting  map[string]bool
}

// NewResolver creates a new resolver with the given project config.
func NewResolver(config *ProjectConfig) *Resolver {
	return &Resolver{
		Config:   config,
		LibDir:   LibDir,
		Resolved: make(map[string]*Version),
		Graph:    make(map[string]*DependencyNode),
		Visiting: make(map[string]bool),
	}
}

// ResolveAll resolves all dependencies (including transitive) to concrete versions.
func (r *Resolver) ResolveAll() error {
	for name, constraint := range r.Config.Dependencies {
		node := &DependencyNode{
			Name:       name,
			Constraint: constraint,
			URL:        constraint,
		}

		resolvedURL, err := ResolvePackage(constraint)
		if err != nil {
			return fmt.Errorf("package %q: %w", name, err)
		}
		node.URL = resolvedURL

		if err := r.resolveNode(node); err != nil {
			return fmt.Errorf("package %q: %w", name, err)
		}

		r.Graph[name] = node
	}
	return nil
}

// resolveNode resolves a dependency and its transitive dependencies.
func (r *Resolver) resolveNode(node *DependencyNode) error {
	if r.Visiting[node.Name] {
		return fmt.Errorf("circular dependency detected: %q", node.Name)
	}

	if _, ok := r.Resolved[node.Name]; ok {
		return nil
	}

	r.Visiting[node.Name] = true
	defer func() { r.Visiting[node.Name] = false }()

	var constraintStr string
	if isURL(node.Constraint) {
		constraintStr = "*"
	} else {
		constraintStr = node.Constraint
	}

	versions, err := r.fetchVersions(node.Name, node.URL)
	if err != nil {
		versions = []Version{}
	}

	if len(versions) == 0 {
		if constraintStr == "*" {
			v := Version{Major: 0, Minor: 1, Patch: 0}
			node.Resolved = &v
			r.Resolved[node.Name] = &v
			return nil
		}
		return fmt.Errorf("no versions found for %q satisfying constraint %q", node.Name, constraintStr)
	}

	c, err := ParseConstraint(constraintStr)
	if err != nil {
		return fmt.Errorf("invalid constraint %q: %w", constraintStr, err)
	}

	best, ok := BestMatch(c, versions)
	if !ok {
		return fmt.Errorf("no version of %q satisfies constraint %q", node.Name, constraintStr)
	}
	node.Resolved = &best
	r.Resolved[node.Name] = &best
	return nil
}

func (r *Resolver) fetchVersions(name, url string) ([]Version, error) {
	pkgDir := filepath.Join(r.LibDir, name)
	if _, err := os.Stat(pkgDir); err == nil {
		tags, err := listGitTags(pkgDir)
		if err == nil && len(tags) > 0 {
			return VersionsFromStrings(tags), nil
		}
	}

	tags, err := r.fetchRemoteTags(url)
	if err != nil {
		return nil, err
	}
	return VersionsFromStrings(tags), nil
}

func (r *Resolver) FetchRemoteTags(url string) ([]string, error) {
	return r.fetchRemoteTags(url)
}

func (r *Resolver) fetchRemoteTags(url string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--tags", "--sort=-v:refname", url)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	seen := make(map[string]bool)
	var tags []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		ref := parts[len(parts)-1]
		ref = strings.TrimPrefix(ref, "refs/tags/")
		ref = strings.TrimSuffix(ref, "^{}")

		if seen[ref] {
			continue
		}
		seen[ref] = true
		tags = append(tags, ref)
	}
	return tags, nil
}

// DependencyTree returns a human-readable tree of the resolved dependency graph.
func (r *Resolver) DependencyTree() string {
	var b strings.Builder
	b.WriteString("Dependency Tree:\n")

	var names []string
	for n := range r.Graph {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		node := r.Graph[name]
		r.printTree(&b, node, "  ")
	}
	return b.String()
}

func (r *Resolver) printTree(b *strings.Builder, node *DependencyNode, indent string) {
	version := "?"
	if node.Resolved != nil {
		version = node.Resolved.String()
	}
	b.WriteString(fmt.Sprintf("%s├── %s@%s [%s]\n", indent, node.Name, version, node.URL))
	for _, dep := range node.Deps {
		r.printTree(b, dep, indent+"  ")
	}
}

// ResolvedVersions returns a map of package name -> resolved version string.
func (r *Resolver) ResolvedVersions() map[string]string {
	result := make(map[string]string)
	for name, v := range r.Resolved {
		result[name] = v.String()
	}
	return result
}

// ResolvedURL returns the URL to use for downloading a dependency node.
func ResolvedURL(node *DependencyNode) string {
	if node == nil {
		return ""
	}
	if node.URL != "" {
		return node.URL
	}
	return node.Constraint
}

// ResolveTransitive resolves transitive dependencies by reading downloaded package manifests.
// Must be called after top-level dependencies have been downloaded to LibDir.
func (r *Resolver) ResolveTransitive() error {
	resolved := 0
	for {
		newDeps := make(map[string]string)
		for _, node := range r.Graph {
			pkgCfg, err := LoadPackageConfig(r.LibDir, node.Name)
			if err != nil {
				continue
			}
			for depName, depConstraint := range pkgCfg.Dependencies {
				if _, already := r.Resolved[depName]; already {
					continue
				}
				if _, queued := newDeps[depName]; queued {
					continue
				}
				resolvedURL, err := ResolvePackage(depConstraint)
				if err != nil {
					continue
				}
				newDeps[depName] = resolvedURL
			}
		}
		if len(newDeps) == 0 {
			break
		}
		for name, url := range newDeps {
			node := &DependencyNode{
				Name:       name,
				Constraint: url,
				URL:        url,
			}
			if err := r.resolveNode(node); err != nil {
				return fmt.Errorf("transitive dep %q: %w", name, err)
			}
			r.Graph[name] = node
			resolved++
		}
		if resolved > 100 {
			return fmt.Errorf("too many transitive resolution rounds (possible circular deps)")
		}
	}
	return nil
}

// listGitTags returns the list of tags from a local git repository.
func listGitTags(dir string) ([]string, error) {
	cmd := exec.Command("git", "-C", dir, "tag", "--sort=-v:refname")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(output)), nil
}
