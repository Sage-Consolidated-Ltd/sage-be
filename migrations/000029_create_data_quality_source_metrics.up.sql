-- Create data_quality_source_metrics table
CREATE TABLE IF NOT EXISTS data_quality_source_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    scan_id UUID NOT NULL REFERENCES data_quality_scans(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    parsing_errors BIGINT NULL DEFAULT 0,
    missing_fields_percentage DECIMAL(5,2) NULL DEFAULT 0.0,
    unmapped_events BIGINT NULL DEFAULT 0,
    duplicate_percentage DECIMAL(5,2) NULL DEFAULT 0.0,
    status VARCHAR(20) NOT NULL CHECK (status IN ('good','warning','partial','error')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_data_quality_source_metrics_scan_id ON data_quality_source_metrics(scan_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_source_metrics_source_id ON data_quality_source_metrics(source_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_source_metrics_status ON data_quality_source_metrics(status);
