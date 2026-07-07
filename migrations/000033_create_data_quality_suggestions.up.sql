-- Create data_quality_suggestions table
CREATE TABLE IF NOT EXISTS data_quality_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source_id UUID NULL REFERENCES data_sources(id) ON DELETE SET NULL,
    parser_id UUID NULL REFERENCES parsers(id) ON DELETE SET NULL,
    summary TEXT NOT NULL,
    recommendation TEXT NOT NULL,
    suggested_fix JSONB NOT NULL,
    confidence DECIMAL(3,2) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','applied','dismissed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    applied_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_data_quality_suggestions_org_id ON data_quality_suggestions(organization_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_suggestions_source_id ON data_quality_suggestions(source_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_suggestions_parser_id ON data_quality_suggestions(parser_id);
CREATE INDEX IF NOT EXISTS idx_data_quality_suggestions_status ON data_quality_suggestions(status);
