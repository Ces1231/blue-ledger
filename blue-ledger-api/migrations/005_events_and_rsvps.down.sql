-- ============================================================
-- Migration: 005_events_and_rsvps (DOWN)
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Drop rsvps and events tables in reverse FK order.
-- WARNING: All event and RSVP data will be permanently lost.
-- ============================================================

REVOKE ALL ON rsvps FROM blue_ledger_app;
REVOKE ALL ON rsvps FROM blue_ledger_admin;
DROP TABLE IF EXISTS rsvps;

REVOKE ALL ON events FROM blue_ledger_app;
REVOKE ALL ON events FROM blue_ledger_admin;
DROP TABLE IF EXISTS events;
