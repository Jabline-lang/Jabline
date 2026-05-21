package cmd

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"jabline/pkg/typechecker"

	"github.com/spf13/cobra"
)

var outputBin string

func determineBuildTags(source string) string {
	tags := []string{"runner"}
	
	if strings.Contains(source, "db\"") || strings.Contains(source, "std/db") || strings.Contains(source, "_db") || strings.Contains(source, "db_open") || strings.Contains(source, "cloud\"") {
		tags = append(tags, "runner_sqlite")
	}
	
	if strings.Contains(source, "http\"") || strings.Contains(source, "net/http") || strings.Contains(source, "std/http") || strings.Contains(source, "std/net/http") || strings.Contains(source, "_http") || strings.Contains(source, "http_serve") || strings.Contains(source, "cloud\"") {
		tags = append(tags, "runner_http")
	}
	
	if strings.Contains(source, "websocket\"") || strings.Contains(source, "net/websocket") || strings.Contains(source, "std/websocket") || strings.Contains(source, "std/net/websocket") || strings.Contains(source, "_websocket") || strings.Contains(source, "wsConnect") {
		tags = append(tags, "runner_websocket")
	}
	
	return strings.Join(tags, ",")
}

var MagicMarker = []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77}

var buildCmd = &cobra.Command{
	Use:   "build [file]",
	Short: "Compile a Jabline program into a standalone executable",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		
		sourceBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading file: %s\n", err)
			os.Exit(1)
		}

		l := lexer.New(string(sourceBytes))
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			fmt.Println("Parser errors:")
			for _, msg := range p.Errors() {
				fmt.Printf("\t%s\n", msg)
			}
			os.Exit(1)
		}

		checker := typechecker.New()
		typeErrors := checker.Check(program)
		if len(typeErrors) > 0 {
			fmt.Println("Type errors:")
			for _, msg := range typeErrors {
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

		bytecodeData, err := compiler.Serialize(comp.Bytecode())
		if err != nil {
			fmt.Printf("Serialization error: %s\n", err)
			os.Exit(1)
		}

		outputName := outputBin
		if outputName == "" {
			ext := filepath.Ext(filename)
			outputName = filename[0 : len(filename)-len(ext)]
		}
		if runtime.GOOS == "windows" && filepath.Ext(outputName) != ".exe" {
			outputName += ".exe"
		}

		selfPath, err := os.Executable()
		if err != nil {
			fmt.Printf("Failed to locate self executable: %s\n", err)
			os.Exit(1)
		}

		buildTags := determineBuildTags(string(sourceBytes))
		safeTags := strings.ReplaceAll(buildTags, ",", "_")
		runnerPath := filepath.Join(os.TempDir(), fmt.Sprintf("jabline_runner_%s_%s_%s", runtime.GOOS, runtime.GOARCH, safeTags))
		if runtime.GOOS == "windows" {
			runnerPath += ".exe"
		}

		runnerBytes, err := ioutil.ReadFile(runnerPath)
		if err != nil {
			runnerSrc := filepath.Join(filepath.Dir(selfPath), "internal", "builder", "runner.go")
			if _, errSrc := os.Stat(runnerSrc); errSrc == nil {
				fmt.Printf("Building internal minimal runner (Tree-Shaking: %s)...\n", buildTags)
				
				cmdBuild := exec.Command("go", "build", "-tags", buildTags, "-ldflags=-s -w", "-trimpath", "-o", runnerPath, runnerSrc)
				cmdBuild.Stdout = os.Stdout
				cmdBuild.Stderr = os.Stderr
				
				if errBuild := cmdBuild.Run(); errBuild != nil {
					fmt.Printf("Failed to compile minimal runner: %v\n", errBuild)
					os.Exit(1)
				}
				
				runnerBytes, err = ioutil.ReadFile(runnerPath)
				if err != nil {
					fmt.Printf("Failed to read compiled minimal runner: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Printf("Warning: runner.go not found, falling back to full compiler binary\n")
				runnerBytes, err = ioutil.ReadFile(selfPath)
				if err != nil {
					fmt.Printf("Failed to read self executable: %s\n", err)
					os.Exit(1)
				}
			}
		}

		f, err := os.OpenFile(outputName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			fmt.Printf("Failed to create output file: %s\n", err)
			os.Exit(1)
		}
		defer f.Close()

		_, err = f.Write(runnerBytes)
		if err != nil {
			fmt.Printf("Failed to write runtime: %s\n", err)
			os.Exit(1)
		}

		_, err = f.Write(bytecodeData)
		if err != nil {
			fmt.Printf("Failed to write bytecode: %s\n", err)
			os.Exit(1)
		}

		sizeBuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(sizeBuf, uint64(len(bytecodeData)))
		_, err = f.Write(sizeBuf)
		if err != nil {
			fmt.Printf("Failed to write size: %s\n", err)
			os.Exit(1)
		}

		_, err = f.Write(MagicMarker)
		if err != nil {
			fmt.Printf("Failed to write marker: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully built binary: %s\n", outputName)
	},
}

func init() {
	buildCmd.Flags().StringVarP(&outputBin, "output", "o", "", "Output binary name")
	rootCmd.AddCommand(buildCmd)
}


