package cmd

import (
	"fmt"
	"jabline/pkg/jpm"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for packages in the Jabline registry",
	Long: `Searches the Jabline package registry for packages matching the given query.
If no query is provided, lists all available packages.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registry, err := jpm.ListPackages()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		if len(registry) == 0 {
			fmt.Println("No packages found in the registry.")
			return
		}

		query := ""
		if len(args) > 0 {
			query = strings.ToLower(args[0])
		}

		found := 0
		for name, entry := range registry {
			if query != "" && !strings.Contains(strings.ToLower(name), query) &&
				!strings.Contains(strings.ToLower(entry.Description), query) {
				continue
			}
			found++
			fmt.Printf("  %s (v%s)\n", name, entry.Version)
			if entry.Description != "" {
				fmt.Printf("    %s\n", entry.Description)
			}
			if entry.Author != "" {
				fmt.Printf("    by %s\n", entry.Author)
			}
			fmt.Println()
		}

		if found == 0 {
			fmt.Printf("No packages found matching '%s'\n", query)
		} else {
			fmt.Printf("Found %d package(s)\n", found)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
