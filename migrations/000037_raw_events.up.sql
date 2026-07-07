CREATE TABLE IF NOT EXISTS raw_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source_id UUID REFERENCES data_sources(id) ON DELETE SET NULL,

    -- provider info
    provider VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,

    -- user details
    user_id TEXT,
    user_name TEXT,
    ip_address INET,

    -- app/source context
    application TEXT,

    -- timestamps
    event_timestamp TIMESTAMPTZ NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- processing state
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'processed', 'error')),

    -- worker locking
    locked_at TIMESTAMPTZ,
    locked_by TEXT,

    -- error handling
    error_message TEXT,

    -- raw provider payload (source of truth)
    raw_payload JSONB NOT NULL
);

CREATE INDEX idx_raw_events_processing
ON raw_events (status, locked_at, collected_at);

CREATE INDEX idx_raw_events_pending
ON raw_events (collected_at)
WHERE status = 'pending';

CREATE INDEX idx_raw_events_org_id
ON raw_events (organization_id);

CREATE INDEX idx_raw_events_source_id
ON raw_events (source_id);

CREATE INDEX idx_raw_events_timestamp
ON raw_events (event_timestamp DESC);

CREATE INDEX idx_raw_events_payload_gin
ON raw_events USING GIN (raw_payload);