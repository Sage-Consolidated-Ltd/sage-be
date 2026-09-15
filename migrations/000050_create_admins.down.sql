DROP TRIGGER IF EXISTS update_admins_modtime ON admins;
DROP INDEX IF EXISTS idx_admins_email_lower;
DROP INDEX IF EXISTS idx_admins_status;
DROP TABLE IF EXISTS admins;
