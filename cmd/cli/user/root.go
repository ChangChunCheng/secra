package usercmd

import "github.com/spf13/cobra"

// Cmd is the parent for all `user` subcommands
var Cmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
}

func init() {
        Cmd.AddCommand(registerLocalCmd)
}
