-- ============================================================
-- Migration: 012_votes
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create votes (chapter voting sessions) and
--          vote_responses (individual member responses).
--          UNIQUE(vote_id, member_id) prevents duplicate votes.
--          is_anonymous=TRUE hides member attribution (enforced
--          at the application layer; DB stores member_id for
--          de-duplication only).
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: votes
-- ============================================================
CREATE TABLE IF NOT EXISTS votes (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id    UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  title         TEXT        NOT NULL,
  description   TEXT,
  vote_type     TEXT        NOT NULL
                CHECK (vote_type IN ('election','referendum','motion')),
  options       TEXT[]      NOT NULL,                            -- ["Option A","Option B"]
  deadline      DATE,
  is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
  is_anonymous  BOOLEAN     NOT NULL DEFAULT TRUE,               -- hides member attribution on results
  created_by    UUID        NOT NULL REFERENCES members(id) ON DELETE RESTRICT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE votes ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON votes
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_votes_chapter
  ON votes(chapter_id);

CREATE INDEX IF NOT EXISTS idx_votes_active
  ON votes(chapter_id, is_active, deadline);

CREATE TRIGGER trg_votes_updated_at
  BEFORE UPDATE ON votes
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE, DELETE ON votes TO blue_ledger_app;
GRANT ALL PRIVILEGES ON votes TO blue_ledger_admin;

-- ============================================================
-- Table: vote_responses
-- UNIQUE(vote_id, member_id) is the primary integrity guarantee.
-- ============================================================
CREATE TABLE IF NOT EXISTS vote_responses (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id    UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  vote_id       UUID        NOT NULL REFERENCES votes(id)   ON DELETE CASCADE,
  member_id     UUID        NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  option_index  INT         NOT NULL,                            -- 0-based index into votes.options
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- Prevents a member from voting more than once
  UNIQUE(vote_id, member_id)
);

ALTER TABLE vote_responses ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON vote_responses
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_vote_responses_vote
  ON vote_responses(vote_id);

CREATE INDEX IF NOT EXISTS idx_vote_responses_member
  ON vote_responses(member_id);

GRANT SELECT, INSERT ON vote_responses TO blue_ledger_app;
GRANT ALL PRIVILEGES ON vote_responses TO blue_ledger_admin;
