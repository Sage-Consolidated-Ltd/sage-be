-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT NOT NULL, -- e.g., 'alerts', 'users', 'automation', 'billing'
    resource TEXT NOT NULL, -- e.g., 'alerts', 'users', 'playbooks'
    action TEXT NOT NULL, -- e.g., 'read', 'write', 'delete', 'execute'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default permissions
INSERT INTO permissions (name, description, category, resource, action) VALUES
-- Alert permissions
('alerts.read', 'View alerts and incidents', 'alerts', 'alerts', 'read'),
('alerts.write', 'Create and edit alerts', 'alerts', 'alerts', 'write'),
('alerts.delete', 'Delete alerts', 'alerts', 'alerts', 'delete'),
('alerts.resolve', 'Resolve alerts', 'alerts', 'alerts', 'resolve'),

-- User permissions
('users.read', 'View users', 'users', 'users', 'read'),
('users.write', 'Create and edit users', 'users', 'users', 'write'),
('users.delete', 'Delete users', 'users', 'users', 'delete'),
('users.invite', 'Invite users', 'users', 'users', 'invite'),

-- Organization permissions
('org.read', 'View organization settings', 'org', 'organization', 'read'),
('org.write', 'Edit organization settings', 'org', 'organization', 'write'),
('org.manage_roles', 'Manage roles and permissions', 'org', 'organization', 'manage_roles'),

-- Automation permissions
('automation.read', 'View playbooks and automations', 'automation', 'playbooks', 'read'),
('automation.write', 'Create and edit playbooks', 'automation', 'playbooks', 'write'),
('automation.execute', 'Execute playbooks', 'automation', 'playbooks', 'execute'),
('automation.delete', 'Delete playbooks', 'automation', 'playbooks', 'delete'),

-- Billing permissions
('billing.read', 'View billing information', 'billing', 'billing', 'read'),
('billing.write', 'Edit billing information', 'billing', 'billing', 'write'),

-- Settings permissions
('settings.read', 'View system settings', 'settings', 'settings', 'read'),
('settings.write', 'Edit system settings', 'settings', 'settings', 'write')
ON CONFLICT (name) DO NOTHING;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_permissions_category ON permissions(category);
CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);
