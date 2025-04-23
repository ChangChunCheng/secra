package importcmd

import "github.com/spf13/cobra"

// Cmd 是 import 根指令
var Cmd = &cobra.Command{
	Use:   "import",
	Short: "Import vulnerability data",
}

func init() {
	nvdCmd.Flags().IntVar(&year, "year", 2024, "Year of NVD CVE data")
	nvdCmd.Flags().StringVar(&sourceName, "source", "", "Name of CVE source (e.g. nvd-cve)")
	nvdCmd.MarkFlagRequired("source")

	Cmd.AddCommand(nvdCmd)
}
