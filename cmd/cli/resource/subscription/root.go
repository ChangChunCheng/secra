package subscription

import "github.com/spf13/cobra"

// Cmd is the parent command for resource operations.
var Cmd = &cobra.Command{
	Use:   "subscription",
	Short: "Manage subscriptions",
}

func init() {
	Cmd.AddCommand(subscriptionCveSourceCmd)
	Cmd.AddCommand(subscriptionVendorCmd)
	Cmd.AddCommand(subscriptionProductCmd)
	Cmd.AddCommand(deleteSubscriptionCmd)
}
