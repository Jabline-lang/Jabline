package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
	"jabline/pkg/vm"

	"github.com/spf13/cobra"
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
	
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.New().GetSymbolTable()
	
	fmt.Println("Jabline REPL v0.5")
	fmt.Println("Type 'exit' to quit.")
	
	for {
		fmt.Print(">>")
		if !scanner.Scan() {
			return
		}
		line := scanner.Text()
		
		if line == "exit" {
			return
		}

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
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		fmt.Fprintf(out, "\t%s\n", msg)
	}
}
