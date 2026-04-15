DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'invite_status') THEN
        CREATE TYPE invite_status AS ENUM ('pending', 'accepted', 'rejected', 'expired', 'revoked');
    END IF;
END $$;
CREATE TABLE IF NOT EXISTS organization_invites(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role_id UUID REFERENCES organization_roles(id) ON DELETE SET NULL,
    status invite_status NOT NULL DEFAULT 'pending',
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    token_hash TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX unique_pending_invite
ON organization_invites (organization_id, LOWER(email))
WHERE status = 'pending';

CREATE INDEX idx_invite_email
ON organization_invites (LOWER(email));

