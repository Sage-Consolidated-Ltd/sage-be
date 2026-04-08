# Sage Backend

Go backend service for the Sage platform.

## Tech Stack

- Go 1.25+
- Fiber v2
- PostgreSQL
- Redis
- Session Auth
- Swagger (Swaggo)

## Project Structure

- `cmd/api/main.go`: main API service entrypoint
- `cmd/offensive/main.go`: offensive service entrypoint (placeholder)
- `cmd/shield/main.go`: shield service entrypoint (placeholder)
- `cmd/vision/main.go`: vision service entrypoint (placeholder)
- `internal/`: domain and shared application code
- `migrations/`: SQL migrations and migration runner
- `docs/`: generated Swagger files

## Prerequisites

- Go installed and available in PATH
- PostgreSQL running
- Redis running

## Environment Variables

Create a `.env` file in the repository root. The API currently requires:

```env
APP_ENV=
DATABASE_URL=
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFE=300
APP_ENCRYPTION_KEY=
REDIS_DB_URL=

JWT_SECRET=
PORT=3333
GOMNI_SECURITY_KEY=

GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=

GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITHUB_REDIRECT_URL=

AZURE_CLIENT_ID=
AZURE_CLIENT_SECRET=
AZURE_REDIRECT_URL=
```

## Install Dependencies

```bash
go mod tidy
```

## Run Database Migrations

From the repository root:

```bash
go run migrations/main.go --up
```

Other migration operations:

```bash
go run migrations/main.go --down
go run migrations/main.go --reset
go run migrations/main.go --force <version>
```

## Run API

```bash
go run cmd/api/main.go
```

Default API base URL:

```text
http://localhost:3333/api/v1
```

## Swagger Docs

Generate Swagger docs:

```bash
swag init -g cmd/api/main.go -d ./,internal/users/handlers,internal/users/models,internal/users/requests --parseDependency --parseInternal
```

Once the API is running, open:

```text
http://localhost:3333/docs
```
