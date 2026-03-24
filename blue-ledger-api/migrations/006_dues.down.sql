-- ============================================================
-- Migration: 006_dues (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop the dues_records table.
-- WARNING: All dues history will be permanently lost.
-- ============================================================

REVOKE ALL ON dues_records FROM blue_ledger_app;
REVOKE ALL ON dues_records FROM blue_ledger_admin;

DROP TABLE IF EXISTS dues_records;
