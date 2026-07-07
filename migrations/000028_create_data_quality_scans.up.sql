-- Create data_quality_scans table
CREATE TABLE IF NOT EXISTS data_quality_scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('running','completed','failed')),
    quality_score INT NULL,
    parsing_errors BIGINT NULL,
    missing_fields_percentage DECIMAL(5,2) NULL,
    duplicate_events_percentage DECIMAL(5,2) NULL,
    unmapped_logs_count BIGINT NULL,
    ai_summary TEXT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_data_quality_scans_org_id ON data_quality_scans(organization_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_scans_status ON data_quality_scans(status);
CREATE INDEX IF NOT EXISTS idx_data_quality_scans_created_at ON data_quality_scans(created_at);
CREATE INDEX IF NOT EXISTS idx_data_quality_scans_updated_at ON data_quality_scans(updated_at);
