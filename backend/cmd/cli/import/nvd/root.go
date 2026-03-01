package nvd

import (
	"context"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
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
	Cmd.AddCommand(v2Nvd)
}

func ensureCveSource(db *bun.DB, name, description, url string) (*model.CVESource, error) {
	ctx := context.Background()
	source := new(model.CVESource)
	err := db.NewSelect().Model(source).Where("name = ?", name).Scan(ctx)
	if err == nil {
		return source, nil
	}

	source = &model.CVESource{
		ID:          uuid.New().String(),
		Name:        name,
		Type:        "nvd",
		URL:         url,
		Description: description,
		Enabled:     true,
	}

	_, err = db.NewInsert().Model(source).Exec(ctx)
	return source, err
}
