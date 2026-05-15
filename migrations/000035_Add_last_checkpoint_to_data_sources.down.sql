ALTER TABLE data_sources
DROP COLUMN IF EXISTS last_checkpoint,
DROP COLUMN IF EXISTS last_checkpoint_at;