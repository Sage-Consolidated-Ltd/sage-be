CREATE TABLE IF NOT EXISTS log_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    s3_key    TEXT NOT NULL UNIQUE,
    file_class TEXT NOT NULL,  -- json, csv, pcap, syslog etc (derived from extension at upload time)

    source_type  TEXT,  -- user-provided at confirm step e.g. "firewall", "nginx"
    source_id    UUID REFERENCES data_sources(id) ON DELETE SET NULL,
    description  TEXT,
    category TEXT,
    app_or_context TEXT,

    status TEXT NOT NULL CHECK (status IN ('pending', 'uploaded', 'submitted', 'failed')) DEFAULT 'pending',
    -- pending   = presign issued, S3 upload not yet confirmed
    -- uploaded  = confirmed in S3, not yet sent to AI bot
    -- submitted = handed off to AI bot successfully
    -- failed    = something went wrong (see error_message)

    error_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_log_files_user_id         ON log_files(user_id);
CREATE INDEX idx_log_files_organization_id ON log_files(organization_id);
CREATE INDEX idx_log_files_status          ON log_files(status);
CREATE INDEX idx_log_files_created_at      ON log_files(created_at);