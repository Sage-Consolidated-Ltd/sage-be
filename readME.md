# Sage Backend

Backend services for the Sage platform, implemented in Go.

## Tech Stack

- Go 1.25
- Fiber v2
- PostgreSQL
- Redis
- Swaggo (Swagger)

## Repository Layout

- `cmd/api/main.go`: main API entrypoint
- `cmd/offensive/main.go`: offensive service entrypoint (placeholder)
- `cmd/shield/main.go`: shield service entrypoint (in progress)
- `cmd/vision/main.go`: vision service entrypoint (placeholder)
- `internal/`: domain modules and shared application code
- `migrations/`: SQL migrations and migration/seed runner
- `seeds/`: SQL seed scripts
- `docs/`: generated Swagger artifacts

## Prerequisites

- Go installed and available in PATH
- PostgreSQL running
- Redis running

## Quick Start

1. Install dependencies:

```bash
go mod tidy
```

2. Create `.env` in the project root (see template below).

3. Run migrations:

```bash
go run migrations/main.go --up
```

4. (Optional) Seed data:

```bash
go run migrations/main.go --seed
```

5. Start the API:

```bash
go run cmd/api/main.go
```

API base URL: `http://localhost:3333/api/v1`

## Environment Variables

Create a `.env` file in the repository root:

```env
APP_ENV=development
DATABASE_URL=postgres://postgres:postgres@localhost:5432/sage?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFE=300
APP_ENCRYPTION_KEY=replace_with_32_byte_key
REDIS_DB_URL=redis://localhost:6379

LOG_LEVEL=info
RESEND_API_KEY=replace_me
RESEND_FROM_EMAIL=onboarding@resend.dev
FRONTEND_BASE_URL=http://localhost:3000

JWT_SECRET=replace_me
PORT=3333
GOMNI_SECURITY_KEY=replace_me

GOOGLE_CLIENT_ID=replace_me
GOOGLE_CLIENT_SECRET=replace_me
GOOGLE_REDIRECT_URL=http://localhost:3333/api/v1/auth/callback/google

GITHUB_CLIENT_ID=replace_me
GITHUB_CLIENT_SECRET=replace_me
GITHUB_REDIRECT_URL=http://localhost:3333/api/v1/auth/callback/github

AZURE_CLIENT_ID=replace_me
AZURE_CLIENT_SECRET=replace_me
AZURE_REDIRECT_URL=http://localhost:3333/api/v1/auth/callback/azure

OFFENSE_PORT=3334
SHIELD_PORT=3335
VISION_PORT=3336
```

Do not commit real credentials or production secrets.

## Migration and Seed Commands

```bash
# Apply all up migrations
go run migrations/main.go --up

# Revert one migration step
go run migrations/main.go --down

# Reset migration state to -1
go run migrations/main.go --reset

# Force specific migration version
go run migrations/main.go --force <version>

# Run all seed files in ./seeds
go run migrations/main.go --seed

# Run a specific seed file
go run migrations/main.go --seed-file seeds/001_seed_industry.sql
```

## Swagger Docs

Generate docs:

```bash
swag init -g cmd/api/main.go --parseDependency --parseInternal
swag init -g cmd/shield/main.go --parseDependency --parseInternal
```

After starting the API, open: `http://localhost:3333/api/v1docs`

## Run Other Services

```bash
go run cmd/shield/main.go
go run cmd/offensive/main.go
go run cmd/vision/main.go
```

These services are currently placeholders/in progress and may not expose HTTP endpoints yet.

## Basic Verification

```bash
go test ./...
```
