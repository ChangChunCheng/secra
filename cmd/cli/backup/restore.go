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

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore from a Parquet-based backup archive",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		tmpDir, err := os.MkdirTemp("", "secra_restore_*")
		if err != nil { log.Fatalf("❌ Failed to create temp dir: %v", err) }
		defer os.RemoveAll(tmpDir)

		log.Printf("📂 Extracting backup archive: %s...", inputFile)
		if err := extractTarGz(inputFile, tmpDir); err != nil {
			log.Fatalf("❌ Failed to extract archive: %v", err)
		}

		tables := []string{"cve_sources", "vendors", "products", "cves"}
		for _, table := range tables {
			parquetFile := filepath.Join(tmpDir, table+".parquet")
			if _, err := os.Stat(parquetFile); os.IsNotExist(err) { continue }
			log.Printf("📥 Importing table [%s]...", table)
			importParquetToTable(cmd.Context(), db, table, parquetFile)
		}

		log.Println("📊 Recalculating daily stats...")
		db.DB.NewRaw(`INSERT INTO daily_cve_counts (day, count)
			SELECT published_at::date as day, count(*) as count FROM cves GROUP BY day
			ON CONFLICT (day) DO UPDATE SET count = EXCLUDED.count`).Exec(cmd.Context())

		log.Println("✅ Restore process completed.")
	},
}

func extractTarGz(gzipFile, destDir string) error {
	f, err := os.Open(gzipFile)
	if err != nil { return err }
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil { return err }
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF { break }
		if err != nil { return err }

		target := filepath.Join(destDir, header.Name)
		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(target, os.FileMode(header.Mode))
			continue
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil { return err }
		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()
	}
	return nil
}

func importParquetToTable(ctx context.Context, db *storage.DBWrapper, tableName string, filePath string) {
	fr, err := local.NewLocalFileReader(filePath)
	if err != nil {
		log.Printf("⚠️ Could not open parquet file: %v", err)
		return
	}
	defer fr.Close()

	switch tableName {
	case "cve_sources":
		pr, _ := reader.NewParquetReader(fr, new(SourceDTO), 4)
		num := int(pr.GetNumRows())
		for i := 0; i < num; i++ {
			row := make([]SourceDTO, 1)
			if err := pr.Read(&row); err != nil || len(row) == 0 { break }
			s := &model.CVESource{ID: row[0].ID, Name: row[0].Name, Type: row[0].Type, URL: row[0].URL, Enabled: true}
			db.DB.NewInsert().Model(s).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
		}
		pr.ReadStop()
	case "vendors":
		pr, _ := reader.NewParquetReader(fr, new(VendorDTO), 4)
		num := int(pr.GetNumRows())
		for i := 0; i < num; i++ {
			row := make([]VendorDTO, 1)
			if err := pr.Read(&row); err != nil || len(row) == 0 { break }
			v := &model.Vendor{ID: row[0].ID, Name: row[0].Name}
			db.DB.NewInsert().Model(v).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
		}
		pr.ReadStop()
	case "products":
		pr, _ := reader.NewParquetReader(fr, new(ProductDTO), 4)
		num := int(pr.GetNumRows())
		for i := 0; i < num; i++ {
			row := make([]ProductDTO, 1)
			if err := pr.Read(&row); err != nil || len(row) == 0 { break }
			p := &model.Product{ID: row[0].ID, VendorID: row[0].VendorID, Name: row[0].Name}
			db.DB.NewInsert().Model(p).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
		}
		pr.ReadStop()
	case "cves":
		pr, _ := reader.NewParquetReader(fr, new(CVEDTO), 4)
		num := int(pr.GetNumRows())
		batchSize := 1000
		var batch []model.CVE
		for i := 0; i < num; i++ {
			row := make([]CVEDTO, 1)
			if err := pr.Read(&row); err != nil || len(row) == 0 { break }
			dto := row[0]
			c := model.CVE{
				ID: dto.ID, SourceID: dto.SourceID, SourceUID: dto.SourceUID, Title: dto.Title, Description: dto.Description,
				PublishedAt: time.UnixMilli(dto.PublishedAt).UTC(), UpdatedAt: time.Now().UTC(),
			}
			if dto.Severity != "" { s := dto.Severity; c.Severity = &s }
			score := dto.CVSSScore; c.CVSSScore = &score
			batch = append(batch, c)
			if len(batch) >= batchSize || i == num-1 {
				db.DB.NewInsert().Model(&batch).On("CONFLICT (id) DO UPDATE SET title = EXCLUDED.title").Exec(ctx)
				batch = nil
			}
		}
		pr.ReadStop()
	}
}
