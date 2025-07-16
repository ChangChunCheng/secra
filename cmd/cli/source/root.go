package source

import "github.com/spf13/cobra"

// Cmd is the parent command for resource operations.
var Cmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage CVE resources",
}

func init() {
	Cmd.AddCommand(createCveCmd)
	Cmd.AddCommand(subscribeCveSourceCmd)
	Cmd.AddCommand(subscribeProductCmd)
	Cmd.AddCommand(subscribeVendorCmd)
	Cmd.AddCommand(createVendorCmd)
	Cmd.AddCommand(getVendorCmd)
	Cmd.AddCommand(listVendorCmd)
	Cmd.AddCommand(updateVendorCmd)
	Cmd.AddCommand(deleteVendorCmd)

	// CVE Source commands
	Cmd.AddCommand(createCveSourceCmd)
	Cmd.AddCommand(getCveSourceCmd)
	Cmd.AddCommand(listCveSourceCmd)
	Cmd.AddCommand(updateCveSourceCmd)
	Cmd.AddCommand(deleteCveSourceCmd)
}
