DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'member_status') THEN
        CREATE TYPE member_status AS ENUM ('active', 'inactive', 'pending', 'suspended');
    END IF;
END $$;

ALTER TABLE organization_members
ADD COLUMN IF NOT EXISTS status member_status NOT NULL DEFAULT 'pending';