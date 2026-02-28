package health

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "health",
	Short: "Health check commands for Docker and orchestration",
}

func init() {
	Cmd.AddCommand(checkCmd)
}
