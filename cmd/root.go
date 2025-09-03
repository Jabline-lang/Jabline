package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jabline",
	Short: "Jabline - Un lenguaje de programación simple",
	Long: `Jabline es un lenguaje de programación interpretado simple y fácil de usar.

Este es el intérprete de línea de comandos para Jabline que te permite:
- Ejecutar archivos de código .jb
- Explorar las características del lenguaje

Para empezar, prueba ejecutando un archivo:
  jabline run mi_archivo.jb`,
	Version: "1.0.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(`{{printf "%s version %s\n" .Name .Version}}`)
}
