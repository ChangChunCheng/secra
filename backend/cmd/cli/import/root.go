package importcmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/cmd/cli/import/nvd"
)

var Cmd = &cobra.Command{
	Use:   "import",
	Short: "Import CLI tool",
}

func init() {
	Cmd.AddCommand(nvd.Cmd)
}
