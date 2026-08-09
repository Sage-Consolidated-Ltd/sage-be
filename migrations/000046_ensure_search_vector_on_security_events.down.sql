DROP INDEX IF EXISTS idx_security_events_search_vector;
ALTER TABLE security_events DROP COLUMN IF EXISTS search_vector;
