-- ============================================================
-- Migration: 012_votes (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop vote_responses and votes tables in FK order.
-- WARNING: All voting history will be permanently lost.
-- ============================================================

REVOKE ALL ON vote_responses FROM blue_ledger_app;
REVOKE ALL ON vote_responses FROM blue_ledger_admin;
DROP TABLE IF EXISTS vote_responses;

REVOKE ALL ON votes FROM blue_ledger_app;
REVOKE ALL ON votes FROM blue_ledger_admin;
DROP TABLE IF EXISTS votes;
