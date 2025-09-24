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
	Use:   "run [archivo.jb]",
	Short: "Ejecuta un archivo de código Jabline",
	Long: `El comando run ejecuta un archivo de código fuente escrito en el lenguaje Jabline.

El archivo debe tener extensión .jb y contener código válido de Jabline.

Ejemplo:
  jabline run ejemplo.jb
  jabline run mi_programa.jb`,
	Args: cobra.ExactArgs(1),
	Run:  runJablineFile,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runJablineFile(cmd *cobra.Command, args []string) {
	filename := args[0]

	// Verificar que el archivo tenga extensión .jb
	if !strings.HasSuffix(filename, ".jb") {
		fmt.Printf("Error: El archivo debe tener extensión .jb\n")
		os.Exit(1)
	}

	// Verificar que el archivo exista
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Error: El archivo '%s' no existe\n", filename)
		os.Exit(1)
	}

	// Leer el contenido del archivo
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error leyendo archivo '%s': %s\n", filename, err)
		os.Exit(1)
	}

	// Ejecutar el código
	if err := executeJablineCode(string(content), filename); err != nil {
		fmt.Printf("Error ejecutando '%s': %s\n", filename, err)
		os.Exit(1)
	}
}

func executeJablineCode(input string, filename string) error {
	// Crear el lexer
	l := lexer.New(input)

	// Crear el parser
	p := parser.New(l)

	// Parsear el programa
	program := p.ParseProgram()

	if program == nil {
		return fmt.Errorf("error parseando el programa")
	}

	// Check for parser errors with enhanced formatting
	errors := p.Errors()
	if len(errors) > 0 {
		// Create error formatter
		formatter := evaluator.NewErrorFormatter(true, true)

		var formattedErrors []evaluator.FormattedError
		for _, errMsg := range errors {
			// Extract line/column info from error message if available
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

		// Check for warnings
		warnings := evaluator.DetectCommonMistakes(input)
		formattedErrors = append(formattedErrors, warnings...)

		errorOutput := formatter.FormatMultipleErrors(formattedErrors, filename)
		fmt.Print(errorOutput)
		return fmt.Errorf("compilation failed")
	}

	// Crear el environment
	env := object.NewEnvironment()

	// Evaluar el programa
	result := evaluator.Eval(program, env)

	// Si hay un error o excepción no capturada, mostrarlo con formato mejorado
	if result != nil {
		if result.Type() == object.ERROR_OBJ {
			formatter := evaluator.NewErrorFormatter(true, true)

			runtimeError := evaluator.FormattedError{
				Level:      evaluator.ErrorLevelError,
				Message:    result.Inspect(),
				Line:       0, // Runtime errors don't have line info yet
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

// extractLineColumnFromError extracts line and column numbers from error message
func extractLineColumnFromError(errMsg string) (int, int) {
	// Look for pattern "line X, column Y:"
	if strings.Contains(errMsg, "line ") && strings.Contains(errMsg, "column ") {
		var line, column int
		fmt.Sscanf(errMsg, "line %d, column %d:", &line, &column)
		return line, column
	}
	return 0, 0
}

// getSuggestionForError provides suggestions based on error message
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
