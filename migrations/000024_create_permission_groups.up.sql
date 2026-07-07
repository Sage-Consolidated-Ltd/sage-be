-- Create permission_groups table
CREATE TABLE IF NOT EXISTS permission_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT NOT NULL, -- e.g., 'Alert Management', 'User Management', 'Automation'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create junction table for permission_groups <-> permissions
CREATE TABLE IF NOT EXISTS permission_group_permissions (
    permission_group_id UUID NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (permission_group_id, permission_id)
);

-- Create indexes
CREATE INDEX idx_permission_groups_category ON permission_groups(category);
CREATE INDEX idx_permission_group_permissions_group ON permission_group_permissions(permission_group_id);
CREATE INDEX idx_permission_group_permissions_permission ON permission_group_permissions(permission_id);
