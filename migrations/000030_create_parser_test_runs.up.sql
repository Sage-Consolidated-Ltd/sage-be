-- Create parser_test_runs table
CREATE TABLE IF NOT EXISTS parser_test_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    parser_id UUID NULL REFERENCES parsers(id) ON DELETE SET NULL,
    sample_log TEXT NULL,
    raw_payload JSONB DEFAULT '{}'::jsonb,
    parsed_output JSONB NOT NULL,
    normalized_output JSONB DEFAULT '{}'::jsonb,
    errors JSONB DEFAULT '[]'::jsonb,
    success BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_parser_test_runs_org_id ON parser_test_runs(organization_id);
CREATE INDEX IF NOT EXISTS idx_parser_test_runs_parser_id ON parser_test_runs(parser_id);
CREATE INDEX IF NOT EXISTS idx_parser_test_runs_created_at ON parser_test_runs(created_at);
