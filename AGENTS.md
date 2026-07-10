# AGENTS.md

## Dev Servers

```
docker compose up -d db          # PostgreSQL on :5432
cd backend && air               # Fiber API on :3000 (hot reload)
cd frontend && npm run dev        # Vite on :5173 (proxies /api → :3000)
```

## Project

Full-stack auth system:
- **Backend**: Go 1.25, Fiber v2, pgx/v5, sqlc, JWT + bcrypt + DB sessions
- **Frontend**: Vite 8, React 19, TanStack Router & Query, shadcn/ui v4, Tailwind v4

See `backend/AGENTS.md` for Go architecture conventions.
