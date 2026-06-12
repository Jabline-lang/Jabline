package cmd

import (
	"encoding/json"
	"fmt"
	"jabline/pkg/cli"
	"jabline/pkg/jpm"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for packages in the Jabline registry",
	Long: `Searches the Jabline package registry for packages matching the given query.
If no query is provided, lists all available packages.

With --json, outputs machine-readable JSON.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registry, err := jpm.ListPackages()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		if len(registry) == 0 {
			if cli.IsJSONOutput(CLICfg) {
				fmt.Println("[]")
			} else {
				fmt.Println("No packages found in the registry.")
			}
			return
		}

		query := ""
		if len(args) > 0 {
			query = strings.ToLower(args[0])
		}

		type jsonEntry struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Author      string `json:"author"`
			URL         string `json:"url"`
		}

		var results []jsonEntry
		found := 0

		for name, entry := range registry {
			if query != "" && !strings.Contains(strings.ToLower(name), query) &&
				!strings.Contains(strings.ToLower(entry.Description), query) {
				continue
			}
			found++

			if cli.IsJSONOutput(CLICfg) {
				results = append(results, jsonEntry{
					Name:        name,
					Version:     entry.Version,
					Description: entry.Description,
					Author:      entry.Author,
					URL:         entry.URL,
				})
			} else {
				fmt.Printf("  %s (v%s)\n", Colorize(name, ColorCyan), entry.Version)
				if entry.Description != "" {
					fmt.Printf("    %s\n", entry.Description)
				}
				if entry.Author != "" {
					fmt.Printf("    by %s\n", entry.Author)
				}
				if entry.Category != "" || len(entry.Tags) > 0 {
					parts := []string{}
					if entry.Category != "" {
						parts = append(parts, fmt.Sprintf("category: %s", entry.Category))
					}
					if len(entry.Tags) > 0 {
						parts = append(parts, fmt.Sprintf("tags: %s", strings.Join(entry.Tags, ", ")))
					}
					if entry.License != "" {
						parts = append(parts, fmt.Sprintf("license: %s", entry.License))
					}
					fmt.Printf("    [%s]\n", strings.Join(parts, " | "))
				}
				fmt.Println()
			}
		}

		if cli.IsJSONOutput(CLICfg) {
			data, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(data))
			return
		}

		if found == 0 {
			fmt.Printf("No packages found matching '%s'\n", query)
		} else {
			fmt.Printf("Found %d package(s)\n", found)
		}
	},
}

func init() {
	searchCmd.Flags().Bool("json", false, "Output in JSON format")
	rootCmd.AddCommand(searchCmd)
}
