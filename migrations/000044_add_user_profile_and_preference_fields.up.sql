-- Add phone_number, backup_email, and password_changed_at to users table
ALTER TABLE users
ADD COLUMN IF NOT EXISTS phone_number TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS backup_email TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ DEFAULT NOW();

-- Add date_format to user_preferences table
ALTER TABLE user_preferences
ADD COLUMN IF NOT EXISTS date_format VARCHAR(20) DEFAULT 'DD/MM/YYYY';

-- Add product_updates and weekly_summary to user_notifications table
ALTER TABLE user_notifications
ADD COLUMN IF NOT EXISTS product_updates BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS weekly_summary BOOLEAN DEFAULT false;
