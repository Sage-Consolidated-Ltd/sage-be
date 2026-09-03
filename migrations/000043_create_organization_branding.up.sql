-- Create organization_branding table for logo URLs and appearance settings
CREATE TABLE IF NOT EXISTS organization_branding (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    logo_light_url TEXT DEFAULT '',
    logo_dark_url TEXT DEFAULT '',
    show_in_reports BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_org_branding_org_id ON organization_branding(organization_id);
