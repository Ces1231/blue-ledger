#!/bin/bash
set -e

echo "Building importer..."
cd /tmp
git clone https://github.com/yourusername/blue-ledger.git || cd blue-ledger && git pull
cd blue-ledger/blue-ledger-api

# Download CSV from user's shared storage or inline it
# For now, using a hardcoded base64 or downloading from a temp URL
# This is a workaround since we can't easily send large files to Fly

go build -o import-csv ./cmd/import-csv

# Assuming CSV is available (you'd need to provide it somehow)
# For this demo, let's assume it's baked into the release or accessed via HTTP

export DATABASE_URL="$DATABASE_URL"
./import-csv --file="/tmp/roster.csv"

echo "Import complete!"
