-- ============================================================
-- Migration: 011_mentorships_and_intake
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create mentorships (mentorship pairings within a
--          chapter) and intake_prospects (the prospect pipeline
--          for new member intake tracking).
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: mentorships
-- mentor_id is required; mentee_id is NULL until paired.
-- focus_areas is a TEXT[] (e.g., ['Career Development','Leadership'])
-- ============================================================
CREATE TABLE IF NOT EXISTS mentorships (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  mentor_id   UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  mentee_id   UUID        REFERENCES members(id) ON DELETE SET NULL,   -- NULL if not yet paired

  focus_areas TEXT[],                                          -- ['Career Development','Leadership']
  bio         TEXT,                                             -- mentor's bio / what they offer
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  started_at  TIMESTAMPTZ,
  ended_at    TIMESTAMPTZ,

  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE mentorships ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON mentorships
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_mentorships_chapter
  ON mentorships(chapter_id);

CREATE INDEX IF NOT EXISTS idx_mentorships_mentor
  ON mentorships(mentor_id);

CREATE INDEX IF NOT EXISTS idx_mentorships_mentee
  ON mentorships(mentee_id)
  WHERE mentee_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_mentorships_active
  ON mentorships(chapter_id, is_active);

CREATE TRIGGER trg_mentorships_updated_at
  BEFORE UPDATE ON mentorships
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON mentorships TO blue_ledger_app;
GRANT ALL PRIVILEGES ON mentorships TO blue_ledger_admin;

-- ============================================================
-- Table: intake_prospects
-- Prospect pipeline for chapter intake. Tracks individuals
-- from first inquiry through acceptance or decline.
-- ============================================================
CREATE TABLE IF NOT EXISTS intake_prospects (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,

  name            TEXT        NOT NULL,
  email           TEXT,
  gpa             TEXT,
  academic_year   TEXT,

  -- Who referred this prospect
  referred_by     UUID        REFERENCES members(id) ON DELETE SET NULL,

  stage           TEXT        NOT NULL DEFAULT 'inquiry'
                  CHECK (stage IN (
                    'inquiry','under_review','interview_scheduled',
                    'invited_to_rush','accepted','declined','inactive'
                  )),

  interest_level  TEXT
                  CHECK (interest_level IN ('low','medium','high','very_high')),

  notes           TEXT,
  intake_date     DATE        NOT NULL DEFAULT CURRENT_DATE,
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE intake_prospects ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON intake_prospects
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_intake_chapter
  ON intake_prospects(chapter_id);

CREATE INDEX IF NOT EXISTS idx_intake_stage
  ON intake_prospects(chapter_id, stage);

CREATE INDEX IF NOT EXISTS idx_intake_active
  ON intake_prospects(chapter_id, is_active, intake_date DESC);

CREATE TRIGGER trg_intake_prospects_updated_at
  BEFORE UPDATE ON intake_prospects
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON intake_prospects TO blue_ledger_app;
GRANT ALL PRIVILEGES ON intake_prospects TO blue_ledger_admin;
