-- ============================================================
-- Migration: 018_messaging (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop messages and message_threads tables in FK order.
-- WARNING: All messaging history will be permanently lost.
-- ============================================================

REVOKE ALL ON messages FROM blue_ledger_app;
REVOKE ALL ON messages FROM blue_ledger_admin;
DROP TABLE IF EXISTS messages;

REVOKE ALL ON message_threads FROM blue_ledger_app;
REVOKE ALL ON message_threads FROM blue_ledger_admin;
DROP TABLE IF EXISTS message_threads;
