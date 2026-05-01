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

-- Create indexes
CREATE INDEX idx_permissions_category ON permissions(category);
CREATE INDEX idx_permissions_resource ON permissions(resource);
