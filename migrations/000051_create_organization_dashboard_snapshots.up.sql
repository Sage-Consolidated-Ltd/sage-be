-- Create organization_dashboard_snapshots table for materialized dashboard state
CREATE TABLE IF NOT EXISTS organization_dashboard_snapshots (
    organization_id   UUID PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    security_score    JSONB NOT NULL DEFAULT '{}'::jsonb,
    vulnerabilities   JSONB NOT NULL DEFAULT '{}'::jsonb,
    identity_health   JSONB NOT NULL DEFAULT '{}'::jsonb,
    endpoint_coverage JSONB NOT NULL DEFAULT '{}'::jsonb,
    threat_intel      JSONB NOT NULL DEFAULT '{}'::jsonb,
    active_incidents  JSONB NOT NULL DEFAULT '[]'::jsonb,
    dangerous_threats JSONB NOT NULL DEFAULT '{}'::jsonb,
    compliance_risks  JSONB NOT NULL DEFAULT '{}'::jsonb,
    threat_trends     JSONB NOT NULL DEFAULT '{}'::jsonb,
    geo_threats       JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dashboard_snapshots_updated ON organization_dashboard_snapshots(updated_at);
