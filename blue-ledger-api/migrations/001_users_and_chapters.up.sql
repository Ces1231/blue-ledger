-- ============================================================
-- Migration: 001_users_and_chapters
-- Date: 2026-03-24
-- Author: Nebula
-- Purpose: Create foundational tables — chapters (tenant root),
--          users (cross-chapter identity), auth_sessions
--          (refresh token store), and magic_link_tokens.
--          Also installs the update_updated_at() trigger function
--          used by all subsequent migrations.
-- Safety: SAFE
-- Rollback: EASY
-- Locking: METADATA ONLY (new tables)
-- Est. duration: <1s
-- ============================================================

-- Enable pgcrypto for gen_random_uuid() on PG < 13.
-- On PG 13+ this is a no-op; safe to run regardless.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================
-- Trigger function: update_updated_at
-- Called by triggers on every table that carries updated_at.
-- Defined once here; referenced in all subsequent migrations.
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- Table: chapters
-- Tenant root. Each row is one Phi Beta Sigma chapter.
-- NOT tenant-scoped — visible to the platform layer only.
-- ============================================================
CREATE TABLE IF NOT EXISTS chapters (
  id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name                    TEXT        NOT NULL,                 -- "Tau Sigma Sigma Chapter"
  greek_letters           TEXT        NOT NULL,                 -- "ΤΣΣ"
  city                    TEXT        NOT NULL,
  state_code              CHAR(2)     NOT NULL,
  university              TEXT,
  district                TEXT,
  charter_date            DATE,
  member_id_prefix        TEXT        NOT NULL DEFAULT 'MBR',   -- used for display IDs
  semester_start          DATE,                                  -- current active semester start
  -- Zeffy dues form (chapter-level)
  zeffy_form_id           TEXT,                                  -- chapter's Zeffy form ID for dues
  -- Subscription / billing
  stripe_customer_id      TEXT        UNIQUE,
  stripe_subscription_id  TEXT        UNIQUE,
  subscription_status     TEXT        NOT NULL DEFAULT 'trialing'
                          CHECK (subscription_status IN ('trialing','active','past_due','canceled','paused')),
  trial_ends_at           TIMESTAMPTZ,
  plan_tier               TEXT        NOT NULL DEFAULT 'starter'
                          CHECK (plan_tier IN ('starter','growth','chapter_pro')),
  -- Settings (JSONB for flexibility without future migrations)
  settings                JSONB       NOT NULL DEFAULT '{}',
  -- Metadata
  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at              TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_chapters_stripe_customer
  ON chapters(stripe_customer_id);

CREATE INDEX IF NOT EXISTS idx_chapters_subscription_status
  ON chapters(subscription_status);

CREATE TRIGGER trg_chapters_updated_at
  BEFORE UPDATE ON chapters
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- Table: users
-- Authentication identity, separated from chapter membership.
-- One user can be a member of multiple chapters.
-- NOT tenant-scoped — no chapter_id, no RLS.
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  email               TEXT        NOT NULL UNIQUE,
  email_verified      BOOLEAN     NOT NULL DEFAULT FALSE,
  password_hash       TEXT,                                      -- NULL for magic-link-only accounts
  -- Canonical cross-chapter identity
  first_name          TEXT        NOT NULL,
  last_name           TEXT        NOT NULL,
  display_name        TEXT,
  avatar_url          TEXT,
  -- Auth metadata
  last_login_at       TIMESTAMPTZ,
  failed_login_count  INT         NOT NULL DEFAULT 0,
  locked_until        TIMESTAMPTZ,
  -- Platform-level roles
  is_sysadmin         BOOLEAN     NOT NULL DEFAULT FALSE,
  -- Timestamps
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE TRIGGER trg_users_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- Table: auth_sessions
-- Refresh token store. One row per active device session.
-- refresh_token column stores the SHA-256 hash of the raw token.
-- NOT tenant-scoped — no RLS.
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_sessions (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  chapter_id    UUID        REFERENCES chapters(id) ON DELETE CASCADE,  -- NULL for sysadmin sessions
  refresh_token TEXT        NOT NULL UNIQUE,                   -- hashed before storage
  device_hint   TEXT,                                           -- "Chrome on macOS"
  ip_address    INET,
  expires_at    TIMESTAMPTZ NOT NULL,
  revoked_at    TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user
  ON auth_sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_token
  ON auth_sessions(refresh_token);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_active
  ON auth_sessions(user_id, expires_at)
  WHERE revoked_at IS NULL;

-- ============================================================
-- Table: magic_link_tokens
-- Short-lived single-use tokens for passwordless login.
-- token_hash = SHA-256 of the raw token sent in the email link.
-- NOT tenant-scoped — no RLS.
-- ============================================================
CREATE TABLE IF NOT EXISTS magic_link_tokens (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT        NOT NULL UNIQUE,                     -- SHA-256 of the raw token
  email       TEXT        NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_magic_link_tokens_hash
  ON magic_link_tokens(token_hash);

CREATE INDEX IF NOT EXISTS idx_magic_link_tokens_user
  ON magic_link_tokens(user_id, expires_at)
  WHERE used_at IS NULL;
