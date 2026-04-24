-- Remove added organization_members fields
ALTER TABLE organization_members
DROP COLUMN IF EXISTS invited_by,
DROP COLUMN IF EXISTS invited_at,
DROP COLUMN IF EXISTS joined_at,
DROP COLUMN IF EXISTS updated_at;
