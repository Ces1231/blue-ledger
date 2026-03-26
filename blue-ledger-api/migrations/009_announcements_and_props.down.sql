-- ============================================================
-- Migration: 009_announcements_and_props (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop props and announcements tables.
-- WARNING: All announcements and props history will be lost.
-- ============================================================

REVOKE ALL ON props FROM blue_ledger_app;
REVOKE ALL ON props FROM blue_ledger_admin;
DROP TABLE IF EXISTS props;

REVOKE ALL ON announcements FROM blue_ledger_app;
REVOKE ALL ON announcements FROM blue_ledger_admin;
DROP TABLE IF EXISTS announcements;
