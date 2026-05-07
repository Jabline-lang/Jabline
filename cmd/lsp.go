package cmd

import (
	"jabline/pkg/lsp"

	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start the Jabline Language Server",
	Long:  `Starts the Jabline Language Server Protocol (LSP) server over Stdio. This is intended to be used by editors like VS Code, Neovim, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		server := lsp.NewServer()
		server.RunStdio()
	},
}

func init() {
	lspCmd.Flags().Bool("stdio", false, "Use stdio (default, kept for compatibility with VSCode LanguageClient)")
	rootCmd.AddCommand(lspCmd)
}
