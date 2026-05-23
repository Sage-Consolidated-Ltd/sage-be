ALTER TABLE security_events
DROP CONSTRAINT security_events_severity_check;

ALTER TABLE security_events
ALTER COLUMN severity SET NOT NULL;

ALTER TABLE security_events
ADD CONSTRAINT security_events_severity_check
CHECK (
  severity IN ('low', 'medium', 'high', 'critical')
);