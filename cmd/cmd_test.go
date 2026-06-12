package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err = root.Execute()
	return buf.String(), err
}

func TestRootCommand(t *testing.T) {
	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Fatalf("help failed: %s", err)
	}
	if !strings.Contains(output, "Jabline") {
		t.Errorf("help output should contain 'Jabline', got: %s", output)
	}
	if !strings.Contains(output, "run") {
		t.Errorf("help output should contain 'run', got: %s", output)
	}
	if !strings.Contains(output, "build") {
		t.Errorf("help output should contain 'build', got: %s", output)
	}
	if !strings.Contains(output, "repl") {
		t.Errorf("help output should contain 'repl', got: %s", output)
	}
	if !strings.Contains(output, "test") {
		t.Errorf("help output should contain 'test', got: %s", output)
	}
	if !strings.Contains(output, "debug") {
		t.Errorf("help output should contain 'debug', got: %s", output)
	}
	if !strings.Contains(output, "fmt") {
		t.Errorf("help output should contain 'fmt', got: %s", output)
	}
	if !strings.Contains(output, "lsp") {
		t.Errorf("help output should contain 'lsp', got: %s", output)
	}

	if strings.Contains(output, "completion") {
		t.Errorf("help output should NOT contain 'completion' (disabled)")
	}
}

func TestVersionCommand(t *testing.T) {
	if rootCmd.Version != "0.6.0" {
		t.Errorf("rootCmd.Version = %q, want %q", rootCmd.Version, "0.6.0")
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	tmpl := rootCmd.VersionTemplate()
	if !strings.Contains(tmpl, "version") {
		t.Errorf("version template should contain 'version', got: %s", tmpl)
	}
}

func TestUnknownFlag(t *testing.T) {
	output, err := executeCommand(rootCmd, "--nonexistent-flag")
	if err == nil {
		t.Errorf("expected error for unknown flag, got output: %s", output)
	}
	if !strings.Contains(output, "unknown flag") && !strings.Contains(output, "unknown") {
		t.Errorf("output should mention unknown flag, got: %s", output)
	}
}

func TestUnknownCommand(t *testing.T) {
	output, err := executeCommand(rootCmd, "nonexistent-cmd")
	if err == nil {
		t.Errorf("expected error for unknown command, got output: %s", output)
	}
	if !strings.Contains(output, "unknown command") {
		t.Errorf("output should mention unknown command, got: %s", output)
	}
}

func TestHasAllSubCommands(t *testing.T) {
	expectedCommands := []string{"run", "build", "test", "repl", "debug", "fmt", "lsp", "init", "get", "install", "uninstall", "publish", "search"}
	cmdMap := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		cmdMap[c.Name()] = true
	}
	for _, name := range expectedCommands {
		if !cmdMap[name] {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestSubCommandHelp(t *testing.T) {
	subCommands := []string{"run", "build", "test", "repl", "debug", "fmt", "lsp", "init", "get", "install", "uninstall", "publish", "search"}
	for _, name := range subCommands {
		t.Run(name, func(t *testing.T) {
			output, err := executeCommand(rootCmd, name, "--help")
			if err != nil {
				t.Fatalf("help for %s failed: %s", name, err)
			}
			if !strings.Contains(output, name) && !strings.Contains(output, "Usage") {
				t.Errorf("help for %q should mention the command name or 'Usage'", name)
			}
		})
	}
}
