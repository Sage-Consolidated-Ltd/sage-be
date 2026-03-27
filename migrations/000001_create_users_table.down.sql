DROP TRIGGER IF EXISTS update_users_modtime ON users;

DROP INDEX IF EXISTS idx_users_not_deleted;

DROP TABLE IF EXISTS users;