package nvd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/importer"
	nvd_v1_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v1"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var v1Nvd = &cobra.Command{
	Use:   "v1",
	Short: "Import CVEs from NVD feeds (v1.1)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg = config.Load()
		db = storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		for year := 2002; year <= time.Now().Year(); year++ {
			log.Printf("Processing year %d...", year)
			if err := ImportNvdYear(cmd.Context(), db.DB, cfg, year); err != nil {
				log.Printf("Failed to import year %d: %v", year, err)
			}
		}
	},
}

func ImportNvdYear(ctx context.Context, db *bun.DB, cfg *config.AppConfig, year int) error {
	filename := fmt.Sprintf("nvdcve-1.1-%d.json.gz", year)
	url := fmt.Sprintf("%snvdcve-1.1-%d.json.gz", cfg.NvdURLv1, year)
	tmpDir := os.TempDir()
	gzPath := filepath.Join(tmpDir, filename)

	if err := downloadFile(url, gzPath); err != nil {
		return err
	}
	defer os.Remove(gzPath)

	f, err := os.Open(gzPath)
	if err != nil { return err }
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil { return err }
	defer gzr.Close()
	
	var feed nvd_v1_parser.Nvdv1CveFeed
	if err := json.NewDecoder(gzr).Decode(&feed); err != nil {
		return err
	}

	source, err := ensureCveSource(db, "nvd-v1", "NVD v1.1 Data Feed", cfg.NvdURLv1)
	if err != nil {
		return err
	}

	cves, err := nvd_v1_parser.ConvertToCVEs(&feed)
	if err != nil {
		return err
	}

	if err := importer.ImportCVEs(db, source.ID, cves); err != nil {
		return err
	}

	v, p, rel := nvd_v1_parser.ExtractVendorsAndProductsFromv1(&feed)
	
	// UUID v5 approach: Use the standardized importer
	return importer.ImportVendorsAndProductsFromv1(db, v, p, rel, nil, nil)
}

func downloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return fmt.Errorf("bad status: %s", resp.Status) }
	out, err := os.Create(dest)
	if err != nil { return err }
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
