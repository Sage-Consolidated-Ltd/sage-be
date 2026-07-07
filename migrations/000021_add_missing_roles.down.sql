-- -- Revert to original roles
-- ALTER TABLE organization_members 
-- DROP CONSTRAINT IF EXISTS organization_members_role_check;

-- ALTER TABLE organization_members 
-- ADD CONSTRAINT organization_members_role_check 
-- CHECK (role IN ('owner', 'admin', 'analyst', 'viewer'));
