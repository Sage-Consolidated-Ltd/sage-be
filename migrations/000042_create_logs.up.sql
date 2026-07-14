ALTER TABLE log_files
    ADD COLUMN IF NOT EXISTS detected_type      TEXT,
    ADD COLUMN IF NOT EXISTS user_selected_type TEXT,
    ADD COLUMN IF NOT EXISTS event_count        INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS processed_at       TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS error_message      TEXT;

ALTER TABLE log_files DROP CONSTRAINT IF EXISTS log_files_status_check;
 
ALTER TABLE log_files
    ADD CONSTRAINT log_files_status_check
    CHECK (
        status IN (
            'submitted',
            'pending',
            'uploaded',
            'processing',
            'parsed',
            'analyzed',
            'failed',
            'retrying'
        )
    );

CREATE TABLE IF NOT EXISTS parsed_logs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID NOT NULL REFERENCES data_sources(id),
    file_id        UUID NOT NULL REFERENCES log_files(id),
    timestamp      TIMESTAMPTZ,
    level          TEXT,
    message        TEXT,
    raw_json       JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_parsed_logs_file_id ON parsed_logs(file_id);
CREATE INDEX IF NOT EXISTS idx_parsed_logs_data_source_id ON parsed_logs(data_source_id);
CREATE INDEX IF NOT EXISTS idx_parsed_logs_timestamp ON parsed_logs(timestamp);

ALTER TABLE data_sources
    ADD COLUMN IF NOT EXISTS total_events   BIGINT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS events_today   BIGINT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_event_at  TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_data_sources_org_provider
    ON data_sources(organization_id, provider);