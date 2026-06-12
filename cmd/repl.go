package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
	"jabline/pkg/vm"

	"github.com/spf13/cobra"
)

var (
	historyFile string
	history     []string
	historyIdx  int
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start the Jabline interactive shell",
	Run: func(cmd *cobra.Command, args []string) {
		startRepl()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}

func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)

	// Load history
	home, _ := os.UserHomeDir()
	if home != "" {
		historyFile = filepath.Join(home, ".jabline_history")
		data, err := os.ReadFile(historyFile)
		if err == nil {
			history = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
			historyIdx = len(history)
		}
	}
	if history == nil {
		history = []string{}
		historyIdx = 0
	}

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.New().GetSymbolTable()

	fmt.Println("Jabline REPL v0.6")
	fmt.Println("Type 'exit' to quit.")

	// Multi-line input buffer
	buffer := ""

	for {
		if buffer == "" {
			fmt.Print("\033[32m>>\033[0m ")
		} else {
			fmt.Print("\033[33m..\033[0m ")
		}

		if !scanner.Scan() {
			saveHistory()
			return
		}
		line := scanner.Text()

		if line == "exit" {
			saveHistory()
			return
		}

		// If line starts with history prefix, show history
		if line == ":history" {
			for i, h := range history {
				fmt.Printf("%4d  %s\n", i+1, h)
			}
			continue
		}

		// Append to buffer
		if buffer != "" {
			buffer = buffer + "\n" + line
		} else {
			buffer = line
		}

		// Check if the buffer looks complete
		if isCompleteInput(buffer) {
			line = buffer
			buffer = ""

			l := lexer.New(line)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				printParserErrors(os.Stdout, p.Errors())
				continue
			}

			comp := compiler.NewWithState(symbolTable, constants)
			err := comp.Compile(program)
			if err != nil {
				fmt.Printf("Woops! Compilation failed:\n %s\n", err)
				continue
			}

			code := comp.Bytecode()
			constants = code.Constants

			machine := vm.NewWithGlobalsStore(code.Instructions, code.Constants, globals, "REPL")
			err = machine.Run()
			if err != nil {
				fmt.Printf("Woops! Executing bytecode failed:\n %s\n", err)
				continue
			}

			lastPopped := machine.LastPoppedElement()
			if lastPopped != nil {
				fmt.Println(lastPopped.Inspect())
			}

			// Save to history (non-empty, non-whitespace lines)
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && trimmed != "exit" {
				history = append(history, trimmed)
				historyIdx = len(history)
			}
		}
		// Else: continue reading more lines
	}
}

// isCompleteInput checks for balanced brackets to detect incomplete input.
func isCompleteInput(input string) bool {
	depth := 0
	for _, ch := range input {
		switch ch {
		case '{', '(', '[':
			depth++
		case '}', ')', ']':
			depth--
		}
	}
	if depth > 0 {
		return false
	}
	// Try parse — if it fails, it might still be incomplete
	l := lexer.New(input)
	p := parser.New(l)
	p.ParseProgram()
	if len(p.Errors()) > 0 {
		// Check if all errors are just "unexpected EOF" style
		for _, e := range p.Errors() {
			if strings.Contains(e, "unexpected end") || strings.Contains(e, "expected") {
				return false
			}
		}
	}
	return true
}

func saveHistory() {
	if historyFile == "" || len(history) == 0 {
		return
	}
	// Keep last 1000 commands
	if len(history) > 1000 {
		history = history[len(history)-1000:]
	}
	data := strings.Join(history, "\n") + "\n"
	os.WriteFile(historyFile, []byte(data), 0644)
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		fmt.Fprintf(out, "\t%s\n", msg)
	}
}
