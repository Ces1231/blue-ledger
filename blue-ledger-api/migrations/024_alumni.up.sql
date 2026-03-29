-- ============================================================
-- Migration: 024_alumni
-- Date: 2026-03-28
-- Purpose: Create alumni table for chapter alumni directory.
-- Safety: SAFE
-- Rollback: EASY
-- ============================================================

CREATE TABLE IF NOT EXISTS alumni (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  name            TEXT        NOT NULL,
  graduation_year INT,
  employer        TEXT,
  title           TEXT,
  city            TEXT,
  linkedin_url    TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE alumni ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON alumni
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_alumni_chapter
  ON alumni(chapter_id);

CREATE INDEX IF NOT EXISTS idx_alumni_grad_year
  ON alumni(chapter_id, graduation_year DESC NULLS LAST, name ASC);

CREATE TRIGGER trg_alumni_updated_at
  BEFORE UPDATE ON alumni
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON alumni TO blue_ledger_app;
GRANT ALL PRIVILEGES ON alumni TO blue_ledger_admin;
