package cve

import "github.com/spf13/cobra"

// cveCmd is the parent command for CVE operations.
var Cmd = &cobra.Command{
	Use:   "cve",
	Short: "Manage CVEs",
}

func init() {
	// CVE commands
	Cmd.AddCommand(createCveCmd)
	Cmd.AddCommand(getCveCmd)
	Cmd.AddCommand(listCveCmd)
	Cmd.AddCommand(updateCveCmd)
	Cmd.AddCommand(deleteCveCmd)
}
