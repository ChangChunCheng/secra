package product

import "github.com/spf13/cobra"

// productCmd is the parent command for product operations.
var Cmd = &cobra.Command{
	Use:   "product",
	Short: "Manage products",
}

func init() {
	Cmd.AddCommand(createProductCmd)
	Cmd.AddCommand(getProductCmd)
	Cmd.AddCommand(listProductCmd)
	Cmd.AddCommand(updateProductCmd)
	Cmd.AddCommand(deleteProductCmd)
}
