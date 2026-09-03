# Cookiecutter-style justfile for SAGE Backend

set shell := ["sh", "-cu"]

# Go bin path (for migrate CLI)
GOBIN := "/Users/navicsteinchinemerem/go/bin"
up args='':
  docker-compose up -d {{args}}

# Stop services
down args='':
  docker-compose down {{args}}

# Rebuild and start services
build service='':
  @if [ -z "{{service}}" ]; then \
    docker-compose build --no-cache && docker-compose up -d; \
  else \
    docker-compose build --no-cache {{service}} && docker-compose up -d {{service}}; \
  fi

# View logs (pass service name as argument, e.g., `just logs postgres`)
logs service='':
  @if [ -z "{{service}}" ]; then \
    docker-compose logs -f; \
  else \
    docker-compose logs -f {{service}}; \
  fi

# Run management/CLI commands inside containers
# Usage: just manage postgres psql -U sage_user -d sage_db
manage container *cmd:
  docker-compose exec {{container}} {{cmd}}

# Run one-off commands with service dependencies
run service *cmd:
  docker-compose run --rm {{service}} {{cmd}}

# PostgreSQL-specific shortcuts
psql *args:
  docker-compose exec postgres psql -U sage_user -d sage_db {{args}}

pg_dump *args:
  docker-compose exec postgres pg_dump -U sage_user -d sage_db {{args}}

# Database migrations and seed runner
migrate args='up':
  go run migrations/main.go --{{args}}

migrate-seed:
  go run migrations/main.go --seed

# Build and run the application locally (outside Docker)
run-local:
  go run ./cmd/api/main.go

# Test suite
test:
  go test -v ./...

test-coverage:
  go test -v -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out -o coverage.html

# Linting and formatting
lint:
  golangci-lint run

fmt:
  go fmt ./...
  goimports -w .

# Development helpers
tidy:
  go mod tidy

generate:
  go generate ./...

swagger:
	@echo "Generating Swagger documentation..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g main.go -d ./cmd/api,./internal/identity,./internal/organization,./internal/shared --parseInternal -o ./docs/users
	swag init -g main.go -d ./cmd/shield,./internal/shield,./internal/shared --parseInternal -o ./docs/shield
	@echo "Swagger docs updated successfully."

# Health check
health:
  docker-compose exec -T postgres pg_isready -U sage_user -d sage_db

# Show service status
status:
  docker-compose ps

# Clean everything (WARNING: destructive)
clean:
  docker-compose down -v --remove-orphans

# Prune unused resources
prune service='':
  @if [ -z "{{service}}" ]; then \
    docker-compose down -v --remove-orphans; \
  else \
    docker-compose down -v {{service}}; \
  fi

# Show help
help:
  @echo "SAGE Backend - Development Commands"
  @echo ""
  @echo "Usage: just <target> [arguments]"
  @echo ""
  @echo "Docker/Compose Commands:"
  @echo "  up [args]            Start services in detached mode"
  @echo "  down [args]          Stop services"
  @echo "  build [service]      Rebuild and start service(s)"
  @echo "  logs [service]       View logs (no service = all)"
  @echo "  status               Show container status"
  @echo "  clean                Stop and remove everything"
  @echo "  prune [service]      Remove specific service or all"
  @echo ""
  @echo "Container Commands:"
  @echo "  manage <c> <cmd>...  Run command in container"
  @echo "  run <s> <cmd>...     Run one-off command in service"
  @echo "  psql                 Connect to PostgreSQL via psql"
  @echo "  pg_dump              Dump database to stdout"
  @echo "  health               Check PostgreSQL health"
  @echo ""
  @echo "Development:"
  @echo "  migrate              Apply incremental migrations (requires migrate CLI)"
  @echo "  migrate-down         Rollback all migrations"
  @echo "  run-local            Run application locally"
  @echo "  test                 Run tests"
  @echo "  test-coverage        Generate coverage report"
  @echo "  lint                 Run linter"
  @echo "  fmt                  Format code"
  @echo "  tidy                 Tidy dependencies"
  @echo "  generate             Run go generate"
  @echo ""
  @echo "Examples:"
  @echo "  just up              # Start all services"
  @echo "  just logs postgres   # Follow postgres logs"
  @echo "  just psql            # Connect to database"
  @echo "  just manage postgres psql -U sage_user -d sage_db"
