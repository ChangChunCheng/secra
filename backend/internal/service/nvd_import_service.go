package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/fetcher"
	"gitlab.com/jacky850509/secra/internal/importer"
	"gitlab.com/jacky850509/secra/internal/model"
	nvd_v2_parser "gitlab.com/jacky850509/secra/internal/parser/nvd/v2"
	"gitlab.com/jacky850509/secra/internal/util"
)

// NVDImportService handles NVD data import with smart interval detection and chunking
type NVDImportService struct {
	db        *bun.DB
	cfg       *config.AppConfig
	lastReqAt time.Time
}

// NewNVDImportService creates a new NVD import service
func NewNVDImportService(db *bun.DB, cfg *config.AppConfig) *NVDImportService {
	return &NVDImportService{
		db:  db,
		cfg: cfg,
	}
}

type interval struct {
	start time.Time
	end   time.Time
}

// ImportDateRange imports CVEs from NVD for the specified date range
// If force=true, reimports all data; otherwise only fills gaps
func (s *NVDImportService) ImportDateRange(ctx context.Context, start, end time.Time, force bool) (int, error) {
	var gaps []interval
	if force {
		gaps = []interval{{start: start, end: end}}
	} else {
		log.Printf("🔍 Scanning database for missing intervals between %s and %s...", start.Format(time.DateOnly), end.Format(time.DateOnly))
		gaps = s.findMissingIntervals(ctx, start, end)
	}

	if len(gaps) == 0 {
		log.Println("✅ No missing intervals found. Data is up to date.")
		return 0, nil
	}

	// Collect all imported CVEs for batch notification
	allImportedCVEs := []model.CVE{}
	totalCount := 0

	for _, gap := range gaps {
		log.Printf("📅 Processing optimized gap: %s to %s", gap.start.Format(time.DateOnly), gap.end.Format(time.DateOnly))
		chunkSize := 30 * 24 * time.Hour
		for currentStart := gap.start; currentStart.Before(gap.end); {
			currentEnd := currentStart.Add(chunkSize)
			if currentEnd.After(gap.end) {
				currentEnd = gap.end
			}
			importedCVEs := s.importChunk(ctx, currentStart, currentEnd)
			allImportedCVEs = append(allImportedCVEs, importedCVEs...)
			totalCount += len(importedCVEs)
			currentStart = currentEnd
		}
	}

	// Send ONE notification with all imported CVEs
	if len(allImportedCVEs) > 0 {
		log.Printf("📧 Sending consolidated notification for %d CVEs...", len(allImportedCVEs))
		notifier := NewNotificationService(s.cfg.SMTPConfig, s.db)
		notifier.ProcessBatch(ctx, allImportedCVEs)
	}

	log.Printf("✅ All requested intervals processed. Total: %d CVEs", totalCount)
	return totalCount, nil
}

// findMissingIntervals detects missing date ranges in the database
func (s *NVDImportService) findMissingIntervals(ctx context.Context, start, end time.Time) []interval {
	type DateRow struct {
		Day time.Time `bun:"day"`
	}
	var existingDates []DateRow
	s.db.NewSelect().Table("cves").ColumnExpr("published_at::date as day").
		Where("published_at >= ? AND published_at <= ?", start, end).
		Group("day").Order("day ASC").Scan(ctx, &existingDates)

	dateMap := make(map[string]bool)
	for _, d := range existingDates {
		dateMap[d.Day.Format(time.DateOnly)] = true
	}

	var gaps []interval
	var currentGap *interval

	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		dayStr := current.Format(time.DateOnly)
		if !dateMap[dayStr] {
			if currentGap == nil {
				currentGap = &interval{start: current}
			}
			currentGap.end = current.AddDate(0, 0, 1)
		} else {
			if currentGap != nil {
				hasHoleSoon := false
				for lookAhead := 1; lookAhead <= 7; lookAhead++ {
					checkDay := current.AddDate(0, 0, lookAhead)
					if checkDay.After(end) {
						break
					}
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
	if currentGap != nil {
		gaps = append(gaps, *currentGap)
	}
	return gaps
}

// importChunk imports a single chunk of data (up to 30 days)
func (s *NVDImportService) importChunk(ctx context.Context, start, end time.Time) []model.CVE {
	pageSize := 2000
	startIndex := 0
	sourceID := util.SourceID("nvd-v2")
	source, _ := s.ensureCveSource(ctx, sourceID)

	allChunkCVEs := []model.CVE{}

	for {
		s.waitThrottle()
		data, err := fetcher.FetchNvdv2Feed(s.cfg.NvdURLv2, fetcher.NvdV2QueryParams{
			PubStartDate:   start,
			PubEndDate:     end,
			StartIndex:     startIndex,
			ResultsPerPage: pageSize,
			ApiKey:         s.cfg.NvdAPIKey,
			MaxRetries:     s.cfg.NvdMaxRetries,
			RetryDelay:     s.cfg.NvdRetryDelay,
		})
		if err != nil {
			log.Printf("⚠️ Failed to fetch NVD feed: %v", err)
			return allChunkCVEs
		}

		var feed nvd_v2_parser.Nvdv2CveFeed
		if err := json.Unmarshal(data, &feed); err != nil {
			log.Printf("⚠️ Failed to parse NVD feed: %v", err)
			return allChunkCVEs
		}
		if len(feed.Vulnerabilities) == 0 {
			break
		}

		cves, _ := nvd_v2_parser.ConvertToCVEsFromV2(&feed)
		v, p, rel, ref, w := nvd_v2_parser.ExtractAllFromV2(&feed)

		cveMap := make(map[string]string)
		for _, c := range cves {
			cveMap[c.SourceUID] = c.ID
		}

		// STEP 1: Import CVEs FIRST (to satisfy Foreign Keys)
		importer.ImportCVEs(s.db, source.ID, cves)

		// STEP 2: Establish all relations
		importer.ImportVendorsAndProductsFromv2(s.db, v, p, rel, nil, nil)
		importer.ImportReferences(s.db, ref, cveMap)
		importer.ImportWeaknesses(s.db, w, cveMap)

		allChunkCVEs = append(allChunkCVEs, cves...)

		startIndex += len(feed.Vulnerabilities)
		if startIndex >= feed.TotalResults {
			break
		}
	}

	return allChunkCVEs
}

// waitThrottle implements API rate limiting
func (s *NVDImportService) waitThrottle() {
	delay := 6 * time.Second
	if s.cfg.NvdAPIKey != "" {
		delay = 1 * time.Second
	}
	elapsed := time.Since(s.lastReqAt)
	if elapsed < delay {
		wait := delay - elapsed
		log.Printf("⏱️ Throttling: waiting %v...", wait)
		time.Sleep(wait)
	}
	s.lastReqAt = time.Now()
}

// ensureCveSource ensures the CVE source exists in database
func (s *NVDImportService) ensureCveSource(ctx context.Context, sourceID string) (*model.CVESource, error) {
	source := new(model.CVESource)
	err := s.db.NewSelect().Model(source).Where("id = ?", sourceID).Scan(ctx)
	if err == nil {
		return source, nil
	}

	source = &model.CVESource{
		ID:          sourceID,
		Name:        "nvd-v2",
		Type:        "nvd",
		URL:         s.cfg.NvdURLv2,
		Description: "NVD v2 API",
		Enabled:     true,
	}
	_, err = s.db.NewInsert().Model(source).Exec(ctx)
	return source, err
}

// GetLastImportDate returns the most recent import date
func (s *NVDImportService) GetLastImportDate(ctx context.Context) (time.Time, error) {
	sourceID := util.SourceID("nvd-v2")
	var job model.ImportJob
	err := s.db.NewSelect().
		Model(&job).
		Where("source_id = ?", sourceID).
		Where("status = ?", model.JobStatusSuccess).
		Order("data_end_date DESC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		return time.Time{}, nil
	}

	return job.DataEndDate, nil
}
