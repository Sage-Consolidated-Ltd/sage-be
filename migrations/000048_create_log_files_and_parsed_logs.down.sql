-- Drop foreign key constraint on analysis_results
ALTER TABLE IF EXISTS analysis_results
DROP CONSTRAINT IF EXISTS fk_analysis_results_log_file_id;

-- Drop tables (cascades indexes and generated columns)
DROP TABLE IF EXISTS parsed_logs;
DROP TABLE IF EXISTS log_files;
