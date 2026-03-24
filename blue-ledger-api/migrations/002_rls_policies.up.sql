-- ============================================================
-- Migration: 002_rls_policies
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create the two database roles used by the API and
--          sysadmin, then apply RLS policies to all tenant tables
--          that exist as of this migration (chapters and users
--          are non-tenant — no RLS policy needed).
--
--          blue_ledger_app  — used by the Go API service account.
--                             Subject to RLS. Sets app.chapter_id
--                             session variable on every request.
--          blue_ledger_admin — used by sysadmin console and
--                              background migration jobs.
--                              Granted BYPASSRLS — sees all rows.
--
-- NOTE: Tables for members, events, etc. are created in later
--       migrations and have their own RLS ENABLE + POLICY
--       statements inline. This migration establishes the roles
--       and applies RLS to the chapter/user-adjacent tables that
--       were bootstrapped in 001.
--
-- Safety: SAFE
-- Rollback: EASY
-- Locking: NONE (DDL on roles; no data modification)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- DB Role: blue_ledger_app
-- Used by the Go API. Connects with a limited-privilege account.
-- RLS policies apply to all queries made under this role.
-- ============================================================
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'blue_ledger_app') THEN
    CREATE ROLE blue_ledger_app NOLOGIN NOINHERIT;
  END IF;
END $$;

-- ============================================================
-- DB Role: blue_ledger_admin
-- Used by sysadmin console and migration runner.
-- BYPASSRLS means RLS policies are ignored — full platform view.
-- ============================================================
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'blue_ledger_admin') THEN
    CREATE ROLE blue_ledger_admin NOLOGIN NOINHERIT BYPASSRLS;
  END IF;
END $$;

-- ============================================================
-- Grant table-level permissions to blue_ledger_app
-- auth_sessions and magic_link_tokens: full DML (login flow)
-- chapters: SELECT + UPDATE only (app can read/update chapters
--           but cannot create/delete via the app role)
-- users: full DML (registration, profile update, login)
-- ============================================================
GRANT SELECT, INSERT, UPDATE, DELETE ON auth_sessions      TO blue_ledger_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON magic_link_tokens  TO blue_ledger_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON users              TO blue_ledger_app;
GRANT SELECT, UPDATE                 ON chapters           TO blue_ledger_app;

-- blue_ledger_admin gets unrestricted access to all current tables
GRANT ALL PRIVILEGES ON chapters          TO blue_ledger_admin;
GRANT ALL PRIVILEGES ON users             TO blue_ledger_admin;
GRANT ALL PRIVILEGES ON auth_sessions     TO blue_ledger_admin;
GRANT ALL PRIVILEGES ON magic_link_tokens TO blue_ledger_admin;

-- Allow use of sequences (for future serial PKs if any are added)
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO blue_ledger_app;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO blue_ledger_admin;

-- ============================================================
-- RLS Policy: auth_sessions
-- Users can only see and manage their own sessions.
-- The API always sets app.current_user_id as a session variable
-- for user-scoped operations.
-- Note: auth_sessions is NOT chapter-scoped — it uses user_id.
-- ============================================================
ALTER TABLE auth_sessions ENABLE ROW LEVEL SECURITY;

-- App role: each user sees only their own sessions
CREATE POLICY user_isolation ON auth_sessions
  USING (user_id = current_setting('app.current_user_id', true)::uuid);

-- ============================================================
-- RLS Policy: magic_link_tokens
-- User-scoped (not chapter-scoped).
-- ============================================================
ALTER TABLE magic_link_tokens ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_isolation ON magic_link_tokens
  USING (user_id = current_setting('app.current_user_id', true)::uuid);

-- ============================================================
-- Note on chapter-scoped tables (members, events, etc.):
-- Those tables are created in migrations 003+ and carry their
-- own ENABLE ROW LEVEL SECURITY + chapter_isolation POLICY
-- inline with CREATE TABLE. This migration only handles the
-- role infrastructure and the non-tenant tables from 001.
-- ============================================================
