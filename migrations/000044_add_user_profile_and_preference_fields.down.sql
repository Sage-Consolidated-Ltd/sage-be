ALTER TABLE users
DROP COLUMN IF EXISTS phone_number,
DROP COLUMN IF EXISTS backup_email,
DROP COLUMN IF EXISTS password_changed_at;

ALTER TABLE user_preferences
DROP COLUMN IF EXISTS date_format;

ALTER TABLE user_notifications
DROP COLUMN IF EXISTS product_updates,
DROP COLUMN IF EXISTS weekly_summary;
