-- ============================================================
-- Migration: 008_badges_and_quests
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create the badge and quest gamification tables.
--
--   badges              — badge definitions per chapter.
--                         criteria is JSONB for flexible badge
--                         engine logic (no schema change needed
--                         to add new badge types).
--   member_badges       — awarded badge instances.
--   quests              — multi-step quest definitions per chapter.
--                         steps is JSONB array.
--   member_quest_progress — per-member quest completion state.
--
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: badges
-- ============================================================
CREATE TABLE IF NOT EXISTS badges (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  name            TEXT        NOT NULL,
  icon            TEXT,                                          -- emoji or icon key
  category        TEXT,
  description     TEXT,
  requirement     TEXT,                                          -- human-readable requirement text
  xp_reward       INT         NOT NULL DEFAULT 0,
  rarity          TEXT        NOT NULL DEFAULT 'common'
                  CHECK (rarity IN ('common','uncommon','rare','legendary')),
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  -- Flexible criteria evaluated by the badge engine
  -- Example: {"type":"xp_threshold","threshold":500}
  -- Example: {"type":"meeting_streak","count":5}
  -- Example: {"type":"service_hours","min_hours":10}
  criteria        JSONB       NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(chapter_id, name)
);

ALTER TABLE badges ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON badges
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_badges_chapter
  ON badges(chapter_id);

CREATE INDEX IF NOT EXISTS idx_badges_active
  ON badges(chapter_id, is_active);

GRANT SELECT, INSERT, UPDATE, DELETE ON badges TO blue_ledger_app;
GRANT ALL PRIVILEGES ON badges TO blue_ledger_admin;

-- ============================================================
-- Table: member_badges
-- ============================================================
CREATE TABLE IF NOT EXISTS member_badges (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id   UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  badge_id    UUID        NOT NULL REFERENCES badges(id)   ON DELETE CASCADE,
  awarded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  awarded_by  UUID        REFERENCES members(id) ON DELETE SET NULL,  -- NULL for system auto-award
  UNIQUE(member_id, badge_id)
);

ALTER TABLE member_badges ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON member_badges
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_member_badges_member
  ON member_badges(member_id);

CREATE INDEX IF NOT EXISTS idx_member_badges_chapter
  ON member_badges(chapter_id);

CREATE INDEX IF NOT EXISTS idx_member_badges_badge
  ON member_badges(badge_id);

GRANT SELECT, INSERT, UPDATE, DELETE ON member_badges TO blue_ledger_app;
GRANT ALL PRIVILEGES ON member_badges TO blue_ledger_admin;

-- ============================================================
-- Table: quests
-- Multi-step challenge definitions seeded per chapter.
-- steps JSONB format: [{"title":"...","description":"...","target_count":3}]
-- ============================================================
CREATE TABLE IF NOT EXISTS quests (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  title           TEXT        NOT NULL,
  description     TEXT,
  xp_reward       INT         NOT NULL DEFAULT 0,
  badge_reward_id UUID        REFERENCES badges(id) ON DELETE SET NULL,
  steps           JSONB       NOT NULL DEFAULT '[]',
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE quests ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON quests
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_quests_chapter
  ON quests(chapter_id);

CREATE INDEX IF NOT EXISTS idx_quests_active
  ON quests(chapter_id, is_active);

GRANT SELECT, INSERT, UPDATE, DELETE ON quests TO blue_ledger_app;
GRANT ALL PRIVILEGES ON quests TO blue_ledger_admin;

-- ============================================================
-- Table: member_quest_progress
-- Tracks each member's per-quest completion state.
-- progress JSONB format: {"0": 2, "1": 1} (step_index: count)
-- ============================================================
CREATE TABLE IF NOT EXISTS member_quest_progress (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id       UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  quest_id        UUID        NOT NULL REFERENCES quests(id)   ON DELETE CASCADE,
  progress        JSONB       NOT NULL DEFAULT '{}',
  completed_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(member_id, quest_id)
);

ALTER TABLE member_quest_progress ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON member_quest_progress
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_mqp_member
  ON member_quest_progress(member_id);

CREATE INDEX IF NOT EXISTS idx_mqp_chapter
  ON member_quest_progress(chapter_id);

CREATE INDEX IF NOT EXISTS idx_mqp_quest
  ON member_quest_progress(quest_id);

CREATE TRIGGER trg_member_quest_progress_updated_at
  BEFORE UPDATE ON member_quest_progress
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON member_quest_progress TO blue_ledger_app;
GRANT ALL PRIVILEGES ON member_quest_progress TO blue_ledger_admin;
