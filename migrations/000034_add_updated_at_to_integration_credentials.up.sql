BEGIN;

-- add updated_at
ALTER TABLE integration_credentials
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- remove old FK to integrations
ALTER TABLE integration_credentials
DROP CONSTRAINT IF EXISTS integration_credentials_integration_id_fkey;

-- create new FK to data_sources
ALTER TABLE integration_credentials
ADD CONSTRAINT integration_credentials_integration_id_fkey
FOREIGN KEY (integration_id)
REFERENCES data_sources(id)
ON DELETE CASCADE;

-- indexes
CREATE INDEX IF NOT EXISTS idx_integration_credentials_integration_id
ON integration_credentials(integration_id);

CREATE INDEX IF NOT EXISTS idx_integration_credentials_updated_at
ON integration_credentials(updated_at);

-- drop old integrations table
DROP TABLE IF EXISTS integrations CASCADE;

COMMIT;