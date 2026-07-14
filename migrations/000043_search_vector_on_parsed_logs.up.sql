-- generated tsvector column, kept in sync automatically
ALTER TABLE parsed_logs ADD COLUMN search_vector tsvector
  GENERATED ALWAYS AS (to_tsvector('english', coalesce(message, ''))) STORED;

CREATE INDEX idx_parsed_logs_search ON parsed_logs USING GIN (search_vector);
CREATE INDEX idx_parsed_logs_raw_json ON parsed_logs USING GIN (raw_json jsonb_path_ops);
CREATE INDEX idx_parsed_logs_filters ON parsed_logs (data_source_id, timestamp DESC, level);