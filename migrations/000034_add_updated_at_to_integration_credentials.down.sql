BEGIN;

DROP INDEX IF EXISTS idx_integration_credentials_integration_id;
DROP INDEX IF EXISTS idx_integration_credentials_updated_at;

ALTER TABLE integration_credentials
DROP CONSTRAINT IF EXISTS integration_credentials_integration_id_fkey;

ALTER TABLE integration_credentials
ADD CONSTRAINT integration_credentials_integration_id_fkey
FOREIGN KEY (integration_id)
REFERENCES integrations(id)
ON DELETE CASCADE;

ALTER TABLE integration_credentials
DROP COLUMN IF EXISTS updated_at;

COMMIT;