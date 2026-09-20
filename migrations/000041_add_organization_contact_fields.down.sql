ALTER TABLE organizations
DROP COLUMN IF EXISTS primary_contact_email,
DROP COLUMN IF EXISTS support_email;
