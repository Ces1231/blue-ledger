-- ============================================================
-- Migration: 017_audit_log (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop the audit_log table.
-- WARNING: All platform audit history will be permanently lost.
--          This is an irreversible operation in production.
--          Only run this in development environments.
-- ============================================================

REVOKE ALL ON audit_log FROM blue_ledger_app;
REVOKE ALL ON audit_log FROM blue_ledger_admin;

DROP TABLE IF EXISTS audit_log;
