-- Create security_events table
CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    parser_id UUID NULL REFERENCES parsers(id) ON DELETE SET NULL,
    source_event_id VARCHAR(255) NULL,
    source VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low','medium','high','critical')),
    actor_email VARCHAR(255) NULL,
    actor_username VARCHAR(255) NULL,
    ip_address VARCHAR(255) NULL,
    geo_country VARCHAR(100) NULL,
    geo_city VARCHAR(100) NULL,
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    normalized_payload JSONB DEFAULT '{}'::jsonb,
    parse_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (parse_status IN ('pending','success','failed','partial')),
    parse_errors JSONB DEFAULT '[]'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_events_org_id ON security_events(organization_id);
CREATE INDEX IF NOT EXISTS idx_security_events_source_id ON security_events(source_id);
CREATE INDEX IF NOT EXISTS idx_security_events_parser_id ON security_events(parser_id);
CREATE INDEX IF NOT EXISTS idx_security_events_type ON security_events(event_type);
CREATE INDEX IF NOT EXISTS idx_security_events_category ON security_events(event_category);
CREATE INDEX IF NOT EXISTS idx_security_events_severity ON security_events(severity);
CREATE INDEX IF NOT EXISTS idx_security_events_actor_email ON security_events(actor_email);
CREATE INDEX IF NOT EXISTS idx_security_events_ip_address ON security_events(ip_address);
CREATE INDEX IF NOT EXISTS idx_security_events_occurred_at ON security_events(occurred_at);
CREATE INDEX IF NOT EXISTS idx_security_events_ingested_at ON security_events(ingested_at);
CREATE INDEX IF NOT EXISTS idx_security_events_updated_at ON security_events(updated_at);
