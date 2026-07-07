-- Add department column to organization_members
ALTER TABLE organization_members
ADD COLUMN IF NOT EXISTS department TEXT;

-- Create index for department filtering
CREATE INDEX IF NOT EXISTS idx_org_members_department ON organization_members(department);
