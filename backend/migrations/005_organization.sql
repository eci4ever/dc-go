CREATE TABLE IF NOT EXISTS "organization" (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    logo TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata TEXT
);
