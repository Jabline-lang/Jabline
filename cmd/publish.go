package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"jabline/pkg/jpm"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	registryOwner = "Jabline-lang"
	registryRepo  = "registry"
	registryPath  = "index.json"
)

type ghContent struct {
	SHA     string `json:"sha"`
	Content string `json:"content"`
}

var publishCmd = &cobra.Command{
	Use:   "publish --url <repo-url>",
	Short: "Register your package in the Jabline registry",
	Long: `Registers your package in the Jabline package catalog (Jabline-lang/registry).

Your package code stays in your own GitHub repository. This command only adds
your package name and URL to the registry so others can install it.

Usage:
  jabline publish --url https://github.com/tu-usuario/tu-paquete

You will be prompted for a GitHub token once (stored in ~/.jabline/config.toml).
Get a token at: https://github.com/settings/tokens (scope: public_repo)`,
	Run: func(cmd *cobra.Command, args []string) {
		repoURL, _ := cmd.Flags().GetString("url")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		tokenFlag, _ := cmd.Flags().GetString("token")

		if repoURL == "" {
			// Try to detect from current git repo
			repoURL = detectGitRemote()
		}

		if repoURL == "" {
			fmt.Println("Usage: jabline publish --url <https://github.com/tu-usuario/tu-paquete>")
			fmt.Println("Your package code stays in your own repository.")
			fmt.Println("This command only registers it in the Jabline registry.")
			os.Exit(1)
		}

		// Read jabline.toml if available (optional)
		config, err := jpm.LoadProject()
		name := filepath.Base(repoURL)
		version := "0.1.0"
		description := ""
		if strings.HasSuffix(name, ".git") {
			name = name[:len(name)-4]
		}

		if err == nil && config.Project.Name != "" {
			name = config.Project.Name
			version = config.Project.Version
			description = config.Project.Description
		}

		author := detectGitAuthor()

		fmt.Println("╔══════════════════════════════════════════╗")
		fmt.Println("║      Jabline Package Publisher          ║")
		fmt.Println("╚══════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("  Package:     %s\n", Colorize(name, ColorCyan))
		fmt.Printf("  Version:     %s\n", version)
		fmt.Printf("  Description: %s\n", description)
		fmt.Printf("  Repository:  %s\n", repoURL)
		fmt.Printf("  Author:      %s\n", author)
		fmt.Println()

		entry := jpm.RegistryEntry{
			Name:        name,
			Description: description,
			URL:         repoURL,
			Version:     version,
			Author:      author,
			Repository:  repoURL,
		}

		fmt.Println()

		if dryRun {
			entryJSON, _ := json.MarshalIndent(entry, "", "  ")
			fmt.Println("📦 Package entry (dry-run):")
			fmt.Println(string(entryJSON))
			fmt.Println()
			fmt.Println("Run without --dry-run to submit.")
			return
		}

		token := resolveToken(tokenFlag)
		submitViaAPI(token, entry)
	},
}

func init() {
	publishCmd.Flags().StringP("url", "u", "", "GitHub URL of your package repository (required)")
	publishCmd.Flags().StringP("token", "t", "", "GitHub token (or set GITHUB_TOKEN env var)")
	publishCmd.Flags().Bool("dry-run", false, "Show what would be submitted without publishing")
	rootCmd.AddCommand(publishCmd)
}

func resolveToken(flagToken string) string {
	if flagToken != "" {
		return flagToken
	}
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	if t := os.Getenv("GH_TOKEN"); t != "" {
		return t
	}

	home, err := os.UserHomeDir()
	if err == nil {
		cfgPath := filepath.Join(home, ".jabline", "config.toml")
		if data, err := os.ReadFile(cfgPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "github_token") || strings.HasPrefix(line, "GITHUB_TOKEN") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						t := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
						if t != "" {
							return t
						}
					}
				}
			}
		}
	}

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        GitHub Token Required             ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("To publish packages, you need a GitHub token.")
	fmt.Println("Get one at: https://github.com/settings/tokens")
	fmt.Println("(Required scope: public_repo)")
	fmt.Println()
	fmt.Print("Enter your GitHub token: ")
	var input string
	fmt.Scanln(&input)
	token := strings.TrimSpace(input)
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: token is required")
		os.Exit(1)
	}

	home, _ = os.UserHomeDir()
	if home != "" {
		cfgDir := filepath.Join(home, ".jabline")
		os.MkdirAll(cfgDir, 0700)
		cfgPath := filepath.Join(cfgDir, "config.toml")
		existing := ""
		if data, err := os.ReadFile(cfgPath); err == nil {
			existing = string(data)
		}
		newConfig := "github_token = \"" + token + "\"\n"
		if strings.Contains(existing, "github_token") {
			lines := strings.Split(existing, "\n")
			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "github_token") {
					lines[i] = newConfig
				}
			}
			os.WriteFile(cfgPath, []byte(strings.Join(lines, "\n")), 0600)
		} else {
			f, _ := os.OpenFile(cfgPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if f != nil {
				f.WriteString(newConfig)
				f.Close()
			}
		}
		fmt.Println("✅ Token saved to ~/.jabline/config.toml")
	}

	return token
}

func submitViaAPI(token string, entry jpm.RegistryEntry) {
	fmt.Println("🤖 Publishing to Jabline registry...")

	client := &http.Client{Timeout: 30 * time.Second}
	authHeader := "Bearer " + token

	// 1. Get current index.json and its SHA
	getURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", registryOwner, registryRepo, registryPath)
	req, _ := http.NewRequest("GET", getURL, nil)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not reach GitHub API: %s\n", err)
		fmt.Fprintln(os.Stderr, "Check your token or network connection.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		fmt.Fprintln(os.Stderr, "Error: registry repository not found.")
		fmt.Fprintln(os.Stderr, "Expected at: https://github.com/Jabline-lang/registry")
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Error: GitHub API returned %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var current ghContent
	json.NewDecoder(resp.Body).Decode(&current)

	// 2. Decode current index.json, add/update entry
	decoded, _ := base64.StdEncoding.DecodeString(current.Content)
	var registry jpm.RegistryIndex
	json.Unmarshal(decoded, &registry)
	registry[entry.Name] = entry

	updated, _ := json.MarshalIndent(registry, "", "  ")
	encoded := base64.StdEncoding.EncodeToString(updated)

	// 3. Commit directly to main
	updatePayload, _ := json.Marshal(map[string]string{
		"message": fmt.Sprintf("Add package %s v%s", entry.Name, entry.Version),
		"content": encoded,
		"sha":     current.SHA,
		"branch":  "main",
	})

	req, _ = http.NewRequest("PUT", getURL, bytes.NewReader(updatePayload))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Error: could not update registry: %s\n", string(body))
		os.Exit(1)
	}
	resp.Body.Close()

	fmt.Printf("✅ Package %s v%s registered!\n", Colorize(entry.Name, ColorCyan), entry.Version)
	fmt.Printf("📦 Now available via: jabline get %s\n", entry.Name)
	fmt.Printf("🔍 Search:            jabline search %s\n", entry.Name)
}

func detectGitRemote() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	result := string(out)
	for len(result) > 0 && (result[len(result)-1] == '\n' || result[len(result)-1] == '\r') {
		result = result[:len(result)-1]
	}
	return result
}

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
