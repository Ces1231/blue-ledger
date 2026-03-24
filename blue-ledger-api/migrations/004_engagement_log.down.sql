-- ============================================================
-- Migration: 004_engagement_log (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop the engagement_log table.
-- WARNING: All XP history will be permanently lost.
-- ============================================================

REVOKE ALL ON engagement_log FROM blue_ledger_app;
REVOKE ALL ON engagement_log FROM blue_ledger_admin;

DROP TABLE IF EXISTS engagement_log;
