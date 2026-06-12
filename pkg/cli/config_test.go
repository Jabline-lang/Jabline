package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Output != FormatText {
		t.Errorf("expected text output, got %s", cfg.Output)
	}
	if cfg.Color != ColorAuto {
		t.Errorf("expected auto color, got %s", cfg.Color)
	}
	if cfg.Registry.URL == "" {
		t.Error("expected registry URL to be set")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Error("expected non-empty config path")
	}
}

func TestRcPath(t *testing.T) {
	path := RcPath()
	if path == "" {
		t.Error("expected non-empty rc path")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	// Config file shouldn't exist in test environment
	cfg := LoadConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config even with missing file")
	}
	if cfg.Output != FormatText {
		t.Error("expected default output format")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Output = FormatJSON
	cfg.Color = ColorAlways
	cfg.Registry.URL = "https://example.com/registry.json"
	cfg.NoVerify = true

	tempDir := t.TempDir()
	savedPath := ConfigPath
	ConfigPath = func() string {
		return filepath.Join(tempDir, "config.toml")
	}
	defer func() { ConfigPath = savedPath }()

	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded := LoadConfig()
	if loaded.Output != FormatJSON {
		t.Errorf("expected JSON output, got %s", loaded.Output)
	}
	if loaded.Color != ColorAlways {
		t.Errorf("expected always color, got %s", loaded.Color)
	}
	if loaded.Registry.URL != "https://example.com/registry.json" {
		t.Errorf("expected custom registry URL, got %s", loaded.Registry.URL)
	}
	if !loaded.NoVerify {
		t.Error("expected no-verify to be true")
	}
}

func TestShouldUseColor(t *testing.T) {
	always := &CLIConfig{Color: ColorAlways}
	if !ShouldUseColor(always) {
		t.Error("ShouldUseColor(always) should be true")
	}

	never := &CLIConfig{Color: ColorNever}
	if ShouldUseColor(never) {
		t.Error("ShouldUseColor(never) should be false")
	}
}

func TestIsJSONOutput(t *testing.T) {
	jsonCfg := &CLIConfig{Output: FormatJSON}
	if !IsJSONOutput(jsonCfg) {
		t.Error("IsJSONOutput should be true for JSON format")
	}

	textCfg := &CLIConfig{Output: FormatText}
	if IsJSONOutput(textCfg) {
		t.Error("IsJSONOutput should be false for text format")
	}
}

func TestDetectDefaultEditor(t *testing.T) {
	editor := detectDefaultEditor()
	if editor == "" {
		t.Error("expected non-empty editor name")
	}
}

func TestSaveConfigCreatesDir(t *testing.T) {
	cfg := DefaultConfig()

	configDir := t.TempDir()
	savedPath := ConfigPath
	ConfigPath = func() string {
		return filepath.Join(configDir, "deep", "nested", "config.toml")
	}
	defer func() { ConfigPath = savedPath }()

	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(configDir, "deep", "nested", "config.toml")); os.IsNotExist(err) {
		t.Error("config file should exist")
	}
}

func TestLoadConfigBadFile(t *testing.T) {
	badPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(badPath, []byte("not valid toml {{{"), 0644); err != nil {
		t.Fatal(err)
	}

	savedPath := ConfigPath
	ConfigPath = func() string { return badPath }
	defer func() { ConfigPath = savedPath }()

	cfg := LoadConfig()
	if cfg.Output != FormatText {
		t.Error("expected default config on bad file")
	}
}
