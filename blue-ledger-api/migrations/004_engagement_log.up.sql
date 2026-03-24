-- ============================================================
-- Migration: 004_engagement_log
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create engagement_log — the immutable, append-only
--          XP audit trail. Every XP award produces one row.
--          member.xp_total is a denormalized cache; this table
--          is the source of truth for all XP calculations.
--          There is intentionally NO updated_at column.
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new table)
-- Est. duration: <1s
-- ============================================================

CREATE TABLE IF NOT EXISTS engagement_log (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id       UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,

  -- What happened
  activity        TEXT        NOT NULL,                         -- "Chapter Meeting (on time)"
  xp_awarded      INT         NOT NULL,
  source          TEXT        NOT NULL
                  CHECK (source IN (
                    'checkin','rsvp','admin','system','quiz',
                    'service','dues','badge','props','mentorship'
                  )),

  -- Optional back-reference to the entity that triggered this award
  reference_id    UUID,                                         -- event_id, service_log_id, etc.
  reference_type  TEXT,                                         -- 'event', 'service_log', 'dues_record', etc.

  -- Who awarded it (NULL for system-generated)
  awarded_by      UUID        REFERENCES members(id) ON DELETE SET NULL,

  note            TEXT,
  semester        TEXT,                                         -- "Fall 2025" for semester XP grouping

  -- created_at only — this table is append-only, no updates
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Row-Level Security ──────────────────────────────────────
ALTER TABLE engagement_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY chapter_isolation ON engagement_log
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

-- ── Indexes ──────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_eng_log_chapter
  ON engagement_log(chapter_id);

CREATE INDEX IF NOT EXISTS idx_eng_log_member
  ON engagement_log(chapter_id, member_id);

CREATE INDEX IF NOT EXISTS idx_eng_log_created
  ON engagement_log(chapter_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_eng_log_source
  ON engagement_log(chapter_id, source);

CREATE INDEX IF NOT EXISTS idx_eng_log_reference
  ON engagement_log(reference_id, reference_type)
  WHERE reference_id IS NOT NULL;

-- ── Grants ───────────────────────────────────────────────────
-- INSERT only for app role — no updates or deletes (immutable log)
GRANT SELECT, INSERT ON engagement_log TO blue_ledger_app;
GRANT ALL PRIVILEGES ON engagement_log TO blue_ledger_admin;
