package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jabline",
	Short: "Jabline - A simple programming language",
	Long: `Jabline is a simple and easy-to-use interpreted programming language.

This is the command-line interpreter for Jabline that allows you to:
- Execute .jb code files
- Explore language features

To start, try running a file:
  jabline run my_file.jb`,
	Version: "0.0.1",
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
