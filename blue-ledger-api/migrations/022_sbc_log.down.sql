-- ============================================================
-- Migration: 022_sbc_log (DOWN)
-- Date: 2026-03-26
-- Purpose: Drop sbc_log table.
-- ============================================================

REVOKE ALL ON sbc_log FROM blue_ledger_app;
REVOKE ALL ON sbc_log FROM blue_ledger_admin;
DROP TABLE IF EXISTS sbc_log;
