# DC-GO

DC-GO is a full-stack authentication and multi-tenant role-based access control (RBAC) application. It combines a Go API, a React single-page application, and PostgreSQL in a production-ready container layout.

The project currently includes credential authentication, persistent sessions, account management, global administration, organizations, invitations, teams, and organization-scoped permissions.

## Tech stack

### Backend

- Go 1.25
- Fiber v2
- PostgreSQL 16
- pgx/v5 connection pooling
- sqlc-generated database access
- JWT access tokens and database-backed refresh sessions
- bcrypt password hashing

### Frontend

- React 19
- TypeScript
- Vite 8
- TanStack Router and TanStack Query
- Tailwind CSS 4
- shadcn/ui

## Features

- Email and password registration and login
- Short-lived JWT access tokens
- Rotating, database-backed refresh sessions
- HttpOnly cookies and CSRF protection
- Profile and password management
- Session listing and revocation
- Global `user` and `admin` roles
- Organization `owner`, `admin`, and `member` roles
- Organization creation, membership, and invitations
- Team creation and membership
- Active organization selection per session
- Admin user-management interface
- Automatic database migrations
- Single-container production application serving both the API and frontend

## Project structure

```text
.
├── backend/
│   ├── cmd/
│   │   ├── api/              # API entry point
│   │   └── reset/            # Development-only database reset command
│   ├── configs/              # Environment configuration
│   ├── internal/
│   │   ├── auth/             # Authentication, sessions, JWT, CSRF
│   │   ├── db/               # sqlc-generated code
│   │   ├── organization/     # Organizations, members, invitations
│   │   ├── team/             # Teams and team membership
│   │   └── user/             # Profiles and global roles
│   ├── migrations/           # PostgreSQL migrations
│   ├── pkg/                  # Shared database, logging, response, validation
│   └── sql/queries/          # Queries consumed by sqlc
├── frontend/
│   ├── src/components/       # Application and shadcn UI components
│   ├── src/hooks/            # Authentication and UI hooks
│   ├── src/lib/              # API client and Query configuration
│   └── src/routes/           # TanStack file-based routes
├── docker-compose.yml
├── Dockerfile
└── dev.sh                    # Starts the complete development environment
```

The backend follows a layered flow:

```text
Route -> Handler -> Service -> Repository -> sqlc -> PostgreSQL
```

## Prerequisites

Install the following before running the development environment:

- Go 1.25 or newer
- Node.js 20 or newer
- npm
- Docker with Docker Compose
- [Air](https://github.com/air-verse/air) for Go hot reload

Install Air with:

```bash
go install github.com/air-verse/air@latest
```

## Environment configuration

Create a local environment file from the committed template:

```bash
cp .env.example .env
```

The template contains:

```dotenv
DATABASE_URL=postgres://postgres:postgres@db:5432/dc_go?sslmode=disable
JWT_SECRET=ReplaceThisWithAStrong32CharacterSecret!123
REDIS_PASSWORD=ReplaceThisWithARedisPassword!123
REDIS_URL=redis://:ReplaceThisWithARedisPassword!123@redis:6379/0
S3_ENDPOINT=http://seaweedfs:8333
S3_ACCESS_KEY=dcgo
S3_SECRET_KEY=ReplaceThisWithAnS3Secret!123
S3_BUCKET=dc-go
S3_REGION=us-east-1
S3_USE_SSL=false
S3_FORCE_PATH_STYLE=true
REDISINSIGHT_ENCRYPTION_KEY=ReplaceThisWithA32CharacterEncryptionKey!123
```

`JWT_SECRET` must be at least 32 characters and contain at least three of these character classes: uppercase letters, lowercase letters, digits, and symbols.

`dev.sh` loads the database, Redis, and S3 variables from the root `.env` file. It rewrites Docker service hostnames to `127.0.0.1` because Air runs the Go API directly on the host. Avatar storage uses the S3 configuration; Redis provides distributed rate limiting and is included in the API health check.

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `3000` | Backend HTTP port |
| `ENVIRONMENT` | `development` | Set to `production` to enable secure cookies |
| `JWT_ISSUER` | `dc-go` | Expected JWT issuer |
| `JWT_AUDIENCE` | `dc-go` | Expected JWT audience |
| `STATIC_DIR` | `./public` | Directory served by the Go application |
| `REDIS_URL` | Required | Password-protected Redis connection URL |
| `S3_ENDPOINT` | `http://seaweedfs:8333` | S3-compatible API endpoint |
| `S3_BUCKET` | `dc-go` | Default object-storage bucket |
| `S3_REGION` | `us-east-1` | S3 signing region |
| `S3_FORCE_PATH_STYLE` | `true` | Required for the local SeaweedFS endpoint |
| `REDISINSIGHT_ENCRYPTION_KEY` | Required | Encrypts Redis Insight connection data at rest |

The `.env` file is ignored by Git. Do not commit real secrets.

## Development

Install frontend dependencies once:

```bash
cd frontend
npm ci
cd ..
```

Make sure Docker is running, then start the complete development environment:

```bash
./dev.sh
```

The script:

1. Starts PostgreSQL, Redis, SeaweedFS, and Redis Insight and waits for them to become healthy.
2. Runs the Fiber API through Air with hot reload.
3. Runs the Vite development server.
4. Stops Air and Vite together when either exits or you press `Ctrl+C`.

Development services are available at:

| Service | URL |
| --- | --- |
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:3000/api/v1 |
| Health check | http://localhost:3000/api/v1/health |
| PostgreSQL | `localhost:5432` |
| Redis | `localhost:6379` |
| Redis Insight | http://localhost:5540 |
| SeaweedFS S3 API | http://localhost:8333 |
| SeaweedFS Filer UI | http://localhost:8888 |
| SeaweedFS Master UI | http://localhost:9333 |
| SeaweedFS Admin UI | http://localhost:23646 |

Vite proxies `/api` requests to the backend. Infrastructure containers remain running after `dev.sh` exits; stop them when needed with:

```bash
docker compose --profile devtools stop db redis seaweedfs redisinsight
```

Redis Insight registers `DC-GO Redis` automatically using `REDIS_PASSWORD`; complete its first-run terms screen if prompted.

Redis Insight is in the `devtools` Compose profile and binds only to `127.0.0.1`, so starting the production `app` service does not expose or start it. Redis Insight state is stored in `redisinsightdata`, Redis data in `redisdata`, and SeaweedFS data in `seaweeddata`. SeaweedFS runs in single-node `weed mini` mode and creates the configured bucket automatically. Generic `S3_*` variables keep the application portable across S3-compatible providers.

## Authentication model

Successful registration or login sets three cookies:

- `access_token`: HttpOnly JWT with a 15-minute lifetime.
- `refresh_token`: HttpOnly database session token with a 7-day lifetime.
- `csrf_token`: readable CSRF token sent back in the `X-CSRF-Token` header for state-changing requests.

When an API request receives `401 Unauthorized`, the frontend attempts one refresh. A successful refresh rotates the database session token and retries the original request.

Sessions can also store the user's active organization and team. Organization access is always checked against current membership rather than trusting frontend state.

Redis-backed fixed-window limits protect sensitive endpoints across all API instances:

| Action | Limit | Key |
| --- | --- | --- |
| Register | 3 per hour | Client IP |
| Login | 5 per 15 minutes | Client IP |
| Refresh session | 30 per minute | Client IP |
| Change password | 5 per hour | User |
| Create invitation | 20 per hour | Organization |

Limited responses return `429 Too Many Requests` with `RateLimit-*` and `Retry-After` headers. Sensitive endpoints return `503 Service Unavailable` if Redis cannot enforce a limit. Identifiers are hashed before being stored in Redis.

## Roles

DC-GO has two separate role scopes:

| Scope | Roles | Purpose |
| --- | --- | --- |
| Application | `user`, `admin` | Platform-wide access such as global user administration |
| Organization | `owner`, `admin`, `member` | Access within a specific organization |

Organization owners can update or delete the organization and manage member roles. Owners and organization admins can manage invitations and teams. Organization membership does not grant global administrator access.

## API overview

All endpoints use the `/api/v1` prefix and return a common response envelope:

```json
{
  "success": true,
  "data": {}
}
```

Errors use `success: false` and a `message` field.

### Authentication and sessions

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/auth/register` | Create an account and session |
| `POST` | `/auth/login` | Authenticate and create a session |
| `POST` | `/auth/refresh` | Rotate the refresh session |
| `GET` | `/auth/session` | Get the current user and session |
| `PUT` | `/auth/session/active-organization` | Change the active organization |
| `PUT` | `/auth/password` | Change the current password |
| `GET` | `/auth/sessions` | List active sessions |
| `DELETE` | `/auth/sessions/:id` | Revoke another session |
| `POST` | `/auth/logout` | Revoke the current session and clear cookies |

### Users

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/users/me` | Get the current profile |
| `PUT` | `/users/me` | Update the current profile |
| `PUT` | `/users/me/avatar` | Upload a JPEG or PNG profile photo |
| `DELETE` | `/users/me/avatar` | Remove the current profile photo |
| `GET` | `/users/:id/avatar` | Read an authenticated user's profile photo |
| `DELETE` | `/users/me` | Delete the current account |
| `GET` | `/users` | List users as a global admin |
| `PUT` | `/users/:id/role` | Change a user's global role |

Avatar uploads use `multipart/form-data` with an `avatar` field. Files are limited to 2 MiB and 2048 × 2048 pixels. The backend validates and re-encodes JPEG or PNG data before storing it in the private S3-compatible bucket.

### Organizations and invitations

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/organizations` | Create an organization |
| `GET` | `/organizations` | List the current user's organizations |
| `GET` | `/organizations/:id` | Get an organization |
| `PUT` | `/organizations/:id` | Update an organization |
| `DELETE` | `/organizations/:id` | Delete an organization |
| `GET` | `/organizations/:id/members` | List organization members |
| `GET` | `/organizations/:id/members/me` | Get the current membership |
| `PUT` | `/organizations/:id/members/:userID/role` | Change a member role |
| `DELETE` | `/organizations/:id/members/:userID` | Remove a member |
| `POST` | `/organizations/:id/invitations` | Create an invitation |
| `GET` | `/organizations/:id/invitations` | List invitations |
| `POST` | `/invitations/:id/accept` | Accept an invitation |
| `POST` | `/invitations/:id/decline` | Decline an invitation |
| `DELETE` | `/invitations/:id` | Cancel an invitation |

### Teams

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/organizations/:orgID/teams` | Create a team |
| `GET` | `/organizations/:orgID/teams` | List organization teams |
| `GET` | `/teams/:id` | Get a team |
| `PUT` | `/teams/:id` | Update a team |
| `DELETE` | `/teams/:id` | Delete a team |
| `POST` | `/teams/:id/members` | Add an organization member to a team |
| `GET` | `/teams/:id/members` | List team members |
| `DELETE` | `/teams/:id/members/:userID` | Remove a team member |

## Database migrations

Migrations in `backend/migrations` run automatically when the API starts. Applied filenames are recorded in the `schema_migrations` table, and each new migration is applied in a transaction.

### Upgrading an existing `dc-express` database

Fresh installations use the `dc_go` database. If an existing Docker volume was created before the project rename, stop the application and rename its database once before starting the updated code:

```bash
docker compose stop app
docker compose exec -T db psql -U postgres -d postgres \
  -c 'ALTER DATABASE dc_express RENAME TO dc_go;'
```

Update `DATABASE_URL` to use `/dc_go` after the rename. The operation preserves the existing schema and data.

Add schema changes as a new, sequentially named SQL file instead of modifying a migration that has already been applied. Regenerate database code after changing the schema or SQL queries:

```bash
cd backend
sqlc generate
```

## Quality checks

Run backend checks:

```bash
cd backend
go test ./...
go vet ./...
```

Run frontend checks:

```bash
cd frontend
npm test
npm run typecheck
npm run lint
npm run format:check
npm run build
```

## Production with Docker Compose

With a valid root `.env` file, build and start the complete application:

```bash
docker compose up --build -d
```

The production container builds the React application, compiles a static Go binary, applies pending migrations, and serves both the SPA and API on http://localhost:3000.

Check the running services and API health:

```bash
docker compose ps
curl http://localhost:3000/api/v1/health
```

Stop the stack without deleting PostgreSQL, Redis, or SeaweedFS data:

```bash
docker compose down
```

## Current scope

The backend already exposes organization, invitation, team management, and S3-backed avatar APIs. The current frontend focuses on authentication, account and session management, profile-photo uploads, organization switching, the protected layout, and global user administration.

The dashboard is still a placeholder. Email verification, password recovery, OAuth buttons, two-factor enrollment, and full organization/team management screens are not implemented yet.
