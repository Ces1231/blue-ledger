#!/bin/bash
# ─────────────────────────────────────────────────────
# The Blue Ledger — GitHub Setup Script
# Run this once from inside the blue-ledger/ folder
# ─────────────────────────────────────────────────────

set -e

echo ""
echo "🔷 The Blue Ledger — GitHub Sync Setup"
echo "═══════════════════════════════════════"
echo ""

# Check if git is initialized
if [ ! -d ".git" ]; then
  git init
  git branch -M main
  echo "✓ Git initialized"
else
  echo "✓ Git already initialized"
fi

# Stage all files
git add .

# Commit if there are changes
if git diff --cached --quiet; then
  echo "✓ Nothing new to commit"
else
  git commit -m "Blue Ledger v1.0.0 — 47 screens, full documentation"
  echo "✓ Changes committed"
fi

echo ""
echo "═══════════════════════════════════════"
echo "Next: create the repo on GitHub"
echo ""
echo "  1. Go to github.com/new"
echo "  2. Name it: blue-ledger"
echo "  3. Set to Private (recommended)"
echo "  4. Do NOT initialize with README (you already have one)"
echo "  5. Click 'Create repository'"
echo ""
echo "Then run:"
echo ""
echo "  git remote add origin https://github.com/YOUR_USERNAME/blue-ledger.git"
echo "  git push -u origin main"
echo ""
echo "Replace YOUR_USERNAME with your GitHub username."
echo "═══════════════════════════════════════"
echo ""
