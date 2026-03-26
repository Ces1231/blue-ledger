FROM golang:1.24-alpine AS builder
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

CMD ["./blue-ledger-api"]
