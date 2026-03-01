-- Add import_jobs table for tracking CVE import scheduler tasks
CREATE TABLE IF NOT EXISTS import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID NOT NULL,
    source_name TEXT NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    status TEXT NOT NULL CHECK (status IN ('running', 'success', 'failed')),
    records_count INTEGER DEFAULT 0,
    error_message TEXT,
    data_start_date TIMESTAMP WITH TIME ZONE,
    data_end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for querying job history by source
CREATE INDEX IF NOT EXISTS idx_import_jobs_source_id ON import_jobs(source_id);
CREATE INDEX IF NOT EXISTS idx_import_jobs_source_name ON import_jobs(source_name);

-- Index for finding last successful import
CREATE INDEX IF NOT EXISTS idx_import_jobs_status_date ON import_jobs(source_id, status, data_end_date DESC);

-- Index for monitoring running jobs
CREATE INDEX IF NOT EXISTS idx_import_jobs_status ON import_jobs(status) WHERE status = 'running';

COMMENT ON TABLE import_jobs IS 'Tracks history and status of scheduled CVE data imports from various sources';
COMMENT ON COLUMN import_jobs.source_id IS 'Foreign key to cve_sources table';
COMMENT ON COLUMN import_jobs.source_name IS 'Denormalized source name for quick access';
COMMENT ON COLUMN import_jobs.status IS 'Import job status: running, success, or failed';
COMMENT ON COLUMN import_jobs.data_start_date IS 'Start date of the CVE data range being imported';
COMMENT ON COLUMN import_jobs.data_end_date IS 'End date of the CVE data range being imported';
