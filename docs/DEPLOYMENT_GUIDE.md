# Blue Ledger Deployment Guide

**Version:** 1.0.0  
**Last Updated:** March 30, 2026  
**Status:** Production Ready

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Local Development](#local-development)
- [Production Deployment](#production-deployment)
- [Docker Deployment](#docker-deployment)
- [Fly.io Deployment](#flyio-deployment)
- [Environment Configuration](#environment-configuration)
- [Database Setup](#database-setup)
- [Health Checks](#health-checks)
- [Troubleshooting](#troubleshooting)
- [Monitoring](#monitoring)

---

## Prerequisites

### System Requirements
- **CPU:** 2+ cores minimum
- **RAM:** 4GB minimum
- **Storage:** 20GB minimum
- **OS:** Linux, macOS, or Windows with WSL2

### Required Tools

| Tool | Version | Purpose |
|---|---|---|
| Docker | 20.10+ | Container runtime |
| Docker Compose | 1.29+ | Multi-container orchestration |
| Go | 1.25+ | Backend compilation (optional) |
| Node.js | 18+ | Frontend build (optional) |
| PostgreSQL | 16 | Database (optional - included in Docker) |

### Installation

**Docker & Docker Compose (macOS/Linux):**
```bash
# macOS
brew install docker docker-compose

# Ubuntu/Debian
sudo apt-get install docker.io docker-compose
sudo usermod -aG docker $USER
```

**Fly.io CLI (Optional - for cloud deployment):**
```bash
curl -L https://fly.io/install.sh | sh
flyctl auth login
```

---

## Local Development

### Quick Start (30 seconds)

```bash
# Clone repository
git clone https://github.com/Ces1231/blue-ledger.git
cd blue-ledger/blue-ledger-api

# Start all services
docker-compose up

# Wait for "healthy" status
# API: http://localhost:8081
# Web: http://localhost:3001
```

### Manual Steps

**1. Build Images**
```bash
cd blue-ledger-api

# Build API
docker build -t blue-ledger-api:latest .

# Build Web
cd ../web
docker build -t blue-ledger-web:latest .
cd ..
```

**2. Create Environment File**
```bash
cat > blue-ledger-api/.env << EOF
DATABASE_URL=postgres://postgres:postgres@postgres:5432/blue_ledger
REDIS_URL=redis://redis:6379
JWT_SECRET=$(openssl rand -hex 32)
API_PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug
EOF
```

**3. Start Services**
```bash
docker-compose up -d

# Monitor logs
docker-compose logs -f
```

**4. Verify Status**
```bash
# Check containers
docker-compose ps

# Test API
curl http://localhost:8081/v1/healthz
```

### Development Workflow

```bash
# View logs
docker-compose logs -f blue-ledger-api

# Stop services
docker-compose down

# Rebuild specific service
docker-compose up -d --build blue-ledger-api

# Access database
docker exec -it blue-ledger-api-postgres-1 psql -U postgres -d blue_ledger

# Run migrations
docker-compose exec blue-ledger-api make migrate
```

---

## Production Deployment

### Pre-Deployment Checklist

- [ ] All tests passing locally
- [ ] Environment variables configured
- [ ] Database backups scheduled
- [ ] SSL certificates obtained
- [ ] Monitoring/alerting configured
- [ ] Logging aggregation set up
- [ ] Deployment plan documented

### Architecture Overview

```
                    ┌─────────────┐
                    │   Clients   │
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
    ┌───▼────┐         ┌───▼────┐        ┌───▼────┐
    │  CDN   │         │ Nginx  │        │  DNS   │
    └────────┘         └───┬────┘        └────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
    ┌───▼──────┐       ┌───▼──────┐      ┌───▼──────┐
    │  Web App │       │    API   │      │  Cache   │
    │ (React)  │       │   (Go)   │      │ (Redis)  │
    └────────┬─┘       └───┬──────┘      └──────────┘
             │             │
             └─────────────┼────────────────┐
                           │                │
                       ┌───▼────────────────▼──┐
                       │  PostgreSQL Database  │
                       │   (Master + Replica)  │
                       └───────────────────────┘
```

---

## Docker Deployment

### Production Docker Compose

```bash
cd blue-ledger-api

# Create production environment
cat > .env.prod << EOF
DATABASE_URL=postgres://user:password@db-prod:5432/blue_ledger_prod
REDIS_URL=redis://redis-prod:6379
JWT_SECRET=$(openssl rand -hex 32)
API_PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info
EOF

# Deploy with custom network
docker network create blue-ledger-network

# Start services
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Production docker-compose.prod.yml

```yaml
version: '3.8'

services:
  api:
    restart: always
    environment:
      - ENVIRONMENT=production
      - LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/v1/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3
    resources:
      limits:
        cpus: '1.0'
        memory: 512M
      reservations:
        cpus: '0.5'
        memory: 256M

  postgres:
    restart: always
    environment:
      - POSTGRES_PASSWORD_FILE=/run/secrets/db_password
    secrets:
      - db_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - /backups:/backups:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    restart: always
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

secrets:
  db_password:
    file: ./secrets/db_password.txt

volumes:
  postgres_data:
    driver: local
```

### Backup Strategy

```bash
# Daily backup script
cat > /usr/local/bin/backup-blue-ledger.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backups/blue-ledger"
DATE=$(date +%Y-%m-%d_%H-%M-%S)

mkdir -p $BACKUP_DIR

# Backup database
docker exec blue-ledger-api-postgres-1 pg_dump -U postgres blue_ledger | \
  gzip > $BACKUP_DIR/db_backup_$DATE.sql.gz

# Keep only last 30 days
find $BACKUP_DIR -name "db_backup_*.sql.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_DIR/db_backup_$DATE.sql.gz"
EOF

chmod +x /usr/local/bin/backup-blue-ledger.sh

# Schedule with cron (2 AM daily)
0 2 * * * /usr/local/bin/backup-blue-ledger.sh
```

---

## Fly.io Deployment

### Prerequisites
```bash
flyctl auth login
flyctl apps list
```

### Deploy API

```bash
cd blue-ledger-api

# Create app (if not exists)
flyctl app create blue-ledger-api

# Set secrets
flyctl secrets set \
  DATABASE_URL="postgres://..." \
  JWT_SECRET="$(openssl rand -hex 32)" \
  --app blue-ledger-api

# Deploy
flyctl deploy --remote-only --app blue-ledger-api

# View logs
flyctl logs --app blue-ledger-api

# Monitor
flyctl status --app blue-ledger-api
```

### Deploy Web

```bash
cd ../web

# Create app
flyctl app create blue-ledger-web

# Deploy
flyctl deploy --remote-only --app blue-ledger-web

# View logs
flyctl logs --app blue-ledger-web
```

### Fly.io fly.toml Configuration

```toml
# fly.toml - API
app = "blue-ledger-api"
primary_region = "sjc"

[build]
  image = "blue-ledger-api:latest"

[[services]]
  protocol = "tcp"
  internal_port = 8080
  processes = ["app"]

  [[services.ports]]
    port = 80
    handlers = ["http"]
    force_https = true

  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]

[env]
  ENVIRONMENT = "production"
  LOG_LEVEL = "info"

[checks]
  [checks.api_health]
    type = "http"
    interval = "10s"
    timeout = "5s"
    grace_period = "30s"
    method = "GET"
    path = "/v1/healthz"
    expected_status_codes = [200]
```

---

## Environment Configuration

### Required Environment Variables

```bash
# Database
DATABASE_URL=postgres://user:password@host:5432/blue_ledger

# Redis (optional but recommended)
REDIS_URL=redis://host:6379

# JWT
JWT_SECRET=your-secret-key-here-minimum-32-chars

# API
API_PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info

# Deployment
DEPLOYMENT_ENV=production
DEPLOYMENT_REGION=us-west-2
```

### Create .env File

```bash
cat > blue-ledger-api/.env << EOF
# Database
DATABASE_URL=postgres://postgres:$(openssl rand -hex 8)@localhost:5432/blue_ledger

# JWT Secret (generate with: openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 32)

# Server
API_PORT=8080
ENVIRONMENT=production

# Logging
LOG_LEVEL=info

# Feature Flags
ENABLE_SIGNUP=false
ENABLE_CSV_IMPORT=true
EOF
```

---

## Database Setup

### PostgreSQL Initialization

```bash
# Create database
createdb blue_ledger

# Run migrations
cd blue-ledger-api
make migrate

# Seed test data
make seed

# Verify
psql -d blue_ledger -c "\dt"
```

### Migration Commands

```bash
# Run migrations
docker-compose exec blue-ledger-api make migrate

# Rollback last migration
docker-compose exec blue-ledger-api make migrate-down

# View migration status
docker-compose exec blue-ledger-api make migrate-status
```

---

## Health Checks

### API Health Endpoint

```bash
curl http://localhost:8081/v1/healthz
# Response: {"status":"ok"}
```

### Container Health

```bash
# Check all containers
docker-compose ps

# Check specific service
docker-compose ps blue-ledger-api

# View health details
docker inspect blue-ledger-api-api-1 | jq '.State.Health'
```

### Database Health

```bash
# Test connection
psql -h localhost -U postgres -d blue_ledger -c "SELECT 1"

# View table count
psql -h localhost -U postgres -d blue_ledger -c "\dt"
```

---

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose logs blue-ledger-api

# Verify image exists
docker images | grep blue-ledger

# Rebuild image
docker-compose down
docker-compose up --build
```

### Database Connection Error

```bash
# Test PostgreSQL
psql -h localhost -U postgres

# View environment
docker-compose config | grep DATABASE_URL

# Verify password
docker exec blue-ledger-api-postgres-1 psql -U postgres -c "SELECT 1"
```

### Port Already in Use

```bash
# Find process on port 8081
lsof -i :8081
# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
```

### API Returns 500 Error

```bash
# Check API logs
docker-compose logs blue-ledger-api | tail -50

# Verify environment variables
docker-compose config | grep -E "JWT_SECRET|DATABASE_URL"

# Test database from container
docker-compose exec blue-ledger-api psql $DATABASE_URL -c "SELECT 1"
```

---

## Monitoring

### Log Aggregation

```bash
# Real-time logs
docker-compose logs -f

# Follow specific service
docker-compose logs -f blue-ledger-api

# Last 100 lines
docker-compose logs --tail=100
```

### Performance Monitoring

```bash
# Container stats
docker stats blue-ledger-api-api-1

# View resource limits
docker inspect blue-ledger-api-api-1 | jq '.HostConfig.Memory'
```

### Application Monitoring

```bash
# API response time
time curl http://localhost:8081/v1/healthz

# Database query performance
docker-compose exec blue-ledger-api psql $DATABASE_URL -c "EXPLAIN SELECT * FROM members LIMIT 1"
```

---

## Rollback Procedures

### Rollback to Previous Image

```bash
# Tag previous version
docker tag blue-ledger-api:v1.0.0 blue-ledger-api:latest

# Restart with previous version
docker-compose up -d --no-build
```

### Database Rollback

```bash
# Restore from backup
gunzip -c /backups/blue-ledger/db_backup_2026-03-30_02-00-00.sql.gz | \
  psql -U postgres -d blue_ledger

# Rollback single migration
docker-compose exec blue-ledger-api make migrate-down
```

---

## Support

For deployment issues:
- **GitHub Issues:** https://github.com/Ces1231/blue-ledger/issues
- **Documentation:** https://github.com/Ces1231/blue-ledger#readme
