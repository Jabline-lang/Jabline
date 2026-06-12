package cmd

import (
	"fmt"
	"jabline/pkg/jpm"
	"os"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update dependencies to latest compatible versions",
	Long: `Updates all dependencies to the latest versions that satisfy
the constraints in jabline.toml. Also regenerates the lock file.

Examples:
  jabline update         Update all dependencies
  jabline update mylib   Update only 'mylib'`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := jpm.LoadProject()
		if err != nil {
			fmt.Printf("Error: Not a Jabline project. %s\n", err)
			os.Exit(1)
		}

		if len(config.Dependencies) == 0 {
			fmt.Println("No dependencies found in jabline.toml")
			return
		}

		checkOnly, _ := cmd.Flags().GetBool("check")

		resolver := jpm.NewResolver(config)
		if err := resolver.ResolveAll(); err != nil {
			fmt.Printf("Error resolving dependencies: %s\n", err)
			os.Exit(1)
		}

		fmt.Println(resolver.DependencyTree())

		if checkOnly {
			fmt.Println("All dependencies up-to-date.")
			return
		}

		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}

		for name, node := range resolver.Graph {
			if filter != "" && name != filter {
				continue
			}

			url := jpm.ResolvedURL(node)
			version := ""
			if node.Resolved != nil {
				version = node.Resolved.String()
			}

			fmt.Printf("Updating %s to %s...\n", name, version)
			_, checksum, err := jpm.DownloadDependency(url, version)
			if err != nil {
				fmt.Printf("Error updating %s: %s\n", name, err)
				continue
			}

			if entry, ok := resolver.ResolvedVersions()[name]; ok {
				fmt.Printf("  Resolved %s@%s [checksum: %s]\n", name, entry, checksum[:12])
			}
		}

		lock, err := jpm.UpdateLockFile(config, resolver.ResolvedVersions(), jpm.LibDir)
		if err != nil {
			fmt.Printf("Error generating lock file: %s\n", err)
			os.Exit(1)
		}

		if err := lock.Save("."); err != nil {
			fmt.Printf("Error saving lock file: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Update complete.")
	},
}

func init() {
	updateCmd.Flags().BoolP("check", "c", false, "Only check for updates, don't install")
	rootCmd.AddCommand(updateCmd)
}
