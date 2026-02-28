package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var outputFile string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Parquet-based backup",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		timestamp := time.Now().Format("20060102_150405")
		if outputFile == "" { outputFile = fmt.Sprintf("secra_backup_%s.tar.gz", timestamp) }

		tmpDir, _ := os.MkdirTemp("", "secra_backup_*")
		defer os.RemoveAll(tmpDir)

		log.Printf("📦 Starting backup process to %s...", outputFile)

		// Added cve_products to the list
		tables := []string{"cve_sources", "vendors", "products", "cves", "cve_products", "users", "subscriptions"}
		for _, table := range tables {
			parquetFile := filepath.Join(tmpDir, table+".parquet")
			log.Printf("📄 Exporting table [%s]...", table)
			exportTableToParquet(cmd.Context(), db, table, parquetFile)
		}

		createTarGz(outputFile, tmpDir)
		log.Printf("✅ Backup successfully created: %s", outputFile)
	},
}

func init() {
	createCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output filename (.tar.gz)")
}

type SourceDTO struct {
	ID   string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Name string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
	Type string `parquet:"name=type, type=BYTE_ARRAY, convertedtype=UTF8"`
	URL  string `parquet:"name=url, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type VendorDTO struct {
	ID   string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Name string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type ProductDTO struct {
	ID       string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	VendorID string `parquet:"name=vendor_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Name     string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type CVEDTO struct {
	ID          string  `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	SourceID    string  `parquet:"name=source_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	SourceUID   string  `parquet:"name=source_uid, type=BYTE_ARRAY, convertedtype=UTF8"`
	Title       string  `parquet:"name=title, type=BYTE_ARRAY, convertedtype=UTF8"`
	Description string  `parquet:"name=description, type=BYTE_ARRAY, convertedtype=UTF8"`
	Severity    string  `parquet:"name=severity, type=BYTE_ARRAY, convertedtype=UTF8"`
	CVSSScore   float64 `parquet:"name=cvss_score, type=DOUBLE"`
	PublishedAt int64   `parquet:"name=published_at, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
}

type CVEProductDTO struct {
	CVEID     string `parquet:"name=cve_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	ProductID string `parquet:"name=product_id, type=BYTE_ARRAY, convertedtype=UTF8"`
}

func exportTableToParquet(ctx context.Context, db *storage.DBWrapper, tableName string, filePath string) {
	fw, _ := local.NewLocalFileWriter(filePath)
	defer fw.Close()

	switch tableName {
	case "cve_sources":
		var items []model.CVESource
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(SourceDTO), 4)
		for _, item := range items {
			pw.Write(SourceDTO{ID: item.ID, Name: item.Name, Type: item.Type, URL: item.URL})
		}
		pw.WriteStop()
	case "vendors":
		var items []model.Vendor
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(VendorDTO), 4)
		for _, item := range items {
			pw.Write(VendorDTO{ID: item.ID, Name: item.Name})
		}
		pw.WriteStop()
	case "products":
		var items []model.Product
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(ProductDTO), 4)
		for _, item := range items {
			pw.Write(ProductDTO{ID: item.ID, VendorID: item.VendorID, Name: item.Name})
		}
		pw.WriteStop()
	case "cves":
		var items []model.CVE
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(CVEDTO), 4)
		for _, item := range items {
			dto := CVEDTO{ID: item.ID, SourceID: item.SourceID, SourceUID: item.SourceUID, Title: item.Title, Description: item.Description, PublishedAt: item.PublishedAt.UnixMilli()}
			if item.Severity != nil { dto.Severity = *item.Severity }
			if item.CVSSScore != nil { dto.CVSSScore = *item.CVSSScore }
			pw.Write(dto)
		}
		pw.WriteStop()
	case "cve_products":
		var items []model.CVEProduct
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(CVEProductDTO), 4)
		for _, item := range items {
			pw.Write(CVEProductDTO{CVEID: item.CVEID, ProductID: item.ProductID})
		}
		pw.WriteStop()
	}
}

func createTarGz(outputFile string, srcDir string) {
	out, _ := os.Create(outputFile)
	defer out.Close()
	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	files, _ := os.ReadDir(srcDir)
	for _, f := range files {
		info, _ := f.Info()
		header, _ := tar.FileInfoHeader(info, "")
		header.Name = f.Name()
		tw.WriteHeader(header)
		file, _ := os.Open(filepath.Join(srcDir, f.Name()))
		io.Copy(tw, file)
		file.Close()
	}
}
