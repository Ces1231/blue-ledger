-- ============================================================
-- Migration: 013_minutes_and_scholarships
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create chapter_minutes (meeting minutes with draft/
--          final workflow) and scholarships (scholarship
--          application tracking with reviewer workflow).
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: chapter_minutes
-- ============================================================
CREATE TABLE IF NOT EXISTS chapter_minutes (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id          UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  meeting_date        DATE        NOT NULL,
  title               TEXT        NOT NULL,
  body                TEXT        NOT NULL,                      -- Markdown or plain text
  recorder_id         UUID        REFERENCES members(id) ON DELETE SET NULL,
  quorum              BOOLEAN     NOT NULL DEFAULT FALSE,
  attendee_ids        UUID[],                                    -- member UUIDs who attended
  xp_for_attendance   INT         NOT NULL DEFAULT 0,            -- XP awarded to attendees on finalize
  status              TEXT        NOT NULL DEFAULT 'draft'
                      CHECK (status IN ('draft','final')),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE chapter_minutes ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON chapter_minutes
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_minutes_chapter
  ON chapter_minutes(chapter_id);

CREATE INDEX IF NOT EXISTS idx_minutes_meeting_date
  ON chapter_minutes(chapter_id, meeting_date DESC);

CREATE INDEX IF NOT EXISTS idx_minutes_status
  ON chapter_minutes(chapter_id, status);

CREATE TRIGGER trg_chapter_minutes_updated_at
  BEFORE UPDATE ON chapter_minutes
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON chapter_minutes TO blue_ledger_app;
GRANT ALL PRIVILEGES ON chapter_minutes TO blue_ledger_admin;

-- ============================================================
-- Table: scholarships
-- Scholarship application records. Applications move through
-- a reviewer workflow: submitted → under_review → finalist
-- → awarded | declined | needs_more_info.
-- ============================================================
CREATE TABLE IF NOT EXISTS scholarships (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id          UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,

  -- Applicant info (not tied to a member — external applicants welcome)
  member_id           UUID        REFERENCES members(id) ON DELETE SET NULL,  -- NULL for external applicants
  applicant_name      TEXT        NOT NULL,
  school              TEXT,
  gpa                 TEXT,
  major               TEXT,
  academic_year       INT,
  city                TEXT,
  semester            TEXT,                                      -- "Spring 2026"
  essay               TEXT,
  submitted_at        DATE,

  amount_cents        INT,                                       -- scholarship amount in cents

  status              TEXT        NOT NULL DEFAULT 'submitted'
                      CHECK (status IN (
                        'submitted','under_review','finalist',
                        'awarded','declined','needs_more_info'
                      )),

  notes               TEXT,
  reviewer_id         UUID        REFERENCES members(id) ON DELETE SET NULL,
  reviewed_at         TIMESTAMPTZ,

  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE scholarships ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON scholarships
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_scholarships_chapter
  ON scholarships(chapter_id);

CREATE INDEX IF NOT EXISTS idx_scholarships_status
  ON scholarships(chapter_id, status);

CREATE INDEX IF NOT EXISTS idx_scholarships_member
  ON scholarships(member_id)
  WHERE member_id IS NOT NULL;

CREATE TRIGGER trg_scholarships_updated_at
  BEFORE UPDATE ON scholarships
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON scholarships TO blue_ledger_app;
GRANT ALL PRIVILEGES ON scholarships TO blue_ledger_admin;
