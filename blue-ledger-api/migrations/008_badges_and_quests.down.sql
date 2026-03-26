-- ============================================================
-- Migration: 008_badges_and_quests (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop all badge and quest tables in reverse FK order.
-- WARNING: All badge definitions, awards, and quest progress
--          will be permanently lost.
-- ============================================================

REVOKE ALL ON member_quest_progress FROM blue_ledger_app;
REVOKE ALL ON member_quest_progress FROM blue_ledger_admin;
DROP TABLE IF EXISTS member_quest_progress;

REVOKE ALL ON quests FROM blue_ledger_app;
REVOKE ALL ON quests FROM blue_ledger_admin;
DROP TABLE IF EXISTS quests;

REVOKE ALL ON member_badges FROM blue_ledger_app;
REVOKE ALL ON member_badges FROM blue_ledger_admin;
DROP TABLE IF EXISTS member_badges;

REVOKE ALL ON badges FROM blue_ledger_app;
REVOKE ALL ON badges FROM blue_ledger_admin;
DROP TABLE IF EXISTS badges;
