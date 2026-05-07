package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"jabline/pkg/stdlib"
	"jabline/pkg/vm"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [file_or_dir]",
	Short: "Discover and run Jabline test files (*_test.jb)",
	Long: `Search and execute all files ending with _test.jb.

Examples:
  jabline test              — run all *_test.jb in the current directory
  jabline test ./my_test.jb — run a specific test file
  jabline test ./tests/     — run all tests in the specified directory`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := "."
		if len(args) == 1 {
			target = args[0]
		}

		// Discover test files
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

		for _, testFile := range testFiles {
			passed, failed := runTestFile(testFile)
			totalPassed += passed
			totalFailed += failed
			totalFiles++
		}

		// Final summary
		fmt.Printf("\n\033[1m══════════ Final Summary ══════════\033[0m\n")
		fmt.Printf("Executed files: %d\n", totalFiles)
		fmt.Printf("Tests passed:  \033[32m%d ✅\033[0m\n", totalPassed)
		fmt.Printf("Tests failed:  \033[31m%d ❌\033[0m\n", totalFailed)

		if totalFailed == 0 {
			fmt.Printf("\n\033[1;32m🎉 All tests passed! 🎉\033[0m\n")
		} else {
			fmt.Printf("\n\033[1;31m💥 %d test(s) failed.\033[0m\n", totalFailed)
			os.Exit(1) // Non-zero exit for CI/CD
		}
	},
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

	src, err := ioutil.ReadFile(filename)
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

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Printf("  \033[31mError del compilador: %s\033[0m\n", err)
		return 0, 1
	}

	bytecode := comp.Bytecode()
	machine := vm.New(bytecode.Instructions, bytecode.Constants, filename)

	// Reset runner stats before execute
	stdlib.ResetTestResults()

	if err := machine.Run(); err != nil {
		fmt.Printf("  \033[31mError de ejecución: %s\033[0m\n", err)
		return 0, 1
	}

	// Read test stats from the native reporter
	return stdlib.TestPassed, stdlib.TestFailed
}

func init() {
	rootCmd.AddCommand(testCmd)
}
