ALTER TABLE organization
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active',
    ADD CONSTRAINT organization_status_check
        CHECK (status IN ('active', 'inactive', 'suspended', 'archived'));

ALTER TABLE member
    ADD COLUMN permissions TEXT[] NOT NULL DEFAULT '{}',
    ADD CONSTRAINT member_permissions_check CHECK (
        permissions <@ ARRAY[
            'members.manage',
            'academic.students.manage',
            'academic.structure.manage',
            'academic.results.manage',
            'audit.view'
        ]::TEXT[]
    );

CREATE TABLE organization_audit_log (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    actor_user_id TEXT REFERENCES "user"(id) ON DELETE SET NULL,
    actor_name TEXT NOT NULL,
    actor_email TEXT NOT NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT,
    details JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX organization_audit_log_org_created_idx
    ON organization_audit_log(organization_id, created_at DESC);
CREATE INDEX organization_audit_log_actor_idx
    ON organization_audit_log(actor_user_id, created_at DESC);
