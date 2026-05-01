-- Remove profile fields from users table
ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS last_login_at;
