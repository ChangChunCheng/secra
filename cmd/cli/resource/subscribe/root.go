package subscribe

import "github.com/spf13/cobra"

// Cmd is the parent command for resource operations.
var Cmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Subscribe CVE",
}

func init() {
	Cmd.AddCommand(subscribeCveSourceCmd)
	Cmd.AddCommand(subscribeVendorCmd)
	Cmd.AddCommand(subscribeProductCmd)
}
