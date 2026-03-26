FROM golang:1.25-alpine AS builder
WORKDIR /app

# Download dependencies first (layer-cached separately from source)
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /blue-ledger-api ./cmd/server

# ── Runtime Stage ────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -u 1001 appuser

WORKDIR /app

# Copy binary and migrations
COPY --from=builder /blue-ledger-api .
COPY --from=builder /app/migrations ./migrations

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/v1/healthz || exit 1

# Write JWT keys from env vars to files at startup (production / Fly.io).
# JWT_PRIVATE_KEY and JWT_PUBLIC_KEY are set as Fly.io secrets.
# Falls back to existing files in dev (docker-compose mounts keys/).
CMD ["sh", "-c", "if [ -n \"$JWT_PRIVATE_KEY\" ]; then mkdir -p /app/keys && printf '%s' \"$JWT_PRIVATE_KEY\" > /app/keys/private.pem && printf '%s' \"$JWT_PUBLIC_KEY\" > /app/keys/public.pem; fi && exec ./blue-ledger-api"]
