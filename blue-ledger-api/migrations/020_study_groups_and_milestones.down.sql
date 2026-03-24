-- ============================================================
-- Migration: 020_study_groups_and_milestones (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop milestones and study_groups tables.
-- WARNING: All study group and milestone data will be lost.
-- ============================================================

REVOKE ALL ON milestones FROM blue_ledger_app;
REVOKE ALL ON milestones FROM blue_ledger_admin;
DROP TABLE IF EXISTS milestones;

REVOKE ALL ON study_groups FROM blue_ledger_app;
REVOKE ALL ON study_groups FROM blue_ledger_admin;
DROP TABLE IF EXISTS study_groups;
