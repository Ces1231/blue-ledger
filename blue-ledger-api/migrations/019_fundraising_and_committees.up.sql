-- ============================================================
-- Migration: 019_fundraising_and_committees
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create fundraising_campaigns (chapter fundraising
--          goal tracking) and committees (chapter committee
--          structure with membership list).
--
--          goal_cents and current_cents use integer cents
--          throughout (no floating-point amounts).
--          member_ids on committees is a UUID[] for lightweight
--          membership without a separate join table.
--
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: fundraising_campaigns
-- ============================================================
CREATE TABLE IF NOT EXISTS fundraising_campaigns (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  title           TEXT        NOT NULL,
  description     TEXT,
  goal_cents      INT         NOT NULL CHECK (goal_cents > 0),
  current_cents   INT         NOT NULL DEFAULT 0 CHECK (current_cents >= 0),
  category        TEXT,
  deadline        DATE,
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  created_by      UUID        NOT NULL REFERENCES members(id) ON DELETE RESTRICT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE fundraising_campaigns ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON fundraising_campaigns
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_fundraising_chapter
  ON fundraising_campaigns(chapter_id);

CREATE INDEX IF NOT EXISTS idx_fundraising_active
  ON fundraising_campaigns(chapter_id, is_active);

CREATE TRIGGER trg_fundraising_campaigns_updated_at
  BEFORE UPDATE ON fundraising_campaigns
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON fundraising_campaigns TO blue_ledger_app;
GRANT ALL PRIVILEGES ON fundraising_campaigns TO blue_ledger_admin;

-- ============================================================
-- Table: committees
-- Chapter committee definitions with membership list.
-- member_ids is a UUID[] of member IDs on the committee.
-- For lightweight read/write — no separate join table at v1 scale.
-- ============================================================
CREATE TABLE IF NOT EXISTS committees (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id       UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  name             TEXT        NOT NULL,
  chair_id         UUID        REFERENCES members(id) ON DELETE SET NULL,
  description      TEXT,
  meeting_schedule TEXT,                                         -- "Every Tuesday at 7pm"
  member_ids       UUID[]      NOT NULL DEFAULT '{}',            -- member UUIDs on this committee
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE committees ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON committees
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_committees_chapter
  ON committees(chapter_id);

CREATE INDEX IF NOT EXISTS idx_committees_chair
  ON committees(chair_id)
  WHERE chair_id IS NOT NULL;

-- GIN index for "committees this member belongs to" queries
CREATE INDEX IF NOT EXISTS idx_committees_members
  ON committees USING GIN (member_ids);

CREATE TRIGGER trg_committees_updated_at
  BEFORE UPDATE ON committees
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON committees TO blue_ledger_app;
GRANT ALL PRIVILEGES ON committees TO blue_ledger_admin;
