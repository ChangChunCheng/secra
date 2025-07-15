package resource

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var createCveResourceCmd = &cobra.Command{
	Use:   "create-cve-resource",
	Short: "Create a new CVE resource",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		cveRepo := repo.NewCVESourceRepo(db.DB)
		svc := service.NewCveSourceService(cveRepo)

		name, _ := cmd.Flags().GetString("name")
		ctype, _ := cmd.Flags().GetString("type")
		url, _ := cmd.Flags().GetString("url")
		desc, _ := cmd.Flags().GetString("description")
		cveResource, err := svc.Create(context.Background(), name, url, ctype, desc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created CVE resource: ID=%s Name=%s URL=%s\n", cveResource.ID, cveResource.Name, cveResource.URL)
	},
}

func init() {
	createCveResourceCmd.Flags().String("name", "", "Resource name")
	createCveResourceCmd.Flags().String("type", "", "CVE resource type")
	createCveResourceCmd.Flags().String("url", "", "Resource URL")
	createCveResourceCmd.Flags().String("description", "", "Description")
	createCveResourceCmd.MarkFlagRequired("name")
	createCveResourceCmd.MarkFlagRequired("type")
	createCveResourceCmd.MarkFlagRequired("url")
	createCveResourceCmd.MarkFlagRequired("description")
}
