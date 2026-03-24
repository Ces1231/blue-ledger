-- Migration 001: Initial Schema
-- Creates all tables for the Blue Ledger SaaS platform.
-- Tables are ordered by dependency (referenced tables first).

-- ── Extensions ───────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── chapters (tenant root — no RLS) ─────────────────────────────────────────
CREATE TABLE chapters (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                  TEXT NOT NULL,
  greek_letters         TEXT NOT NULL,
  city                  TEXT NOT NULL,
  state_code            CHAR(2) NOT NULL,
  university            TEXT,
  district              TEXT,
  charter_date          DATE,
  member_id_prefix      TEXT NOT NULL DEFAULT 'MBR',
  semester_start        DATE,
  stripe_customer_id    TEXT UNIQUE,
  stripe_subscription_id TEXT UNIQUE,
  subscription_status   TEXT NOT NULL DEFAULT 'trialing'
                        CHECK (subscription_status IN ('trialing','active','past_due','canceled','paused')),
  trial_ends_at         TIMESTAMPTZ,
  plan_tier             TEXT NOT NULL DEFAULT 'starter'
                        CHECK (plan_tier IN ('starter','growth','chapter_pro')),
  settings              JSONB NOT NULL DEFAULT '{}',
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_chapters_stripe_customer ON chapters(stripe_customer_id);
CREATE INDEX idx_chapters_subscription_status ON chapters(subscription_status);

-- ── users (cross-chapter authentication identity) ────────────────────────────
CREATE TABLE users (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email               TEXT NOT NULL UNIQUE,
  email_verified      BOOLEAN NOT NULL DEFAULT FALSE,
  password_hash       TEXT,
  first_name          TEXT NOT NULL,
  last_name           TEXT NOT NULL,
  display_name        TEXT,
  avatar_url          TEXT,
  last_login_at       TIMESTAMPTZ,
  failed_login_count  INT NOT NULL DEFAULT 0,
  locked_until        TIMESTAMPTZ,
  is_sysadmin         BOOLEAN NOT NULL DEFAULT FALSE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email);

-- ── members (tenant-scoped chapter membership) ───────────────────────────────
CREATE TABLE members (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id          UUID NOT NULL REFERENCES chapters(id),
  user_id             UUID NOT NULL REFERENCES users(id),
  member_display_id   TEXT NOT NULL,
  role                TEXT NOT NULL DEFAULT 'member'
                      CHECK (role IN ('member','chair','pia','admin','sysadmin')),
  inducted_year       INT,
  status              TEXT NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active','inactive','alumni','suspended','pledging')),
  employer            TEXT,
  job_title           TEXT,
  city                TEXT,
  linkedin_url        TEXT,
  xp_total            INT NOT NULL DEFAULT 0,
  xp_semester         INT NOT NULL DEFAULT 0,
  level               TEXT NOT NULL DEFAULT 'Neophyte',
  level_key           TEXT NOT NULL DEFAULT 'neo',
  dues_status         TEXT NOT NULL DEFAULT 'unpaid'
                      CHECK (dues_status IN ('paid','unpaid','late','waived','outstanding')),
  avatar_bg           TEXT,
  avatar_fg           TEXT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at          TIMESTAMPTZ,
  UNIQUE(chapter_id, user_id),
  UNIQUE(chapter_id, member_display_id)
);

CREATE INDEX idx_members_chapter ON members(chapter_id);
CREATE INDEX idx_members_user ON members(user_id);
CREATE INDEX idx_members_chapter_user ON members(chapter_id, user_id);
CREATE INDEX idx_members_role ON members(chapter_id, role);
CREATE INDEX idx_members_xp ON members(chapter_id, xp_total DESC);

-- ── auth_sessions (refresh token store) ─────────────────────────────────────
CREATE TABLE auth_sessions (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  chapter_id    UUID REFERENCES chapters(id),
  refresh_token TEXT NOT NULL UNIQUE,
  device_hint   TEXT,
  ip_address    INET,
  expires_at    TIMESTAMPTZ NOT NULL,
  revoked_at    TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_sessions_user ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_token ON auth_sessions(refresh_token);

-- ── magic_link_tokens ────────────────────────────────────────────────────────
CREATE TABLE magic_link_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL UNIQUE,
  email       TEXT NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_magic_link_tokens_hash ON magic_link_tokens(token_hash);

-- ── events ───────────────────────────────────────────────────────────────────
CREATE TABLE events (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id    UUID NOT NULL REFERENCES chapters(id),
  name          TEXT NOT NULL,
  description   TEXT,
  event_date    DATE NOT NULL,
  event_time    TEXT,
  location      TEXT,
  event_type    TEXT NOT NULL CHECK (event_type IN ('Meeting','Service','Social','Conference','Committee','SBC','Other')),
  xp_attend     INT NOT NULL DEFAULT 50,
  xp_rsvp       INT NOT NULL DEFAULT 10,
  rsvp_deadline DATE,
  capacity      INT,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  qr_code_token TEXT UNIQUE,
  created_by    UUID REFERENCES members(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_events_chapter ON events(chapter_id);
CREATE INDEX idx_events_chapter_date ON events(chapter_id, event_date DESC);
CREATE INDEX idx_events_qr_token ON events(qr_code_token);

-- ── rsvps ────────────────────────────────────────────────────────────────────
CREATE TABLE rsvps (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  event_id        UUID NOT NULL REFERENCES events(id),
  member_id       UUID NOT NULL REFERENCES members(id),
  status          TEXT NOT NULL CHECK (status IN ('yes','no','maybe')),
  attended        BOOLEAN NOT NULL DEFAULT FALSE,
  no_show         BOOLEAN NOT NULL DEFAULT FALSE,
  checked_in_at   TIMESTAMPTZ,
  checked_in_by   UUID REFERENCES members(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(event_id, member_id)
);

CREATE INDEX idx_rsvps_chapter_event ON rsvps(chapter_id, event_id);
CREATE INDEX idx_rsvps_member ON rsvps(member_id);

-- ── engagement_log (append-only XP audit trail) ──────────────────────────────
CREATE TABLE engagement_log (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  member_id       UUID NOT NULL REFERENCES members(id),
  activity        TEXT NOT NULL,
  xp_awarded      INT NOT NULL,
  source          TEXT NOT NULL CHECK (source IN ('checkin','rsvp','admin','system','quiz','service','dues','badge','props','mentorship')),
  reference_id    UUID,
  reference_type  TEXT,
  awarded_by      UUID REFERENCES members(id),
  note            TEXT,
  semester        TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_eng_log_chapter ON engagement_log(chapter_id);
CREATE INDEX idx_eng_log_member ON engagement_log(chapter_id, member_id);
CREATE INDEX idx_eng_log_created ON engagement_log(chapter_id, created_at DESC);
CREATE INDEX idx_eng_log_source ON engagement_log(chapter_id, source);

-- ── service_log ──────────────────────────────────────────────────────────────
CREATE TABLE service_log (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  member_id       UUID NOT NULL REFERENCES members(id),
  event_name      TEXT NOT NULL,
  organization    TEXT,
  service_date    DATE NOT NULL,
  hours           NUMERIC(5,2) NOT NULL CHECK (hours > 0),
  xp_awarded      INT NOT NULL DEFAULT 0,
  verified        BOOLEAN NOT NULL DEFAULT FALSE,
  verified_by     UUID REFERENCES members(id),
  verified_at     TIMESTAMPTZ,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_service_log_chapter ON service_log(chapter_id);
CREATE INDEX idx_service_log_member ON service_log(chapter_id, member_id);

-- ── dues_records ─────────────────────────────────────────────────────────────
CREATE TABLE dues_records (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id            UUID NOT NULL REFERENCES chapters(id),
  member_id             UUID NOT NULL REFERENCES members(id),
  semester              TEXT NOT NULL,
  amount_cents          INT NOT NULL,
  due_date              DATE NOT NULL,
  paid_at               TIMESTAMPTZ,
  payment_method        TEXT,
  zeffy_form_id         TEXT,
  zeffy_transaction_id  TEXT,
  status                TEXT NOT NULL DEFAULT 'unpaid'
                        CHECK (status IN ('unpaid','paid','late','waived','outstanding')),
  xp_awarded            BOOLEAN NOT NULL DEFAULT FALSE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(chapter_id, member_id, semester)
);

CREATE INDEX idx_dues_chapter ON dues_records(chapter_id);
CREATE INDEX idx_dues_member ON dues_records(chapter_id, member_id);
CREATE INDEX idx_dues_semester ON dues_records(chapter_id, semester);

-- ── badges ───────────────────────────────────────────────────────────────────
CREATE TABLE badges (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  name            TEXT NOT NULL,
  icon            TEXT,
  category        TEXT,
  description     TEXT,
  requirement     TEXT,
  xp_reward       INT NOT NULL DEFAULT 0,
  rarity          TEXT NOT NULL DEFAULT 'common'
                  CHECK (rarity IN ('common','uncommon','rare','legendary')),
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  criteria        JSONB NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(chapter_id, name)
);

CREATE INDEX idx_badges_chapter ON badges(chapter_id);

-- ── member_badges ────────────────────────────────────────────────────────────
CREATE TABLE member_badges (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  member_id   UUID NOT NULL REFERENCES members(id),
  badge_id    UUID NOT NULL REFERENCES badges(id),
  awarded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  awarded_by  UUID REFERENCES members(id),
  UNIQUE(member_id, badge_id)
);

CREATE INDEX idx_member_badges_member ON member_badges(member_id);
CREATE INDEX idx_member_badges_chapter ON member_badges(chapter_id);

-- ── announcements ────────────────────────────────────────────────────────────
CREATE TABLE announcements (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  title       TEXT NOT NULL,
  body        TEXT NOT NULL,
  category    TEXT,
  posted_by   UUID NOT NULL REFERENCES members(id),
  is_pinned   BOOLEAN NOT NULL DEFAULT FALSE,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_announcements_chapter ON announcements(chapter_id);
CREATE INDEX idx_announcements_pinned ON announcements(chapter_id, is_pinned, created_at DESC);

-- ── props ────────────────────────────────────────────────────────────────────
CREATE TABLE props (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  from_id     UUID NOT NULL REFERENCES members(id),
  to_id       UUID NOT NULL REFERENCES members(id),
  category    TEXT NOT NULL CHECK (category IN ('Leadership','Brotherhood','Service','Academic','Professionalism','Other')),
  message     TEXT,
  xp_awarded  INT NOT NULL DEFAULT 10,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_props_chapter ON props(chapter_id, created_at DESC);
CREATE INDEX idx_props_to_member ON props(to_id);

-- ── quests ───────────────────────────────────────────────────────────────────
CREATE TABLE quests (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  title           TEXT NOT NULL,
  description     TEXT,
  xp_reward       INT NOT NULL DEFAULT 0,
  badge_reward_id UUID REFERENCES badges(id),
  steps           JSONB NOT NULL DEFAULT '[]',
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE member_quest_progress (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  member_id       UUID NOT NULL REFERENCES members(id),
  quest_id        UUID NOT NULL REFERENCES quests(id),
  progress        JSONB NOT NULL DEFAULT '{}',
  completed_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(member_id, quest_id)
);

-- ── notifications ────────────────────────────────────────────────────────────
CREATE TABLE notifications (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  user_id     UUID NOT NULL REFERENCES users(id),
  type        TEXT NOT NULL CHECK (type IN ('badge','prop','xp','dues','event','announcement','level','system')),
  title       TEXT NOT NULL,
  body        TEXT NOT NULL,
  link        TEXT,
  is_read     BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);

-- ── mentorships ──────────────────────────────────────────────────────────────
CREATE TABLE mentorships (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  mentor_id   UUID NOT NULL REFERENCES members(id),
  mentee_id   UUID REFERENCES members(id),
  focus_areas TEXT[],
  bio         TEXT,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  started_at  TIMESTAMPTZ,
  ended_at    TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── intake_prospects ─────────────────────────────────────────────────────────
CREATE TABLE intake_prospects (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  name            TEXT NOT NULL,
  email           TEXT,
  gpa             TEXT,
  academic_year   TEXT,
  referred_by     UUID REFERENCES members(id),
  stage           TEXT NOT NULL DEFAULT 'inquiry'
                  CHECK (stage IN ('inquiry','under_review','interview_scheduled','invited_to_rush','accepted','declined','inactive')),
  interest_level  TEXT CHECK (interest_level IN ('low','medium','high','very_high')),
  notes           TEXT,
  intake_date     DATE NOT NULL DEFAULT CURRENT_DATE,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_intake_chapter ON intake_prospects(chapter_id);
CREATE INDEX idx_intake_stage ON intake_prospects(chapter_id, stage);

-- ── votes ────────────────────────────────────────────────────────────────────
CREATE TABLE votes (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id    UUID NOT NULL REFERENCES chapters(id),
  title         TEXT NOT NULL,
  description   TEXT,
  vote_type     TEXT NOT NULL CHECK (vote_type IN ('election','referendum','motion')),
  options       TEXT[] NOT NULL,
  deadline      DATE,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  is_anonymous  BOOLEAN NOT NULL DEFAULT TRUE,
  created_by    UUID NOT NULL REFERENCES members(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vote_responses (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id    UUID NOT NULL REFERENCES chapters(id),
  vote_id       UUID NOT NULL REFERENCES votes(id),
  member_id     UUID NOT NULL REFERENCES members(id),
  option_index  INT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(vote_id, member_id)
);

-- ── chapter_minutes ──────────────────────────────────────────────────────────
CREATE TABLE chapter_minutes (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id          UUID NOT NULL REFERENCES chapters(id),
  meeting_date        DATE NOT NULL,
  title               TEXT NOT NULL,
  body                TEXT NOT NULL,
  recorder_id         UUID REFERENCES members(id),
  quorum              BOOLEAN NOT NULL DEFAULT FALSE,
  attendee_ids        UUID[],
  xp_for_attendance   INT NOT NULL DEFAULT 0,
  status              TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','final')),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── scholarships ─────────────────────────────────────────────────────────────
CREATE TABLE scholarships (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id            UUID NOT NULL REFERENCES chapters(id),
  applicant_name        TEXT NOT NULL,
  school                TEXT,
  gpa                   TEXT,
  essay                 TEXT,
  academic_year         INT,
  city                  TEXT,
  submitted_at          DATE,
  status                TEXT NOT NULL DEFAULT 'submitted'
                        CHECK (status IN ('submitted','under_review','finalist','awarded','declined','needs_more_info')),
  reviewer_id           UUID REFERENCES members(id),
  notes                 TEXT,
  award_amount_cents    INT,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── fundraising_campaigns ────────────────────────────────────────────────────
CREATE TABLE fundraising_campaigns (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id    UUID NOT NULL REFERENCES chapters(id),
  name          TEXT NOT NULL,
  description   TEXT,
  goal_cents    INT NOT NULL,
  raised_cents  INT NOT NULL DEFAULT 0,
  category      TEXT,
  deadline      DATE,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── store_items ──────────────────────────────────────────────────────────────
CREATE TABLE store_items (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  name        TEXT NOT NULL,
  description TEXT,
  xp_cost     INT NOT NULL,
  category    TEXT,
  stock       INT NOT NULL DEFAULT 0,
  icon        TEXT,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE store_orders (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  member_id   UUID NOT NULL REFERENCES members(id),
  item_id     UUID NOT NULL REFERENCES store_items(id),
  xp_spent    INT NOT NULL,
  status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','fulfilled','cancelled')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── chapter_goals ────────────────────────────────────────────────────────────
CREATE TABLE chapter_goals (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  title       TEXT NOT NULL,
  description TEXT,
  category    TEXT,
  target      NUMERIC(10,2) NOT NULL,
  current     NUMERIC(10,2) NOT NULL DEFAULT 0,
  unit        TEXT,
  deadline    DATE,
  owner_id    UUID REFERENCES members(id),
  status      TEXT NOT NULL DEFAULT 'in_progress'
              CHECK (status IN ('in_progress','achieved','missed','paused')),
  xp_reward   INT NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── job_board ────────────────────────────────────────────────────────────────
CREATE TABLE job_board (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  posted_by   UUID NOT NULL REFERENCES members(id),
  company     TEXT NOT NULL,
  job_title   TEXT NOT NULL,
  job_type    TEXT CHECK (job_type IN ('Internship','Full-Time','Part-Time','Contract','Fellowship')),
  location    TEXT,
  link        TEXT,
  description TEXT,
  fields      TEXT[],
  deadline    DATE,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── point_economy ────────────────────────────────────────────────────────────
CREATE TABLE point_economy (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  activity    TEXT NOT NULL,
  xp          INT NOT NULL,
  category    TEXT,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE(chapter_id, activity)
);

-- ── audit_log (platform-wide) ────────────────────────────────────────────────
CREATE TABLE audit_log (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id    UUID,
  actor_type  TEXT NOT NULL CHECK (actor_type IN ('user','system','webhook')),
  chapter_id  UUID,
  action      TEXT NOT NULL,
  target_type TEXT,
  target_id   UUID,
  metadata    JSONB DEFAULT '{}',
  ip_address  INET,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_actor ON audit_log(actor_id);
CREATE INDEX idx_audit_log_chapter ON audit_log(chapter_id);
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);
