CREATE TABLE IF NOT EXISTS industries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_industries_id ON industries(id) WHERE deleted_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_industries_name ON industries(name);

CREATE TABLE IF NOT EXISTS organization_industries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    industry_id UUID NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_industries_org_industry ON organization_industries(organization_id, industry_id);