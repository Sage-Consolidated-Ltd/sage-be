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

-- Insert default permission groups
INSERT INTO permission_groups (name, description, category) VALUES
('Alert Management', 'Full access to alerts and incidents', 'Alerts'),
('User Management', 'Full access to user management', 'Users'),
('Organization Management', 'Full access to organization settings', 'Organization'),
('Automation', 'Full access to playbooks and automations', 'Automation'),
('Billing', 'Full access to billing information', 'Billing'),
('Settings', 'Full access to system settings', 'Settings');

-- Link permissions to groups
-- Alert Management
INSERT INTO permission_group_permissions (permission_group_id, permission_id)
SELECT pg.id, p.id FROM permission_groups pg, permissions p 
WHERE pg.name = 'Alert Management' AND p.category = 'alerts';

-- User Management
INSERT INTO permission_group_permissions (permission_group_id, permission_id)
SELECT pg.id, p.id FROM permission_groups pg, permissions p 
WHERE pg.name = 'User Management' AND p.category = 'users';

-- Organization Management
INSERT INTO permission_group_permissions (permission_group_id, permission_id)
SELECT pg.id, p.id FROM permission_groups pg, permissions p 
WHERE pg.name = 'Organization Management' AND p.category = 'org';

-- Automation
INSERT INTO permission_group_permissions (permission_group_id, permission_id)
SELECT pg.id, p.id FROM permission_groups pg, permissions p 
WHERE pg.name = 'Automation' AND p.category = 'automation';

-- Billing
INSERT INTO permission_group_permissions (permission_group_id, permission_id)
SELECT pg.id, p.id FROM permission_groups pg, permissions p 
WHERE pg.name = 'Billing' AND p.category = 'billing';

-- Settings
INSERT INTO permission_group_permissions (permission_group_id, permission_id)
SELECT pg.id, p.id FROM permission_groups pg, permissions p 
WHERE pg.name = 'Settings' AND p.category = 'settings';

-- Create indexes
CREATE INDEX idx_permission_groups_category ON permission_groups(category);
CREATE INDEX idx_permission_group_permissions_group ON permission_group_permissions(permission_group_id);
CREATE INDEX idx_permission_group_permissions_permission ON permission_group_permissions(permission_id);
