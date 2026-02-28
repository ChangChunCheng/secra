package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var inputFile string

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore from a Parquet-based backup archive",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile = args[0]
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		tmpDir, err := os.MkdirTemp("", "secra_restore_*")
		if err != nil {
			log.Fatalf("❌ Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		log.Printf("📂 Extracting backup archive: %s...", inputFile)
		if err := extractTarGz(inputFile, tmpDir); err != nil {
			log.Fatalf("❌ Failed to extract archive: %v", err)
		}

		// 1. Restore Tables (Order matters for FK constraints)
		tables := []string{"vendors", "products", "cves"}
		for _, table := range tables {
			parquetFile := filepath.Join(tmpDir, table+".parquet")
			if _, err := os.Stat(parquetFile); os.IsNotExist(err) {
				continue
			}
			log.Printf("📥 Importing table [%s]...", table)
			if err := importParquetToTable(cmd.Context(), db, table, parquetFile); err != nil {
				log.Printf("⚠️ Failed to import %s: %v", table, err)
			}
		}

		log.Println("✅ Restore process completed.")
	},
}

func extractTarGz(gzipFile, destDir string) error {
	f, err := os.Open(gzipFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(destDir, header.Name)
		f, _ := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		io.Copy(f, tr)
		f.Close()
	}
	return nil
}

func importParquetToTable(ctx context.Context, db *storage.DBWrapper, tableName string, filePath string) error {
	fr, err := local.NewLocalFileReader(filePath)
	if err != nil {
		return err
	}
	defer fr.Close()

	switch tableName {
	case "vendors":
		pr, _ := reader.NewParquetReader(fr, new(VendorDTO), 4)
		num := int(pr.GetNumRows())
		for i := 0; i < num; i++ {
			row := make([]VendorDTO, 1)
			pr.Read(&row)
			v := &model.Vendor{ID: row[0].ID, Name: row[0].Name}
			db.DB.NewInsert().Model(v).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
		}
		pr.ReadStop()
	case "products":
		pr, _ := reader.NewParquetReader(fr, new(ProductDTO), 4)
		num := int(pr.GetNumRows())
		for i := 0; i < num; i++ {
			row := make([]ProductDTO, 1)
			pr.Read(&row)
			p := &model.Product{ID: row[0].ID, VendorID: row[0].VendorID, Name: row[0].Name}
			db.DB.NewInsert().Model(p).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
		}
		pr.ReadStop()
	case "cves":
		pr, _ := reader.NewParquetReader(fr, new(CVEDTO), 4)
		num := int(pr.GetNumRows())
		for i := 0; i < num; i++ {
			row := make([]CVEDTO, 1)
			pr.Read(&row)
			dto := row[0]
			c := &model.CVE{
				ID:          dto.ID,
				SourceID:    dto.SourceID,
				SourceUID:   dto.SourceUID,
				Title:       dto.Title,
				Description: dto.Description,
				PublishedAt: time.UnixMilli(dto.PublishedAt),
				UpdatedAt:   time.Now(),
			}
			if dto.Severity != "" { c.Severity = &dto.Severity }
			f := dto.CVSSScore
			c.CVSSScore = &f
			db.DB.NewInsert().Model(c).On("CONFLICT (id) DO UPDATE SET title = EXCLUDED.title").Exec(ctx)
		}
		pr.ReadStop()
	}
	return nil
}
