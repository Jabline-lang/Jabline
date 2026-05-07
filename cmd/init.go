package cmd

import (
	"fmt"
	"jabline/pkg/jpm"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new Jabline project",
	Long: `Creates a new Jabline project with a jabline.toml configuration file, 
a default main.jb, and a .gitignore. 
If no name is provided, the current directory name is used.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := ""
		targetPath := "."

		if len(args) > 0 {
			projectName = args[0]
			targetPath = projectName
		} else {
			projectName = jpm.GetDefaultProjectName()
		}

		// Check if jabline.toml already exists in target path
		configPath := filepath.Join(targetPath, jpm.ModFileName)
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("Error: %s already exists in %s\n", jpm.ModFileName, targetPath)
			os.Exit(1)
		}

		err := jpm.InitProject(projectName, targetPath)
		if err != nil {
			fmt.Printf("Error initializing project: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully initialized Jabline project: %s\n", projectName)
		if targetPath != "." {
			fmt.Printf("Project created in directory: %s\n", targetPath)
		}
		fmt.Printf("Generated files:\n")
		fmt.Printf("  - %s\n", filepath.Join(targetPath, jpm.ModFileName))
		fmt.Printf("  - %s\n", filepath.Join(targetPath, "main.jb"))
		fmt.Printf("  - %s\n", filepath.Join(targetPath, ".gitignore"))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
