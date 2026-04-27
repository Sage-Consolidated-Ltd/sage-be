-- Create parser_versions table
CREATE TABLE IF NOT EXISTS parser_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    parser_id UUID NOT NULL REFERENCES parsers(id) ON DELETE CASCADE,
    version_number INT NOT NULL,
    logic JSONB NOT NULL,
    mappings JSONB DEFAULT '[]'::jsonb,
    changed_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    change_note TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(parser_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_parser_versions_parser_id ON parser_versions(parser_id);
CREATE INDEX IF NOT EXISTS idx_parser_versions_org_id ON parser_versions(organization_id);
CREATE INDEX IF NOT EXISTS idx_parser_versions_created_at ON parser_versions(created_at);
