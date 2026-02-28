package nvd

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/fetcher"
	"gitlab.com/jacky850509/secra/internal/importer"
	nvd_v2_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v2"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	startDate string
	endDate   string
	apiKey    string
	force     bool
	lastReqAt time.Time
)

func init() {
	v2Nvd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD) [required]")
	v2Nvd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	v2Nvd.Flags().StringVar(&apiKey, "apikey", "", "Optional NVD API key")
	v2Nvd.Flags().BoolVarP(&force, "force", "f", false, "Force re-import even if data exists for the date")
	v2Nvd.MarkFlagRequired("start")
}

type interval struct {
	start time.Time
	end   time.Time
}

var v2Nvd = &cobra.Command{
	Use:   "v2",
	Short: "Import CVEs from NVD API v2 with smart interval merging",
	Run: func(cmd *cobra.Command, args []string) {
		start, _ := time.Parse(time.DateOnly, startDate)
		end := time.Now().UTC()
		if endDate != "" {
			end, _ = time.Parse(time.DateOnly, endDate)
		}

		cfg = config.Load()
		db = storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		var gaps []interval
		if force {
			gaps = []interval{{start: start, end: end}}
		} else {
			log.Printf("🔍 Scanning database for missing intervals between %s and %s...", start.Format(time.DateOnly), end.Format(time.DateOnly))
			gaps = findMissingIntervals(cmd.Context(), db.DB, start, end)
		}

		if len(gaps) == 0 {
			log.Println("✅ No missing intervals found. Data is up to date.")
			return
		}

		for _, gap := range gaps {
			log.Printf("📅 Processing optimized gap: %s to %s", gap.start.Format(time.DateOnly), gap.end.Format(time.DateOnly))
			
			chunkSize := 30 * 24 * time.Hour
			for currentStart := gap.start; currentStart.Before(gap.end); {
				currentEnd := currentStart.Add(chunkSize)
				if currentEnd.After(gap.end) {
					currentEnd = gap.end
				}

				ImportNvdv2Chunk(cmd.Context(), db.DB, cfg, currentStart, currentEnd)
				currentStart = currentEnd
			}
		}
		log.Println("✅ All requested intervals processed.")
	},
}

func waitThrottle(cfg *config.AppConfig) {
	delay := 6 * time.Second
	if apiKey != "" || cfg.NvdAPIKey != "" {
		delay = 1 * time.Second
	}
	
	elapsed := time.Since(lastReqAt)
	if elapsed < delay {
		wait := delay - elapsed
		log.Printf("⏱️ Throttling: waiting %v...", wait)
		time.Sleep(wait)
	}
	lastReqAt = time.Now()
}

func findMissingIntervals(ctx context.Context, db *bun.DB, start, end time.Time) []interval {
	type DateRow struct { Day time.Time `bun:"day"` }
	var existingDates []DateRow
	db.NewSelect().Table("cves").ColumnExpr("published_at::date as day").
		Where("published_at >= ? AND published_at <= ?", start, end).
		Group("day").Order("day ASC").Scan(ctx, &existingDates)

	dateMap := make(map[string]bool)
	for _, d := range existingDates { dateMap[d.Day.Format(time.DateOnly)] = true }

	var gaps []interval
	var currentGap *interval

	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		dayStr := current.Format(time.DateOnly)
		if !dateMap[dayStr] {
			if currentGap == nil { currentGap = &interval{start: current} }
			currentGap.end = current.AddDate(0, 0, 1)
		} else {
			// GREEDY MERGE: Only close the gap if we have more than 7 days of existing data
			// This reduces the number of small API requests.
			if currentGap != nil {
				// Check if there's another missing day within next 7 days
				hasHoleSoon := false
				for lookAhead := 1; lookAhead <= 7; lookAhead++ {
					checkDay := current.AddDate(0, 0, lookAhead)
					if checkDay.After(end) { break }
					if !dateMap[checkDay.Format(time.DateOnly)] {
						hasHoleSoon = true
						break
					}
				}
				
				if !hasHoleSoon {
					gaps = append(gaps, *currentGap)
					currentGap = nil
				}
			}
		}
	}
	if currentGap != nil { gaps = append(gaps, *currentGap) }
	return gaps
}

func ImportNvdv2Chunk(ctx context.Context, db *bun.DB, cfg *config.AppConfig, start, end time.Time) {
	pageSize := 2000
	startIndex := 0

	source, _ := ensureCveSource(db, "nvd-v2", "NVD v2 API", cfg.NvdURLv2)

	for {
		waitThrottle(cfg) // Ensure every fetch call is throttled

		data, err := fetcher.FetchNvdv2Feed(cfg.NvdURLv2, fetcher.NvdV2QueryParams{
			PubStartDate:   start,
			PubEndDate:     end,
			StartIndex:     startIndex,
			ResultsPerPage: pageSize,
			ApiKey:         apiKey,
			MaxRetries:     cfg.NvdMaxRetries,
			RetryDelay:     cfg.NvdRetryDelay,
		})
		
		if err != nil {
			log.Printf("❌ Skipping chunk %s-%s due to error: %v", start.Format(time.DateOnly), end.Format(time.DateOnly), err)
			return
		}

		var feed nvd_v2_parser.Nvdv2CveFeed
		json.Unmarshal(data, &feed)

		if len(feed.Vulnerabilities) == 0 { break }

		cves, _ := nvd_v2_parser.ConvertToCVEsFromV2(&feed)
		importer.ImportCVEs(db, source.ID, cves)

		v, p, rel, ref, w := nvd_v2_parser.ExtractAllFromV2(&feed)
		importer.ImportVendorsAndProductsFromv2(db, v, nil, nil, nil, nil)
		
		vendorMap, _ := importer.BuildVendorMap(db)
		for i := range p {
			if realID, ok := vendorMap[p[i].VendorID]; ok { p[i].VendorID = realID }
		}
		importer.ImportVendorsAndProductsFromv2(db, nil, p, nil, nil, nil)

		productMap, _ := importer.BuildProductMap(db)
		uids := make([]string, 0, len(cves))
		for _, c := range cves { uids = append(uids, c.SourceUID) }
		cveMap, _ := importer.BuildCveMap(db, uids)

		importer.ImportVendorsAndProductsFromv2(db, nil, nil, rel, cveMap, productMap)
		importer.ImportReferences(db, ref, cveMap)
		importer.ImportWeaknesses(db, w, cveMap)

		startIndex += len(feed.Vulnerabilities)
		if startIndex >= feed.TotalResults { break }
	}
}
