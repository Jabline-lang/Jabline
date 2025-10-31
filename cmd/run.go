package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"jabline/pkg/evaluator"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [file.jb]",
	Short: "Execute a Jabline code file",
	Long: `The run command executes a source code file written in the Jabline language.

The file must have a .jb extension and contain valid Jabline code.

Example:
  jabline run example.jb
  jabline run my_program.jb`,
	Args: cobra.ExactArgs(1),
	Run:  runJablineFile,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runJablineFile(cmd *cobra.Command, args []string) {
	filename := args[0]

	if !strings.HasSuffix(filename, ".jb") {
		fmt.Printf("Error: The file must have a .jb extension.\n")
		os.Exit(1)
	}
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Error: The file '%s' does not exist\n", filename)
		os.Exit(1)
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file '%s': %s\n", filename, err)
		os.Exit(1)
	}

	if err := executeJablineCode(string(content), filename); err != nil {
		fmt.Printf("Error running '%s': %s\n", filename, err)
		os.Exit(1)
	}
}

func executeJablineCode(input string, filename string) error {

	l := lexer.New(input)

	p := parser.New(l)

	program := p.ParseProgram()

	if program == nil {
		return fmt.Errorf("error parsing the program")
	}
	errors := p.Errors()
	if len(errors) > 0 {

		formatter := evaluator.NewErrorFormatter(true, true)

		var formattedErrors []evaluator.FormattedError
		for _, errMsg := range errors {

			line, column := extractLineColumnFromError(errMsg)
			sourceLine := evaluator.ExtractSourceLine(input, line)

			formattedError := evaluator.FormattedError{
				Level:      evaluator.ErrorLevelError,
				Message:    errMsg,
				Line:       line,
				Column:     column,
				Filename:   filename,
				SourceLine: sourceLine,
				Suggestion: getSuggestionForError(errMsg),
			}
			formattedErrors = append(formattedErrors, formattedError)
		}

		warnings := evaluator.DetectCommonMistakes(input)
		formattedErrors = append(formattedErrors, warnings...)

		errorOutput := formatter.FormatMultipleErrors(formattedErrors, filename)
		fmt.Print(errorOutput)
		return fmt.Errorf("compilation failed")
	}

	env := object.NewEnvironment()

	result := evaluator.Eval(program, env)

	if result != nil {
		if result.Type() == object.ERROR_OBJ {
			formatter := evaluator.NewErrorFormatter(true, true)

			runtimeError := evaluator.FormattedError{
				Level:      evaluator.ErrorLevelError,
				Message:    result.Inspect(),
				Line:       0,
				Column:     0,
				Filename:   filename,
				SourceLine: "",
				Suggestion: getSuggestionForError(result.Inspect()),
			}

			errorOutput := formatter.FormatError(runtimeError)
			fmt.Print(errorOutput)
			return fmt.Errorf("runtime error")
		}
		if result.Type() == object.EXCEPTION_OBJ {
			formatter := evaluator.NewErrorFormatter(true, true)

			exceptionError := evaluator.FormattedError{
				Level:      evaluator.ErrorLevelError,
				Message:    "Uncaught " + result.Inspect(),
				Line:       0,
				Column:     0,
				Filename:   filename,
				SourceLine: "",
				Suggestion: "Add a try-catch block to handle this exception",
			}

			errorOutput := formatter.FormatError(exceptionError)
			fmt.Print(errorOutput)
			return fmt.Errorf("uncaught exception")
		}
	}

	return nil
}

func extractLineColumnFromError(errMsg string) (int, int) {

	if strings.Contains(errMsg, "line ") && strings.Contains(errMsg, "column ") {
		var line, column int
		fmt.Sscanf(errMsg, "line %d, column %d:", &line, &column)
		return line, column
	}
	return 0, 0
}

func getSuggestionForError(errMsg string) string {
	if strings.Contains(errMsg, "no prefix parse function") {
		if strings.Contains(errMsg, "ILLEGAL") {
			return "Check for invalid characters or unsupported syntax"
		}
		if strings.Contains(errMsg, ";") {
			return "Remove the semicolon or add an expression before it"
		}
	}

	if strings.Contains(errMsg, "expected next token") {
		if strings.Contains(errMsg, "RPAREN") {
			return "Add a closing parenthesis ')'"
		}
		if strings.Contains(errMsg, "RBRACE") {
			return "Add a closing brace '}'"
		}
	}

	if strings.Contains(errMsg, "identifier not found") {
		return "Check if the variable is declared before using it"
	}

	if strings.Contains(errMsg, "wrong number of arguments") {
		return "Check the function documentation for the correct number of parameters"
	}

	return ""
}
