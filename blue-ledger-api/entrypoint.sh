#!/bin/sh
set -e

# Write JWT keys from env vars to files at startup (production / Fly.io).
# JWT_PRIVATE_KEY and JWT_PUBLIC_KEY are set as Fly.io secrets.
# Falls back to existing files in dev (docker-compose mounts keys/).
if [ -n "$JWT_PRIVATE_KEY" ]; then
    mkdir -p /app/keys
    printf '%s' "$JWT_PRIVATE_KEY" > /app/keys/private.pem
    printf '%s' "$JWT_PUBLIC_KEY" > /app/keys/public.pem
fi

# Run importer if DATABASE_URL is set and chapters table is empty
# This works for both dev and production - idempotent operation
if [ -n "$DATABASE_URL" ]; then
    echo "🔄 Checking if seed data exists..."
    # Try to import - the importer is idempotent, only imports if chapters table is empty
    if ./import-csv-embedded; then
        echo "✅ Seed data imported successfully"
    else
        echo "⚠️  Seed data import failed or already exists"
    fi
fi

# Start the main API server
exec ./blue-ledger-api
