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

	// Check for parser errors
	errors := p.Errors()
	if len(errors) > 0 {
		return fmt.Errorf("parser errors: %s", errors[0])
	}

	// Crear el environment
	env := object.NewEnvironment()

	// Evaluar el programa
	result := evaluator.Eval(program, env)

	// Si hay un error o excepción no capturada, mostrarlo
	if result != nil {
		if result.Type() == object.ERROR_OBJ {
			return fmt.Errorf("%s", result.Inspect())
		}
		if result.Type() == object.EXCEPTION_OBJ {
			return fmt.Errorf("Uncaught %s", result.Inspect())
		}
	}

	return nil
}
