-- ============================================================
-- Migration: 015_job_board_and_resources (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop resources and job_board tables.
-- WARNING: All job postings and resource library data lost.
--          Cloudflare R2 objects referenced by file_key are
--          NOT automatically deleted — clean those separately.
-- ============================================================

REVOKE ALL ON resources FROM blue_ledger_app;
REVOKE ALL ON resources FROM blue_ledger_admin;
DROP TABLE IF EXISTS resources;

REVOKE ALL ON job_board FROM blue_ledger_app;
REVOKE ALL ON job_board FROM blue_ledger_admin;
DROP TABLE IF EXISTS job_board;
