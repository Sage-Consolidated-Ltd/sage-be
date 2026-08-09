# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25

# ============================================================
# Dependencies
# ============================================================

FROM golang:${GO_VERSION}-alpine AS deps

RUN apk add --no-cache \
    git \
    ca-certificates

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download


# ============================================================
# Builder
# ============================================================

FROM deps AS builder

ARG BUILD_PATH

RUN test -n "$BUILD_PATH" || \
    (echo "BUILD_PATH build-arg is required" && exit 1)

# ------------------------------------------------------------
# Install Swagger generator
# ------------------------------------------------------------

RUN go install github.com/swaggo/swag/cmd/swag@latest

# ------------------------------------------------------------
# Copy source
# ------------------------------------------------------------

COPY . .

# ------------------------------------------------------------
# Generate Users Swagger
# ------------------------------------------------------------

RUN swag init \
    -g main.go \
    -d ./cmd/api,./internal/identity,./internal/organization,./internal/shared \
    --parseInternal \
    -o ./docs/users

# ------------------------------------------------------------
# Generate Shield Swagger
# ------------------------------------------------------------

RUN swag init \
    -g main.go \
    -d ./cmd/shield,./internal/shield,./internal/shared \
    --parseInternal \
    -o ./docs/shield

# ------------------------------------------------------------
# Verify Swagger generation
# ------------------------------------------------------------

RUN echo "========================================" && \
    echo "Generated Swagger files" && \
    echo "========================================" && \
    find ./docs -type f -print

# ------------------------------------------------------------
# Build application
# ------------------------------------------------------------

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/service \
    ./${BUILD_PATH}


# ============================================================
# Runtime
# ============================================================

FROM alpine:3.19 AS runtime

RUN apk add --no-cache \
    ca-certificates \
    tzdata

WORKDIR /app

# Application
COPY --from=builder /out/service ./service

# Swagger
COPY --from=builder /src/docs ./docs

# Migrations
COPY --from=builder /src/migrations ./migrations

# Seeds
COPY --from=builder /src/seeds ./seeds

ENTRYPOINT ["./service"]