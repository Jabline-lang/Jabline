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
	Long: `Installs all project dependencies specified in jabline.toml.

With --lock, also generates a jabline.lock file with resolved versions.
With --frozen-lockfile, fails if the lock file is not up-to-date.`,
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

		lockOnly, _ := cmd.Flags().GetBool("lock")
		frozenLock, _ := cmd.Flags().GetBool("frozen-lockfile")

		existingLock, _ := jpm.LoadLockFile(".")

		if frozenLock && existingLock != nil {
			if !jpm.LockFileCompatible(existingLock, config) {
				fmt.Println("Error: jabline.lock is not up-to-date with jabline.toml")
				fmt.Println("Run 'jabline install' without --frozen-lockfile to update")
				os.Exit(1)
			}
		}

		resolver := jpm.NewResolver(config)
		if err := resolver.ResolveAll(); err != nil {
			fmt.Printf("Warning: could not resolve all dependencies: %s\n", err)
		}

		installDeps := func(deps map[string]string, kind string) {
			if len(deps) == 0 {
				return
			}

			fmt.Printf("Installing %d %s...\n", len(deps), kind)

			for name, url := range deps {
				resolvedURL, err := jpm.ResolvePackage(url)
				if err != nil {
					fmt.Printf("Warning: could not resolve %s: %s\n", name, err)
					resolvedURL = url
				}

				bestVersion := ""
				if node, ok := resolver.Graph[name]; ok && node.Resolved != nil {
					bestVersion = node.Resolved.String()
				}

				if existingLock != nil {
					if entry, ok := existingLock.Dependencies[name]; ok && entry.Resolved != "" {
						bestVersion = entry.Resolved
					}
				}

				fmt.Printf("Installing %s (%s)...\n", name, bestVersion)
				_, checksum, err := jpm.DownloadDependency(resolvedURL, bestVersion)
				if err != nil {
					fmt.Printf("Error installing %s: %s\n", name, err)
					continue
				}
				if checksum != "" {
					fmt.Printf("  Checksum: %s\n", checksum[:16])
				}

				if existingLock != nil {
					if entry, ok := existingLock.Dependencies[name]; ok && entry.Hash != "" {
						if err := jpm.VerifyPackageIntegrity(jpm.LibDir, name, entry.Hash); err != nil {
							fmt.Printf("Warning: integrity check failed for %s: %s\n", name, err)
						} else {
							fmt.Printf("  Integrity: verified\n")
						}
					}
				}
			}
		}

		installDeps(config.Dependencies, "dependencies")
		installDeps(config.DevDependencies, "dev dependencies")

		// Resolve and install transitive dependencies
		if err := resolver.ResolveTransitive(); err != nil {
			fmt.Printf("Warning: could not resolve transitive deps: %s\n", err)
		}
		transitiveDeps := make(map[string]string)
		for name, node := range resolver.Graph {
			if _, ok := config.Dependencies[name]; ok {
				continue
			}
			if _, ok := config.DevDependencies[name]; ok {
				continue
			}
			transitiveDeps[name] = jpm.ResolvedURL(node)
		}
		installDeps(transitiveDeps, "transitive dependencies")

		if lockOnly {
			lock, err := jpm.UpdateLockFile(config, resolver.ResolvedVersions(), jpm.LibDir)
			if err != nil {
				fmt.Printf("Error generating lock file: %s\n", err)
				os.Exit(1)
			}
			if err := lock.Save("."); err != nil {
				fmt.Printf("Error saving lock file: %s\n", err)
				os.Exit(1)
			}
			fmt.Println("Lock file generated.")
		}

		fmt.Println("Dependency installation complete.")
	},
}

func init() {
	installCmd.Flags().Bool("lock", false, "Generate jabline.lock after installing")
	installCmd.Flags().Bool("frozen-lockfile", false, "Fail if lock file is stale")
	rootCmd.AddCommand(installCmd)
}
