-- Create organization_settings table per Organization module spec
CREATE TABLE IF NOT EXISTS organization_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Alert and detection settings
    default_alert_severity_threshold TEXT DEFAULT 'medium',
    auto_containment_enabled BOOLEAN DEFAULT false,
    auto_containment_threshold INTEGER DEFAULT 80,
    
    -- Security settings
    allowed_ip_ranges JSONB DEFAULT '[]'::jsonb,
    session_timeout_minutes INTEGER DEFAULT 60,
    audit_logging_enabled BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create index on organization_id for fast lookups
CREATE INDEX IF NOT EXISTS idx_org_settings_org_id ON organization_settings(organization_id);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_org_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_org_settings_updated_at ON organization_settings;
CREATE TRIGGER trigger_org_settings_updated_at
    BEFORE UPDATE ON organization_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_org_settings_updated_at();
