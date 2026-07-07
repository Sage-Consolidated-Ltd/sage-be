-- Remove department column from organization_members
ALTER TABLE organization_members
DROP COLUMN IF EXISTS department;
