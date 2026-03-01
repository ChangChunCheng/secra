package health

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "health",
	Short: "Health check and system testing",
}

func init() {
	Cmd.AddCommand(checkCmd)
}
