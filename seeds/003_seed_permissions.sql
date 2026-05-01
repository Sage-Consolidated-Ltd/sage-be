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
('settings.write', 'Edit system settings', 'settings', 'settings', 'write');

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
