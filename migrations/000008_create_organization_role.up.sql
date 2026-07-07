CREATE TABLE IF NOT EXISTS organization_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE organization_members 
ADD COLUMN IF NOT EXISTS role_id UUID 
REFERENCES organization_roles(id) ON DELETE SET NULL;

ALTER TABLE organization_members 
DROP COLUMN IF EXISTS role;

ALTER TABLE organization_roles
ADD CONSTRAINT organization_roles_name_unique UNIQUE (name);