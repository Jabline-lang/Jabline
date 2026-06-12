package cmd

import (
	"fmt"
	"os"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"jabline/pkg/typechecker"
	"jabline/pkg/vm"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	hot           bool
	evalCode      string
	astDump       bool
	bcDump        bool
	sandboxLevel  string
)

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Execute a Jabline program",
	RunE: func(cmd *cobra.Command, args []string) error {
		var code string
		var filename string

		if evalCode != "" {
			code = evalCode
			filename = "<inline>"
		} else {
			if len(args) == 0 {
				return fmt.Errorf("requires at least 1 arg(s) or -e flag")
			}
			filename = args[0]
			bytes, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("error reading file: %s", err)
			}
			code = string(bytes)
		}

		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			for _, msg := range p.Errors() {
				fmt.Printf("Parser error: %s\n", msg)
			}
			return fmt.Errorf("parse failed")
		}

		checker := typechecker.New()
		typeErrors := checker.Check(program)
		if len(typeErrors) > 0 {
			for _, msg := range typeErrors {
				fmt.Printf("Type error: %s\n", msg)
			}
			return fmt.Errorf("type check failed")
		}

		if astDump {
			fmt.Println("=== AST Dump ===")
			fmt.Println(program.String())
		}

		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			return fmt.Errorf("compiler error: %s", err)
		}

		bytecode := comp.Bytecode()

		if bcDump {
			fmt.Println("=== Bytecode Dump ===")
			fmt.Println(bytecode.Instructions.String())
			fmt.Println("--- Constants ---")
			for i, obj := range bytecode.Constants {
				fmt.Printf("[%04d] %s\n", i, obj.Inspect())
			}
		}

		machine := vm.New(bytecode.Instructions, bytecode.Constants, filename)
		vm.GlobalVM = machine

		if sandboxLevel != "" {
			if err := machine.SetSandboxLevel(sandboxLevel); err != nil {
				return fmt.Errorf("invalid sandbox level %q: %s", sandboxLevel, err)
			}
		}

		if hot {
			go watchFile(filename, machine)
		}

		if err := machine.Run(); err != nil {
			machine.Cleanup()
			return fmt.Errorf("VM runtime error: %s", err)
		}

		machine.PrintTelemetry()
		machine.Cleanup()
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVarP(&hot, "hot", "H", false, "Enable hot code reloading")
	runCmd.Flags().StringVarP(&evalCode, "eval", "e", "", "Evaluate string as Jabline code")
	runCmd.Flags().BoolVarP(&astDump, "ast", "a", false, "Print AST after parsing")
	runCmd.Flags().BoolVarP(&bcDump, "bytecode", "b", false, "Print bytecode instructions before executing")
	runCmd.Flags().StringVarP(&sandboxLevel, "sandbox", "s", "", "Sandbox level (none, secure, restrictive, isolated)")
	rootCmd.AddCommand(runCmd)
}

func watchFile(filename string, machine *vm.VM) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher: %s\n", err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("\n[HOT] Change detected, reloading...")
					reload(filename, machine)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Watcher error: %s\n", err)
			}
		}
	}()

	err = watcher.Add(filename)
	if err != nil {
		fmt.Printf("Error adding file to watcher: %s\n", err)
		return
	}
	<-done
}

func reload(filename string, machine *vm.VM) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file during reload: %s\n", err)
		return
	}

	code := string(bytes)
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Reload parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		return
	}

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Printf("Reload compiler error: %s\n", err)
		return
	}

	bytecode := comp.Bytecode()
	machine.Reload(bytecode.Instructions, bytecode.Constants)
}
