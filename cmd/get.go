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
  jabline get https://github.com/user/repo.git

With --save-dev, marks as a development dependency.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nameOrURL := args[0]

		config, err := jpm.LoadProject()
		if err != nil {
			fmt.Printf("Error: Not a Jabline project (jabline.toml not found or invalid). %s\n", err)
			os.Exit(1)
		}

		saveDev, _ := cmd.Flags().GetBool("save-dev")

		url, err := jpm.ResolvePackage(nameOrURL)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		pkgName := jpm.DerivePackageName(url)

		resolver := jpm.NewResolver(config)
		if err := resolver.ResolveAll(); err != nil {
			fmt.Printf("Warning: could not resolve all dependencies: %s\n", err)
		}

		bestVersion := ""
		if node, ok := resolver.Graph[pkgName]; ok && node.Resolved != nil {
			bestVersion = node.Resolved.String()
		}
		if bestVersion == "" {
			tags, err := resolver.FetchRemoteTags(url)
			if err == nil && len(tags) > 0 {
				versions := jpm.VersionsFromStrings(tags)
				best := jpm.FindLatest(versions)
				bestVersion = best.String()
			}
		}

		pkgName, checksum, err := jpm.DownloadDependency(url, bestVersion)
		if err != nil {
			fmt.Printf("Error downloading dependency: %s\n", err)
			os.Exit(1)
		}

		if saveDev {
			config.AddDevDependency(url)
		} else {
			config.AddDependency(url)
		}

		err = jpm.SaveProject(config)
		if err != nil {
			fmt.Printf("Error saving jabline.toml: %s\n", err)
			os.Exit(1)
		}

		resolvedVersions := make(map[string]string)
		resolvedVersions[pkgName] = bestVersion
		for name, v := range resolver.ResolvedVersions() {
			resolvedVersions[name] = v
		}

		lock, err := jpm.UpdateLockFile(config, resolvedVersions, jpm.LibDir)
		if err != nil {
			fmt.Printf("Warning: could not generate lock file: %s\n", err)
		} else {
			if err := lock.Save("."); err != nil {
				fmt.Printf("Warning: could not save lock file: %s\n", err)
			}
		}

		fmt.Printf("Successfully added dependency '%s' (%s)\n", pkgName, url)
		if bestVersion != "" {
			fmt.Printf("  Resolved version: %s\n", bestVersion)
		}
		if checksum != "" {
			fmt.Printf("  Checksum: %s\n", checksum[:16])
		}
	},
}

func init() {
	getCmd.Flags().Bool("save-dev", false, "Save as a development dependency")
	rootCmd.AddCommand(getCmd)
}
