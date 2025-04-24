package nvd

import (
	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var cfg *config.AppConfig
var db *storage.DBWrapper

var Cmd = &cobra.Command{
	Use:   "nvd",
	Short: "NVD CLI tool",
}

func init() {
	Cmd.AddCommand(v1Nvd)
}
