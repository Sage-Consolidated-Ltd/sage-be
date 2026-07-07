-- Drop indexes first (safe practice)
DROP INDEX IF EXISTS idx_analysis_results_log_file_id;
DROP INDEX IF EXISTS idx_analysis_results_json_input_id;
DROP INDEX IF EXISTS idx_analysis_results_request_type;
DROP INDEX IF EXISTS idx_threats_analysis_id;
DROP INDEX IF EXISTS idx_threats_organization_id;
DROP INDEX IF EXISTS idx_threats_severity;
DROP INDEX IF EXISTS idx_threats_mitre;

DROP TABLE IF EXISTS threats;

DROP TABLE IF EXISTS analysis_results;

DROP TABLE IF EXISTS json_inputs;