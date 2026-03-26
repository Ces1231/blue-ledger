-- ============================================================
-- Migration: 014_store_and_goals (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop chapter_goals, store_orders, store_items in
--          reverse FK order.
-- WARNING: All store and goal data will be permanently lost.
-- ============================================================

REVOKE ALL ON chapter_goals FROM blue_ledger_app;
REVOKE ALL ON chapter_goals FROM blue_ledger_admin;
DROP TABLE IF EXISTS chapter_goals;

REVOKE ALL ON store_orders FROM blue_ledger_app;
REVOKE ALL ON store_orders FROM blue_ledger_admin;
DROP TABLE IF EXISTS store_orders;

REVOKE ALL ON store_items FROM blue_ledger_app;
REVOKE ALL ON store_items FROM blue_ledger_admin;
DROP TABLE IF EXISTS store_items;
