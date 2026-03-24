-- ============================================================
-- Migration: 010_notifications
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create the notifications table. Notifications are
--          addressed to users (not members) so they survive
--          chapter role changes and cross-chapter sessions.
--          chapter_id is still present for RLS isolation.
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new table)
-- Est. duration: <1s
-- ============================================================

CREATE TABLE IF NOT EXISTS notifications (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,

  -- Addressed to user, not member — survives chapter hops
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  type        TEXT        NOT NULL
              CHECK (type IN (
                'badge','prop','xp','dues','event',
                'announcement','level','system'
              )),

  title       TEXT        NOT NULL,
  body        TEXT        NOT NULL,
  link        TEXT,                                              -- frontend route to navigate on click

  is_read     BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- Intentionally no updated_at — only is_read changes and
  -- that is handled by a targeted UPDATE; no trigger needed.
);

-- ── Row-Level Security ──────────────────────────────────────
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;

CREATE POLICY chapter_isolation ON notifications
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

-- ── Indexes ──────────────────────────────────────────────────
-- Primary query pattern: "my unread notifications, newest first"
CREATE INDEX IF NOT EXISTS idx_notifications_user
  ON notifications(user_id, is_read, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_chapter_recent
  ON notifications(chapter_id, created_at DESC);

-- ── Grants ───────────────────────────────────────────────────
GRANT SELECT, INSERT, UPDATE ON notifications TO blue_ledger_app;
GRANT ALL PRIVILEGES ON notifications TO blue_ledger_admin;
