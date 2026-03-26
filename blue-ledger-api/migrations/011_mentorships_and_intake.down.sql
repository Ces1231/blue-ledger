-- ============================================================
-- Migration: 011_mentorships_and_intake (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop intake_prospects and mentorships tables.
-- WARNING: All intake pipeline and mentorship data will be lost.
-- ============================================================

REVOKE ALL ON intake_prospects FROM blue_ledger_app;
REVOKE ALL ON intake_prospects FROM blue_ledger_admin;
DROP TABLE IF EXISTS intake_prospects;

REVOKE ALL ON mentorships FROM blue_ledger_app;
REVOKE ALL ON mentorships FROM blue_ledger_admin;
DROP TABLE IF EXISTS mentorships;
