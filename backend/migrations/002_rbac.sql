UPDATE "user"
SET role = 'user'
WHERE role IS NULL OR role NOT IN ('user', 'admin');

ALTER TABLE "user"
    ALTER COLUMN role SET DEFAULT 'user',
    ALTER COLUMN role SET NOT NULL,
    ADD COLUMN two_factor_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD CONSTRAINT user_role_check CHECK (role IN ('user', 'admin'));

UPDATE "session" s
SET active_organization_id = NULL
WHERE active_organization_id IS NOT NULL
  AND (
      NOT EXISTS (
          SELECT 1 FROM "organization" o WHERE o.id = s.active_organization_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM "member" m
          WHERE m.organization_id = s.active_organization_id AND m.user_id = s.user_id
      )
  );

ALTER TABLE "session"
    ADD CONSTRAINT session_active_organization_fk
    FOREIGN KEY (active_organization_id)
    REFERENCES "organization"(id)
    ON DELETE SET NULL;
