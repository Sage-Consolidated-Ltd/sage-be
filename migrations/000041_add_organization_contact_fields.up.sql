-- Add primary_contact_email and support_email to organizations table
ALTER TABLE organizations
ADD COLUMN IF NOT EXISTS primary_contact_email TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS support_email TEXT DEFAULT '';
