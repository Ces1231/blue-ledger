-- ============================================================
-- Migration: 022_sbc_log
-- Date: 2026-03-26
-- Author: Nebula
-- Purpose: Create sbc_log — the Service / Brotherhood / Conduct
--          log table. Each row records a categorical note about
--          a member, entered by an admin or chair.
--
--          category CHECK constrains entries to one of three
--          predefined categories: service, brotherhood, conduct.
--          recorded_by is the member who authored the entry
--          (separate from member_id, the subject of the entry).
--
--          This table has no updated_at column — entries are
--          replaced (delete + re-create) rather than edited.
--          UPDATE in the service only patches category/note
--          in-place to avoid orphaned rows.
--
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new table)
-- Est. duration: <1s
-- ============================================================

CREATE TABLE IF NOT EXISTS sbc_log (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id   UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id    UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  category     TEXT        NOT NULL
               CHECK (category IN ('service','brotherhood','conduct')),
  note         TEXT        NOT NULL,
  recorded_by  UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- No updated_at — entries are append-only or patched in-place
);

-- ── Row-Level Security ──────────────────────────────────────
ALTER TABLE sbc_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY chapter_isolation ON sbc_log
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

-- ── Indexes ──────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_sbc_log_chapter
  ON sbc_log(chapter_id);

CREATE INDEX IF NOT EXISTS idx_sbc_log_member
  ON sbc_log(chapter_id, member_id);

CREATE INDEX IF NOT EXISTS idx_sbc_log_category
  ON sbc_log(chapter_id, category);

CREATE INDEX IF NOT EXISTS idx_sbc_log_created
  ON sbc_log(chapter_id, created_at DESC);

-- ── Grants ───────────────────────────────────────────────────
GRANT SELECT, INSERT, UPDATE, DELETE ON sbc_log TO blue_ledger_app;
GRANT ALL PRIVILEGES ON sbc_log TO blue_ledger_admin;
