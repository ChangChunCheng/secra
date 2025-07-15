package user

import (
	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var cfg *config.AppConfig
var db *storage.DBWrapper

var Cmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
}

func init() {
	Cmd.AddCommand(registerCmd)
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(updateProfileCmd)
}
