-- tracks raw JSON submissions (the other endpoint)
CREATE TABLE IF NOT EXISTS json_inputs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    payload         JSONB NOT NULL,
    source_type     TEXT,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS analysis_results (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- only one will be set depending on request_type
    log_file_id  UUID REFERENCES log_files(id) ON DELETE SET NULL,
    json_input_id UUID REFERENCES json_inputs(id) ON DELETE SET NULL,

    -- enforces that exactly one source is always set
    CONSTRAINT chk_single_source CHECK (
        (log_file_id IS NOT NULL AND json_input_id IS NULL) OR
        (log_file_id IS NULL AND json_input_id IS NOT NULL)
    ),

    request_type TEXT NOT NULL CHECK (request_type IN ('file', 'json')),
    log_type     TEXT NOT NULL,
    approach     TEXT NOT NULL,
    overall      TEXT NOT NULL,
    summary      JSONB NOT NULL DEFAULT '{}',
    outcome      JSONB NOT NULL DEFAULT '{}',

    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE IF NOT EXISTS threats (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_id     UUID NOT NULL REFERENCES analysis_results(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    source          TEXT NOT NULL,
    title           TEXT NOT NULL,
    category        TEXT NOT NULL,
    severity        TEXT NOT NULL CHECK (severity IN ('Low', 'Medium', 'High', 'Critical')),
    mitre           TEXT NOT NULL,
    event_count     INT NOT NULL DEFAULT 0,
    time_range      TEXT NOT NULL,
    what_happened   TEXT NOT NULL,
    evidence        JSONB NOT NULL DEFAULT '[]',
    recommendation  TEXT NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_analysis_results_log_file_id  ON analysis_results(log_file_id);
CREATE INDEX idx_analysis_results_json_input_id ON analysis_results(json_input_id);
CREATE INDEX idx_analysis_results_request_type  ON analysis_results(request_type);
CREATE INDEX idx_threats_analysis_id            ON threats(analysis_id);
CREATE INDEX idx_threats_organization_id        ON threats(organization_id);
CREATE INDEX idx_threats_severity               ON threats(severity);
CREATE INDEX idx_threats_mitre                  ON threats(mitre);