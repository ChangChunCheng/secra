package scheduler

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/storage"
	"gitlab.com/jacky850509/secra/internal/util"
)

type NVDImporter struct {
	db      *storage.DBWrapper
	config  *config.AppConfig
	service *service.NVDImportService
}

func NewNVDImporter(db *storage.DBWrapper, cfg *config.AppConfig) *NVDImporter {
	return &NVDImporter{
		db:      db,
		config:  cfg,
		service: service.NewNVDImportService(db.DB, cfg),
	}
}

func (n *NVDImporter) GetSourceName() string {
	return "nvd"
}

func (n *NVDImporter) GetSourceID(ctx context.Context, db *bun.DB) (string, error) {
	return util.SourceID("nvd-v2"), nil
}

func (n *NVDImporter) Import(ctx context.Context, start, end time.Time) (int, error) {
	// Use shared import service with smart interval detection and chunking
	return n.service.ImportDateRange(ctx, start, end, false)
}

func (n *NVDImporter) GetLastImportDate(ctx context.Context, db *bun.DB) (time.Time, error) {
	return n.service.GetLastImportDate(ctx)
}
