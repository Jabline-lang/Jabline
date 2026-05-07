package cmd

import (
	"fmt"
	"jabline/pkg/jpm"
	"os"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dependencies from jabline.toml into the lib folder",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load project
		config, err := jpm.LoadProject()
		if err != nil {
			fmt.Printf("Error: Not a Jabline project. %s\n", err)
			os.Exit(1)
		}

		if len(config.Dependencies) == 0 {
			fmt.Println("No dependencies found in jabline.toml")
			return
		}

		fmt.Printf("Installing %d dependencies...\n", len(config.Dependencies))

		for name, url := range config.Dependencies {
			fmt.Printf("Installing %s (%s)...\n", name, url)
			_, err := jpm.DownloadDependency(url)
			if err != nil {
				fmt.Printf("Error installing %s: %s\n", name, err)
				// Continue with others?
			}
		}

		fmt.Println("Dependency installation complete.")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
