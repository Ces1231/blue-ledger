-- ============================================================
-- Migration: 018_messaging
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create message_threads (conversation containers) and
--          messages (individual messages within a thread).
--
--          participant_ids and read_by are UUID arrays.
--          participant_ids stores who belongs to the thread.
--          read_by stores which users have read each message
--          (append-only from the application layer).
--
--          Both tables are chapter-scoped with RLS.
--
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: message_threads
-- A thread is a conversation between 2+ chapter members.
-- participant_ids is a UUID[] of member IDs.
-- ============================================================
CREATE TABLE IF NOT EXISTS message_threads (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id       UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  participant_ids  UUID[]      NOT NULL,                         -- member UUIDs in this thread
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE message_threads ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON message_threads
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_message_threads_chapter
  ON message_threads(chapter_id);

-- GIN index for efficient "threads containing this member" queries
CREATE INDEX IF NOT EXISTS idx_message_threads_participants
  ON message_threads USING GIN (participant_ids);

GRANT SELECT, INSERT, UPDATE ON message_threads TO blue_ledger_app;
GRANT ALL PRIVILEGES ON message_threads TO blue_ledger_admin;

-- ============================================================
-- Table: messages
-- Individual messages within a thread.
-- read_by is a UUID[] of member IDs who have read this message.
-- Append-only from the application perspective (no edits/deletes
-- exposed in the API).
-- ============================================================
CREATE TABLE IF NOT EXISTS messages (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  thread_id   UUID        NOT NULL REFERENCES message_threads(id) ON DELETE CASCADE,
  chapter_id  UUID        NOT NULL REFERENCES chapters(id)        ON DELETE CASCADE,
  sender_id   UUID        NOT NULL REFERENCES members(id)         ON DELETE CASCADE,
  body        TEXT        NOT NULL,
  read_by     UUID[]      NOT NULL DEFAULT '{}',                  -- member UUIDs who have read this
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- No updated_at — messages are not editable
);

ALTER TABLE messages ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON messages
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX IF NOT EXISTS idx_messages_thread
  ON messages(thread_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_messages_sender
  ON messages(sender_id);

CREATE INDEX IF NOT EXISTS idx_messages_chapter_recent
  ON messages(chapter_id, created_at DESC);

-- GIN index for efficient "unread by this member" queries
CREATE INDEX IF NOT EXISTS idx_messages_read_by
  ON messages USING GIN (read_by);

GRANT SELECT, INSERT, UPDATE ON messages TO blue_ledger_app;
GRANT ALL PRIVILEGES ON messages TO blue_ledger_admin;
