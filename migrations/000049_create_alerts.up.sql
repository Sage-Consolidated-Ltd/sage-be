CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    threat_label VARCHAR(100) NOT NULL,
    mitre VARCHAR(50) NOT NULL,
    log_source VARCHAR(50) NOT NULL,
    event_id VARCHAR(50) NOT NULL,
    entity_host VARCHAR(255) NULL,
    entity_account VARCHAR(255) NULL,
    entity_ip VARCHAR(100) NULL,
    security_event_id UUID NULL REFERENCES security_events(id) ON DELETE SET NULL,
    context JSONB NOT NULL DEFAULT '{}',
    detected_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_org_detected ON alerts(organization_id, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_threat_label ON alerts(threat_label);
CREATE INDEX IF NOT EXISTS idx_alerts_entity_host ON alerts(entity_host);
CREATE INDEX IF NOT EXISTS idx_alerts_entity_account ON alerts(entity_account);
CREATE INDEX IF NOT EXISTS idx_alerts_entity_ip ON alerts(entity_ip);

