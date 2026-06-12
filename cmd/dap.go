package cmd

import (
	"os"

	"jabline/pkg/dap"

	"github.com/spf13/cobra"
)

var dapCmd = &cobra.Command{
	Use:   "dap",
	Short: "Start the Debug Adapter Protocol server (for VS Code debugger)",
	Run: func(cmd *cobra.Command, args []string) {
		adapter := dap.NewDebugAdapter(os.Stdin, os.Stdout)
		adapter.Run()
	},
}

func init() {
	rootCmd.AddCommand(dapCmd)
}
