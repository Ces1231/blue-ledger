.PHONY: up down build logs migrate keys dev test

# ── Docker Compose ────────────────────────────────────────────────────────────

## Start everything (postgres + redis + api + web)
up:
	docker compose up --build

## Start in background
up-d:
	docker compose up --build -d

## Stop all containers
down:
	docker compose down

## Rebuild images without cache
build:
	docker compose build --no-cache

## Tail logs (all services)
logs:
	docker compose logs -f

## Tail only API logs
logs-api:
	docker compose logs -f api

# ── Database ──────────────────────────────────────────────────────────────────

## Run migrations (requires golang-migrate installed or use docker)
migrate-up:
	migrate -path ./migrations -database "postgres://blue_ledger:localpassword@localhost:5432/blue_ledger?sslmode=disable" up

migrate-down:
	migrate -path ./migrations -database "postgres://blue_ledger:localpassword@localhost:5432/blue_ledger?sslmode=disable" down 1

# ── Local dev (no Docker for the API) ────────────────────────────────────────

## Run API locally against dockerized postgres+redis
dev:
	docker compose up -d postgres redis
	go run ./cmd/server

# ── Keys ─────────────────────────────────────────────────────────────────────

## Generate RSA key pair for JWT (only needed once)
keys:
	mkdir -p keys
	openssl genrsa -out keys/private.pem 2048
	openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo "Keys written to keys/private.pem and keys/public.pem"

# ── Testing ───────────────────────────────────────────────────────────────────

## Run Go tests
test:
	go test ./...

## Run React tests
test-web:
	cd ../web && npm test
