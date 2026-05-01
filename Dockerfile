# Multi-stage build for Go application
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates make gcc musl-dev

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both API and worker binaries
WORKDIR /app/cmd/api
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/api .

WORKDIR /app/cmd/worker
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/worker .

# Final stage - minimal runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the compiled binaries from builder stage
COPY --from=builder /app/api .
COPY --from=builder /app/worker .
COPY --from=builder /app/docs ./docs

# Run as non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser

# No default CMD - each service defines its own in docker-compose
