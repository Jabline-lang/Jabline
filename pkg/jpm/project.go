package jpm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	ModFileName    = "jabline.toml"
	LockFileName   = "jabline.lock"
	DefaultVersion = "0.1.0"
)

type ProjectConfig struct {
	Project      ProjectMetadata   `toml:"project"`
	Dependencies map[string]string `toml:"dependencies"`
}

type ProjectMetadata struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
}

// InitProject creates a new Jabline project with default files in the specified path.
func InitProject(projectName string, targetPath string) error {
	// 1. Create target directory if it doesn't exist
	if targetPath != "" && targetPath != "." {
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	config := ProjectConfig{
		Project: ProjectMetadata{
			Name:        projectName,
			Version:     DefaultVersion,
			Description: "A new Jabline project",
		},
		Dependencies: make(map[string]string),
	}

	// 2. Save jabline.toml
	configPath := filepath.Join(targetPath, ModFileName)
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", ModFileName, err)
	}

	// 3. Create main.jb
	mainPath := filepath.Join(targetPath, "main.jb")
	mainContent := `echo("Hello World!");
`
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		err = os.WriteFile(mainPath, []byte(mainContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to create main.jb: %w", err)
		}
	}

	// 4. Create .gitignore
	gitignorePath := filepath.Join(targetPath, ".gitignore")
	gitignoreContent := `# Jabline binaries
*.exe
jabline
jabline_debug

# Dependency directory
lib/

# Local cache
.jb_cache/
`
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		err = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	return nil
}

func LoadProject() (*ProjectConfig, error) {
	data, err := os.ReadFile(ModFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", ModFileName, err)
	}

	var config ProjectConfig
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", ModFileName, err)
	}

	if config.Dependencies == nil {
		config.Dependencies = make(map[string]string)
	}

	return &config, nil
}

func SaveProject(config *ProjectConfig) error {
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	err = os.WriteFile(ModFileName, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", ModFileName, err)
	}

	return nil
}

// GetProjectName returns the name of the current directory if not specified.
func GetDefaultProjectName() string {
	dir, err := os.Getwd()
	if err != nil {
		return "my-project"
	}
	return filepath.Base(dir)
}

// AddDependency adds a dependency to the project config.
func (c *ProjectConfig) AddDependency(url string) {
	if c.Dependencies == nil {
		c.Dependencies = make(map[string]string)
	}
	// Use the last part of the URL as the package name for now
	name := filepath.Base(url)
	name = strings.TrimSuffix(name, ".git")
	c.Dependencies[name] = url
}

// RemoveDependency removes a dependency by name.
func (c *ProjectConfig) RemoveDependency(name string) bool {
	if c.Dependencies == nil {
		return false
	}
	if _, ok := c.Dependencies[name]; ok {
		delete(c.Dependencies, name)
		return true
	}
	return false
}
