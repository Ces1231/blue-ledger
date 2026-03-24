-- ============================================================
-- Migration: 016_point_economy_and_config
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create point_economy (per-chapter configurable XP
--          values per activity) and chapter_config (key/value
--          store for chapter-level feature flags and settings).
--
--          Also seeds default point_economy rows using a helper
--          function that can be called per chapter on onboarding.
--          The function seed_default_point_economy(chapter_id)
--          inserts the 12 default activities from the prototype
--          with INSERT ... ON CONFLICT DO NOTHING (idempotent).
--
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables + function)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: point_economy
-- Per-chapter configurable XP values per activity type.
-- UNIQUE(chapter_id, activity) enforces one XP value per action.
-- ============================================================
CREATE TABLE IF NOT EXISTS point_economy (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  activity    TEXT        NOT NULL,
  xp          INT         NOT NULL CHECK (xp >= 0),
  category    TEXT,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  UNIQUE(chapter_id, activity)
);

ALTER TABLE point_economy ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON point_economy
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_point_economy_chapter
  ON point_economy(chapter_id);

GRANT SELECT, INSERT, UPDATE, DELETE ON point_economy TO blue_ledger_app;
GRANT ALL PRIVILEGES ON point_economy TO blue_ledger_admin;

-- ============================================================
-- Table: chapter_config
-- Flexible key/value configuration store per chapter.
-- Used for Zeffy form IDs, feature flags, display preferences,
-- XP level thresholds, and other settings that change rarely.
-- UNIQUE(chapter_id, key) enforces one value per key per chapter.
-- ============================================================
CREATE TABLE IF NOT EXISTS chapter_config (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  key         TEXT        NOT NULL,
  value       TEXT        NOT NULL,
  description TEXT,                                              -- human-readable description of this key
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(chapter_id, key)
);

ALTER TABLE chapter_config ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON chapter_config
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_chapter_config_chapter
  ON chapter_config(chapter_id);

CREATE TRIGGER trg_chapter_config_updated_at
  BEFORE UPDATE ON chapter_config
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON chapter_config TO blue_ledger_app;
GRANT ALL PRIVILEGES ON chapter_config TO blue_ledger_admin;

-- ============================================================
-- Function: seed_default_point_economy(p_chapter_id UUID)
-- Inserts the 12 default XP activities from the prototype into
-- point_economy for a given chapter. Uses ON CONFLICT DO NOTHING
-- so it is safe to call multiple times (idempotent).
--
-- Called during chapter onboarding by the API's registration
-- handler after the chapter row is created.
-- ============================================================
CREATE OR REPLACE FUNCTION seed_default_point_economy(p_chapter_id UUID)
RETURNS VOID AS $$
BEGIN
  INSERT INTO point_economy (chapter_id, activity, xp, category) VALUES
    (p_chapter_id, 'Chapter Meeting (on time)',        50,  'Attendance'),
    (p_chapter_id, 'Chapter Meeting (late)',           25,  'Attendance'),
    (p_chapter_id, 'RSVP (attending)',                 10,  'Attendance'),
    (p_chapter_id, 'Community Service (per hour)',     20,  'Service'),
    (p_chapter_id, 'Props Given',                      0,   'Recognition'),
    (p_chapter_id, 'Props Received',                   10,  'Recognition'),
    (p_chapter_id, 'Dues Paid (on time)',              100, 'Dues'),
    (p_chapter_id, 'Dues Paid (late)',                  50, 'Dues'),
    (p_chapter_id, 'Badge Earned',                      0,  'Achievement'),
    (p_chapter_id, 'Quest Completed',                   0,  'Achievement'),
    (p_chapter_id, 'Mentorship Session',               30,  'Development'),
    (p_chapter_id, 'Admin Award',                       0,  'Admin')
  ON CONFLICT (chapter_id, activity) DO NOTHING;
END;
$$ LANGUAGE plpgsql;
