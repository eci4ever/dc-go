ALTER TABLE "user"
    ADD COLUMN avatar_key TEXT,
    ADD COLUMN avatar_content_type TEXT,
    ADD COLUMN avatar_updated_at TIMESTAMPTZ;
