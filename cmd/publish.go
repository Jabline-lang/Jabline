package cmd

import (
	"encoding/json"
	"fmt"
	"jabline/pkg/jpm"
	"net/url"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish your package to the Jabline registry",
	Long: `Publishes the current project to the Jabline package registry so other
developers can install it with 'jabline get <name>'.

This command reads your jabline.toml, generates the registry entry, and opens
a GitHub issue on the Jabline registry repository for review. Once approved by
a maintainer, your package will appear in 'jabline search'.

Requirements:
  - A valid jabline.toml in the current directory
  - Your project must be hosted on a public git repository`,
	Run: func(cmd *cobra.Command, args []string) {
		repoURL, _ := cmd.Flags().GetString("url")

		// 1. Load project config
		config, err := jpm.LoadProject()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			fmt.Println("Make sure you're in a Jabline project directory with a valid jabline.toml")
			os.Exit(1)
		}

		name := config.Project.Name
		version := config.Project.Version
		description := config.Project.Description

		if name == "" {
			fmt.Println("Error: Project name is empty in jabline.toml")
			os.Exit(1)
		}

		// 2. Try to detect git remote URL if not provided
		if repoURL == "" {
			repoURL = detectGitRemote()
		}

		if repoURL == "" {
			fmt.Println("Error: Could not detect git remote URL.")
			fmt.Println("Please provide it with: jabline publish --url <git-url>")
			os.Exit(1)
		}

		// 3. Generate the registry entry
		entry := jpm.RegistryEntry{
			Name:        name,
			Description: description,
			URL:         repoURL,
			Version:     version,
			Author:      detectGitAuthor(),
		}

		entryJSON, _ := json.MarshalIndent(map[string]jpm.RegistryEntry{
			name: entry,
		}, "", "  ")

		// 4. Create a GitHub issue URL with pre-filled content
		issueTitle := fmt.Sprintf("[Package] %s v%s", name, version)
		issueBody := fmt.Sprintf(`## New Package Submission

**Package Name:** %s
**Version:** %s
**Description:** %s
**Repository:** %s
**Author:** %s

### Registry Entry (copy to index.json)
`+"```json\n%s\n```"+`

---
*Submitted via* `+"`jabline publish`", name, version, description, repoURL, entry.Author, string(entryJSON))

		issueURL := fmt.Sprintf(
			"https://github.com/Jabline-lang/registry/issues/new?title=%s&body=%s",
			url.QueryEscape(issueTitle),
			url.QueryEscape(issueBody),
		)

		// 5. Show summary and open browser
		fmt.Println("╔══════════════════════════════════════════╗")
		fmt.Println("║      📦 Jabline Package Publisher        ║")
		fmt.Println("╚══════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("  Package:     %s\n", name)
		fmt.Printf("  Version:     %s\n", version)
		fmt.Printf("  Description: %s\n", description)
		fmt.Printf("  Repository:  %s\n", repoURL)
		fmt.Printf("  Author:      %s\n", entry.Author)
		fmt.Println()
		fmt.Println("Opening GitHub to submit your package for review...")
		fmt.Println()

		if err := openBrowser(issueURL); err != nil {
			fmt.Println("Could not open browser automatically.")
			fmt.Println("Please open this URL manually:")
			fmt.Println(issueURL)
		} else {
			fmt.Println("A GitHub issue has been opened in your browser.")
			fmt.Println("Once a maintainer approves it, your package will be available via:")
			fmt.Printf("  jabline get %s\n", name)
		}
	},
}

func init() {
	publishCmd.Flags().StringP("url", "u", "", "Git repository URL (auto-detected if not provided)")
	rootCmd.AddCommand(publishCmd)
}

// detectGitRemote tries to get the origin remote URL from git.
func detectGitRemote() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	// Trim whitespace/newlines
	result := string(out)
	for len(result) > 0 && (result[len(result)-1] == '\n' || result[len(result)-1] == '\r') {
		result = result[:len(result)-1]
	}
	return result
}

// detectGitAuthor tries to get the git user name.
func detectGitAuthor() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return "Unknown"
	}
	result := string(out)
	for len(result) > 0 && (result[len(result)-1] == '\n' || result[len(result)-1] == '\r') {
		result = result[:len(result)-1]
	}
	if result == "" {
		return "Unknown"
	}
	return result
}

// openBrowser opens the default browser with the given URL.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
