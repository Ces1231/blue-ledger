-- ============================================================
-- Migration: 007_service_log
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create service_log — records community service hours
--          submitted by members and verified by chairs/admins.
--          Verification triggers XP award via the XP engine.
--          hours is NUMERIC(5,2) with CHECK hours > 0.
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new table)
-- Est. duration: <1s
-- ============================================================

CREATE TABLE IF NOT EXISTS service_log (
  id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID          NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id       UUID          NOT NULL REFERENCES members(id)  ON DELETE CASCADE,

  event_name      TEXT          NOT NULL,
  organization    TEXT,
  service_date    DATE          NOT NULL,
  hours           NUMERIC(5,2)  NOT NULL CHECK (hours > 0),
  xp_awarded      INT           NOT NULL DEFAULT 0,

  -- Verification workflow
  verified        BOOLEAN       NOT NULL DEFAULT FALSE,
  verified_by     UUID          REFERENCES members(id) ON DELETE SET NULL,
  verified_at     TIMESTAMPTZ,

  notes           TEXT,

  created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- ── Row-Level Security ──────────────────────────────────────
ALTER TABLE service_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY chapter_isolation ON service_log
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

-- ── Indexes ──────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_service_log_chapter
  ON service_log(chapter_id);

CREATE INDEX IF NOT EXISTS idx_service_log_member
  ON service_log(chapter_id, member_id);

CREATE INDEX IF NOT EXISTS idx_service_log_unverified
  ON service_log(chapter_id, verified, service_date DESC)
  WHERE verified = FALSE;

-- ── updated_at trigger ───────────────────────────────────────
CREATE TRIGGER trg_service_log_updated_at
  BEFORE UPDATE ON service_log
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ── Grants ───────────────────────────────────────────────────
GRANT SELECT, INSERT, UPDATE, DELETE ON service_log TO blue_ledger_app;
GRANT ALL PRIVILEGES ON service_log TO blue_ledger_admin;
