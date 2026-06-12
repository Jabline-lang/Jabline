package cmd

import (
	"fmt"
	"jabline/pkg/cli"
	"os"

	"github.com/spf13/cobra"
)

var (
	// AppVersion is the current version of Jabline.
	AppVersion = "0.6.0"

	// Global CLI config
	CLICfg *cli.CLIConfig

	// Persistent flags
	flagJSON    bool
	flagColor   string
	flagVerbose bool
)

var rootCmd = &cobra.Command{
	Use:   "jabline",
	Short: "Jabline - A compiled, cloud-native programming language",
	Long: `Jabline is a simple and easy-to-use VM-based programming language.

This is the command-line interpreter for Jabline that allows you to:
- Execute .jb code files
- Explore language features
- Manage packages with the built-in package manager

To start, try running a file:
  jabline run my_file.jb`,
	Version: AppVersion,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		CLICfg = cli.LoadConfig()

		if flagJSON {
			CLICfg.Output = cli.FormatJSON
		}
		switch flagColor {
		case "always":
			CLICfg.Color = cli.ColorAlways
		case "never":
			CLICfg.Color = cli.ColorNever
		case "auto", "":
			CLICfg.Color = cli.ColorAuto
		default:
			return fmt.Errorf("invalid color mode %q (use auto, always, or never)", flagColor)
		}
		CLICfg.Verbose = flagVerbose
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(`{{printf "%s version %s\n" .Name .Version}}`)

	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringVar(&flagColor, "color", "auto", "Color output mode (auto, always, never)")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose output")

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add config command
	rootCmd.AddCommand(configCmd)
}

// configCmd manages CLI configuration.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a CLI configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cli.LoadConfig()
		return displayConfigValue(cfg, args[0])
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a CLI configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cli.LoadConfig()
		key, val := args[0], args[1]
		if err := setConfigValue(cfg, key, val); err != nil {
			return err
		}
		if err := cli.SaveConfig(cfg); err != nil {
			return fmt.Errorf("error saving config: %s", err)
		}
		fmt.Printf("Set %s = %s\n", key, val)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all CLI configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := cli.LoadConfig()
		fmt.Printf("output      = %s\n", cfg.Output)
		fmt.Printf("color       = %s\n", cfg.Color)
		fmt.Printf("no-verify   = %v\n", cfg.NoVerify)
		fmt.Printf("verbose     = %v\n", cfg.Verbose)
		fmt.Printf("registry.url      = %s\n", cfg.Registry.URL)
		fmt.Printf("registry.cache-ttl = %d\n", cfg.Registry.CacheTTL)
		fmt.Printf("editor.command    = %s\n", cfg.Editor.Command)
		fmt.Printf("editor.args       = %s\n", cfg.Editor.Args)
		fmt.Printf("build.output      = %s\n", cfg.Build.Output)
		fmt.Printf("build.tags        = %s\n", cfg.Build.Tags)
		fmt.Printf("\nConfig path: %s\n", cli.ConfigPath())
	},
}

func displayConfigValue(cfg *cli.CLIConfig, key string) error {
	switch key {
	case "output":
		fmt.Println(cfg.Output)
	case "color":
		fmt.Println(cfg.Color)
	case "no-verify":
		fmt.Println(cfg.NoVerify)
	case "verbose":
		fmt.Println(cfg.Verbose)
	case "registry.url":
		fmt.Println(cfg.Registry.URL)
	case "registry.cache-ttl":
		fmt.Println(cfg.Registry.CacheTTL)
	case "editor.command":
		fmt.Println(cfg.Editor.Command)
	case "editor.args":
		fmt.Println(cfg.Editor.Args)
	case "build.output":
		fmt.Println(cfg.Build.Output)
	case "build.tags":
		fmt.Println(cfg.Build.Tags)
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func setConfigValue(cfg *cli.CLIConfig, key, val string) error {
	switch key {
	case "output":
		if val != "text" && val != "json" {
			return fmt.Errorf("invalid output format %q (use text or json)", val)
		}
		cfg.Output = cli.OutputFormat(val)
	case "color":
		if val != "auto" && val != "always" && val != "never" {
			return fmt.Errorf("invalid color mode %q (use auto, always, or never)", val)
		}
		cfg.Color = cli.ColorMode(val)
	case "no-verify":
		cfg.NoVerify = val == "true" || val == "yes" || val == "1"
	case "verbose":
		cfg.Verbose = val == "true" || val == "yes" || val == "1"
	case "registry.url":
		cfg.Registry.URL = val
	case "registry.cache-ttl":
		fmt.Sscanf(val, "%d", &cfg.Registry.CacheTTL)
	case "editor.command":
		cfg.Editor.Command = val
	case "editor.args":
		cfg.Editor.Args = val
	case "build.output":
		cfg.Build.Output = val
	case "build.tags":
		cfg.Build.Tags = val
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

// Colorize returns a colorized string if the config allows color output.
func Colorize(text, colorCode string) string {
	if CLICfg != nil && cli.ShouldUseColor(CLICfg) {
		return fmt.Sprintf("\033[%sm%s\033[0m", colorCode, text)
	}
	return text
}

// Color constants
const (
	ColorRed    = "31"
	ColorGreen  = "32"
	ColorYellow = "33"
	ColorBlue   = "34"
	ColorCyan   = "36"
	ColorBold   = "1"
)
