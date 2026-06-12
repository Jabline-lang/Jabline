package cmd

import (
	"fmt"
	"jabline/pkg/jpm"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	initTemplate string
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new Jabline project",
	Long: `Creates a new Jabline project with a jabline.toml configuration file, 
a default main.jb, and a .gitignore. 
If no name is provided, the current directory name is used.

Templates:
  hello         Basic Hello World (default)
  http-server   HTTP server with health checks and Prometheus metrics
  api           REST API with CORS middleware
  worker        Concurrent worker pool with backpressure
  microservice  Production-ready microservice
`,
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

		configPath := filepath.Join(targetPath, jpm.ModFileName)
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("Error: %s already exists in %s\n", jpm.ModFileName, targetPath)
			os.Exit(1)
		}

		err := jpm.InitProject(projectName, targetPath, initTemplate)
		if err != nil {
			fmt.Printf("Error initializing project: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully initialized Jabline project: %s\n", projectName)
		fmt.Printf("Template: %s\n", initTemplate)
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
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "hello", "Project template (hello, http-server, api, worker, microservice)")
	rootCmd.AddCommand(initCmd)
}
