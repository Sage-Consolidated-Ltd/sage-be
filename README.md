# Sage Backend (Clean Architecture)

Backend services for the Sage platform, implemented in Go following Clean Architecture and SOLID principles. A comprehensive security and data management platform with multi-tenant organization support, log ingestion, parsing, and data quality analysis.

## Tech Stack

- Go 1.25
- Fiber v2
- PostgreSQL
- Redis
- Swaggo (Swagger)
- Session & JWT Authentication
- Resend (Email Service)

## Architecture & Repository Layout

The codebase strictly enforces **Clean Architecture** (Separation of Concerns, Dependency Inversion) and **SOLID Principles**:

- `cmd/`: Microservice entry points (Composition Roots)
  - `cmd/api/main.go`: Users API service entrypoint (`app.NewAPIApp()`)
  - `cmd/shield/main.go`: Security/Log Processing service entrypoint (`app.NewShieldApp()`)
  - `cmd/worker/`: Background async worker service
- `internal/`: Domain modules and application logic
  - `app/`: Modular application bootstrap and lifecycle runners
    - `app.go`: HTTP server lifecycle and graceful shutdown runner
    - `api/app.go`: Assembles dependencies and routes for the Users API service
    - `shield/app.go`: Assembles dependencies and routes for the Shield API service
  - `users/`: Users domain layer (Entities, Use Cases, Ports, Inbound/Outbound Adapters)
  - `shield/`: Shield domain layer (Entities, Use Cases, Ports, Inbound/Outbound Adapters)
  - `shared/`: Cross-cutting concerns (Config, DB, Logger, Mailer, Storage, Middlewares)
- `migrations/`: SQL migrations and migration runner
- `seeds/`: SQL seed scripts
- `docs/`: Generated Swagger API artifacts
  - `docs/users/`: Swagger documentation for Users API
  - `docs/shield/`: Swagger documentation for Shield API

## Features

### Users & Authentication Service (`cmd/api`)

- **Authentication**
  - User registration with organization creation
  - Email/password login with JWT tokens & Redis Session store
  - OAuth 2.0 integrations (Google, GitHub, Azure)
  - Two-factor authentication (TOTP-based)
  - Password management (forgot password, reset, change)
  - Email verification flow

- **Profile & Identity**
  - Profile management (name, email, avatar upload to S3)
  - User preferences (theme, notifications, etc.)
  - Activity/audit logging with IP tracking
  - Session management (create, list, revoke sessions)

- **Organization & Access Control**
  - Create and update multi-tenant organizations
  - Member management (bulk & single email invitations)
  - Role & Permissions system (Owner, Admin, Member, and Custom Roles)

### Shield Service - Security & Log Processing (`cmd/shield`)

- **Log/Event Ingestion**
  - Single and bulk event ingestion
  - Advanced event search and filtering
  - Event detail retrieval with parser history

- **Data Source Management**
  - Log data source registration and health monitoring
  - Ingestion metrics (volume, rates, errors)
  - Sync and disconnect sources

- **Advanced Parser Engine**
  - Parser types: Regex, JSON, CSV, Key-Value, AI/NLP
  - Create, read, update, list, delete, and test parsers
  - Preview parsing logic without saving
  - Parser validation, import/export

- **Data Quality Management**
  - Comprehensive quality analysis scans
  - Overall quality score, error tracking, unmapped logs
  - AI-powered insights & automated fixes with diff preview
  - Export quality reports (CSV, PDF, JSON)

## Prerequisites

- Go 1.25+ installed
- PostgreSQL running
- Redis running

## Environment Variables

Create a `.env` file in the repository root:

```env
APP_ENV=development
DATABASE_URL=postgres://sage_user:sage_dev_password@localhost:5432/sage_db?sslmode=disable
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
PORT=4000
SHIELD_PORT=3335
GOMNI_SECURITY_KEY=replace_me

GOOGLE_CLIENT_ID=replace_me
GOOGLE_CLIENT_SECRET=replace_me
GOOGLE_REDIRECT_URL=http://localhost:4000/api/v1/auth/callback/google

GITHUB_CLIENT_ID=replace_me
GITHUB_CLIENT_SECRET=replace_me
GITHUB_REDIRECT_URL=http://localhost:4000/api/v1/auth/callback/github

AZURE_CLIENT_ID=replace_me
AZURE_CLIENT_SECRET=replace_me
AZURE_REDIRECT_URL=http://localhost:4000/api/v1/auth/callback/azure

S3_BUCKET=sage-uploads
S3_REGION=us-east-1
AWS_ACCESS_KEY_ID=replace_me
AWS_SECRET_ACCESS_KEY=replace_me
```

## Quick Start

1. Install dependencies:

```bash
go mod tidy
```

2. Run migrations:

```bash
go run migrations/main.go --up
```

3. (Optional) Seed data:

```bash
go run migrations/main.go --seed
```

4. Start the Users API service:

```bash
go run ./cmd/api
```

Users API base URL: `http://localhost:4000/api/v1`

5. Start the Shield API service:

```bash
go run ./cmd/shield
```

Shield API base URL: `http://localhost:3335/api/v1`

## Swagger Documentation

### Generate Swagger Specs

To regenerate Swagger documentation for individual services:

```bash
# Generate Users API docs
swag init -g main.go -d ./cmd/api,./internal/users,./internal/shared -o ./docs/users

# Generate Shield API docs
swag init -g main.go -d ./cmd/shield,./internal/shield,./internal/shared -o ./docs/shield
```

### Accessing Swagger UI

- **Users API Documentation**: `http://localhost:4000/docs/api-docs`
- **Shield API Documentation**: `http://localhost:3335/docs/shield-docs`

## Testing & Linting

```bash
# Run test suite
go test -v ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run
```
