-- ============================================================
-- Migration: 013_minutes_and_scholarships (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop scholarships and chapter_minutes tables.
-- WARNING: All minutes and scholarship application data lost.
-- ============================================================

REVOKE ALL ON scholarships FROM blue_ledger_app;
REVOKE ALL ON scholarships FROM blue_ledger_admin;
DROP TABLE IF EXISTS scholarships;

REVOKE ALL ON chapter_minutes FROM blue_ledger_app;
REVOKE ALL ON chapter_minutes FROM blue_ledger_admin;
DROP TABLE IF EXISTS chapter_minutes;
