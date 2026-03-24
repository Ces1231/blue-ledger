-- ============================================================
-- Migration: 009_announcements_and_props
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create announcements (chapter-wide broadcast posts)
--          and props (peer-to-peer recognition with XP award).
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: announcements
-- ============================================================
CREATE TABLE IF NOT EXISTS announcements (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  title       TEXT        NOT NULL,
  body        TEXT        NOT NULL,
  category    TEXT        CHECK (category IN (
                'General','Meeting','Event','Dues','Urgent','Brotherhood','Other'
              )),
  is_pinned   BOOLEAN     NOT NULL DEFAULT FALSE,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  author_id   UUID        NOT NULL REFERENCES members(id) ON DELETE RESTRICT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at  TIMESTAMPTZ
);

ALTER TABLE announcements ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON announcements
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_announcements_chapter
  ON announcements(chapter_id);

CREATE INDEX IF NOT EXISTS idx_announcements_pinned
  ON announcements(chapter_id, is_pinned, created_at DESC)
  WHERE is_active = TRUE AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_announcements_active_recent
  ON announcements(chapter_id, created_at DESC)
  WHERE is_active = TRUE AND deleted_at IS NULL;

CREATE TRIGGER trg_announcements_updated_at
  BEFORE UPDATE ON announcements
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON announcements TO blue_ledger_app;
GRANT ALL PRIVILEGES ON announcements TO blue_ledger_admin;

-- ============================================================
-- Table: props
-- Peer recognition with a fixed XP award (default 10 XP).
-- append-only — no updated_at.
-- ============================================================
CREATE TABLE IF NOT EXISTS props (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  from_id     UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  to_id       UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  category    TEXT        NOT NULL
              CHECK (category IN (
                'Leadership','Brotherhood','Service','Academic','Professionalism','Other'
              )),
  message     TEXT,
  xp_awarded  INT         NOT NULL DEFAULT 10,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- Intentionally no updated_at — props are immutable once given
);

ALTER TABLE props ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON props
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_props_chapter
  ON props(chapter_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_props_to_member
  ON props(to_id);

CREATE INDEX IF NOT EXISTS idx_props_from_member
  ON props(from_id);

GRANT SELECT, INSERT ON props TO blue_ledger_app;
GRANT ALL PRIVILEGES ON props TO blue_ledger_admin;
