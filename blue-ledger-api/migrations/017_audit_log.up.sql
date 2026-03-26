-- ============================================================
-- Migration: 017_audit_log
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create the platform-wide audit_log table. This table
--          is NOT tenant-scoped and NOT protected by RLS.
--          It is readable only by the sysadmin role (blue_ledger_admin).
--          Records all material actions: member CRUD, dues events,
--          billing events, role changes, danger zone operations,
--          and webhook receipts.
--
--          actor_type distinguishes human users from system jobs
--          and webhook-triggered events.
--          metadata is JSONB for flexible per-action context.
--
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new table)
-- Est. duration: <1s
-- ============================================================

CREATE TABLE IF NOT EXISTS audit_log (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id    UUID,                                              -- user_id or NULL for system
  actor_type  TEXT        NOT NULL
              CHECK (actor_type IN ('user','system','webhook')),
  chapter_id  UUID,                                              -- NULL for platform-level actions
  action      TEXT        NOT NULL,                             -- 'member.created', 'dues.paid', etc.
  target_type TEXT,                                              -- 'member', 'event', 'chapter', etc.
  target_id   UUID,
  metadata    JSONB       NOT NULL DEFAULT '{}',
  ip_address  INET,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- No updated_at — audit log is append-only
);

-- ── No RLS — intentionally platform-wide ─────────────────────
-- Sysadmin (blue_ledger_admin role with BYPASSRLS) is the only
-- consumer of full audit_log data.
-- The app role can INSERT only (write audit events during request
-- handling) but cannot SELECT (no user should read another's audit
-- trail via the app role).

-- ── Indexes ──────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_audit_log_actor
  ON audit_log(actor_id)
  WHERE actor_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_audit_log_chapter
  ON audit_log(chapter_id)
  WHERE chapter_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_audit_log_created
  ON audit_log(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_log_action
  ON audit_log(action, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_log_target
  ON audit_log(target_id, target_type)
  WHERE target_id IS NOT NULL;

-- ── Grants ───────────────────────────────────────────────────
-- App role: INSERT only (write audit events, cannot read back)
GRANT INSERT ON audit_log TO blue_ledger_app;
-- Admin role: full access (sysadmin console reads audit log)
GRANT ALL PRIVILEGES ON audit_log TO blue_ledger_admin;
