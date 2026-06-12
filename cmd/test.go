package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"jabline/pkg/stdlib"
	"jabline/pkg/typechecker"
	"jabline/pkg/vm"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	coverFlag         bool
	testWatchFlag     bool
	coverAllCoverage  map[string]map[int]int // aggregated across all test files
)

var testCmd = &cobra.Command{
	Use:   "test [file_or_dir]",
	Short: "Discover and run Jabline test files (*_test.jb)",
	Long: `Search and execute all files ending with _test.jb.

Examples:
  jabline test              — run all *_test.jb in the current directory
  jabline test ./my_test.jb — run a specific test file
  jabline test ./tests/     — run all tests in the specified directory
  jabline test --coverage   — run tests and show coverage report`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := "."
		if len(args) == 1 {
			target = args[0]
		}

		if testWatchFlag {
			watchTests(target)
			return
		}

		runTests(target)
	},
}

func printCoverageReport() {
	totalLines := 0
	coveredLines := 0

	// Collect unique source files from the coverage data
	sourceFiles := make(map[string]bool)
	for filename := range coverAllCoverage {
		sourceFiles[filename] = true
	}

	// If no filenames recorded, show per-file stats from coverage
	if len(sourceFiles) == 0 {
		fmt.Println("\n\033[1mCoverage: no data collected\033[0m")
		return
	}

	fmt.Printf("\n\033[1m══════════ Coverage Report ══════════\033[0m\n")

	var fileNames []string
	for f := range coverAllCoverage {
		fileNames = append(fileNames, f)
	}
	sort.Strings(fileNames)

	for _, filename := range fileNames {
		lines := coverAllCoverage[filename]
		covered := len(lines)
		totalLines += covered
		coveredLines += covered

		// Count how many unique lines were hit in this file
		lineNums := make([]int, 0, len(lines))
		for line := range lines {
			lineNums = append(lineNums, line)
		}
		sort.Ints(lineNums)

		fmt.Printf("  \033[36m%s\033[0m: %d lines hit\n", filename, covered)
		if len(lineNums) > 0 {
			// Show first few and last few
			show := lineNums
			if len(show) > 20 {
				show = append(show[:10], show[len(show)-10:]...)
			}
			lineStr := make([]string, len(show))
			for i, l := range show {
				lineStr[i] = fmt.Sprintf("%d", l)
			}
			fmt.Printf("    Lines: %s\n", strings.Join(lineStr, ", "))
		}
	}

	if totalLines > 0 {
		fmt.Printf("\n  Total: %d lines executed\n", coveredLines)
	}
}

func discoverTestFiles(target string) ([]string, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	// If it's a single file, just return it
	if !info.IsDir() {
		if strings.HasSuffix(target, "_test.jb") || strings.HasSuffix(target, ".jb") {
			return []string{target}, nil
		}
		return nil, fmt.Errorf("the file '%s' is not a Jabline test", target)
	}

	// Walk the directory recursively
	var files []string
	err = filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() && strings.HasSuffix(path, "_test.jb") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// runTestFile compiles and runs one test file, returning (passed, failed) counts.
func runTestFile(filename string) (passed, failed int) {
	fmt.Printf("\033[33m▶ %s\033[0m\n", filename)

	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("  \033[31mError leyendo archivo: %s\033[0m\n", err)
		return 0, 1
	}

	// The assert.jb module exports describe, it, assertEqual, etc.
	// We prepend the import so all test helpers are available.
	preamble := `import { describe, it, assertEqual, assertNotEqual, assertTrue, assertFalse, assertNull, assertNotNull, assertType, assertLength, assertGreaterThan, assertLessThan, roughlyEqual, includes, fails, showFinalReport } from "testing/assert";` + "\n\n"
	// Auto-inject showFinalReport() so Go always reads the correct counts.
	source := preamble + string(src) + "\n\nshowFinalReport();\n"

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("  \033[31mErrores del parser:\033[0m\n")
		for _, msg := range p.Errors() {
			fmt.Printf("    %s\n", msg)
		}
		return 0, 1
	}

	checker := typechecker.New()
	typeErrors := checker.Check(program)
	if len(typeErrors) > 0 {
		fmt.Printf("  \033[33mType warnings:\033[0m\n")
		for _, msg := range typeErrors {
			fmt.Printf("    %s\n", msg)
		}
	}

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Printf("  \033[31mCompiler error: %s\033[0m\n", err)
		return 0, 1
	}

	bytecode := comp.Bytecode()
	machine := vm.New(bytecode.Instructions, bytecode.Constants, filename)

	// Enable coverage if --coverage flag is set
	if coverFlag {
		machine.RecordingCoverage = true
		machine.ResetCoverage()
	}

	// Reset runner stats before execute
	stdlib.ResetTestResults()

	if err := machine.Run(); err != nil {
		fmt.Printf("  \033[31mRuntime error: %s\033[0m\n", err)
		return 0, 1
	}

	// Aggregate coverage data
	if coverFlag {
		report := machine.GetCoverageReport()
		for fname, lines := range report {
			if coverAllCoverage[fname] == nil {
				coverAllCoverage[fname] = make(map[int]int)
			}
			for line, count := range lines {
				coverAllCoverage[fname][line] += count
			}
		}
	}

	// Read test stats from the native reporter
	return stdlib.TestPassed, stdlib.TestFailed
}

func runTests(target string) {
	testFiles, err := discoverTestFiles(target)
	if err != nil {
		fmt.Printf("Error searching for tests: %s\n", err)
		os.Exit(1)
	}

	if len(testFiles) == 0 {
		fmt.Println("⚠️  No test files (*_test.jb) found.")
		os.Exit(0)
	}

	fmt.Printf("\033[1;34m╔══════════════════════════════════╗\033[0m\n")
	fmt.Printf("\033[1;34m║   Jabline Test Runner            ║\033[0m\n")
	fmt.Printf("\033[1;34m╚══════════════════════════════════╝\033[0m\n")
	fmt.Printf("\nFound \033[1m%d\033[0m test files\n\n", len(testFiles))

	totalPassed := 0
	totalFailed := 0
	totalFiles := 0
	if coverFlag {
		coverAllCoverage = make(map[string]map[int]int)
	}

	for _, testFile := range testFiles {
		passed, failed := runTestFile(testFile)
		totalPassed += passed
		totalFailed += failed
		totalFiles++
	}

	fmt.Printf("\n\033[1m══════════ Final Summary ══════════\033[0m\n")
	fmt.Printf("Executed files: %d\n", totalFiles)
	fmt.Printf("Tests passed:  \033[32m%d ✅\033[0m\n", totalPassed)
	fmt.Printf("Tests failed:  \033[31m%d ❌\033[0m\n", totalFailed)

	if coverFlag && coverAllCoverage != nil {
		printCoverageReport()
	}

	if totalFailed == 0 {
		fmt.Printf("\n\033[1;32m🎉 All tests passed! 🎉\033[0m\n")
	} else {
		fmt.Printf("\n\033[1;31m💥 %d test(s) failed.\033[0m\n", totalFailed)
		os.Exit(1)
	}
}

func watchTests(target string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "watcher error: %s\n", err)
		os.Exit(1)
	}
	defer watcher.Close()

	// Add directories recursively
	filepath.Walk(target, func(p string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if info.IsDir() {
			base := info.Name()
			if base == "lib" || base == ".jb_cache" || base == ".git" || base == "node_modules" {
				return filepath.SkipDir
			}
			watcher.Add(p)
		}
		return nil
	})

	fmt.Printf("Watching %s for changes... (Ctrl+C to stop)\n", target)

	// Run tests once at start
	clearConsole()
	fmt.Printf("🔁 Running tests (%s)...\n", time.Now().Format("15:04:05"))
	runTests(target)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok { return }
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 && (strings.HasSuffix(event.Name, ".jb") || strings.HasSuffix(event.Name, ".go")) {
				clearConsole()
				fmt.Printf("🔁 Change detected: %s\n", event.Name)
				fmt.Printf("🔁 Running tests (%s)...\n", time.Now().Format("15:04:05"))
				runTests(target)
			}
		case err, ok := <-watcher.Errors:
			if !ok { return }
			fmt.Fprintf(os.Stderr, "watcher error: %s\n", err)
		}
	}
}

func clearConsole() {
	fmt.Print("\033[2J\033[H")
}

func init() {
	testCmd.Flags().BoolVar(&coverFlag, "coverage", false, "Report code coverage")
	testCmd.Flags().BoolVarP(&testWatchFlag, "watch", "w", false, "Watch for changes and re-run tests")
	rootCmd.AddCommand(testCmd)
}
