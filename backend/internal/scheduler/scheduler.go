package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/storage"
)

// Importer defines the interface that all CVE source importers must implement
type Importer interface {
	// GetSourceName returns the human-readable name of the source (e.g., "nvd", "osv")
	GetSourceName() string

	// GetSourceID returns the database ID for this source
	GetSourceID(ctx context.Context, db *bun.DB) (string, error)

	// Import executes the import for the specified date range
	// Returns the number of records imported and any error
	Import(ctx context.Context, start, end time.Time) (int, error)

	// GetLastImportDate returns the most recent successful import date
	GetLastImportDate(ctx context.Context, db *bun.DB) (time.Time, error)
}

// Scheduler manages scheduled CVE imports from multiple sources
type Scheduler struct {
	db        *storage.DBWrapper
	config    *config.AppConfig
	cron      *cron.Cron
	importers map[string]Importer
	mu        sync.RWMutex
	isRunning bool
}

// NewScheduler creates a new scheduler instance
func NewScheduler(db *storage.DBWrapper, cfg *config.AppConfig) *Scheduler {
	return &Scheduler{
		db:        db,
		config:    cfg,
		cron:      cron.New(cron.WithSeconds()), // Support seconds in cron format
		importers: make(map[string]Importer),
		isRunning: false,
	}
}

// RegisterImporter adds a new importer to the scheduler
func (s *Scheduler) RegisterImporter(imp Importer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sourceName := imp.GetSourceName()
	s.importers[sourceName] = imp
	log.Printf("✅ Registered importer: %s", sourceName)
}

// Start begins the scheduled import process
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	log.Println("🚀 Starting CVE Import Scheduler...")

	// Run backfill for all sources before starting regular schedule
	if err := s.backfillAllSources(ctx); err != nil {
		log.Printf("⚠️ Backfill encountered errors: %v", err)
	}

	// Schedule regular imports
	_, err := s.cron.AddFunc(s.config.ImportSchedule, func() {
		s.runAllImports(context.Background())
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.cron.Start()
	log.Printf("✅ Scheduler started with schedule: %s", s.config.ImportSchedule)

	return nil
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	log.Println("🛑 Stopping CVE Import Scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.isRunning = false
	log.Println("✅ Scheduler stopped")
}

// backfillAllSources checks each source for missing data and imports it
func (s *Scheduler) backfillAllSources(ctx context.Context) error {
	s.mu.RLock()
	importers := make([]Importer, 0, len(s.importers))
	for _, imp := range s.importers {
		importers = append(importers, imp)
	}
	s.mu.RUnlock()

	for _, imp := range importers {
		if err := s.backfillSource(ctx, imp); err != nil {
			log.Printf("⚠️ Backfill failed for %s: %v", imp.GetSourceName(), err)
		}
	}

	return nil
}

// backfillSource imports missing data for a single source
func (s *Scheduler) backfillSource(ctx context.Context, imp Importer) error {
	sourceName := imp.GetSourceName()
	log.Printf("🔍 Checking backfill for source: %s", sourceName)

	// Get last successful import date
	lastImport, err := imp.GetLastImportDate(ctx, s.db.DB)
	if err != nil {
		return fmt.Errorf("failed to get last import date: %w", err)
	}

	now := time.Now().UTC()

	// If never imported, start from 1 year ago (reasonable default for security monitoring)
	if lastImport.IsZero() {
		lastImport = now.AddDate(-1, 0, 0)
		log.Printf("📅 No previous import found for %s, starting from %s", sourceName, lastImport.Format(time.DateOnly))
	}

	// Calculate days since last import
	daysSince := int(now.Sub(lastImport).Hours() / 24)

	if daysSince <= 0 {
		log.Printf("✅ No backfill needed for %s (last import: %s)", sourceName, lastImport.Format(time.DateOnly))
		return nil
	}

	log.Printf("📥 Backfilling %d days for %s (%s to %s)", daysSince, sourceName,
		lastImport.Format(time.DateOnly), now.Format(time.DateOnly))

	// Execute backfill import
	return s.executeImport(ctx, imp, lastImport, now)
}

// runAllImports executes scheduled import for all registered sources
func (s *Scheduler) runAllImports(ctx context.Context) {
	s.mu.RLock()
	importers := make([]Importer, 0, len(s.importers))
	for _, imp := range s.importers {
		importers = append(importers, imp)
	}
	s.mu.RUnlock()

	log.Println("⏰ Scheduled import triggered")

	for _, imp := range importers {
		// Import today's CVEs
		now := time.Now().UTC()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		if err := s.executeImport(ctx, imp, today, now); err != nil {
			log.Printf("⚠️ Scheduled import failed for %s: %v", imp.GetSourceName(), err)
		}
	}
}

// executeImport runs an import and records it in import_jobs
func (s *Scheduler) executeImport(ctx context.Context, imp Importer, start, end time.Time) error {
	sourceName := imp.GetSourceName()
	sourceID, err := imp.GetSourceID(ctx, s.db.DB)
	if err != nil {
		return fmt.Errorf("failed to get source ID: %w", err)
	}

	// Create import job record
	job := &model.ImportJob{
		SourceID:      sourceID,
		SourceName:    sourceName,
		Status:        model.JobStatusRunning,
		DataStartDate: start,
		DataEndDate:   end,
		StartTime:     time.Now(),
	}

	if _, err := s.db.DB.NewInsert().Model(job).Exec(ctx); err != nil {
		return fmt.Errorf("failed to create import job: %w", err)
	}

	// Execute import
	log.Printf("🔄 Starting import for %s (%s to %s)", sourceName, start.Format(time.DateOnly), end.Format(time.DateOnly))

	count, importErr := imp.Import(ctx, start, end)

	// Update job record
	endTime := time.Now()
	job.EndTime = &endTime
	job.RecordsCount = count

	if importErr != nil {
		job.Status = model.JobStatusFailed
		job.ErrorMessage = importErr.Error()
		log.Printf("❌ Import failed for %s: %v", sourceName, importErr)
	} else {
		job.Status = model.JobStatusSuccess
		log.Printf("✅ Import completed for %s: %d records", sourceName, count)
	}

	if _, err := s.db.DB.NewUpdate().Model(job).WherePK().Exec(ctx); err != nil {
		log.Printf("⚠️ Failed to update import job record: %v", err)
	}

	return importErr
}
