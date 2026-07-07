-- Create data_sources table
CREATE TABLE IF NOT EXISTS data_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    type VARCHAR(100) NOT NULL,
    provider VARCHAR(100) NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active','delayed','error','disabled','disconnected')),
    events_today BIGINT DEFAULT 0,
    total_events BIGINT DEFAULT 0,
    last_event_at TIMESTAMPTZ NULL,
    last_sync_at TIMESTAMPTZ NULL,
    error_count BIGINT DEFAULT 0,
    delayed_by_minutes INT DEFAULT 0,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_data_sources_org_id ON data_sources(organization_id);
CREATE INDEX IF NOT EXISTS idx_data_sources_status ON data_sources(status);
CREATE INDEX IF NOT EXISTS idx_data_sources_type ON data_sources(type);
CREATE INDEX IF NOT EXISTS idx_data_sources_last_event_at ON data_sources(last_event_at);
