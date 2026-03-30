# Blue Ledger Development Setup Guide

**Version:** 1.0.0  
**Last Updated:** March 30, 2026  
**Status:** Production Ready

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Backend Setup](#backend-setup)
- [Frontend Setup](#frontend-setup)
- [Database Setup](#database-setup)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Debugging](#debugging)
- [Common Issues](#common-issues)
- [IDE Setup](#ide-setup)

---

## Prerequisites

### System Requirements

- **OS**: macOS 10.15+, Ubuntu 20.04+, or Windows 10+ (WSL2)
- **RAM**: 8GB minimum (16GB recommended)
- **Storage**: 20GB free space
- **CPU**: 4+ cores

### Required Software

| Tool | Version | Command |
|---|---|---|
| Git | 2.30+ | `git --version` |
| Docker | 20.10+ | `docker --version` |
| Docker Compose | 1.29+ | `docker-compose --version` |
| Go | 1.25+ | `go version` |
| Node.js | 18+ | `node --version` |
| npm | 9+ | `npm --version` |

### Optional Tools

- **PostgreSQL CLI** - for direct database access
- **Redis CLI** - for cache testing
- **Postman** - for API testing
- **VS Code** - recommended editor

---

## Installation

### macOS

```bash
# Install Homebrew (if not already installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install dependencies
brew install git docker docker-compose go node

# Start Docker daemon
open -a Docker
```

### Ubuntu/Debian

```bash
# Update packages
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y git docker.io docker-compose golang-1.25 nodejs npm

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Start Docker
sudo systemctl start docker
sudo systemctl enable docker
```

### Windows (WSL2)

```bash
# Install WSL2
# Go to Microsoft Store → Install Ubuntu 20.04 or 22.04

# In Ubuntu terminal:
sudo apt update && sudo apt upgrade -y
sudo apt install -y git docker docker.io docker-compose golang-1.25 nodejs npm

# Configure Docker
sudo usermod -aG docker $USER
```

---

## Project Structure

### Repository Layout

```
blue-ledger/
│
├── blue-ledger-api/              # Go backend
│   ├── cmd/
│   │   ├── server/main.go        # API entry point
│   │   ├── migrate/main.go       # Migrations
│   │   └── seed/main.go          # Seed data
│   ├── internal/                 # Private packages
│   │   ├── members/              # Member module
│   │   ├── auth/                 # Auth module
│   │   ├── platform/             # Config module
│   │   └── [20+ modules]
│   ├── pkg/                      # Shared packages
│   │   ├── config/
│   │   ├── db/
│   │   ├── logger/
│   │   └── middleware/
│   ├── migrations/               # SQL migrations
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
│
├── web/                          # React frontend
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── services/
│   │   ├── context/
│   │   ├── types/
│   │   └── App.tsx
│   ├── public/
│   ├── Dockerfile
│   ├── nginx.conf
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   ├── package.json
│   └── package-lock.json
│
├── docs/                         # Documentation
│   ├── API_DOCUMENTATION.md
│   ├── DEPLOYMENT_GUIDE.md
│   ├── ARCHITECTURE.md
│   ├── USER_GUIDE.md
│   └── DEVELOPMENT_SETUP.md
│
├── data/                         # Sample data
└── README.md
```

---

## Backend Setup

### Clone Repository

```bash
git clone https://github.com/Ces1231/blue-ledger.git
cd blue-ledger
```

### Navigate to Backend

```bash
cd blue-ledger-api
```

### Install Go Dependencies

```bash
# Download dependencies
go mod download

# Verify dependencies
go mod tidy
```

### Create Environment File

```bash
cat > .env << EOF
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/blue_ledger

# Redis
REDIS_URL=redis://localhost:6379

# JWT Secret (generate: openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 32)

# Server
API_PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug

# Chapter (optional)
CHAPTER_ID=8f22b21e-624a-413a-a6d9-67756233ca7f
CHAPTER_NAME=Tau Sigma Sigma

# Features
ENABLE_SIGNUP=true
ENABLE_CSV_IMPORT=true
EOF
```

### Build Backend

```bash
# Build the application
go build -o bin/server ./cmd/server

# Run the application
./bin/server
```

Or using Docker:

```bash
# Build Docker image
docker build -t blue-ledger-api:latest .

# Run container
docker run -p 8081:8080 --env-file .env blue-ledger-api:latest
```

### Verify Backend

```bash
# Test health endpoint
curl http://localhost:8080/v1/healthz

# Expected response:
# {"status":"ok"}
```

---

## Frontend Setup

### Navigate to Frontend

```bash
cd ../web
```

### Install Node Dependencies

```bash
npm install
```

### Create Environment File

```bash
cat > .env.local << EOF
VITE_API_URL=http://localhost:8080
VITE_APP_NAME=Blue Ledger
VITE_APP_ENV=development
EOF
```

### Start Development Server

```bash
npm run dev
```

Output will show:
```
VITE v4.x.x  ready in xxx ms

➜  Local:   http://localhost:5173/
➜  press h to show help
```

### Build for Production

```bash
# Build optimized bundle
npm run build

# Preview production build
npm run preview
```

### Verify Frontend

Open browser to: `http://localhost:5173`

You should see the Blue Ledger login screen.

---

## Database Setup

### Using Docker Compose (Recommended)

```bash
cd ../blue-ledger-api

# Start all services
docker-compose up -d

# Verify services
docker-compose ps

# Check logs
docker-compose logs -f postgres
```

### Manual Setup

#### 1. Install PostgreSQL

**macOS:**
```bash
brew install postgresql@16
brew services start postgresql@16
```

**Ubuntu:**
```bash
sudo apt install -y postgresql-16
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### 2. Create Database

```bash
# Login to PostgreSQL
psql -U postgres

# Create database and user
CREATE DATABASE blue_ledger;
CREATE USER blue_ledger_user WITH ENCRYPTED PASSWORD 'password';
ALTER ROLE blue_ledger_user SET client_encoding TO 'utf8';
ALTER ROLE blue_ledger_user SET default_transaction_isolation TO 'read committed';
ALTER ROLE blue_ledger_user SET default_transaction_deferrable TO on;
ALTER ROLE blue_ledger_user SET default_transaction_read_only TO off;
GRANT ALL PRIVILEGES ON DATABASE blue_ledger TO blue_ledger_user;

\q
```

#### 3. Run Migrations

```bash
cd blue-ledger-api

# Set environment variable
export DATABASE_URL="postgres://blue_ledger_user:password@localhost:5432/blue_ledger"

# Run migrations
make migrate
```

#### 4. Seed Data (Optional)

```bash
make seed
```

### Verify Database

```bash
# Connect to database
psql -U blue_ledger_user -d blue_ledger

# List tables
\dt

# Count members
SELECT COUNT(*) FROM members;

# Exit
\q
```

---

## Development Workflow

### Starting Development

```bash
# Terminal 1: Backend
cd blue-ledger-api
docker-compose up -d
go run ./cmd/server

# Terminal 2: Frontend
cd web
npm run dev

# Terminal 3: Monitor logs
docker-compose logs -f
```

### Making Changes

**Backend:**
```bash
# Edit Go files
# The application will NOT auto-reload
# Stop (Ctrl+C) and restart

go run ./cmd/server
```

**Frontend:**
```bash
# Edit React files
# The app WILL auto-reload with HMR
# Changes appear instantly in browser
```

### Database Changes

**Create Migration:**
```bash
# Create migration file
cd blue-ledger-api
touch migrations/001_add_new_column.up.sql
touch migrations/001_add_new_column.down.sql

# Edit files with SQL
nano migrations/001_add_new_column.up.sql
nano migrations/001_add_new_column.down.sql

# Run migration
make migrate
```

**Migration Template:**
```sql
-- UP
ALTER TABLE members ADD COLUMN new_field VARCHAR(255);

-- DOWN
ALTER TABLE members DROP COLUMN new_field;
```

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/my-feature

# Make changes
# ... edit files ...

# Commit changes
git add .
git commit -m "feat: add new feature

- Added member profile update
- Added validation
- Added tests"

# Push to GitHub
git push origin feature/my-feature

# Create pull request on GitHub
```

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `style` - Code style (no logic change)
- `refactor` - Code refactoring
- `perf` - Performance improvement
- `test` - Tests added/modified
- `chore` - Build, dependencies, etc.

**Example:**
```
feat(members): add reactivate endpoint

- Added POST /members/:id/reactivate endpoint
- Clears deleted_at timestamp to restore member
- Returns 200 OK with restored member data

Fixes #42
```

---

## Testing

### Backend Tests

```bash
cd blue-ledger-api

# Run all tests
go test ./...

# Run specific package tests
go test ./internal/members

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend Tests

```bash
cd web

# Run tests (Vitest)
npm test

# Run with coverage
npm test -- --coverage

# Watch mode
npm test -- --watch
```

### Integration Tests

```bash
cd blue-ledger-api

# Run integration tests with database
go test -tags=integration ./...
```

### Manual Testing

**Test Login:**
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@tausigmasigma.org",
    "password": "BlueLedger2026!"
  }'
```

**Test Create Member:**
```bash
TOKEN="<access_token_from_login>"

curl -X POST http://localhost:8080/v1/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f"
  }'
```

---

## Debugging

### Backend Debugging

#### Using Delve (Go Debugger)

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger
dlv debug ./cmd/server

# Set breakpoint
(dlv) break main.main

# Continue
(dlv) continue

# Print variable
(dlv) print variableName
```

#### Using Logging

```go
import "log"

// Debug log
log.Printf("Debug: member = %+v", member)

// Error log
log.Printf("Error: %v", err)
```

Set `LOG_LEVEL=debug` in `.env` for verbose logging.

### Frontend Debugging

**Browser DevTools:**
1. Open browser → F12 (or Cmd+Option+I on macOS)
2. Use **Console** tab to log messages
3. Use **Debugger** tab to set breakpoints
4. Use **Network** tab to inspect API calls

**VS Code Debugging:**
```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "chrome",
      "request": "launch",
      "name": "Launch Chrome",
      "url": "http://localhost:5173",
      "webRoot": "${workspaceFolder}/web/src",
      "sourceMapPathOverride": {
        "/src/*": "${webspaceFolder}/src/*"
      }
    }
  ]
}
```

### Database Debugging

```bash
# Connect to database
psql -U postgres -d blue_ledger

# List tables
\dt

# View table structure
\d members

# Run query
SELECT * FROM members LIMIT 5;

# Explain query performance
EXPLAIN ANALYZE SELECT * FROM members WHERE id = '...';
```

---

## Common Issues

### Issue: Port Already in Use

```bash
# Find process on port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in .env
API_PORT=8081
```

### Issue: Database Connection Error

```bash
# Check PostgreSQL is running
pg_isready

# Verify DATABASE_URL
echo $DATABASE_URL

# Test connection
psql $DATABASE_URL -c "SELECT 1"

# Check credentials
psql -U postgres -h localhost
```

### Issue: Module Not Found

```bash
# Backend
go mod download
go mod tidy

# Frontend
npm install
npm ci  # Clean install
```

### Issue: Docker Container Won't Start

```bash
# Check logs
docker-compose logs blue-ledger-api

# Rebuild image
docker-compose down
docker-compose up --build

# Check disk space
df -h
```

### Issue: Hot Reload Not Working

**Frontend:**
```bash
# Restart dev server
npm run dev

# Clear Vite cache
rm -rf node_modules/.vite
```

**Backend:**
```bash
# Backend doesn't have hot reload
# Stop and restart manually
go run ./cmd/server
```

---

## IDE Setup

### VS Code Configuration

**Install Extensions:**
1. Go (golang.go)
2. TypeScript Vue Plugin
3. Volar
4. REST Client
5. Docker
6. PostgreSQL

**Create .vscode/settings.json:**
```json
{
  "go.lintOnSave": "package",
  "go.lintTool": "golangci-lint",
  "go.lintArgs": ["--fast"],
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "search.exclude": {
    "**/node_modules": true,
    "**/dist": true,
    "**/.next": true
  }
}
```

### Goland/IntelliJ Setup

1. Open project → Configure SDK
2. Set Go SDK to: `/usr/local/go` or `$(brew --prefix go)/bin/go`
3. Enable Go modules: Settings → Go → Go Modules
4. Configure run configuration for main.go

---

## Performance Tips

### Backend

```go
// Use database connection pooling
// Use caching for frequently accessed data
// Use indexes on frequently queried columns
// Use prepared statements for queries
```

### Frontend

```javascript
// Lazy load components
// Memoize expensive computations
// Use virtualization for long lists
// Optimize images and assets
```

### Database

```sql
-- Add indexes
CREATE INDEX idx_members_chapter_id ON members(chapter_id);
CREATE INDEX idx_engagement_member_id ON engagement_log(member_id);

-- View query plans
EXPLAIN ANALYZE SELECT ...
```

---

## Additional Resources

### Documentation
- **API Docs**: [API_DOCUMENTATION.md](API_DOCUMENTATION.md)
- **Deployment**: [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)
- **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)
- **User Guide**: [USER_GUIDE.md](USER_GUIDE.md)

### Official Documentation
- **Go**: https://golang.org/doc/
- **Echo**: https://echo.labstack.com/
- **React**: https://react.dev
- **PostgreSQL**: https://www.postgresql.org/docs/
- **Docker**: https://docs.docker.com/

### Helpful Tools
- **Postman**: https://www.postman.com/
- **DBeaver**: https://dbeaver.io/ (Database GUI)
- **TablePlus**: https://tableplus.com/ (Database GUI)
- **Insomnia**: https://insomnia.rest/ (API testing)

---

## Getting Help

- **GitHub Issues**: https://github.com/Ces1231/blue-ledger/issues
- **GitHub Discussions**: https://github.com/Ces1231/blue-ledger/discussions
- **Email**: admin@tausigmasigma.org

---

## Next Steps

1. ✅ Complete backend setup
2. ✅ Complete frontend setup
3. ✅ Complete database setup
4. ✅ Run development servers
5. ✅ Test login and create member
6. ✅ Read API documentation
7. ✅ Make your first feature!

Happy coding! 🚀
