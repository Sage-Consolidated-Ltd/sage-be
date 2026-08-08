# syntax=docker/dockerfile:1
ARG GO_VERSION=1.25

# ---- shared deps layer --------------------------------------------------
FROM golang:${GO_VERSION}-alpine AS deps
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

# ---- build whichever binary was asked for --------------------------------
FROM deps AS builder
ARG BUILD_PATH
RUN test -n "$BUILD_PATH" || (echo "BUILD_PATH build-arg is required" && exit 1)

# Install swag CLI
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy full source
COPY . .

# Regenerate Swagger specs directly in source tree before compilation
RUN swag init -g main.go -d ./cmd/api,./internal/identity,./internal/organization,./internal/shared --parseInternal -o ./docs/users
RUN swag init -g main.go -d ./cmd/shield,./internal/shield,./internal/shared --parseInternal -o ./docs/shield

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/service ./${BUILD_PATH}

# ---- slim runtime ---------------------------------------------------------
FROM alpine:3.19 AS runtime
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/service ./service
COPY --from=builder /src/docs ./docs
COPY --from=builder /src/migrations ./migrations
COPY --from=builder /src/seeds ./seeds
ENTRYPOINT ["./service"]
