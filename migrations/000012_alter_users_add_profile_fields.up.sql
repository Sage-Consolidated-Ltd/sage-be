-- Add profile fields to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS avatar_url TEXT,
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;

-- Create index for last_login_at for activity queries
CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON users(last_login_at);
