-- Create analysis_results table
CREATE TABLE IF NOT EXISTS analysis_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    log_file_id UUID NULL,
    json_input_id UUID NULL,
    request_type VARCHAR(50) NOT NULL DEFAULT 'file',
    log_type VARCHAR(50) NOT NULL DEFAULT '',
    approach TEXT NOT NULL DEFAULT '',
    overall TEXT NOT NULL DEFAULT '',
    summary JSONB NOT NULL DEFAULT '{}'::jsonb,
    outcome JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analysis_results_log_file_id ON analysis_results(log_file_id);
CREATE INDEX IF NOT EXISTS idx_analysis_results_created_at ON analysis_results(created_at);

-- Create threats table
CREATE TABLE IF NOT EXISTS threats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_id UUID NOT NULL REFERENCES analysis_results(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source VARCHAR(255) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    mitre VARCHAR(100) NOT NULL DEFAULT '',
    event_count INT NOT NULL DEFAULT 0,
    time_range VARCHAR(255) NOT NULL DEFAULT '',
    what_happened TEXT NOT NULL DEFAULT '',
    evidence JSONB DEFAULT '[]'::jsonb,
    recommendation TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_threats_analysis_id ON threats(analysis_id);
CREATE INDEX IF NOT EXISTS idx_threats_organization_id ON threats(organization_id);
CREATE INDEX IF NOT EXISTS idx_threats_severity ON threats(severity);
CREATE INDEX IF NOT EXISTS idx_threats_created_at ON threats(created_at);
