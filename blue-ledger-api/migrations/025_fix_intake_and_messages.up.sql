-- ============================================================
-- Migration: 025_fix_intake_and_messages
-- Date: 2026-03-25
-- Author: Blue Ledger
-- Purpose: Align intake_prospects and message_threads schemas
--          with the service layer expectations.
--          - intake_prospects: add first_name, last_name, phone,
--            university, grad_year, added_by, archived_at
--          - message_threads: add subject, created_by, is_group_chat
--          - Create thread_participants join table
-- Safety: SAFE
-- Rollback: EASY (see 025_fix_intake_and_messages.down.sql)
-- Locking: METADATA ONLY (new columns / new table)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Fix: intake_prospects
-- Add columns expected by the intake service layer.
-- ============================================================
ALTER TABLE intake_prospects
    ADD COLUMN IF NOT EXISTS first_name  TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_name   TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS phone       TEXT,
    ADD COLUMN IF NOT EXISTS university  TEXT,
    ADD COLUMN IF NOT EXISTS grad_year   INTEGER,
    ADD COLUMN IF NOT EXISTS added_by    UUID         REFERENCES members(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

-- Seed first_name from existing name column (full name → first_name)
UPDATE intake_prospects
    SET first_name = name, last_name = ''
    WHERE first_name = '' AND name IS NOT NULL AND name <> '';

-- ============================================================
-- Fix: message_threads
-- Add subject, created_by, is_group_chat missing from original schema.
-- ============================================================
ALTER TABLE message_threads
    ADD COLUMN IF NOT EXISTS subject       TEXT    NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS created_by    UUID    REFERENCES members(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS is_group_chat BOOLEAN NOT NULL DEFAULT FALSE;

-- ============================================================
-- Table: thread_participants
-- Explicit join table for thread membership, required by the
-- messages service layer.  Any existing participant_ids arrays
-- are migrated into rows here.
-- ============================================================
CREATE TABLE IF NOT EXISTS thread_participants (
    thread_id  UUID        NOT NULL REFERENCES message_threads(id) ON DELETE CASCADE,
    member_id  UUID        NOT NULL REFERENCES members(id)         ON DELETE CASCADE,
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (thread_id, member_id)
);

ALTER TABLE thread_participants ENABLE ROW LEVEL SECURITY;

CREATE POLICY chapter_isolation ON thread_participants
    USING (
        EXISTS (
            SELECT 1 FROM message_threads mt
            WHERE mt.id = thread_participants.thread_id
              AND mt.chapter_id = current_setting('app.chapter_id')::uuid
        )
    );

CREATE INDEX IF NOT EXISTS idx_thread_participants_thread
    ON thread_participants(thread_id);

CREATE INDEX IF NOT EXISTS idx_thread_participants_member
    ON thread_participants(member_id);

GRANT SELECT, INSERT, DELETE ON thread_participants TO blue_ledger_app;
GRANT ALL PRIVILEGES ON thread_participants TO blue_ledger_admin;

-- Migrate any existing participant_ids array rows to thread_participants
INSERT INTO thread_participants (thread_id, member_id)
SELECT mt.id, unnest(mt.participant_ids)
FROM message_threads mt
WHERE mt.participant_ids IS NOT NULL
  AND array_length(mt.participant_ids, 1) > 0
ON CONFLICT DO NOTHING;
