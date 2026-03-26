-- ============================================================
-- Migration: 006_dues
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create dues_records table. Tracks per-member, per-
--          semester dues obligations. Includes Zeffy integration
--          columns (zeffy_form_id, zeffy_transaction_id) for the
--          0%-fee dues collection flow via webhook/Zapier.
--          UNIQUE(chapter_id, member_id, semester) prevents
--          duplicate dues records per billing period.
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new table)
-- Est. duration: <1s
-- ============================================================

CREATE TABLE IF NOT EXISTS dues_records (
  id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id            UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id             UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,

  semester              TEXT        NOT NULL,                    -- "Fall 2025"
  amount_cents          INT         NOT NULL,                    -- stored in cents (no floats)
  due_date              DATE        NOT NULL,
  paid_at               TIMESTAMPTZ,

  -- Payment method: zeffy is the primary dues collection method (0% fees)
  -- Other methods remain for cash/admin overrides
  payment_method        TEXT
                        CHECK (payment_method IN (
                          'zeffy','cash','zelle','venmo','check','waived','admin_override'
                        )),

  -- Zeffy integration (populated via Zapier webhook POST /webhooks/zeffy)
  zeffy_form_id         TEXT,                                    -- Zeffy form ID for this chapter
  zeffy_transaction_id  TEXT,                                    -- Zeffy transaction reference

  status                TEXT        NOT NULL DEFAULT 'unpaid'
                        CHECK (status IN ('unpaid','paid','late','waived','outstanding')),

  xp_awarded            BOOLEAN     NOT NULL DEFAULT FALSE,      -- true after XP has been granted

  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(chapter_id, member_id, semester)
);

-- ── Row-Level Security ──────────────────────────────────────
ALTER TABLE dues_records ENABLE ROW LEVEL SECURITY;

CREATE POLICY chapter_isolation ON dues_records
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

-- ── Indexes ──────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_dues_chapter
  ON dues_records(chapter_id);

CREATE INDEX IF NOT EXISTS idx_dues_member
  ON dues_records(chapter_id, member_id);

CREATE INDEX IF NOT EXISTS idx_dues_semester
  ON dues_records(chapter_id, semester);

CREATE INDEX IF NOT EXISTS idx_dues_status
  ON dues_records(chapter_id, status);

CREATE INDEX IF NOT EXISTS idx_dues_zeffy_transaction
  ON dues_records(zeffy_transaction_id)
  WHERE zeffy_transaction_id IS NOT NULL;

-- ── updated_at trigger ───────────────────────────────────────
CREATE TRIGGER trg_dues_records_updated_at
  BEFORE UPDATE ON dues_records
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ── Grants ───────────────────────────────────────────────────
GRANT SELECT, INSERT, UPDATE, DELETE ON dues_records TO blue_ledger_app;
GRANT ALL PRIVILEGES ON dues_records TO blue_ledger_admin;
