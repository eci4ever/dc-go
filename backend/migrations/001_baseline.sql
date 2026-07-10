CREATE TABLE "user" (
    id TEXT PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT false, image TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    role TEXT, banned BOOLEAN NOT NULL DEFAULT false, ban_reason TEXT, ban_expires TIMESTAMPTZ
);
CREATE TABLE "session" (
    id TEXT PRIMARY KEY, expires_at TIMESTAMPTZ NOT NULL, token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address TEXT, user_agent TEXT, user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    impersonated_by TEXT, active_organization_id TEXT, active_team_id TEXT
);
CREATE INDEX session_user_id_idx ON "session"(user_id);
CREATE TABLE "account" (
    id TEXT PRIMARY KEY, account_id TEXT NOT NULL, provider_id TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE, access_token TEXT, refresh_token TEXT,
    id_token TEXT, access_token_expires_at TIMESTAMPTZ, refresh_token_expires_at TIMESTAMPTZ,
    scope TEXT, password TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider_id, account_id)
);
CREATE INDEX account_user_id_idx ON "account"(user_id);
CREATE TABLE "verification" (id TEXT PRIMARY KEY, identifier TEXT NOT NULL, value TEXT NOT NULL, expires_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE INDEX verification_identifier_idx ON "verification"(identifier);
CREATE TABLE "organization" (id TEXT PRIMARY KEY, name TEXT NOT NULL, slug TEXT NOT NULL UNIQUE, logo TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), metadata TEXT);
CREATE TABLE "member" (
    id TEXT PRIMARY KEY, organization_id TEXT NOT NULL REFERENCES "organization"(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE(organization_id, user_id)
);
CREATE INDEX member_user_org_idx ON "member"(user_id, organization_id);
CREATE TABLE "team" (id TEXT PRIMARY KEY, name TEXT NOT NULL, organization_id TEXT NOT NULL REFERENCES "organization"(id) ON DELETE CASCADE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE INDEX team_organization_id_idx ON "team"(organization_id);
CREATE TABLE "team_member" (id TEXT PRIMARY KEY, team_id TEXT NOT NULL REFERENCES "team"(id) ON DELETE CASCADE, user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE(team_id, user_id));
CREATE INDEX team_member_user_team_idx ON "team_member"(user_id, team_id);
CREATE TABLE "invitation" (
    id TEXT PRIMARY KEY, organization_id TEXT NOT NULL REFERENCES "organization"(id) ON DELETE CASCADE,
    email TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
    team_id TEXT REFERENCES "team"(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'expired')),
    expires_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), inviter_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX invitation_pending_org_email_idx ON "invitation"(organization_id, lower(email)) WHERE status = 'pending';
CREATE INDEX invitation_email_status_idx ON "invitation"(lower(email), status);
