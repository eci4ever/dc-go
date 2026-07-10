CREATE TABLE IF NOT EXISTS "invitation" (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES "organization"(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT,
    team_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    inviter_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS invitation_organizationId_idx ON "invitation"(organization_id);
CREATE INDEX IF NOT EXISTS invitation_email_idx ON "invitation"(email);
