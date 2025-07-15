package resource

import "github.com/spf13/cobra"

// Cmd is the parent command for resource operations.
var Cmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage CVE resources",
}

func init() {
	Cmd.AddCommand(createCveResourceCmd)
	Cmd.AddCommand(createCveCmd)
	Cmd.AddCommand(subscribeCveResourceCmd)
	Cmd.AddCommand(subscribeProductCmd)
	Cmd.AddCommand(subscribeVendorCmd)
}
