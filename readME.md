# Sage Backend

Backend services for the Sage platform, implemented in Go. A comprehensive security and data management platform with multi-tenant organization support, log ingestion, parsing, and data quality analysis.

## Tech Stack

- Go 1.25
- Fiber v2
- PostgreSQL
- Redis
- Swaggo (Swagger)
- Session Authentication
- Resend (Email Service)

## Repository Layout

- `cmd/api/main.go`: main API entrypoint (port 3333)
- `cmd/shield/main.go`: security/log processing service (port 3335)
- `cmd/offensive/main.go`: offensive security service (port 3334, placeholder)
- `cmd/vision/main.go`: vision/analytics service (port 3336, placeholder)
- `internal/`: domain modules and shared application code
  - `users/`: User authentication, profiles, and session management
  - `shared/`: Organization, roles, permissions, and multi-tenancy
  - `shield/`: Log ingestion, parsing, quality analysis, and integrations
  - `offensive/`: Offensive security module (placeholder)
  - `vision/`: Vision/analytics module (placeholder)
- `migrations/`: SQL migrations (33+ migrations) and migration/seed runner
- `seeds/`: SQL seed scripts
- `docs/`: generated Swagger artifacts

## Features

### Users & Authentication Service

- **Authentication**
  - User registration with organization creation
  - Email/password login with JWT tokens
  - OAuth 2.0 integrations (Google, GitHub, Azure)
  - Two-factor authentication (TOTP-based)
  - Password management (forgot password, reset, change)
  - Email verification flow

- **Profile & Identity**
  - Profile management (name, email, avatar)
  - User preferences (theme, notifications, etc.)
  - Activity/audit logging with IP tracking
  - Session management (create, list, revoke sessions)

### Organization & Access Control

- **Organization Management**
  - Create and update organizations
  - Organization metadata and settings
  - Industries catalog
  - Multi-tenant support

- **Member Management**
  - Invite members by email (bulk and single)
  - List members with pagination and filtering
  - Update member roles and departments
  - Remove members
  - Accept invitations with token validation

- **Role & Permissions System**
  - Built-in roles: Owner, Admin, Member
  - Custom roles with granular permissions
  - Permission groups and access control
  - Role-based authorization throughout the platform

### Shield Service - Security & Log Processing

- **Log/Event Ingestion**
  - Single and bulk event ingestion
  - Advanced event search and filtering
  - Event detail retrieval with parser history
  - Batch processing capabilities

- **Data Source Management**
  - Create and register log data sources
  - Health monitoring with metrics (events/day, total events, errors)
  - Per-source log retrieval and analysis
  - Aggregated health dashboard
  - Sync and disconnect sources

- **Advanced Parser Engine**
  - Multiple parser types: Regex, JSON, CSV, Key-Value, AI/NLP
  - Create, read, update, list, delete parsers
  - Test parsers with sample logs
  - Preview parsing logic without saving
  - Parser validation and versioning
  - Import/Export parser configurations
  - Track 24h parse count and error rates
  - Parser status tracking (active, warning, error, disabled)

- **Data Quality Management**
  - Run comprehensive quality analysis scans
  - Metrics: overall quality score, parsing errors, missing fields, duplicates, unmapped logs
  - AI-powered insights and recommendations
  - Automated fixes with diff preview
  - Export quality reports (CSV, PDF, JSON)
  - Track quality metrics over time

- **Ingestion Job Management**
  - Track job status (queued, running, completed, failed, cancelled)
  - Job types: parse, sync, validate
  - Event processing metrics and error tracking
  - Job history and scheduling

- **Health Monitoring & Notifications**
  - Real-time ingestion health dashboard
  - Total events, active sources, delayed sources, error sources
  - Event volume analytics (by hour/day)
  - Download health reports (CSV, PDF, JSON)
  - Alert notifications on ingestion issues

- **Integration Management**
  - Create provider-specific integrations (Kafka, S3, API, etc.)
  - Encrypted credential storage per integration
  - Stream tracking with last offset/checkpoint
  - Integration status monitoring
  - Event buffering for reliability
  - Connection testing before activation

### Cross-Service Features

- **Security & Authentication**
  - JWT authentication middleware
  - Session management with Redis
  - Audit logging with IP tracking

- **Email & Notifications**
  - Transactional emails via Resend API
  - User notification preferences
  - Multi-channel support (Email, Push, Slack)

- **API Documentation**
  - Swagger/OpenAPI integration
  - Auto-generated API documentation

- **Health & Status**
  - Health check endpoints on all services
  - Structured logging with configurable levels

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
swag init -g main.go -d ./cmd/api,./internal/users,./internal/shared --parseInternal -o ./docs/users
swag init -g main.go -d ./cmd/shield,./internal/shield,./internal/shared --parseInternal -o ./docs/shield
```

backend.sageconsolidated.com - Deployment

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
