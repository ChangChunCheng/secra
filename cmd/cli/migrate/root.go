package migrate

import "github.com/spf13/cobra"

// Cmd is the parent for all `migrate` subcommands
var Cmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage DB migrations",
}

func init() {
	Cmd.AddCommand(UpCmd)
	Cmd.AddCommand(StatusCmd)
}
