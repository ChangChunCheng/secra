package cvesource

import "github.com/spf13/cobra"

// cveSourceCmd is the parent command for CVE source operations.
var Cmd = &cobra.Command{
	Use:   "cve-source",
	Short: "Manage CVE sources",
}

func init() {
	Cmd.AddCommand(createCveSourceCmd)
	Cmd.AddCommand(getCveSourceCmd)
	Cmd.AddCommand(listCveSourceCmd)
	Cmd.AddCommand(updateCveSourceCmd)
	Cmd.AddCommand(deleteCveSourceCmd)
}
