CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'support_admin',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    failed_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ NULL,
    last_login_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_admins_email_lower ON admins(LOWER(email));
CREATE INDEX IF NOT EXISTS idx_admins_status ON admins(status);

DROP TRIGGER IF EXISTS update_admins_modtime ON admins;
CREATE TRIGGER update_admins_modtime 
BEFORE UPDATE ON admins 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();
