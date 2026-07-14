-- Drop indexes
DROP INDEX IF EXISTS idx_parsed_logs_filters;
DROP INDEX IF EXISTS idx_parsed_logs_raw_json;
DROP INDEX IF EXISTS idx_parsed_logs_search;

-- Drop generated column
ALTER TABLE parsed_logs
DROP COLUMN IF EXISTS search_vector;