FROM golang:1.25-alpine AS builder
WORKDIR /app

# Download dependencies first (layer-cached separately from source)
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /blue-ledger-api ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /import-csv-embedded ./cmd/import-csv-embedded

# ── Runtime Stage ────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -u 1001 appuser

WORKDIR /app

# Copy binary and migrations
COPY --from=builder /blue-ledger-api .
COPY --from=builder /import-csv-embedded .
COPY --from=builder /app/migrations ./migrations

# Pre-create keys dir with correct ownership so the startup CMD can write
# JWT key files from env vars when running in production (Fly.io).
RUN mkdir -p /app/keys && chown -R appuser:appuser /app/keys

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/v1/healthz || exit 1

# Startup: write JWT keys, then start server (migrations + importer handled by server)
CMD ["sh", "-c", "if [ -n \"$JWT_PRIVATE_KEY\" ]; then mkdir -p /app/keys && printf '%s' \"$JWT_PRIVATE_KEY\" > /app/keys/private.pem && printf '%s' \"$JWT_PUBLIC_KEY\" > /app/keys/public.pem; fi; exec ./blue-ledger-api"]
