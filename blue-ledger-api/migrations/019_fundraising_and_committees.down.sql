-- ============================================================
-- Migration: 019_fundraising_and_committees (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop committees and fundraising_campaigns tables.
-- WARNING: All fundraising and committee data will be lost.
-- ============================================================

REVOKE ALL ON committees FROM blue_ledger_app;
REVOKE ALL ON committees FROM blue_ledger_admin;
DROP TABLE IF EXISTS committees;

REVOKE ALL ON fundraising_campaigns FROM blue_ledger_app;
REVOKE ALL ON fundraising_campaigns FROM blue_ledger_admin;
DROP TABLE IF EXISTS fundraising_campaigns;
