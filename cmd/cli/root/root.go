package root

import (
	"github.com/spf13/cobra"
	importcmd "gitlab.com/jacky850509/secra/cmd/cli/import"
	"gitlab.com/jacky850509/secra/cmd/cli/migrate"
	"gitlab.com/jacky850509/secra/cmd/cli/source"
	"gitlab.com/jacky850509/secra/cmd/cli/user"
)

// rootCmd is the base command for secra CLI
var rootCmd = &cobra.Command{
	Use:   "secra",
	Short: "Secra CLI tool",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// User commands
	rootCmd.AddCommand(user.Cmd)
	// Resource commands
	rootCmd.AddCommand(source.Cmd)
	rootCmd.AddCommand(migrate.Cmd)
	rootCmd.AddCommand(importcmd.Cmd)
}
