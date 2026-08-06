-- Create organization_environments table for system-generated, read-only metadata
CREATE TABLE IF NOT EXISTS organization_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id TEXT UNIQUE,
    subscription_id TEXT,
    region TEXT,
    deployment_mode TEXT NOT NULL DEFAULT 'saas',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_org_env_org_id ON organization_environments(organization_id);
