CREATE TABLE IF NOT EXISTS "member" (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES "organization"(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS member_organizationId_idx ON "member"(organization_id);
CREATE INDEX IF NOT EXISTS member_userId_idx ON "member"(user_id);
