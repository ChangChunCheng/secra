package resource

import "github.com/spf13/cobra"

// Cmd is the parent command for resource operations.
var Cmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage CVE resources",
}

func init() {
	// Subcommands will be added here
	// e.g., Cmd.AddCommand(createCveResourceCmd)
}
