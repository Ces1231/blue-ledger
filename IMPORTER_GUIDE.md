# CSV Member Importer - Production Deployment Guide

## Overview
The Blue Ledger API includes an embedded CSV importer that can populate the database with chapter and member data on fresh deployments.

##  Files
- **Embedded Importer**: `cmd/import-csv-embedded/main.go` - Standalone Go binary with hardcoded member roster
- **File-Based Importer**: `cmd/import-csv/main.go` - Reads from CSV file (for future custom rosters)

## Quick Start - Local Testing

### Prerequisites
- Go 1.25+
- PostgreSQL 16 (running locally or in Docker)
- DATABASE_URL environment variable set

### Run Importer

```bash
cd blue-ledger-api

# Option 1: Using Go directly
go run ./cmd/import-csv-embedded

# Option 2: Build and run binary
go build -o import-csv-embedded ./cmd/import-csv-embedded
./import-csv-embedded
```

### Expected Output
```
Loaded 35 members from embedded CSV
Chapter ID: <uuid>
Admin user created: admin@tausigmasigma.org

✅ Imported 35 members
⏭️  Skipped 0 incomplete records

🔑 Credentials for testing:
   Email: admin@tausigmasigma.org
   Password: BlueLedger2026!
```

## Production Deployment

### Method 1: Manual Pre-Deploy (Recommended)

1. Deploy API migrations to production database
2. Run importer BEFORE bringing up web frontend:
   ```bash
   DATABASE_URL="postgres://user:pass@host/blue_ledger" \
   ./import-csv-embedded
   ```
3. Deploy web frontend

### Method 2: Docker One-Off Task

```bash
docker build -t blue-ledger-importer:latest -f Dockerfile.importer .
docker run \
  -e DATABASE_URL="postgres://user:pass@host/blue_ledger" \
  blue-ledger-importer:latest
```

### Method 3: Fly.io Machine Run

```bash
flyctl machines run -e DATABASE_URL="$DATABASE_URL" \
  registry.fly.io/blue-ledger-api:latest \
  /app/import-csv-embedded
```

## Data Imported

### Chapter
- **Name**: Tau Sigma Sigma
- **Greek Letters**: ΤΣΣ
- **Location**: Atlanta, GA
- **University**: Georgia Tech

### Users & Members
- **Admin User**: admin@tausigmasigma.org (password: BlueLedger2026!)
- **35 Chapter Members**: Full roster with names, emails, contact info
- **Display IDs**: ΤΣΣ-001 through ΤΣΣ-035

##  Customization

To use different member data:

### Option A: Modify Embedded CSV (rebuild required)
1. Edit `cmd/import-csv-embedded/main.go`
2. Update the `rosterData` constant with new CSV
3. Rebuild: `go build -o import-csv-embedded ./cmd/import-csv-embedded`

### Option B: Use File-Based Importer (no rebuild)
1. Prepare CSV file with same format
2. Run: `go run ./cmd/import-csv -- file="members.csv"`

### CSV Format
```
PBS#,First Name,Last Name,Preferred Name,Email,Phone,T-Shirt,Polo,Blazer,Birthday,Sigmaversary
1001,George,Smith,George,george.smith@email.com,404-555-0101,L,L,40R,January 15,2020
```

## Verification

After import, verify with:

```bash
# Check chapter exists
curl -X GET https://blue-ledger-api.fly.dev/v1/chapters \
  -H "Authorization: Bearer $TOKEN"

# List members (requires chapter context)
curl -X GET https://blue-ledger-api.fly.dev/v1/members \
  -H "Authorization: Bearer $TOKEN"

# Login test
curl -X POST https://blue-ledger-api.fly.dev/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@tausigmasigma.org","password":"BlueLedger2026!"}'
```

## Idempotency

The importer is safe to run multiple times:
- Checks if database is empty (row count > 0 in chapters table)
- Skips import if data already exists
- Uses ON CONFLICT for safe upserts

## Troubleshooting

### Import fails with "relation chapters does not exist"
- Migrations haven't run yet
- Solution: Run API migrations first: `go run ./cmd/migrate`

### "password authentication failed"
- DATABASE_URL credentials are incorrect
- Verify connection string format: `postgres://user:pass@host:port/db?sslmode=disable`

### Members not showing in API
- Check RLS (Row-Level Security) policies - admin user may need chapter association
- Members list endpoint requires chapter context in JWT

## Next Steps

1. ✅ Deploy API to production with embedded importer
2. ✅ Run importer on fresh database
3. ✅ Test admin login credentials
4. ✅ Verify member roster in directory
5. ✅ Enable frontend access

