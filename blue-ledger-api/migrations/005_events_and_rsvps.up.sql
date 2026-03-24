-- ============================================================
-- Migration: 005_events_and_rsvps
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create events and rsvps tables. Events support QR
--          check-in via a signed token. RSVPs track attendance,
--          no-shows, and who performed the check-in scan.
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- ============================================================
-- Table: events
-- ============================================================
CREATE TABLE IF NOT EXISTS events (
  id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id              UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  name                    TEXT        NOT NULL,
  description             TEXT,
  event_date              DATE        NOT NULL,
  event_time              TEXT,                                  -- "7:00 PM" — flexible string
  location                TEXT,
  event_type              TEXT        NOT NULL
                          CHECK (event_type IN (
                            'Meeting','Service','Social','Conference',
                            'Committee','SBC','Other'
                          )),
  xp_attend               INT         NOT NULL DEFAULT 50,       -- XP for attending
  xp_rsvp                 INT         NOT NULL DEFAULT 10,       -- XP for RSVPing yes in advance
  rsvp_deadline           DATE,
  capacity                INT,
  is_active               BOOLEAN     NOT NULL DEFAULT TRUE,
  qr_code_token           TEXT        UNIQUE,                    -- signed token for QR scanner check-in
  check_in_window_minutes INT         NOT NULL DEFAULT 120,      -- how long after start QR is valid

  -- Who created this event (logical ref — members table must exist)
  created_by              UUID        REFERENCES members(id) ON DELETE SET NULL,

  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at              TIMESTAMPTZ
);

-- ── RLS ──────────────────────────────────────────────────────
ALTER TABLE events ENABLE ROW LEVEL SECURITY;

CREATE POLICY chapter_isolation ON events
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

-- ── Indexes ──────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_events_chapter
  ON events(chapter_id);

CREATE INDEX IF NOT EXISTS idx_events_chapter_date
  ON events(chapter_id, event_date DESC);

CREATE INDEX IF NOT EXISTS idx_events_qr_token
  ON events(qr_code_token)
  WHERE qr_code_token IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_events_active
  ON events(chapter_id, is_active, event_date DESC)
  WHERE deleted_at IS NULL;

-- ── updated_at trigger ───────────────────────────────────────
CREATE TRIGGER trg_events_updated_at
  BEFORE UPDATE ON events
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ── Grants ───────────────────────────────────────────────────
GRANT SELECT, INSERT, UPDATE, DELETE ON events TO blue_ledger_app;
GRANT ALL PRIVILEGES ON events TO blue_ledger_admin;

-- ============================================================
-- Table: rsvps
-- One row per (event, member) pair. UNIQUE enforced at DB level.
-- ============================================================
CREATE TABLE IF NOT EXISTS rsvps (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  event_id        UUID        NOT NULL REFERENCES events(id)   ON DELETE CASCADE,
  member_id       UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,

  status          TEXT        NOT NULL
                  CHECK (status IN ('yes','no','maybe')),

  attended        BOOLEAN     NOT NULL DEFAULT FALSE,
  no_show         BOOLEAN     NOT NULL DEFAULT FALSE,

  -- Check-in tracking
  checked_in_at   TIMESTAMPTZ,
  checked_in_by   UUID        REFERENCES members(id) ON DELETE SET NULL,  -- scanner/chair

  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(event_id, member_id)
);

-- ── RLS ──────────────────────────────────────────────────────
ALTER TABLE rsvps ENABLE ROW LEVEL SECURITY;

CREATE POLICY chapter_isolation ON rsvps
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

-- ── Indexes ──────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_rsvps_chapter_event
  ON rsvps(chapter_id, event_id);

CREATE INDEX IF NOT EXISTS idx_rsvps_member
  ON rsvps(member_id);

CREATE INDEX IF NOT EXISTS idx_rsvps_event_status
  ON rsvps(event_id, status);

-- ── updated_at trigger ───────────────────────────────────────
CREATE TRIGGER trg_rsvps_updated_at
  BEFORE UPDATE ON rsvps
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ── Grants ───────────────────────────────────────────────────
GRANT SELECT, INSERT, UPDATE, DELETE ON rsvps TO blue_ledger_app;
GRANT ALL PRIVILEGES ON rsvps TO blue_ledger_admin;
