# AGENTS.md

## Project

This project is a production-ready REST API built with:

- Go 1.25+
- Fiber
- PostgreSQL
- sqlc
- JWT Authentication
- Clean Modular Architecture

The codebase should always prioritize:

- Readability
- Maintainability
- Testability
- Performance

---

# Architecture

Follow this architecture strictly.

```
cmd/
    api/
        main.go

internal/
    auth/
    user/
    organization/
    attendance/
    billing/

pkg/
    response/
    logger/
    jwt/
    database/
    validator/

configs/

migrations/

docs/
```

Each feature owns its own code.

Example:

```
internal/user/

handler.go
service.go
repository.go
dto.go
model.go
routes.go
errors.go
```

Never create a global handlers folder.

---

# Layer Responsibilities

## Handler

Responsible only for HTTP.

Allowed:

- Read request
- Bind JSON
- Validate request
- Call service
- Return HTTP response

Not allowed:

- SQL
- Business logic
- Password hashing
- JWT creation
- External API calls

---

## Service

Contains business logic.

Allowed:

- Validation
- Transactions
- Authorization
- Password hashing
- Calling repositories
- Calling external services

Must never know Fiber.

Only receive:

```
context.Context
```

---

## Repository

Responsible only for database.

Allowed:

- SQL
- CRUD
- Transactions

Not allowed:

- Validation
- Business rules
- HTTP logic

---

# Context

Always propagate context.

Example

```go
func (s *UserService) Create(
    ctx context.Context,
    req dto.CreateUserRequest,
)
```

Never use context.Background() inside handlers.

---

# Dependency Injection

No global variables.

Always inject dependencies.

```
Repository
↓

Service
↓

Handler
```

Example

```go
repo := repository.New(db)

service := service.New(repo)

handler := handler.New(service)
```

---

# DTO

Never expose database models.

Always use DTOs.

Example

```
CreateUserRequest

UpdateUserRequest

UserResponse
```

Database model should never be returned directly.

---

# Validation

Use Fiber request parsing with `go-playground/validator`.

Example

```go
Email string `json:"email" validate:"required,email"`
```

Parse request bodies in Fiber handlers with `c.BodyParser(&req)`, then validate DTOs before service execution.

---

# Responses

Use a consistent response format.

Success

```json
{
  "success": true,
  "data": {}
}
```

Error

```json
{
  "success": false,
  "message": "...",
  "errors": []
}
```

Do not invent different response formats.

---

# Error Handling

Never panic.

Always return errors.

Convert errors into HTTP responses inside handlers.

---

# Logging

Use structured logging.

Preferred:

- slog
- zap

Never use fmt.Println() in production code.

---

# Authentication

Authentication middleware should:

- Verify JWT
- Load current user
- Store user inside context

Handlers should access current user via helper functions.

---

# Routing

Each module owns its routes.

Example

```
internal/user/routes.go

internal/auth/routes.go

internal/billing/routes.go
```

main.go only registers modules.

---

# Database

Use PostgreSQL.

Repositories should use:

- sqlc (preferred)
- GORM (acceptable)

Avoid raw SQL in handlers or services.

---

# Transactions

Transactions belong in the service layer.

Repositories should receive tx when required.

---

# Configuration

Configuration comes only from environment variables.

Example

```
DATABASE_URL

JWT_SECRET

REDIS_URL

PORT
```

Never hardcode secrets.

---

# Middleware

Common middleware:

- Recovery
- Logger
- Request ID
- CORS
- Rate Limiter
- Authentication

Keep middleware reusable.

---

# API Versioning

Use:

```
/api/v1
```

Future versions:

```
/api/v2
```

Never remove existing endpoints without versioning.

---

# Naming

Use singular package names.

Good

```
user

auth

billing
```

Avoid

```
users

handlers

services
```

---

# Testing

Every service should have unit tests.

Repositories should have integration tests.

Business logic should never require HTTP to test.

---

# Performance

Prefer:

- Prepared queries
- Connection pooling
- Batch operations
- Pagination
- Context timeout

Avoid:

- N+1 queries
- Large memory allocations
- Loading unnecessary columns

---

# Code Style

Prefer early returns.

Keep functions small.

Target:

- <50 lines per function

Avoid nested if statements.

Extract helper functions when necessary.

---

# Comments

Write comments only when explaining WHY.

Do not comment obvious code.

Good

```go
// Prevent duplicate attendance within configured interval.
```

Bad

```go
// Increment i.
i++
```

---

# AI Instructions

When generating code:

- Follow existing architecture.
- Reuse existing packages.
- Never duplicate logic.
- Prefer composition over inheritance.
- Keep handlers thin.
- Keep services focused.
- Return typed errors.
- Keep code production-ready.
- Write idiomatic Go.
- Avoid unnecessary abstractions.
- If unsure, follow the existing project conventions instead of introducing new patterns.
