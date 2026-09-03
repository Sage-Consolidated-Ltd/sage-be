ALTER TABLE security_events
ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
    to_tsvector('english', 
        coalesce(source, '') || ' ' || 
        coalesce(event_type, '') || ' ' || 
        coalesce(event_category, '') || ' ' || 
        coalesce(severity, '') || ' ' || 
        coalesce(actor_email, '') || ' ' || 
        coalesce(actor_username, '') || ' ' || 
        coalesce(ip_address, '') || ' ' || 
        coalesce(raw_payload::text, '')
    )
) STORED;

CREATE INDEX IF NOT EXISTS idx_security_events_search_vector ON security_events USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_security_events_org_occurred ON security_events (organization_id, occurred_at DESC);
