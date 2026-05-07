package cmd

import (
	"fmt"
	"jabline/pkg/jpm"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [package name]",
	Short: "Remove a dependency from the project",
	Long:  `Removes a dependency from your jabline.toml and deletes its source code from the lib directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// 1. Load project
		config, err := jpm.LoadProject()
		if err != nil {
			fmt.Printf("Error: Not a Jabline project. %s\n", err)
			os.Exit(1)
		}

		// 2. Remove from config
		if !config.RemoveDependency(name) {
			fmt.Printf("Error: Dependency '%s' not found in jabline.toml\n", name)
			os.Exit(1)
		}

		// 3. Save jabline.toml
		err = jpm.SaveProject(config)
		if err != nil {
			fmt.Printf("Error saving jabline.toml: %s\n", err)
			os.Exit(1)
		}

		// 4. Delete from lib/
		destPath := filepath.Join(jpm.LibDir, name)
		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf("Deleting %s...\n", destPath)
			err = os.RemoveAll(destPath)
			if err != nil {
				fmt.Printf("Warning: Could not delete directory %s: %s\n", destPath, err)
			}
		}

		fmt.Printf("Successfully uninstalled dependency '%s'\n", name)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
