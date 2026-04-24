-- Add missing organization fields per Organization module spec
ALTER TABLE organizations
ADD COLUMN IF NOT EXISTS slug TEXT UNIQUE,
ADD COLUMN IF NOT EXISTS industry TEXT,
ADD COLUMN IF NOT EXISTS country TEXT,
ADD COLUMN IF NOT EXISTS timezone TEXT DEFAULT 'UTC',
ADD COLUMN IF NOT EXISTS risk_threshold_default INTEGER DEFAULT 50;

-- Create index on slug for lookups
CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);

-- Update existing organizations with generated slugs
UPDATE organizations 
SET slug = LOWER(REGEXP_REPLACE(name, '[^a-zA-Z0-9]+', '-', 'g'))
WHERE slug IS NULL;
