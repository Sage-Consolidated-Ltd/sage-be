ALTER TABLE security_events
ALTER COLUMN severity DROP NOT NULL;

ALTER TABLE security_events
DROP CONSTRAINT security_events_severity_check;

ALTER TABLE security_events
ADD CONSTRAINT security_events_severity_check
CHECK (
  severity IS NULL OR
  severity IN ('low', 'medium', 'high', 'critical')
);