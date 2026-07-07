# syntax=docker/dockerfile:1
ARG GO_VERSION=1.25

# ---- shared deps layer --------------------------------------------------
FROM golang:${GO_VERSION}-alpine AS deps
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# ---- regenerate swagger docs on every image build -----------------------
# Only reruns when source actually changes, since docker/buildkit caches
# this stage's layers and every service (api/shield/offensive/vision/
# worker/migrate) shares it.
FROM deps AS docsgen
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g main.go -d ./cmd/api,./internal/users,./internal/shared \
    --parseInternal -o ./docs/users
RUN swag init -g main.go -d ./cmd/shield,./internal/shield,./internal/shared \
    --parseInternal -o ./docs/shield

# ---- build whichever binary was asked for --------------------------------
FROM deps AS builder
ARG BUILD_PATH
RUN test -n "$BUILD_PATH" || (echo "BUILD_PATH build-arg is required" && exit 1)
COPY --from=docsgen /src/docs ./docs
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/service ./${BUILD_PATH}

# ---- slim runtime ---------------------------------------------------------
FROM alpine:3.19 AS runtime
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/service ./service
COPY --from=builder /src/docs ./docs
ENTRYPOINT ["./service"]