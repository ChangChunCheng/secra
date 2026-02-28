package root

import (
	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/cmd/cli/backup"
	"gitlab.com/jacky850509/secra/cmd/cli/health"
	importcmd "gitlab.com/jacky850509/secra/cmd/cli/import"
	"gitlab.com/jacky850509/secra/cmd/cli/migrate"
	"gitlab.com/jacky850509/secra/cmd/cli/resource"
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
	rootCmd.AddCommand(resource.Cmd)
	rootCmd.AddCommand(migrate.Cmd)
	rootCmd.AddCommand(importcmd.Cmd)
	rootCmd.AddCommand(backup.BackupCmd)
	rootCmd.AddCommand(health.Cmd)
}
