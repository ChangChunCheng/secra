package model

import (
	"time"

	"github.com/uptrace/bun"
)

// ImportJob records the import task history for each CVE source
type ImportJob struct {
	bun.BaseModel `bun:"table:import_jobs"`

	ID            string     `bun:",pk,notnull,nullzero,default:gen_random_uuid()" json:"id"`
	SourceID      string     `bun:",notnull" json:"source_id"`        // FK to cve_sources
	SourceName    string     `bun:",notnull" json:"source_name"`      // Denormalized for quick access
	StartTime     time.Time  `bun:",notnull" json:"start_time"`       // Import start time
	EndTime       *time.Time `bun:",nullzero" json:"end_time"`        // Import end time (null if still running)
	Status        string     `bun:",notnull" json:"status"`           // running, success, failed
	RecordsCount  int        `bun:",default:0" json:"records_count"`  // Number of CVEs imported
	ErrorMessage  string     `bun:",nullzero" json:"error_message"`   // Error details if failed
	DataStartDate time.Time  `bun:",nullzero" json:"data_start_date"` // Data range start (e.g., 2024-01-01)
	DataEndDate   time.Time  `bun:",nullzero" json:"data_end_date"`   // Data range end (e.g., 2024-01-31)
	CreatedAt     time.Time  `bun:",default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time  `bun:",default:current_timestamp" json:"updated_at"`
}

// Import job statuses
const (
	JobStatusRunning = "running"
	JobStatusSuccess = "success"
	JobStatusFailed  = "failed"
)
