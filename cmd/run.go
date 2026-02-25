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

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Execute a Jabline program",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading file: %s\n", err)
			os.Exit(1)
		}

		code := string(bytes)
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			fmt.Println("Parser errors:")
			for _, msg := range p.Errors() {
				fmt.Printf("\t%s\n", msg)
			}
			os.Exit(1)
		}

		comp := compiler.New()
		err = comp.Compile(program)
		if err != nil {
			fmt.Printf("Compiler error: %s\n", err)
			os.Exit(1)
		}

		bytecode := comp.Bytecode()
		machine := vm.New(bytecode.Instructions, bytecode.Constants, filename)
		err = machine.Run()

		if err != nil {
			fmt.Printf("VM runtime error: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
