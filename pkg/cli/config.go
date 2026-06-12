package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pelletier/go-toml/v2"
)

// ConfigPath returns the path to the user's CLI config file.
// This is a variable so tests can override it.
var ConfigPath = func() string {
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), ".config")
		}
		return filepath.Join(base, "jabline", "config.toml")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Preferences", "jabline", "config.toml")
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "jabline", "config.toml")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "jabline", "config.toml")
	}
}

// RcPath returns the path to the user's .jablinerc file.
func RcPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jablinerc")
}

// OutputFormat represents the format for CLI output.
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// ColorMode represents color output preference.
type ColorMode string

const (
	ColorAuto  ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever ColorMode = "never"
)

// CLIConfig holds all CLI configuration values.
type CLIConfig struct {
	Output   OutputFormat `toml:"output"`
	Color    ColorMode    `toml:"color"`
	NoVerify bool         `toml:"no-verify"`
	Verbose  bool         `toml:"verbose"`

	Registry struct {
		URL      string `toml:"url"`
		CacheTTL int    `toml:"cache-ttl"`
	} `toml:"registry"`

	Editor struct {
		Command string `toml:"command"`
		Args    string `toml:"args"`
	} `toml:"editor"`

	Build struct {
		Output string `toml:"output"`
		Tags   string `toml:"tags"`
	} `toml:"build"`
}

// DefaultConfig returns the default CLI configuration.
func DefaultConfig() *CLIConfig {
	cfg := &CLIConfig{
		Output: FormatText,
		Color:  ColorAuto,
	}
	cfg.Registry.URL = "https://raw.githubusercontent.com/Jabline-lang/registry/main/index.json"
	cfg.Registry.CacheTTL = 60
	cfg.Editor.Command = detectDefaultEditor()
	return cfg
}

func detectDefaultEditor() string {
	switch runtime.GOOS {
	case "windows":
		return "notepad"
	case "darwin":
		return "open -W -n"
	default:
		if ed := os.Getenv("EDITOR"); ed != "" {
			return ed
		}
		return "vi"
	}
}

// LoadConfig loads the CLI configuration from the user's config file.
// If the file doesn't exist, returns the default config without error.
func LoadConfig() *CLIConfig {
	cfg := DefaultConfig()

	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var fileCfg CLIConfig
	if err := toml.Unmarshal(data, &fileCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: invalid config file %s: %s\n", path, err)
		return cfg
	}

	// Merge: only override non-zero values
	if fileCfg.Output != "" {
		cfg.Output = fileCfg.Output
	}
	if fileCfg.Color != "" {
		cfg.Color = fileCfg.Color
	}
	cfg.NoVerify = fileCfg.NoVerify
	cfg.Verbose = fileCfg.Verbose
	if fileCfg.Registry.URL != "" {
		cfg.Registry.URL = fileCfg.Registry.URL
	}
	if fileCfg.Registry.CacheTTL > 0 {
		cfg.Registry.CacheTTL = fileCfg.Registry.CacheTTL
	}
	if fileCfg.Editor.Command != "" {
		cfg.Editor.Command = fileCfg.Editor.Command
	}
	if fileCfg.Editor.Args != "" {
		cfg.Editor.Args = fileCfg.Editor.Args
	}
	if fileCfg.Build.Output != "" {
		cfg.Build.Output = fileCfg.Build.Output
	}
	if fileCfg.Build.Tags != "" {
		cfg.Build.Tags = fileCfg.Build.Tags
	}

	return cfg
}

// SaveConfig writes the CLI configuration to the user's config file.
func SaveConfig(cfg *CLIConfig) error {
	path := ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// ShouldUseColor determines if colored output should be used based on config and terminal.
func ShouldUseColor(cfg *CLIConfig) bool {
	switch cfg.Color {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default:
		// Auto: check if stdout is a terminal
		stat, _ := os.Stdout.Stat()
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
}

// IsJSONOutput returns true if the output format is JSON.
func IsJSONOutput(cfg *CLIConfig) bool {
	return cfg.Output == FormatJSON
}
