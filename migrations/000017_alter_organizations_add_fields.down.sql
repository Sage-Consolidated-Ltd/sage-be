-- Remove added organization fields
ALTER TABLE organizations
DROP COLUMN IF EXISTS slug,
DROP COLUMN IF EXISTS industry,
DROP COLUMN IF EXISTS country,
DROP COLUMN IF EXISTS timezone,
DROP COLUMN IF EXISTS risk_threshold_default;
