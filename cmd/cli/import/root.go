package importcmd

import (
	"log"

	"github.com/spf13/cobra"
)

var year int
var sourceName string

func init() {
	Cmd.PersistentFlags().IntVar(&year, "year", 2025, "Year of feed (NVD only)")
	Cmd.PersistentFlags().StringVar(&sourceName, "source", "", "Name of vulnerability source")
	Cmd.MarkPersistentFlagRequired("source")
}

// Cmd 是 import 根指令
var Cmd = &cobra.Command{
	Use:   "import",
	Short: "Import vulnerability data",
	Run: func(cmd *cobra.Command, args []string) {
		if sourceName == "" {
			log.Fatalf("❌ --source is required.")
		}

		switch sourceName {
		case "nvd-cve":
			nvdCmd.Run(cmd, args)
		default:
			log.Fatalf("❌ Unsupported source: %s", sourceName)
		}
	},
}
