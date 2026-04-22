DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'integration_status') THEN
        CREATE TYPE integration_status AS ENUM ('active', 'inactive', 'error');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    connection_type VARCHAR(50) NOT NULL,
    status integration_status DEFAULT 'inactive',
    config JSONB NOT NULL, -- provider-specific config
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE integration_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    encrypted_value TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE integration_streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    stream_name VARCHAR(255) NOT NULL, -- topic / channel / endpoint
    last_offset TEXT, -- kafka offset / cursor / checkpoint
    last_event_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(integration_id, stream_name)
);

CREATE TABLE ingestion_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL, -- running, stopped, failed
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    error TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE integration_events_buffer (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending', -- pending, processed, failed
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_integrations_tenant 
ON integrations(tenant_id);

CREATE INDEX idx_streams_integration 
ON integration_streams(integration_id);

CREATE INDEX idx_sessions_integration 
ON ingestion_sessions(integration_id);

