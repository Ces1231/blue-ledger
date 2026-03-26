-- ============================================================
-- Migration: 014_store_and_goals
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create the XP reward store (store_items, store_orders)
--          and chapter_goals (chapter-level progress tracking).
--          XP amounts are stored as integers throughout.
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: store_items
-- Items members can redeem using XP points.
-- ============================================================
CREATE TABLE IF NOT EXISTS store_items (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  name        TEXT        NOT NULL,
  description TEXT,
  xp_cost     INT         NOT NULL CHECK (xp_cost > 0),
  category    TEXT,
  stock       INT         NOT NULL DEFAULT 0 CHECK (stock >= 0),  -- 0 = unlimited if is_active
  icon        TEXT,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE store_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON store_items
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_store_items_chapter
  ON store_items(chapter_id);

CREATE INDEX IF NOT EXISTS idx_store_items_active
  ON store_items(chapter_id, is_active);

GRANT SELECT, INSERT, UPDATE, DELETE ON store_items TO blue_ledger_app;
GRANT ALL PRIVILEGES ON store_items TO blue_ledger_admin;

-- ============================================================
-- Table: store_orders
-- Records when a member redeems XP for a store item.
-- xp_spent is captured at order time (item cost may change later).
-- ============================================================
CREATE TABLE IF NOT EXISTS store_orders (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id   UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  item_id     UUID        NOT NULL REFERENCES store_items(id) ON DELETE RESTRICT,
  xp_spent    INT         NOT NULL CHECK (xp_spent > 0),
  status      TEXT        NOT NULL DEFAULT 'pending'
              CHECK (status IN ('pending','fulfilled','cancelled')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE store_orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON store_orders
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_store_orders_chapter
  ON store_orders(chapter_id);

CREATE INDEX IF NOT EXISTS idx_store_orders_member
  ON store_orders(member_id);

CREATE INDEX IF NOT EXISTS idx_store_orders_status
  ON store_orders(chapter_id, status);

GRANT SELECT, INSERT, UPDATE ON store_orders TO blue_ledger_app;
GRANT ALL PRIVILEGES ON store_orders TO blue_ledger_admin;

-- ============================================================
-- Table: chapter_goals
-- Chapter-level progress goals (e.g., service hours, GPA, etc.)
-- target and current use NUMERIC(10,2) to support fractional
-- values (hours, GPA, percentages).
-- ============================================================
CREATE TABLE IF NOT EXISTS chapter_goals (
  id          UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID           NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  title       TEXT           NOT NULL,
  description TEXT,
  category    TEXT,
  target      NUMERIC(10,2)  NOT NULL CHECK (target > 0),
  current     NUMERIC(10,2)  NOT NULL DEFAULT 0 CHECK (current >= 0),
  unit        TEXT,                                              -- "hours", "members", "%", etc.
  deadline    DATE,
  owner_id    UUID           REFERENCES members(id) ON DELETE SET NULL,
  status      TEXT           NOT NULL DEFAULT 'in_progress'
              CHECK (status IN ('in_progress','achieved','missed','paused')),
  xp_reward   INT            NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

ALTER TABLE chapter_goals ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON chapter_goals
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_chapter_goals_chapter
  ON chapter_goals(chapter_id);

CREATE INDEX IF NOT EXISTS idx_chapter_goals_status
  ON chapter_goals(chapter_id, status);

CREATE TRIGGER trg_chapter_goals_updated_at
  BEFORE UPDATE ON chapter_goals
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON chapter_goals TO blue_ledger_app;
GRANT ALL PRIVILEGES ON chapter_goals TO blue_ledger_admin;
