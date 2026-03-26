-- ============================================================
-- Migration: 007_service_log (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop the service_log table.
-- WARNING: All service hour records will be permanently lost.
-- ============================================================

REVOKE ALL ON service_log FROM blue_ledger_app;
REVOKE ALL ON service_log FROM blue_ledger_admin;

DROP TABLE IF EXISTS service_log;
