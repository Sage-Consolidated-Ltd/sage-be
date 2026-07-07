ALTER TABLE security_events
ADD CONSTRAINT unique_source_event
UNIQUE (source_id, source_event_id);