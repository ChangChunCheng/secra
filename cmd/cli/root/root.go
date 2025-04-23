package root

import (
	"github.com/spf13/cobra"
	importcmd "gitlab.com/jacky850509/secra/cmd/cli/import"
	"gitlab.com/jacky850509/secra/cmd/cli/migrate"
)

var rootCmd = &cobra.Command{
	Use:   "secra",
	Short: "Secra CLI tool",
}

// Execute 是 CLI 的進入點
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(migrate.Cmd)
	rootCmd.AddCommand(importcmd.Cmd)
}
