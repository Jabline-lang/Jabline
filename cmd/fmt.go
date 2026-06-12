package cmd

import (
	"fmt"
	"runtime"
	"strings"

	jfmt "jabline/pkg/fmt"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var newline = "\n"

func init() {
	if runtime.GOOS == "windows" {
		newline = "\r\n"
	}
}

var (
	watchMode bool
	checkMode bool
	lintMode  bool
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [file or directory]",
	Short: "Format and lint Jabline source files",
	Long: `Formats and optionally lints Jabline source files (.jb) in-place.

If no path is provided, the current directory is formatted recursively.
Skips lib/ and .jb_cache/ directories automatically.

Flags:
  -w, --watch   Watch for file changes and reformat automatically.
  -c, --check   Check if files are already formatted (exit 1 if not). Useful for CI.
  -l, --lint    Run linter for style issues (indentation, naming, line length, etc.).`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		if lintMode {
			lintIssuesTotal := 0
			filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil { return nil }
				if info.IsDir() {
					base := info.Name()
					if base == "lib" || base == ".jb_cache" || base == ".git" || base == "node_modules" {
						return filepath.SkipDir
					}
					return nil
				}
				if filepath.Ext(p) != ".jb" { return nil }

				src, err := os.ReadFile(p)
				if err != nil { return nil }
				issues := jfmt.Lint(p, string(src))
				if len(issues) > 0 {
					result := jfmt.LintResult{Filename: p, Issues: issues}
					fmt.Print(jfmt.FormatLintResult(result))
					lintIssuesTotal += len(issues)
				}
				return nil
			})
			if lintIssuesTotal > 0 {
				fmt.Fprintf(os.Stderr, "\n%d lint issue(s) found.\n", lintIssuesTotal)
				os.Exit(1)
			}
			fmt.Println("No lint issues found.")
			return
		}

		if watchMode {
			fmt.Printf("Watching %s for changes...\n", path)
			if err := watchPath(path); err != nil {
				fmt.Fprintf(os.Stderr, "watcher error: %s\n", err)
				os.Exit(1)
			}
			return
		}

		stats, err := formatPath(path, checkMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}

		if checkMode {
			if stats.dirty > 0 {
				fmt.Fprintf(os.Stderr, "%d file(s) are not formatted. Run 'jabline fmt' to fix.\n", stats.dirty)
				os.Exit(1)
			}
			fmt.Printf("All %d file(s) are properly formatted.\n", stats.checked)
		} else if stats.formatted > 0 {
			fmt.Printf("Formatted %d file(s).\n", stats.formatted)
		} else {
			fmt.Printf("All %d file(s) are already formatted.\n", stats.checked)
		}
	},
}

type fmtStats struct {
	checked   int
	formatted int
	dirty     int
	errors    int
}

func formatPath(path string, checkOnly bool) (fmtStats, error) {
	var stats fmtStats
	info, err := os.Stat(path)
	if err != nil {
		return stats, err
	}

	if !info.IsDir() {
		changed, err := formatFile(path, checkOnly)
		if err != nil {
			stats.errors++
			return stats, err
		}
		stats.checked++
		if changed {
			if checkOnly {
				stats.dirty++
				fmt.Printf("  not formatted: %s\n", path)
			} else {
				stats.formatted++
			}
		}
		return stats, nil
	}

	walkErr := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == "lib" || base == ".jb_cache" || base == ".git" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(p) != ".jb" {
			return nil
		}

		changed, ferr := formatFile(p, checkOnly)
		stats.checked++
		if ferr != nil {
			stats.errors++
			fmt.Fprintf(os.Stderr, "  error in %s: %s\n", p, ferr)
			return nil // continue walking
		}
		if changed {
			if checkOnly {
				stats.dirty++
				fmt.Printf("  not formatted: %s\n", p)
			} else {
				fmt.Printf("  formatted: %s\n", p)
				stats.formatted++
			}
		}
		return nil
	})

	return stats, walkErr
}

// formatFile parses and formats a single .jb file.
// Returns (changed, error). If checkOnly=true, does not write back.
func formatFile(filename string, checkOnly bool) (bool, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return false, fmt.Errorf("cannot read %s: %w", filename, err)
	}

	original := string(src)
	// Normalize CRLF to LF for comparison
	normalized := strings.ReplaceAll(original, "\r\n", "\n")

	l := lexer.New(normalized)
	p := parser.New(l)
	program := p.ParseProgram()

	if errs := p.Errors(); len(errs) > 0 {
		return false, fmt.Errorf("parse errors in %s:\n  %s", filename, strings.Join(errs, "\n  "))
	}

	formatted := jfmt.Format(program) + "\n"

	if formatted == normalized {
		return false, nil
	}

	if !checkOnly {
		output := formatted
		if newline == "\r\n" {
			output = strings.ReplaceAll(output, "\n", "\r\n")
		}
		if err := os.WriteFile(filename, []byte(output), 0644); err != nil {
			return true, fmt.Errorf("cannot write %s: %w", filename, err)
		}
	}

	return true, nil
}

func watchPath(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) && filepath.Ext(event.Name) == ".jb" {
					if !strings.Contains(event.Name, "lib") &&
						!strings.Contains(event.Name, ".jb_cache") {
						changed, err := formatFile(event.Name, false)
						if err != nil {
							fmt.Fprintf(os.Stderr, "  error: %s\n", err)
						} else if changed {
							fmt.Printf("  formatted: %s\n", event.Name)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Fprintf(os.Stderr, "watcher error: %s\n", err)
			}
		}
	}()

	if err = addRecursive(watcher, path); err != nil {
		return err
	}
	<-done
	return nil
}

func addRecursive(watcher *fsnotify.Watcher, path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == "lib" || base == ".git" || base == ".jb_cache" || base == "node_modules" {
				return filepath.SkipDir
			}
			return watcher.Add(p)
		}
		return nil
	})
}

func init() {
	fmtCmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "Watch and re-format files on save")
	fmtCmd.Flags().BoolVarP(&checkMode, "check", "c", false, "Check formatting without writing; exits 1 if any file is unformatted")
	fmtCmd.Flags().BoolVarP(&lintMode, "lint", "l", false, "Run linter for style issues")
	rootCmd.AddCommand(fmtCmd)
}
