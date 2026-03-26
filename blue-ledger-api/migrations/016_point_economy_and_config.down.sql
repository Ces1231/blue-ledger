-- ============================================================
-- Migration: 016_point_economy_and_config (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop chapter_config, point_economy, and the
--          seed_default_point_economy helper function.
-- WARNING: All XP configuration and chapter config data lost.
-- ============================================================

REVOKE ALL ON chapter_config FROM blue_ledger_app;
REVOKE ALL ON chapter_config FROM blue_ledger_admin;
DROP TABLE IF EXISTS chapter_config;

REVOKE ALL ON point_economy FROM blue_ledger_app;
REVOKE ALL ON point_economy FROM blue_ledger_admin;
DROP TABLE IF EXISTS point_economy;

DROP FUNCTION IF EXISTS seed_default_point_economy(UUID);
