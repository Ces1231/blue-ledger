-- ============================================================
-- Migration: 015_job_board_and_resources
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create job_board (member-posted job/internship
--          listings) and resources (file/link library for
--          chapter documents, guides, and media).
--          tags on resources is TEXT[] for multi-tag search.
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: job_board
-- ============================================================
CREATE TABLE IF NOT EXISTS job_board (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  posted_by   UUID        NOT NULL REFERENCES members(id)  ON DELETE RESTRICT,
  company     TEXT        NOT NULL,
  job_title   TEXT        NOT NULL,
  job_type    TEXT
              CHECK (job_type IN (
                'Internship','Full-Time','Part-Time','Contract','Fellowship'
              )),
  location    TEXT,
  link        TEXT,
  description TEXT,
  fields      TEXT[],                                            -- ["Finance","Tech","Healthcare"]
  deadline    DATE,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE job_board ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON job_board
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_job_board_chapter
  ON job_board(chapter_id);

CREATE INDEX IF NOT EXISTS idx_job_board_active_recent
  ON job_board(chapter_id, is_active, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_job_board_posted_by
  ON job_board(posted_by);

CREATE TRIGGER trg_job_board_updated_at
  BEFORE UPDATE ON job_board
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON job_board TO blue_ledger_app;
GRANT ALL PRIVILEGES ON job_board TO blue_ledger_admin;

-- ============================================================
-- Table: resources
-- Chapter resource library — links, uploaded files, guides.
-- file_key stores the Cloudflare R2 object key when a file
-- has been uploaded. url is used for external links.
-- ============================================================
CREATE TABLE IF NOT EXISTS resources (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  title       TEXT        NOT NULL,
  description TEXT,
  type        TEXT        NOT NULL
              CHECK (type IN ('link','pdf','video','image','document','other')),
  url         TEXT,                                              -- external URL or R2 presigned URL
  file_key    TEXT,                                              -- Cloudflare R2 object key
  category    TEXT,
  tags        TEXT[],                                            -- ["forms","training","finance"]
  posted_by   UUID        NOT NULL REFERENCES members(id) ON DELETE RESTRICT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- Intentionally no updated_at — resources are replaced, not edited
);

ALTER TABLE resources ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON resources
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_resources_chapter
  ON resources(chapter_id);

CREATE INDEX IF NOT EXISTS idx_resources_type
  ON resources(chapter_id, type);

CREATE INDEX IF NOT EXISTS idx_resources_category
  ON resources(chapter_id, category);

GRANT SELECT, INSERT, UPDATE, DELETE ON resources TO blue_ledger_app;
GRANT ALL PRIVILEGES ON resources TO blue_ledger_admin;
