package resource

import (
	"github.com/spf13/cobra"

	"gitlab.com/jacky850509/secra/cmd/cli/resource/cve"
	"gitlab.com/jacky850509/secra/cmd/cli/resource/cvesource"
	"gitlab.com/jacky850509/secra/cmd/cli/resource/product"
	"gitlab.com/jacky850509/secra/cmd/cli/resource/subscribe"
	"gitlab.com/jacky850509/secra/cmd/cli/resource/vendor"
)

// Cmd is the parent command for resource operations.
var Cmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage CVE resources",
}

func init() {
	Cmd.AddCommand(vendor.Cmd)
	Cmd.AddCommand(product.Cmd)
	Cmd.AddCommand(cvesource.Cmd)
	Cmd.AddCommand(cve.Cmd)
	Cmd.AddCommand(subscribe.Cmd)
}
