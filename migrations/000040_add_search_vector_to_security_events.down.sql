DROP INDEX IF EXISTS idx_security_events_search_vector;
DROP INDEX IF EXISTS idx_security_events_org_occurred;

ALTER TABLE security_events
DROP COLUMN IF EXISTS search_vector,
DROP COLUMN IF EXISTS embedding;
