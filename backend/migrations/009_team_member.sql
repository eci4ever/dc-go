CREATE TABLE IF NOT EXISTS "team_member" (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL REFERENCES "team"(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS teamMember_teamId_idx ON "team_member"(team_id);
CREATE INDEX IF NOT EXISTS teamMember_userId_idx ON "team_member"(user_id);
