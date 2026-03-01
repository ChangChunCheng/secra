package vendor

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for vendor operations.
var Cmd = &cobra.Command{
	Use:   "vendor",
	Short: "Manage vendors",
}

func init() {
	// Vendor commands
	Cmd.AddCommand(createVendorCmd)
	Cmd.AddCommand(getVendorCmd)
	Cmd.AddCommand(listVendorCmd)
	Cmd.AddCommand(updateVendorCmd)
	Cmd.AddCommand(deleteVendorCmd)
}
