package cmd

import (
	"fmt"
	"jabline/pkg/jpm"
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [package]",
	Short: "Add a dependency to the project",
	Long: `Downloads a dependency and adds it to your jabline.toml.

You can use a package name from the Jabline registry:
  jabline get http-router

Or a full git URL:
  jabline get https://github.com/user/repo.git`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nameOrURL := args[0]

		// 1. Load project
		config, err := jpm.LoadProject()
		if err != nil {
			fmt.Printf("Error: Not a Jabline project (jabline.toml not found or invalid). %s\n", err)
			os.Exit(1)
		}

		// 2. Resolve package name → URL via registry
		url, err := jpm.ResolvePackage(nameOrURL)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		// 3. Download dependency
		pkgName, err := jpm.DownloadDependency(url)
		if err != nil {
			fmt.Printf("Error downloading dependency: %s\n", err)
			os.Exit(1)
		}

		// 4. Update jabline.toml
		config.AddDependency(url)
		err = jpm.SaveProject(config)
		if err != nil {
			fmt.Printf("Error saving jabline.toml: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully added dependency '%s' (%s)\n", pkgName, url)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
