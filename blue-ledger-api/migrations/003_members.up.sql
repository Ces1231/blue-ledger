-- ============================================================
-- Migration: 003_members
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create the members table — the per-chapter profile
--          for a user's membership in a specific chapter.
--          One user can have one members row per chapter.
--          Includes XP, dues status, role, gamification fields,
--          and privacy settings as JSONB.
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new table)
-- Est. duration: <1s
-- ============================================================

CREATE TABLE IF NOT EXISTS members (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id          UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  user_id             UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,

  -- Chapter-scoped display ID (e.g., ΤΤΣ-001 or MBR-042)
  -- Combination of chapter.member_id_prefix + sequence
  display_id          TEXT        NOT NULL,

  -- Denormalized display fields (source of truth remains on users,
  -- but chapter-specific name overrides are allowed)
  name                TEXT        NOT NULL,
  initials            TEXT,                                        -- "J.W." — derived or overridden
  email               TEXT,                                        -- copied from users.email at creation; may diverge

  -- Chapter role
  role                TEXT        NOT NULL DEFAULT 'member'
                      CHECK (role IN ('member','chair','pia','admin','sysadmin')),

  -- Membership metadata
  inducted_year       INT,
  status              TEXT        NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active','inactive','alumni','suspended','pledging')),

  -- Professional profile
  employer            TEXT,
  title               TEXT,                                        -- job title
  city                TEXT,
  linkedin            TEXT,
  bio                 TEXT,

  -- Avatar display
  avatar_url          TEXT,
  avatar_bg           TEXT,                                        -- hex color, e.g. "#1a3a6b"
  avatar_fg           TEXT,                                        -- hex color, e.g. "#ffffff"

  -- XP & gamification
  xp_total            INT         NOT NULL DEFAULT 0,
  xp_semester         INT         NOT NULL DEFAULT 0,             -- reset each semester
  service_hours       NUMERIC(6,2) NOT NULL DEFAULT 0,

  -- Level system (denormalized cache; recalculated by XP engine)
  level_key           TEXT        NOT NULL DEFAULT 'neo'
                      CHECK (level_key IN ('neo','scholar','leader','sage','legend','icon')),

  streak              INT         NOT NULL DEFAULT 0,              -- consecutive meetings attended
  last_meeting_date   DATE,

  -- Dues status (denormalized cache; source of truth is dues_records)
  dues_status         TEXT        NOT NULL DEFAULT 'unpaid'
                      CHECK (dues_status IN ('paid','unpaid','late','waived','outstanding')),

  -- Privacy settings (JSONB for granular field-level visibility)
  -- Example: {"email":true,"employer":true,"linkedin":false,"city":true}
  privacy_settings    JSONB       NOT NULL DEFAULT '{"email":true,"employer":true,"linkedin":true,"city":true}',

  -- Timestamps
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at          TIMESTAMPTZ,

  -- Constraints
  UNIQUE(chapter_id, user_id),
  UNIQUE(chapter_id, display_id)
);

-- ── Row-Level Security ──────────────────────────────────────
ALTER TABLE members ENABLE ROW LEVEL SECURITY;

CREATE POLICY chapter_isolation ON members
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

-- ── Indexes ──────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_members_chapter
  ON members(chapter_id);

CREATE INDEX IF NOT EXISTS idx_members_user
  ON members(user_id);

CREATE INDEX IF NOT EXISTS idx_members_chapter_user
  ON members(chapter_id, user_id);

CREATE INDEX IF NOT EXISTS idx_members_role
  ON members(chapter_id, role);

CREATE INDEX IF NOT EXISTS idx_members_xp
  ON members(chapter_id, xp_total DESC);

CREATE INDEX IF NOT EXISTS idx_members_status
  ON members(chapter_id, status)
  WHERE deleted_at IS NULL;

-- ── updated_at trigger ───────────────────────────────────────
CREATE TRIGGER trg_members_updated_at
  BEFORE UPDATE ON members
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ── Grants ───────────────────────────────────────────────────
GRANT SELECT, INSERT, UPDATE, DELETE ON members TO blue_ledger_app;
GRANT ALL PRIVILEGES ON members TO blue_ledger_admin;
