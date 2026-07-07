-- Create custom_roles table
CREATE TABLE IF NOT EXISTS custom_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT FALSE, -- For built-in roles like admin, analyst
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, name)
);

-- Create junction table for custom_roles <-> permission_groups
CREATE TABLE IF NOT EXISTS custom_role_permission_groups (
    custom_role_id UUID NOT NULL REFERENCES custom_roles(id) ON DELETE CASCADE,
    permission_group_id UUID NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (custom_role_id, permission_group_id)
);

-- Create junction table for custom_roles <-> permissions (for granular control)
CREATE TABLE IF NOT EXISTS custom_role_permissions (
    custom_role_id UUID NOT NULL REFERENCES custom_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (custom_role_id, permission_id)
);

-- Create indexes
CREATE INDEX idx_custom_roles_org ON custom_roles(organization_id);
CREATE INDEX idx_custom_roles_system ON custom_roles(is_system_role);
CREATE INDEX idx_custom_role_permission_groups_role ON custom_role_permission_groups(custom_role_id);
CREATE INDEX idx_custom_role_permission_groups_group ON custom_role_permission_groups(permission_group_id);
CREATE INDEX idx_custom_role_permissions_role ON custom_role_permissions(custom_role_id);
CREATE INDEX idx_custom_role_permissions_permission ON custom_role_permissions(permission_id);
