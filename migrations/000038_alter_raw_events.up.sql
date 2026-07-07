ALTER TABLE raw_events
ADD COLUMN provider_status VARCHAR(20) NOT NULL DEFAULT 'pending';