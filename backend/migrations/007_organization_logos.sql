ALTER TABLE "organization"
    ADD COLUMN IF NOT EXISTS logo_key TEXT,
    ADD COLUMN IF NOT EXISTS logo_content_type TEXT,
    ADD COLUMN IF NOT EXISTS logo_updated_at TIMESTAMPTZ;
