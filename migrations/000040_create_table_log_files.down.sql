-- Drop indexes first (safe practice)
DROP INDEX IF EXISTS idx_log_files_created_at;
DROP INDEX IF EXISTS idx_log_files_s3_key;
DROP INDEX IF EXISTS idx_log_files_status;
DROP INDEX IF EXISTS idx_log_files_user_id;

-- Drop table
DROP TABLE IF EXISTS log_files;