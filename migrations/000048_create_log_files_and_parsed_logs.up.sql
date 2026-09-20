CREATE TABLE IF NOT EXISTS log_files (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id    UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    s3_key             TEXT NOT NULL UNIQUE,
    file_class         TEXT NOT NULL,
    event_count        INT DEFAULT 0,
    processed_at       TIMESTAMPTZ,
    source_type        TEXT,
    source_id          UUID REFERENCES data_sources(id) ON DELETE SET NULL,
    description        TEXT,
    category           TEXT,
    app_or_context     TEXT,
    status             TEXT NOT NULL CHECK (
        status IN (
            'pending',
            'uploaded',
            'submitted',
            'processing',
            'parsed',
            'analyzed',
            'failed',
            'retrying'
        )
    ) DEFAULT 'pending',
    error_message      TEXT,
    detected_type      TEXT,
    user_selected_type TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_log_files_user_id         ON log_files(user_id);
CREATE INDEX IF NOT EXISTS idx_log_files_organization_id ON log_files(organization_id);
CREATE INDEX IF NOT EXISTS idx_log_files_status          ON log_files(status);
CREATE INDEX IF NOT EXISTS idx_log_files_created_at      ON log_files(created_at);

CREATE TABLE IF NOT EXISTS parsed_logs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    file_id        UUID NOT NULL REFERENCES log_files(id) ON DELETE CASCADE,
    timestamp      TIMESTAMPTZ,
    level          TEXT,
    message        TEXT,
    raw_json       JSONB,
    search_vector  tsvector GENERATED ALWAYS AS (to_tsvector('english', coalesce(message, ''))) STORED,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_parsed_logs_file_id        ON parsed_logs(file_id);
CREATE INDEX IF NOT EXISTS idx_parsed_logs_data_source_id ON parsed_logs(data_source_id);
CREATE INDEX IF NOT EXISTS idx_parsed_logs_timestamp      ON parsed_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_parsed_logs_search         ON parsed_logs USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_parsed_logs_raw_json       ON parsed_logs USING GIN (raw_json jsonb_path_ops);
CREATE INDEX IF NOT EXISTS idx_parsed_logs_filters        ON parsed_logs (data_source_id, timestamp DESC, level);

-- Add foreign key constraint to analysis_results if it exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_analysis_results_log_file_id'
    ) THEN
        ALTER TABLE analysis_results
        ADD CONSTRAINT fk_analysis_results_log_file_id
        FOREIGN KEY (log_file_id) REFERENCES log_files(id) ON DELETE SET NULL;
    END IF;
END $$;
