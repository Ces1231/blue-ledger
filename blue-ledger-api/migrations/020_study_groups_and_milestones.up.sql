-- ============================================================
-- Migration: 020_study_groups_and_milestones
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create study_groups (collaborative study sessions
--          organized by chapter members) and milestones
--          (significant life events for brotherhood recognition).
--
--          study_groups.member_ids is UUID[] for attendee list.
--          milestones.type CHECK ensures only defined life event
--          types are stored: birthday, graduation, new_job,
--          engagement, other.
--
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: study_groups
-- ============================================================
CREATE TABLE IF NOT EXISTS study_groups (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  topic       TEXT        NOT NULL,
  host_id     UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  date        DATE        NOT NULL,
  location    TEXT,
  member_ids  UUID[]      NOT NULL DEFAULT '{}',                -- attending member UUIDs
  xp_reward   INT         NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- No updated_at — study groups are not edited after creation
);

ALTER TABLE study_groups ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON study_groups
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_study_groups_chapter
  ON study_groups(chapter_id);

CREATE INDEX IF NOT EXISTS idx_study_groups_host
  ON study_groups(host_id);

CREATE INDEX IF NOT EXISTS idx_study_groups_date
  ON study_groups(chapter_id, date DESC);

-- GIN index for "study groups this member attended" queries
CREATE INDEX IF NOT EXISTS idx_study_groups_members
  ON study_groups USING GIN (member_ids);

GRANT SELECT, INSERT, UPDATE, DELETE ON study_groups TO blue_ledger_app;
GRANT ALL PRIVILEGES ON study_groups TO blue_ledger_admin;

-- ============================================================
-- Table: milestones
-- Life event milestones for brotherhood recognition.
-- member_id references the brother being celebrated.
-- ============================================================
CREATE TABLE IF NOT EXISTS milestones (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id   UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  type        TEXT        NOT NULL
              CHECK (type IN (
                'birthday','graduation','new_job','engagement','other'
              )),
  title       TEXT        NOT NULL,
  date        DATE,
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- No updated_at — milestones are record-of-fact, immutable
);

ALTER TABLE milestones ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON milestones
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_milestones_chapter
  ON milestones(chapter_id);

CREATE INDEX IF NOT EXISTS idx_milestones_member
  ON milestones(member_id);

CREATE INDEX IF NOT EXISTS idx_milestones_type_date
  ON milestones(chapter_id, type, date DESC);

GRANT SELECT, INSERT, UPDATE, DELETE ON milestones TO blue_ledger_app;
GRANT ALL PRIVILEGES ON milestones TO blue_ledger_admin;
