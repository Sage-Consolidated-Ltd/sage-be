-- Create parsers table
CREATE TABLE IF NOT EXISTS parsers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source_id UUID NULL REFERENCES data_sources(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    parser_type VARCHAR(50) NOT NULL CHECK (parser_type IN ('regex','json','csv','key_value','ai_nlp')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','warning','error','disabled')),
    tags JSONB DEFAULT '[]'::jsonb,
    logic JSONB NOT NULL,
    mappings JSONB DEFAULT '[]'::jsonb,
    events_parsed_24h BIGINT DEFAULT 0,
    error_rate DECIMAL(5,2) DEFAULT 0.0,
    owner_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_parsers_org_id ON parsers(organization_id);
CREATE INDEX IF NOT EXISTS idx_parsers_source_id ON parsers(source_id);
CREATE INDEX IF NOT EXISTS idx_parsers_status ON parsers(status);
CREATE INDEX IF NOT EXISTS idx_parsers_type ON parsers(parser_type);
CREATE INDEX IF NOT EXISTS idx_parsers_updated_at ON parsers(updated_at);
