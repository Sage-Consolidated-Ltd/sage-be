CREATE TABLE IF NOT EXISTS physical_device_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES physical_devices(id) ON DELETE CASCADE,
    credential_type VARCHAR(255) NOT NULL,
    credential_value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);