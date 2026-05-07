package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"jabline/pkg/vm"

	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug [file]",
	Short: "Execute a Jabline program in interactive debugging mode",
	Long: `In interactive debug mode you can inspect variable state and control execution line by line.

Available commands in debugger:
  n / next      — execute next line
  c / continue  — continue execution until next breakpoint or end
  b <line>      — set a breakpoint at line
  d <line>      — delete a breakpoint
  locals        — show local variables of current frame
  stack         — show call stack
  list          — show code around current line
  q / quit      — exit debugger`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		src, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error leyendo archivo: %s\n", err)
			os.Exit(1)
		}

		source := string(src)
		l := lexer.New(source)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			fmt.Println("Errores del parser:")
			for _, msg := range p.Errors() {
				fmt.Printf("\t%s\n", msg)
			}
			os.Exit(1)
		}

		comp := compiler.New()
		err = comp.Compile(program)
		if err != nil {
			fmt.Printf("Error del compilador: %s\n", err)
			os.Exit(1)
		}

		bytecode := comp.Bytecode()
		machine := vm.New(bytecode.Instructions, bytecode.Constants, filename)

		// Attach debugger
		session := vm.NewDebugSession(machine, source)
		machine.Debug = session

		fmt.Printf("\033[1;35m╔══════════════════════════════════╗\033[0m\n")
		fmt.Printf("\033[1;35m║   Jabline Debugger  — %s\033[0m\n", filename)
		fmt.Printf("\033[1;35m╚══════════════════════════════════╝\033[0m\n")
		fmt.Println("Escribe 'h' para ver los comandos disponibles.")
		fmt.Println()

		if err := machine.Run(); err != nil {
			fmt.Printf("\n\033[31mError en el VM: %s\033[0m\n", err)
			os.Exit(1)
		}

		machine.PrintTelemetry()
		fmt.Println("\n\033[32m[debug] Programa finalizado.\033[0m")
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
