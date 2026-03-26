-- ============================================================
-- Migration: 003_members (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop the members table and all associated objects.
-- WARNING: All member profile data will be permanently lost.
-- ============================================================

REVOKE ALL ON members FROM blue_ledger_app;
REVOKE ALL ON members FROM blue_ledger_admin;

DROP TABLE IF EXISTS members;
