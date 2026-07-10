# Roles

Global roles and organization roles are independent:

- Global: `user`, `admin`
- Organization: `owner`, `admin`, `member`

Every registered account starts with the global `user` role. Bootstrap the first
global administrator once, after that account has registered:

```sql
UPDATE "user"
SET role = 'admin', updated_at = NOW()
WHERE lower(email) = lower('admin@example.com');
```

Run this statement directly against the deployment database, then sign in again
or reload the current session. Subsequent global role changes are available to
admins through `PUT /api/v1/users/:id/role`.

A global admin does not automatically gain access to an organization. Access to
organization data continues to require a corresponding membership and its
organization-scoped role.
