package backup

import (
	"github.com/spf13/cobra"
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Database backup and restore management (Parquet + Tar.gz)",
}

func init() {
	BackupCmd.AddCommand(createCmd)
	BackupCmd.AddCommand(restoreCmd)
}
