-- ============================================================
-- Migration: 002_rls_policies (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop all RLS policies created in 002 and revoke
--          role permissions. Disable RLS on affected tables.
--          Drop the two application database roles.
-- Safety: SAFE
-- Rollback: N/A (this IS the rollback)
-- ============================================================

-- Revoke grants from application roles
REVOKE ALL ON auth_sessions      FROM blue_ledger_app;
REVOKE ALL ON magic_link_tokens  FROM blue_ledger_app;
REVOKE ALL ON users              FROM blue_ledger_app;
REVOKE ALL ON chapters           FROM blue_ledger_app;

REVOKE ALL ON chapters          FROM blue_ledger_admin;
REVOKE ALL ON users             FROM blue_ledger_admin;
REVOKE ALL ON auth_sessions     FROM blue_ledger_admin;
REVOKE ALL ON magic_link_tokens FROM blue_ledger_admin;

REVOKE USAGE ON ALL SEQUENCES IN SCHEMA public FROM blue_ledger_app;
REVOKE USAGE ON ALL SEQUENCES IN SCHEMA public FROM blue_ledger_admin;

-- Drop RLS policies
DROP POLICY IF EXISTS user_isolation ON auth_sessions;
DROP POLICY IF EXISTS user_isolation ON magic_link_tokens;

-- Disable RLS on tables touched in this migration
ALTER TABLE auth_sessions     DISABLE ROW LEVEL SECURITY;
ALTER TABLE magic_link_tokens DISABLE ROW LEVEL SECURITY;

-- Drop roles last (after all grants are revoked)
DROP ROLE IF EXISTS blue_ledger_app;
DROP ROLE IF EXISTS blue_ledger_admin;
