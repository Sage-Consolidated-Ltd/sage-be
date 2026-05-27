DROP INDEX IF EXISTS idx_raw_events_payload_gin;
DROP INDEX IF EXISTS idx_raw_events_timestamp;
DROP INDEX IF EXISTS idx_raw_events_source_id;
DROP INDEX IF EXISTS idx_raw_events_org_id;
DROP INDEX IF EXISTS idx_raw_events_pending;
DROP INDEX IF EXISTS idx_raw_events_processing;

DROP TABLE IF EXISTS raw_events;