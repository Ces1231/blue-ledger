-- ============================================================
-- Migration: 001_users_and_chapters (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Fully reverse migration 001. Drops all tables created
--          in the up migration in reverse dependency order, then
--          drops the shared trigger function.
-- Safety: SAFE (no production data exists at this stage)
-- Rollback: N/A (this IS the rollback)
-- WARNING: Dropping these tables removes all user, session, and
--          chapter data. Only run this in development.
-- ============================================================

DROP TABLE IF EXISTS magic_link_tokens;
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS chapters;

-- Drop the shared trigger function last — other migrations depend on it
-- Only safe to drop here because all tables referencing it are gone.
DROP FUNCTION IF EXISTS update_updated_at();

DROP EXTENSION IF EXISTS pgcrypto;
