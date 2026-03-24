-- ============================================================
-- Migration: 010_notifications (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop the notifications table.
-- WARNING: All notification history will be permanently lost.
-- ============================================================

REVOKE ALL ON notifications FROM blue_ledger_app;
REVOKE ALL ON notifications FROM blue_ledger_admin;

DROP TABLE IF EXISTS notifications;
