# The Blue Ledger — SaaS Architecture & Implementation Spec

## Meta
| Field | Value |
|-------|-------|
| Document Type | SaaS Architecture Spec |
| Project | The Blue Ledger — Chapter Engagement Platform |
| Organization | Phi Beta Sigma Fraternity |
| Spec Author | J.A.R.V.I.S. |
| Created | 2026-03-24 |
| Status | Draft — Ready for Review |
| project_stage | pre-production |
| Estimated Build | 280–340 hrs total across 6 phases |

---

## 1. Executive Summary

The Blue Ledger is a multi-tenant SaaS platform where each Phi Beta Sigma chapter
is a fully isolated tenant. The prototype is a single-file static SPA (~4,400 lines)
with in-memory data and client-side role enforcement. This spec defines the complete
architecture to transform it into a production SaaS product with:

- Real authentication (email/password + magic link)
- PostgreSQL persistence with row-level security per chapter
- Go REST API backend
- React frontend wired to the API (existing UI design preserved)
- Zeffy for member dues collection (0% fees, embeddable, nonprofit-eligible)
- Stripe Billing for platform subscription (chapters pay monthly SaaS fee)
- Sysadmin super-console for platform management

### Non-Goals for v1
- Mobile native app (the web app is mobile-responsive)
- Single sign-on / SAML integration (Year 2)
- Analytics dashboard for Phi Beta Sigma national HQ (Year 2)
- White-labeling for other Greek organizations (Year 2)

---

## 2. Tech Stack Recommendation

### Backend
| Concern | Choice | Rationale |
|---------|--------|-----------|
| Language | **Go 1.23** | Compile to single binary, low memory, excellent stdlib HTTP support, strong concurrency for future real-time features |
| Framework | **Echo v4** | Lightweight, first-class middleware, clean route grouping, built-in request binding/validation |
| ORM | **sqlc** (generated) + **pgx/v5** | Type-safe, no reflection overhead, queries verified at compile time |
| Migrations | **golang-migrate** | Widely used, supports up/down, integrates with CI |
| Auth tokens | **JWT (RS256)** + **refresh tokens** stored in `auth_sessions` table | RS256 allows public key verification without DB lookup on every request |
| Magic links | Short-lived HMAC-signed tokens stored in `magic_link_tokens` table | No third-party dependency |
| Password hashing | **bcrypt** (cost 12) | Industry standard |
| Validation | echo built-in + `go-playground/validator` | Tag-based, readable struct validation |

### Frontend
| Concern | Choice | Rationale |
|---------|--------|-----------|
| Framework | **React 18 + TypeScript** | The existing SPA logic maps cleanly to React components; TypeScript catches API contract drift early |
| Build | **Vite 5** | Fast HMR, smaller bundles than CRA |
| Routing | **React Router v6** | Client-side routing matching existing 47 screens |
| API client | **TanStack Query v5** + **Axios** | Request deduplication, caching, optimistic updates for XP/leaderboard |
| State | TanStack Query (server state) + **Zustand** (UI state: current user, active view) | Avoids Redux overhead for a domain this size |
| Styling | **Tailwind CSS v3** + existing CSS variables preserved | The existing design tokens (navy, gold, cream) map directly to Tailwind config |
| Forms | **React Hook Form** | Low re-render overhead, good for the many admin forms |

### Database
| Concern | Choice | Rationale |
|---------|--------|-----------|
| Primary DB | **PostgreSQL 16** | Row-level security, JSONB for flexible data (settings, badge criteria), excellent full-text search |
| Multi-tenancy | **Row-level security (RLS) + `chapter_id` FK on every tenant table** | Single schema, chapter data isolated via Postgres policies, simplest operational model for v1 |
| Caching | **Redis 7** | Session management, rate limiting, notification queues, future real-time leaderboard |
| Full-text search | PostgreSQL `tsvector` (no separate search service needed at this scale) | |

### Infrastructure
| Concern | Choice | Rationale |
|---------|--------|-----------|
| Hosting | **Fly.io** | Simple deployments, per-region Postgres, free TLS, scales to zero, no DevOps overhead |
| Database hosting | **Fly.io Postgres** (managed) or **Supabase** (if RLS tooling is preferred) | Either works; Fly.io keeps it in one platform |
| File storage | **Cloudflare R2** | S3-compatible, free egress, used for avatars and resource file uploads |
| Email | **Resend** | Modern API, excellent deliverability, generous free tier (3k/mo), magic link emails |
| Member Dues | **Zeffy** | 0% fees, embeddable iframe/popup forms, recurring memberships — chapters use this to collect dues from brothers; 501(c) nonprofits qualify |
| Platform Billing | **Stripe Billing** | Full API + webhooks for charging chapters their monthly SaaS subscription fee; manages trials, upgrades, cancellations |
| CDN | **Cloudflare** (included with R2) | |
| Monitoring | **Sentry** (errors) + **Fly.io built-in metrics** | |

### Local Development
```
docker-compose.yml: postgres, redis
Go API: go run ./cmd/server
React: npm run dev (Vite on :3001)
API: :8080
```

---

## 3. Multi-Tenancy Model

### Design: Row-Level Security on a shared schema

Every tenant table carries a `chapter_id UUID NOT NULL` foreign key that references
`chapters.id`. PostgreSQL Row-Level Security policies enforce that API connections
scoped to a chapter can only read/write their own rows.

**Why not separate schemas per tenant?**
- Separate schemas would require 1 schema per chapter, making migrations complex
  (run migration N times per tenant). With ~20-200 chapters in Year 1, RLS on a
  shared schema is far simpler to operate and migrate.
- RLS with `chapter_id` is the industry standard for B2B SaaS of this scale
  (Notion, Linear, and thousands of others use this pattern).

### How it works in practice

1. On login, the API issues a JWT containing `{ user_id, chapter_id, role }`.
2. Every API request sets a Postgres session variable: `SET LOCAL app.chapter_id = '<uuid>'`.
3. RLS policies check `chapter_id = current_setting('app.chapter_id')::uuid`.
4. The application layer also enforces `chapter_id` in every WHERE clause as a defense-in-depth measure.
5. The `sysadmin` role bypasses RLS entirely (uses a separate service account connection
   with `SET ROLE blue_ledger_admin`).

### RLS Policy Pattern (applied to all tenant tables)
```sql
-- Enable RLS
ALTER TABLE members ENABLE ROW LEVEL SECURITY;

-- Policy: tenants see only their rows
CREATE POLICY chapter_isolation ON members
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

-- Policy: sysadmin bypasses (sysadmin uses a role with BYPASSRLS)
-- No policy needed — sysadmin role has BYPASSRLS privilege
```

### Tenant Hierarchy
```
Organization (Phi Beta Sigma — implicit, single org for now)
  └── Chapter (e.g., Tau Tau Sigma)
        ├── members
        ├── events
        ├── engagement_log
        ├── service_log
        ├── dues_records
        └── ... (all tenant tables)
```

---

## 4. Database Schema

### Conventions
- All PKs: `UUID` (generated with `gen_random_uuid()`)
- All timestamps: `TIMESTAMPTZ` (UTC stored, displayed in client timezone)
- Soft deletes: `deleted_at TIMESTAMPTZ NULL` on entities that need audit trail
- `chapter_id` is indexed on every tenant table
- All migration files: `{NNN}_{description}.up.sql` and `{NNN}_{description}.down.sql`

---

### Table: `chapters` (tenant root)
```sql
CREATE TABLE chapters (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name              TEXT NOT NULL,                          -- "Tau Tau Sigma Chapter"
  greek_letters     TEXT NOT NULL,                          -- "ΤΤΣ"
  city              TEXT NOT NULL,
  state_code        CHAR(2) NOT NULL,
  university        TEXT,
  district          TEXT,
  charter_date      DATE,
  member_id_prefix  TEXT NOT NULL DEFAULT 'MBR',            -- used for display IDs
  semester_start    DATE,                                   -- current active semester start
  -- Subscription / billing
  stripe_customer_id    TEXT UNIQUE,
  stripe_subscription_id TEXT UNIQUE,
  subscription_status   TEXT NOT NULL DEFAULT 'trialing'   -- trialing | active | past_due | canceled
                        CHECK (subscription_status IN ('trialing','active','past_due','canceled','paused')),
  trial_ends_at         TIMESTAMPTZ,
  plan_tier             TEXT NOT NULL DEFAULT 'starter'    -- starter | growth | chapter_pro
                        CHECK (plan_tier IN ('starter','growth','chapter_pro')),
  -- Settings (JSONB for flexibility without migrations)
  settings          JSONB NOT NULL DEFAULT '{}',
  -- Metadata
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_chapters_stripe_customer ON chapters(stripe_customer_id);
CREATE INDEX idx_chapters_subscription_status ON chapters(subscription_status);
```

---

### Table: `users` (authentication identity — separated from chapter membership)
```sql
-- One user can be a member of multiple chapters (a brother who transfers)
CREATE TABLE users (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email             TEXT NOT NULL UNIQUE,
  email_verified    BOOLEAN NOT NULL DEFAULT FALSE,
  password_hash     TEXT,                                   -- NULL for magic-link-only accounts
  -- Profile (canonical cross-chapter identity)
  first_name        TEXT NOT NULL,
  last_name         TEXT NOT NULL,
  display_name      TEXT,
  avatar_url        TEXT,
  -- Auth metadata
  last_login_at     TIMESTAMPTZ,
  failed_login_count INT NOT NULL DEFAULT 0,
  locked_until      TIMESTAMPTZ,
  -- Platform-level roles
  is_sysadmin       BOOLEAN NOT NULL DEFAULT FALSE,
  -- Timestamps
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email);
```

---

### Table: `members` (chapter membership — a user's presence in a specific chapter)
```sql
-- This is the tenant-scoped profile.
-- One user can have one membership row per chapter.
CREATE TABLE members (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id          UUID NOT NULL REFERENCES chapters(id),
  user_id             UUID NOT NULL REFERENCES users(id),
  -- Chapter-scoped display ID (e.g., ΤΤΣ-001)
  member_display_id   TEXT NOT NULL,
  -- Chapter-specific profile fields
  role                TEXT NOT NULL DEFAULT 'member'
                      CHECK (role IN ('member','chair','pia','admin','sysadmin')),
  inducted_year       INT,
  status              TEXT NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active','inactive','alumni','suspended','pledging')),
  -- Professional profile
  employer            TEXT,
  job_title           TEXT,
  city                TEXT,
  linkedin_url        TEXT,
  -- XP & gamification
  xp_total            INT NOT NULL DEFAULT 0,
  xp_semester         INT NOT NULL DEFAULT 0,              -- reset each semester
  level               TEXT NOT NULL DEFAULT 'Neophyte',
  level_key           TEXT NOT NULL DEFAULT 'neo',
  -- Dues
  dues_status         TEXT NOT NULL DEFAULT 'unpaid'
                      CHECK (dues_status IN ('paid','unpaid','late','waived','outstanding')),
  -- Avatar display preferences
  avatar_bg           TEXT,
  avatar_fg           TEXT,
  -- Timestamps
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at          TIMESTAMPTZ,
  UNIQUE(chapter_id, user_id),
  UNIQUE(chapter_id, member_display_id)
);

ALTER TABLE members ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON members
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_members_chapter ON members(chapter_id);
CREATE INDEX idx_members_user ON members(user_id);
CREATE INDEX idx_members_chapter_user ON members(chapter_id, user_id);
CREATE INDEX idx_members_role ON members(chapter_id, role);
CREATE INDEX idx_members_xp ON members(chapter_id, xp_total DESC);
```

---

### Table: `auth_sessions` (refresh token store)
```sql
CREATE TABLE auth_sessions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  chapter_id      UUID REFERENCES chapters(id),            -- NULL for sysadmin sessions
  refresh_token   TEXT NOT NULL UNIQUE,                    -- hashed before storage
  device_hint     TEXT,                                    -- "Chrome on macOS"
  ip_address      INET,
  expires_at      TIMESTAMPTZ NOT NULL,
  revoked_at      TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_sessions_user ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_token ON auth_sessions(refresh_token);
```

---

### Table: `magic_link_tokens`
```sql
CREATE TABLE magic_link_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL UNIQUE,                        -- SHA-256 of the raw token
  email       TEXT NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_magic_link_tokens_hash ON magic_link_tokens(token_hash);
```

---

### Table: `events`
```sql
CREATE TABLE events (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id    UUID NOT NULL REFERENCES chapters(id),
  name          TEXT NOT NULL,
  description   TEXT,
  event_date    DATE NOT NULL,
  event_time    TEXT,                                      -- "7:00 PM" — stored as string for flexibility
  location      TEXT,
  event_type    TEXT NOT NULL CHECK (event_type IN ('Meeting','Service','Social','Conference','Committee','SBC','Other')),
  xp_attend     INT NOT NULL DEFAULT 50,
  xp_rsvp       INT NOT NULL DEFAULT 10,
  rsvp_deadline DATE,
  capacity      INT,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  qr_code_token TEXT UNIQUE,                               -- signed token for QR scanner check-in
  created_by    UUID REFERENCES members(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at    TIMESTAMPTZ
);

ALTER TABLE events ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON events
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_events_chapter ON events(chapter_id);
CREATE INDEX idx_events_chapter_date ON events(chapter_id, event_date DESC);
CREATE INDEX idx_events_qr_token ON events(qr_code_token);
```

---

### Table: `rsvps`
```sql
CREATE TABLE rsvps (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  event_id    UUID NOT NULL REFERENCES events(id),
  member_id   UUID NOT NULL REFERENCES members(id),
  status      TEXT NOT NULL CHECK (status IN ('yes','no','maybe')),
  attended    BOOLEAN NOT NULL DEFAULT FALSE,
  no_show     BOOLEAN NOT NULL DEFAULT FALSE,
  checked_in_at TIMESTAMPTZ,
  checked_in_by UUID REFERENCES members(id),              -- the scanner/chair who checked them in
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(event_id, member_id)
);

ALTER TABLE rsvps ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON rsvps
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_rsvps_chapter_event ON rsvps(chapter_id, event_id);
CREATE INDEX idx_rsvps_member ON rsvps(member_id);
```

---

### Table: `engagement_log` (immutable XP audit trail)
```sql
-- Every XP award creates an immutable row. XP totals are derived sums.
-- This is the source of truth for XP — member.xp_total is a denormalized cache.
CREATE TABLE engagement_log (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  member_id       UUID NOT NULL REFERENCES members(id),
  activity        TEXT NOT NULL,                           -- "Chapter Meeting (on time)"
  xp_awarded      INT NOT NULL,
  source          TEXT NOT NULL CHECK (source IN ('checkin','rsvp','admin','system','quiz','service','dues','badge','props','mentorship')),
  reference_id    UUID,                                    -- event_id, service_log_id, etc.
  reference_type  TEXT,                                    -- 'event', 'service_log', 'dues_record', etc.
  awarded_by      UUID REFERENCES members(id),             -- NULL for system-generated
  note            TEXT,
  semester        TEXT,                                    -- "Fall 2025" — for semester XP grouping
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- NO updated_at — this is append-only
);

ALTER TABLE engagement_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON engagement_log
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_eng_log_chapter ON engagement_log(chapter_id);
CREATE INDEX idx_eng_log_member ON engagement_log(chapter_id, member_id);
CREATE INDEX idx_eng_log_created ON engagement_log(chapter_id, created_at DESC);
CREATE INDEX idx_eng_log_source ON engagement_log(chapter_id, source);
```

---

### Table: `service_log`
```sql
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

ALTER TABLE service_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON service_log
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_service_log_chapter ON service_log(chapter_id);
CREATE INDEX idx_service_log_member ON service_log(chapter_id, member_id);
```

---

### Table: `dues_records`
```sql
CREATE TABLE dues_records (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id          UUID NOT NULL REFERENCES chapters(id),
  member_id           UUID NOT NULL REFERENCES members(id),
  semester            TEXT NOT NULL,                       -- "Fall 2025"
  amount_cents        INT NOT NULL,                        -- stored in cents
  due_date            DATE NOT NULL,
  paid_at             TIMESTAMPTZ,
  payment_method      TEXT,                                -- 'zeffy', 'cash', 'zelle', 'waived', 'admin_override'
  zeffy_form_id       TEXT,                                -- Zeffy form ID for this chapter's dues form
  zeffy_transaction_id TEXT,                               -- Zeffy transaction reference (via webhook/Zapier)
  status              TEXT NOT NULL DEFAULT 'unpaid'
                      CHECK (status IN ('unpaid','paid','late','waived','outstanding')),
  xp_awarded          BOOLEAN NOT NULL DEFAULT FALSE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(chapter_id, member_id, semester)
);

ALTER TABLE dues_records ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON dues_records
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_dues_chapter ON dues_records(chapter_id);
CREATE INDEX idx_dues_member ON dues_records(chapter_id, member_id);
CREATE INDEX idx_dues_semester ON dues_records(chapter_id, semester);
```

---

### Payment Architecture: Dual-Layer Model

```
┌─────────────────────────────────────────────────────────────────┐
│                    PAYMENT FLOWS                                 │
│                                                                  │
│  Brothers → Chapter (DUES)          Platform → Chapter (SAAS)  │
│  ─────────────────────────          ──────────────────────────  │
│  Tool:    Zeffy                     Tool:  Stripe Billing       │
│  Fee:     0%                        Fee:   2.9% + 30¢          │
│  API:     Embed + Zapier webhook    API:   Full REST + webhooks │
│  Flow:    Brother clicks            Flow:  Admin card on file   │
│           embedded Zeffy form               auto-charged monthly│
│           → Zeffy processes         → Stripe fires webhook      │
│           → Zapier fires webhook    → API marks chapter active  │
│           → API marks dues paid     → Access revoked if lapsed  │
│           → XP awarded              │                           │
└─────────────────────────────────────────────────────────────────┘
```

#### Zeffy — Member Dues Collection
- Each chapter admin pastes their Zeffy form URL into System Console → App Config
- Blue Ledger stores the `zeffy_form_id` on the `chapters` record
- The member-facing "Pay Dues" button opens an embedded Zeffy popup
- Dues confirmation is received via **Zapier webhook** → Blue Ledger API endpoint `POST /webhooks/zeffy`
- The webhook handler marks `dues_records.status = 'paid'` and awards XP
- No Zeffy API key required — webhook payload carries `member_email` + `amount` + `transaction_id`

#### Stripe — Platform Subscription Billing
- Chapter admins enter a card when signing up for Blue Ledger
- Stripe Billing handles monthly charge, retries, and dunning
- Webhook `customer.subscription.updated` → `chapters.subscription_status`
- Access gates: if `subscription_status = 'past_due'` for > 7 days, chapter enters read-only mode

#### New API Endpoints (payments)
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/webhooks/zeffy` | Zeffy → Zapier → this endpoint: mark dues paid, award XP |
| `POST` | `/webhooks/stripe` | Stripe subscription events: activate, suspend, cancel |
| `GET`  | `/chapters/:id/billing` | Current subscription status + next billing date |
| `POST` | `/chapters/:id/zeffy-config` | Admin saves Zeffy form ID |

---

### Table: `badges`
```sql
-- Badge definitions are chapter-level (chapters can customize, but seeded from defaults)
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
  -- Criteria stored as JSONB for flexible badge engine
  criteria        JSONB NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(chapter_id, name)
);

ALTER TABLE badges ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON badges
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_badges_chapter ON badges(chapter_id);
```

---

### Table: `member_badges`
```sql
CREATE TABLE member_badges (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  member_id   UUID NOT NULL REFERENCES members(id),
  badge_id    UUID NOT NULL REFERENCES badges(id),
  awarded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  awarded_by  UUID REFERENCES members(id),                -- NULL for system auto-award
  UNIQUE(member_id, badge_id)
);

ALTER TABLE member_badges ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON member_badges
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_member_badges_member ON member_badges(member_id);
CREATE INDEX idx_member_badges_chapter ON member_badges(chapter_id);
```

---

### Table: `announcements`
```sql
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

ALTER TABLE announcements ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON announcements
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_announcements_chapter ON announcements(chapter_id);
CREATE INDEX idx_announcements_pinned ON announcements(chapter_id, is_pinned, created_at DESC);
```

---

### Table: `props`
```sql
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

ALTER TABLE props ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON props
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_props_chapter ON props(chapter_id, created_at DESC);
CREATE INDEX idx_props_to_member ON props(to_id);
```

---

### Table: `quests`
```sql
-- Quest definitions per chapter (seeded from defaults)
CREATE TABLE quests (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  title           TEXT NOT NULL,
  description     TEXT,
  xp_reward       INT NOT NULL DEFAULT 0,
  badge_reward_id UUID REFERENCES badges(id),
  steps           JSONB NOT NULL DEFAULT '[]',             -- [{title, description, target_count}]
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE quests ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON quests
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `member_quest_progress`
```sql
CREATE TABLE member_quest_progress (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  member_id       UUID NOT NULL REFERENCES members(id),
  quest_id        UUID NOT NULL REFERENCES quests(id),
  progress        JSONB NOT NULL DEFAULT '{}',             -- {step_index: current_count}
  completed_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(member_id, quest_id)
);

ALTER TABLE member_quest_progress ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON member_quest_progress
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `notifications`
```sql
CREATE TABLE notifications (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  user_id     UUID NOT NULL REFERENCES users(id),          -- NOT member_id — notifications survive chapter hops
  type        TEXT NOT NULL CHECK (type IN ('badge','prop','xp','dues','event','announcement','level','system')),
  title       TEXT NOT NULL,
  body        TEXT NOT NULL,
  link        TEXT,                                        -- frontend route to navigate to on click
  is_read     BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON notifications
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);
```

---

### Table: `mentorships`
```sql
CREATE TABLE mentorships (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  mentor_id   UUID NOT NULL REFERENCES members(id),
  mentee_id   UUID REFERENCES members(id),                 -- NULL if no mentee yet
  focus_areas TEXT[],                                      -- ['Career Development','Leadership']
  bio         TEXT,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  started_at  TIMESTAMPTZ,
  ended_at    TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE mentorships ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON mentorships
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `intake_prospects`
```sql
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

ALTER TABLE intake_prospects ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON intake_prospects
  USING (chapter_id = current_setting('app.chapter_id')::uuid);

CREATE INDEX idx_intake_chapter ON intake_prospects(chapter_id);
CREATE INDEX idx_intake_stage ON intake_prospects(chapter_id, stage);
```

---

### Table: `votes`
```sql
CREATE TABLE votes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  title       TEXT NOT NULL,
  description TEXT,
  vote_type   TEXT NOT NULL CHECK (vote_type IN ('election','referendum','motion')),
  options     TEXT[] NOT NULL,
  deadline    DATE,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  is_anonymous BOOLEAN NOT NULL DEFAULT TRUE,
  created_by  UUID NOT NULL REFERENCES members(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE votes ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON votes
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `vote_responses`
```sql
CREATE TABLE vote_responses (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  vote_id     UUID NOT NULL REFERENCES votes(id),
  member_id   UUID NOT NULL REFERENCES members(id),
  option_index INT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(vote_id, member_id)
);

ALTER TABLE vote_responses ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON vote_responses
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `chapter_minutes`
```sql
CREATE TABLE chapter_minutes (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID NOT NULL REFERENCES chapters(id),
  meeting_date    DATE NOT NULL,
  title           TEXT NOT NULL,
  body            TEXT NOT NULL,
  recorder_id     UUID REFERENCES members(id),
  quorum          BOOLEAN NOT NULL DEFAULT FALSE,
  attendee_ids    UUID[],
  xp_for_attendance INT NOT NULL DEFAULT 0,
  status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','final')),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE chapter_minutes ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON chapter_minutes
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `scholarships`
```sql
CREATE TABLE scholarships (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id    UUID NOT NULL REFERENCES chapters(id),
  applicant_name TEXT NOT NULL,
  school        TEXT,
  gpa           TEXT,
  essay         TEXT,
  academic_year INT,
  city          TEXT,
  submitted_at  DATE,
  status        TEXT NOT NULL DEFAULT 'submitted'
                CHECK (status IN ('submitted','under_review','finalist','awarded','declined','needs_more_info')),
  reviewer_id   UUID REFERENCES members(id),
  notes         TEXT,
  award_amount_cents INT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE scholarships ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON scholarships
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `fundraising_campaigns`
```sql
CREATE TABLE fundraising_campaigns (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id        UUID NOT NULL REFERENCES chapters(id),
  name              TEXT NOT NULL,
  description       TEXT,
  goal_cents        INT NOT NULL,
  raised_cents      INT NOT NULL DEFAULT 0,
  category          TEXT,
  deadline          DATE,
  is_active         BOOLEAN NOT NULL DEFAULT TRUE,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE fundraising_campaigns ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON fundraising_campaigns
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `store_items`
```sql
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

ALTER TABLE store_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON store_items
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `store_orders`
```sql
CREATE TABLE store_orders (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  member_id   UUID NOT NULL REFERENCES members(id),
  item_id     UUID NOT NULL REFERENCES store_items(id),
  xp_spent    INT NOT NULL,
  status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','fulfilled','cancelled')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE store_orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON store_orders
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `chapter_goals`
```sql
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

ALTER TABLE chapter_goals ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON chapter_goals
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `job_board`
```sql
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

ALTER TABLE job_board ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON job_board
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `point_economy`
```sql
-- Per-chapter configurable XP awards
CREATE TABLE point_economy (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id),
  activity    TEXT NOT NULL,
  xp          INT NOT NULL,
  category    TEXT,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE(chapter_id, activity)
);

ALTER TABLE point_economy ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON point_economy
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

---

### Table: `audit_log` (platform-wide, not tenant-scoped)
```sql
-- Sysadmin actions and billing events. Not RLS-protected (read by sysadmin only).
CREATE TABLE audit_log (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id    UUID,                                        -- user_id or NULL for system
  actor_type  TEXT NOT NULL CHECK (actor_type IN ('user','system','webhook')),
  chapter_id  UUID,                                        -- NULL for platform-level actions
  action      TEXT NOT NULL,                               -- 'member.created', 'dues.paid', etc.
  target_type TEXT,
  target_id   UUID,
  metadata    JSONB DEFAULT '{}',
  ip_address  INET,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_actor ON audit_log(actor_id);
CREATE INDEX idx_audit_log_chapter ON audit_log(chapter_id);
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);
```

---

## 5. Auth Design

### Authentication Flow

**Email/Password Login:**
```
1. POST /auth/login { email, password }
2. Server: bcrypt.Compare(password, user.password_hash)
3. If valid: issue access_token (JWT, 15min) + refresh_token (opaque, 30 days)
4. refresh_token stored hashed in auth_sessions table
5. Tokens returned in response body (no httpOnly cookie — SPA needs them for mobile PWA too)
   -- NOTE: Implement PKCE flow later for mobile; for now, store in memory/sessionStorage
6. Frontend stores access_token in memory (Zustand), refresh_token in localStorage
```

**Magic Link Login:**
```
1. POST /auth/magic-link { email }
2. Server generates random 32-byte token, stores SHA-256 hash in magic_link_tokens (expires 15min)
3. Resend sends email: "Sign in to The Blue Ledger → https://app.blueledger.io/auth/magic?token=<raw>"
4. User clicks link → GET /auth/magic?token=<raw>
5. Server: SHA-256(raw) → lookup → validate expiry → mark used_at
6. Issue access_token + refresh_token same as password flow
```

**Token Refresh:**
```
1. POST /auth/refresh { refresh_token }
2. Server: lookup hashed refresh token in auth_sessions
3. If valid and not expired: issue new access_token (15min) + rotate refresh_token
4. Old refresh_token marked revoked
```

### JWT Structure
```json
{
  "sub": "user-uuid",
  "chapter_id": "chapter-uuid",
  "member_id": "member-uuid",
  "role": "admin",
  "email": "m.williams@chapter.org",
  "iat": 1711234567,
  "exp": 1711235467
}
```
Signed with RS256 (private key on server, public key verifiable without DB).

### Middleware Stack (per-request)
```
Request
  → RateLimiter (Redis: 100 req/min per IP, 1000 req/min per user)
  → RequestLogger (structured JSON: method, path, status, duration, user_id)
  → CORSMiddleware (allow: app domain only)
  → JWTMiddleware (parse + validate access_token; set ctx.user)
  → TenantSetter (SET LOCAL app.chapter_id = ctx.chapter_id — applies RLS)
  → RoleGate (per-route: roles allowed list)
  → Handler
```

### Role Permission Matrix
| Action | member | chair | pia | admin | sysadmin |
|--------|--------|-------|-----|-------|----------|
| View own profile | Y | Y | Y | Y | Y |
| Edit own profile | Y | Y | Y | Y | Y |
| View all members | Y | Y | Y | Y | Y |
| Edit any member | N | N | N | Y | Y |
| Create events | N | Y | N | Y | Y |
| QR check-in scan | N | Y | N | Y | Y |
| Verify service | N | Y | N | Y | Y |
| Award XP manually | N | N | N | Y | Y |
| Manage dues | N | N | N | Y | Y |
| PIA reports | N | N | Y | Y | Y |
| Chapter health | N | N | Y | Y | Y |
| Manage intake | N | Y | N | Y | Y |
| Manage scholarships | N | N | Y | Y | Y |
| Danger zone ops | N | N | N | Y | Y |
| View all chapters | N | N | N | N | Y |
| Platform billing | N | N | N | N | Y |

---

## 6. API Endpoints — Full REST Surface

### Base URL: `https://api.blueledger.io/v1`

All authenticated endpoints require: `Authorization: Bearer <access_token>`
All responses use: `Content-Type: application/json`

Error response envelope:
```json
{ "error": { "code": "MEMBER_NOT_FOUND", "message": "Member not found", "status": 404 } }
```

Success response envelope (list):
```json
{ "data": [...], "meta": { "total": 42, "page": 1, "per_page": 25, "pages": 2 } }
```

Success response envelope (single):
```json
{ "data": { ... } }
```

---

### Auth Endpoints (`/auth`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| POST | `/auth/register` | None | — | Register a chapter + admin user |
| POST | `/auth/login` | None | — | Email/password login |
| POST | `/auth/magic-link` | None | — | Request magic link email |
| GET | `/auth/magic` | None | — | Consume magic link token |
| POST | `/auth/refresh` | None | — | Rotate refresh token |
| POST | `/auth/logout` | JWT | Any | Revoke current session |
| GET | `/auth/me` | JWT | Any | Current user + member profile |

---

### Members Endpoints (`/members`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/members` | JWT | Any | List all chapter members (paginated) |
| GET | `/members/:id` | JWT | Any | Get member profile |
| PUT | `/members/:id` | JWT | self or admin | Update member profile |
| PUT | `/members/:id/role` | JWT | admin | Change member role |
| PUT | `/members/:id/status` | JWT | admin | Change member status |
| DELETE | `/members/:id` | JWT | admin | Soft-delete member |
| GET | `/members/:id/xp-history` | JWT | self or admin | XP engagement log |
| GET | `/members/:id/badges` | JWT | Any | Member's earned badges |
| GET | `/members/:id/service` | JWT | Any | Member's service log |

---

### XP & Leaderboard Endpoints

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/leaderboard` | JWT | Any | XP leaderboard (alltime, semester) |
| POST | `/xp/award` | JWT | admin | Manually award XP to a member |
| POST | `/xp/recalculate` | JWT | admin | Recalculate all XP from engagement_log |
| GET | `/engagement-log` | JWT | admin, chair | Full chapter engagement log (paginated) |

---

### Events Endpoints (`/events`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/events` | JWT | Any | List events (upcoming, past) |
| POST | `/events` | JWT | admin, chair | Create event |
| GET | `/events/:id` | JWT | Any | Get event detail + RSVPs |
| PUT | `/events/:id` | JWT | admin, chair | Update event |
| DELETE | `/events/:id` | JWT | admin | Cancel/delete event |
| POST | `/events/:id/rsvp` | JWT | Any | RSVP to event |
| DELETE | `/events/:id/rsvp` | JWT | Any | Remove RSVP |
| POST | `/events/:id/checkin` | JWT | admin, chair | Manual check-in a member |
| POST | `/events/:id/checkin/qr` | JWT | admin, chair | QR code check-in (validates token) |
| GET | `/events/:id/qr-token` | JWT | admin, chair | Get/generate QR check-in token |

---

### Service Log Endpoints (`/service`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/service` | JWT | Any | List service log entries |
| POST | `/service` | JWT | Any | Submit service hours |
| GET | `/service/:id` | JWT | Any | Get service entry |
| PUT | `/service/:id/verify` | JWT | admin, chair | Verify service hours (awards XP) |
| DELETE | `/service/:id` | JWT | admin | Delete service entry |

---

### Dues Endpoints (`/dues`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/dues` | JWT | admin | All dues records for semester |
| GET | `/dues/me` | JWT | Any | My dues records |
| POST | `/dues` | JWT | admin | Create dues record for member |
| PUT | `/dues/:id/mark-paid` | JWT | admin | Mark dues paid (cash/Zelle) |
| POST | `/dues/:id/stripe-session` | JWT | self or admin | Create Stripe checkout for dues payment |
| POST | `/dues/webhooks/stripe` | None (sig verify) | — | Stripe payment webhook |

---

### Announcements Endpoints (`/announcements`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/announcements` | JWT | Any | List announcements (active, pinned first) |
| POST | `/announcements` | JWT | admin, chair | Create announcement |
| PUT | `/announcements/:id` | JWT | admin, chair | Edit announcement |
| DELETE | `/announcements/:id` | JWT | admin | Archive/delete announcement |
| PUT | `/announcements/:id/pin` | JWT | admin | Toggle pin |

---

### Props Endpoints (`/props`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/props` | JWT | Any | List props (to/from me, or all for admin) |
| POST | `/props` | JWT | Any | Give props to a brother (+10 XP to recipient) |
| GET | `/props/feed` | JWT | Any | Props social feed |

---

### Notifications Endpoints (`/notifications`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/notifications` | JWT | Any | My notifications (unread first) |
| PUT | `/notifications/:id/read` | JWT | Any | Mark notification read |
| PUT | `/notifications/read-all` | JWT | Any | Mark all read |
| DELETE | `/notifications/:id` | JWT | Any | Dismiss notification |

---

### Badges & Quests Endpoints

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/badges` | JWT | Any | All badge definitions for chapter |
| POST | `/badges` | JWT | admin | Create custom badge |
| POST | `/badges/:id/award` | JWT | admin | Manually award badge to member |
| GET | `/quests` | JWT | Any | All quests + my progress |
| GET | `/quests/:id/progress` | JWT | Any | My progress on a quest |

---

### Intake Endpoints (`/intake`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/intake` | JWT | admin, chair | List all prospects |
| POST | `/intake` | JWT | admin, chair | Add prospect |
| GET | `/intake/:id` | JWT | admin, chair | Get prospect detail |
| PUT | `/intake/:id` | JWT | admin, chair | Update prospect stage/notes |
| DELETE | `/intake/:id` | JWT | admin | Remove prospect |

---

### Votes Endpoints (`/votes`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/votes` | JWT | Any | List active votes |
| POST | `/votes` | JWT | admin | Create vote |
| PUT | `/votes/:id` | JWT | admin | Edit/close vote |
| POST | `/votes/:id/respond` | JWT | Any | Submit vote response |
| GET | `/votes/:id/results` | JWT | admin (live), Any (after close) | Vote results |

---

### Mentorship Endpoints (`/mentorship`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/mentorship` | JWT | Any | List active mentorship pairs |
| POST | `/mentorship` | JWT | Any | Register as mentor / request pairing |
| PUT | `/mentorship/:id` | JWT | admin or mentor | Update mentorship |

---

### Minutes Endpoints (`/minutes`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/minutes` | JWT | Any | List meeting minutes |
| POST | `/minutes` | JWT | admin, chair | Create minutes |
| GET | `/minutes/:id` | JWT | Any | Get minutes detail |
| PUT | `/minutes/:id` | JWT | admin, chair | Edit minutes |
| PUT | `/minutes/:id/finalize` | JWT | admin | Mark minutes as Final |

---

### Scholarships Endpoints (`/scholarships`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/scholarships` | JWT | pia, admin | List applications |
| POST | `/scholarships` | JWT | pia, admin | Add application |
| GET | `/scholarships/:id` | JWT | pia, admin | Get application |
| PUT | `/scholarships/:id` | JWT | pia, admin | Update status/notes |

---

### PIA / Chapter Health Endpoints

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/health` | JWT | pia, admin | Chapter health dashboard data |
| GET | `/health/attendance` | JWT | pia, admin | Attendance analytics |
| GET | `/health/xp` | JWT | pia, admin | XP distribution analytics |
| GET | `/health/service` | JWT | pia, admin | Service hours analytics |
| GET | `/health/dues` | JWT | pia, admin | Dues collection analytics |

---

### Store Endpoints (`/store`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/store` | JWT | Any | List store items |
| POST | `/store` | JWT | admin | Create store item |
| POST | `/store/:id/purchase` | JWT | Any | Purchase item with XP |
| GET | `/store/orders` | JWT | admin | All chapter store orders |

---

### Chapter Goals Endpoints (`/goals`)

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/goals` | JWT | Any | List chapter goals |
| POST | `/goals` | JWT | admin | Create goal |
| PUT | `/goals/:id` | JWT | admin | Update goal progress |

---

### Chapter Settings & Config Endpoints

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/settings` | JWT | admin | Chapter settings |
| PUT | `/settings` | JWT | admin | Update chapter settings |
| GET | `/settings/point-economy` | JWT | Any | XP point economy table |
| PUT | `/settings/point-economy` | JWT | admin | Update XP values |
| GET | `/settings/billing` | JWT | admin | Subscription + billing status |
| POST | `/settings/billing/portal` | JWT | admin | Create Stripe billing portal session |

---

### Sysadmin Endpoints (`/sys`) — sysadmin role only

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| GET | `/sys/chapters` | JWT | sysadmin | List all chapters |
| POST | `/sys/chapters` | JWT | sysadmin | Create new chapter tenant |
| GET | `/sys/chapters/:id` | JWT | sysadmin | Get chapter detail + stats |
| PUT | `/sys/chapters/:id/subscription` | JWT | sysadmin | Override subscription status |
| GET | `/sys/users` | JWT | sysadmin | Platform user list |
| GET | `/sys/audit-log` | JWT | sysadmin | Platform audit log |
| POST | `/sys/xp/recalculate-all` | JWT | sysadmin | Recalculate XP for all chapters |
| DELETE | `/sys/chapters/:id` | JWT | sysadmin | Deactivate chapter |

---

### QR Check-In Flow Detail

The QR check-in is a key feature. Here is the precise server-side flow:

```
1. Chair opens scanner → GET /events/:id/qr-token
   Server: if no qr_code_token exists, generate signed token (HMAC-SHA256, event_id + chapter_id + timestamp)
   Store token in events.qr_code_token. Return it to the scanner.

2. Member shows Digital ID QR → QR encodes: "BL:<member_display_id>"
   (QR is generated client-side using member_display_id — no secret needed)

3. Scanner reads QR → POST /events/:id/checkin/qr
   Body: { member_display_id: "ΤΤΣ-001", scanner_token: "<event qr token>" }

4. Server validates:
   a. scanner_token matches events.qr_code_token for this event
   b. Event is active and within check-in window
   c. Member exists in chapter
   d. No duplicate check-in

5. On success:
   a. Upsert RSVP (if no RSVP exists, create one with status='yes', attended=true)
   b. Mark rsvps.attended = true, checked_in_at = NOW(), checked_in_by = current member
   c. Insert engagement_log row (source='checkin', xp=event.xp_attend)
   d. UPDATE members SET xp_total = xp_total + event.xp_attend
   e. Run badge engine check (async/background job)
   f. Create notification for member
   g. Return { member_name, xp_awarded, level, new_badges }
```

---

## 7. XP & Badge Engine

### XP Calculation Rules
- `members.xp_total` is a denormalized cache. The true source of truth is `SUM(engagement_log.xp_awarded)` for that member.
- On any XP award: INSERT into `engagement_log`, then `UPDATE members SET xp_total = xp_total + N`.
- `POST /xp/recalculate` is a safety valve: it recalculates all member XP from scratch using engagement_log, correcting any drift.
- Semester XP (`members.xp_semester`) resets at semester rollover (triggered by admin via `PUT /settings`).

### Level Thresholds (from prototype, carried over)
```
0    – 249   XP → Neophyte    (neo)
250  – 749   XP → Bronze Varsity (bronze)
750  – 1499  XP → Silver Elite   (silver)
1500 – 2499  XP → Gold Legend    (gold)
2500+        XP → Chapter Icon   (icon)
```

Stored in `chapters.settings` JSONB:
```json
{
  "xp_levels": [
    { "key": "neo",    "label": "Neophyte",      "min": 0    },
    { "key": "bronze", "label": "Bronze Varsity", "min": 250  },
    { "key": "silver", "label": "Silver Elite",   "min": 750  },
    { "key": "gold",   "label": "Gold Legend",    "min": 1500 },
    { "key": "icon",   "label": "Chapter Icon",   "min": 2500 }
  ]
}
```

### Badge Engine
Badge criteria stored in `badges.criteria` JSONB. The engine runs after every XP-awarding event:

```json
// Example: "25-Hour Builder" badge
{
  "type": "service_hours_gte",
  "value": 25
}

// Example: "Punctual" badge
{
  "type": "on_time_meetings_gte",
  "value": 10
}

// Example: "Faithful" badge
{
  "type": "dues_on_time_consecutive_semesters_gte",
  "value": 3
}

// Example: "Chapter Legend" badge
{
  "type": "xp_gte",
  "value": 2500
}
```

Badge engine function (runs async after check-in, XP award, service verify):
```go
// CheckAndAwardBadges(ctx, chapterID, memberID) — checks all badge criteria for a member
// For each badge not already earned:
//   1. Query relevant engagement_log/service_log/dues_records data
//   2. Evaluate criteria
//   3. If met: INSERT member_badges, INSERT engagement_log (xp_reward), UPDATE member xp, create notification
```

---

## 8. Billing & Subscription Model

### Plans
| Plan | Price | Included Members | Features |
|------|-------|-----------------|----------|
| Starter | $29/mo | Up to 15 | All core features |
| Growth | $59/mo | Up to 50 | All features + PIA reports |
| Chapter Pro | $99/mo | Unlimited | All features + priority support + custom branding |

### Stripe Integration
- Create Stripe Customer at chapter registration: `stripe.Customer.create({ email, name })`
- Create Stripe Subscription at activation (free trial first 30 days)
- `stripe_customer_id` and `stripe_subscription_id` stored on `chapters` table
- Webhook handler: `POST /dues/webhooks/stripe` (for dues payments) + `POST /billing/webhooks/stripe` (for subscription events)
- Subscription webhook events to handle:
  - `customer.subscription.updated` → update `chapters.subscription_status`
  - `customer.subscription.deleted` → set `canceled`, restrict API access
  - `invoice.payment_failed` → set `past_due`, send warning email
  - `invoice.payment_succeeded` → set `active`

### Subscription Gate Middleware
```go
// For all chapter-scoped endpoints (not auth, not billing portal):
// If chapter.subscription_status IN ('past_due', 'canceled') → 402 Payment Required
// Allow grace period: past_due chapters get 7 days before restriction
// Sysadmin bypasses all subscription gates
```

---

## 9. Frontend Architecture (React Migration)

### Folder Structure
```
web/
  src/
    api/            — Axios instance + all API client functions
    components/     — Shared UI components (Button, Card, Modal, Toast, Badge)
    features/       — One folder per feature domain
      auth/
      dashboard/
      members/
      events/
      xp-leaderboard/
      dues/
      service/
      quests-badges/
      announcements/
      props/
      intake/
      voting/
      mentorship/
      minutes/
      scholarships/
      pia-reports/
      store/
      goals/
      notifications/
      sysadmin/     — System console (sysadmin only)
    hooks/          — Shared hooks (useAuth, useChapter, useCurrentMember)
    stores/         — Zustand stores (authStore, uiStore)
    types/          — TypeScript interfaces matching API response shapes
    utils/          — Date formatting, XP level helpers, etc.
    App.tsx         — Route definitions
    main.tsx        — React root
  index.html
  vite.config.ts
  tailwind.config.ts
  tsconfig.json
```

### Route → Screen Mapping (from existing 47 screens)
```
/ → login
/dashboard → dashboard
/profile → my profile
/profile/edit → edit profile
/members → member directory
/members/:id → member profile
/digital-id → digital ID + QR
/leaderboard → XP leaderboard
/events → events + RSVP
/service → service log
/dues → dues/financials
/announcements → announcements
/props → props feed
/quests → quests + badges
/notifications → notifications
/messages → direct messages
/mentorship → mentorship pairs
/intake → intake pipeline
/votes → voting
/minutes → meeting minutes
/scholarships → scholarship management
/pia → PIA builder / chapter health
/store → XP store
/goals → chapter goals
/sbc → Sigma Beta Club log
/job-board → job board
/alumni → alumni directory
/fundraising → fundraising campaigns
/committees → committees
/resources → resources
/milestones → milestones
/history → chapter history
/network → network map
/announcements/create → create announcement
/events/create → create event
/admin → admin panel
/admin/members → member management
/admin/dues → dues management
/admin/xp → XP management
/admin/settings → chapter settings
/sys → system console (sysadmin only)
/sys/chapters → all chapters
/sys/audit-log → audit log
/sys/billing → billing overview
```

### API Client Pattern
```typescript
// api/client.ts — singleton Axios instance
// api/auth.ts   — auth endpoints
// api/members.ts
// api/events.ts
// etc.

// All API functions are typed:
export async function getMembers(params?: { page?: number; per_page?: number }): Promise<PaginatedResponse<Member>>

// TanStack Query hooks wrap API calls:
export function useMembers(params?) {
  return useQuery({ queryKey: ['members', params], queryFn: () => getMembers(params) })
}
```

### Design Token Preservation
The existing CSS variables (--navy, --gold, --cream, etc.) are ported verbatim to
`tailwind.config.ts` as custom colors. All existing class names remain functional.
The full CSS block from the prototype's `<style>` tag becomes `src/styles/base.css`.
Component-level styles use Tailwind utilities supplemented by existing class names.

---

## 10. Migration Plan — Prototype to SaaS

### Phase 0: Foundation (2 weeks — do first, blocks everything)
1. Set up Go project structure, Echo framework, Postgres, Redis
2. Write all migration files (001 through 020)
3. Implement auth service (password + magic link + JWT + refresh)
4. Deploy to Fly.io dev environment (DB + API)
5. Set up Vite + React + TypeScript + Tailwind project
6. Implement auth screens (login, magic link) wired to API

### Phase 1: Core Member & XP (2 weeks)
7. Members API + React members screens
8. XP engine + engagement log + leaderboard API + screens
9. Dashboard wired to real data

### Phase 2: Events & Check-In (1 week)
10. Events API + QR check-in endpoint
11. React events screens + QR scanner integration
12. RSVP flow + XP award on check-in

### Phase 3: Dues & Billing (1.5 weeks)
13. Dues API + Stripe payment integration
14. Stripe billing subscription + webhook handler
15. Chapter registration + subscription management screens

### Phase 4: Engagement Features (3 weeks)
16. Announcements, Props, Notifications API + screens
17. Service log + verification flow
18. Badges + Badge engine
19. Quests + progress tracking
20. Votes, Minutes, Intake, Mentorship, Scholarships APIs + screens

### Phase 5: PIA, Admin & Sysadmin (2 weeks)
21. Chapter health / PIA analytics endpoints
22. Admin panel (member management, XP management, settings)
23. Sysadmin console (chapter list, audit log, billing override)
24. Store, Goals, Job Board, SBC Log, Alumni, Fundraising, etc.

### Phase 6: Polish & Launch (1 week)
25. Notification emails (dues reminder, event reminder, badge earned)
26. Rate limiting, error monitoring (Sentry)
27. Load testing, security review
28. Production deploy + DNS cutover
29. Seed production data from existing prototype (one-time import script)

### Data Seed Script
A one-time `cmd/seed/main.go` script reads the prototype's hardcoded data objects
and inserts them into the production database for the founding chapter. The script:
- Creates the first chapter tenant (Tau Tau Sigma)
- Creates user accounts for all 7 members
- Creates their member records with existing XP, level, dues status, badges
- Imports engagement_log, service_log, events, announcements, etc.
- Sets up default badges, quests, and point_economy config

---

## 11. Project Structure (Go API)

```
blue-ledger-api/
  cmd/
    server/
      main.go           — entrypoint, wire dependencies, start Echo
    seed/
      main.go           — one-time data seed from prototype
    migrate/
      main.go           — run golang-migrate (can also be called from main with --migrate flag)
  internal/
    auth/
      service.go        — AuthService: Login, MagicLink, Refresh, Logout
      handler.go        — HTTP handlers
      middleware.go     — JWTMiddleware, TenantSetter, RoleGate
      tokens.go         — JWT generation/validation, refresh token management
    members/
      service.go
      handler.go
      repository.go     — sqlc-generated queries + custom queries
    xp/
      service.go        — AwardXP, RecalculateXP, GetLeaderboard
      handler.go
      badge_engine.go   — CheckAndAwardBadges
    events/
      service.go
      handler.go
      qr.go             — QR token generation/validation
    dues/
      service.go
      handler.go
      stripe.go         — Stripe checkout + webhook
    notifications/
      service.go        — CreateNotification, MarkRead
      handler.go
    ... (one package per domain)
    platform/
      sys_handler.go    — sysadmin endpoints
      billing_handler.go — subscription management
  pkg/
    db/
      db.go             — pgx pool initialization + RLS session setter
    redis/
      client.go
    email/
      resend.go         — SendMagicLink, SendDuesReminder, SendBadgeEarned
    storage/
      r2.go             — Cloudflare R2 avatar upload
    config/
      config.go         — env var loading (viper or envconfig)
    logger/
      logger.go         — zerolog structured logger
    validator/
      validator.go      — go-playground/validator setup
  migrations/
    001_initial_schema.up.sql
    001_initial_schema.down.sql
    002_rls_policies.up.sql
    ...
  sqlc/
    schema.sql          — aggregated schema (for sqlc)
    queries/
      members.sql
      events.sql
      engagement_log.sql
      ...
    sqlc.yaml
  Dockerfile
  docker-compose.yml    — local dev: postgres + redis
  .env.example
  go.mod
  go.sum
```

---

## 12. Infrastructure Specification

### Production Topology

```
Users (browser/mobile PWA)
  → Cloudflare CDN (static assets, R2 images, DDoS protection)
  → Fly.io Load Balancer (TLS termination)
      ├── API instances (Go, 2x shared-cpu-1x, 256MB RAM, auto-scale 1–4)
      └── Static file serving (React build, or served from Cloudflare Pages)
  → Fly.io Postgres (postgres-flex-1x, 1GB RAM, daily snapshots)
  → Fly.io Redis (upstash-redis or valkey, 256MB)

External:
  → Resend (transactional email)
  → Stripe (payments)
  → Cloudflare R2 (file storage: avatars, resource uploads)
  → Sentry (error monitoring)
```

### Environment Variables (`.env.example`)
```bash
# Server
APP_ENV=production
PORT=8080
API_BASE_URL=https://api.blueledger.io

# Database
DATABASE_URL=postgres://blue_ledger:password@hostname:5432/blue_ledger?sslmode=require

# Redis
REDIS_URL=redis://:password@hostname:6379

# Auth
JWT_PRIVATE_KEY_PATH=/run/secrets/jwt_private_key
JWT_PUBLIC_KEY_PATH=/run/secrets/jwt_public_key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=720h     # 30 days
MAGIC_LINK_EXPIRY=15m
MAGIC_LINK_HMAC_SECRET=<32-byte-hex>

# Email (Resend)
RESEND_API_KEY=re_...
EMAIL_FROM=noreply@blueledger.io
EMAIL_FROM_NAME=The Blue Ledger

# Storage (Cloudflare R2)
R2_ACCOUNT_ID=...
R2_ACCESS_KEY_ID=...
R2_SECRET_ACCESS_KEY=...
R2_BUCKET=blue-ledger-assets
R2_PUBLIC_URL=https://assets.blueledger.io

# Stripe
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_STARTER_PRICE_ID=price_...
STRIPE_GROWTH_PRICE_ID=price_...
STRIPE_PRO_PRICE_ID=price_...

# Sentry
SENTRY_DSN=https://...@sentry.io/...

# Frontend (Vite build)
VITE_API_BASE_URL=https://api.blueledger.io/v1
```

### Fly.io Configuration (`fly.toml`)
```toml
app = "blue-ledger-api"
primary_region = "iad"    # Washington DC — closest to Atlanta target market

[build]
  dockerfile = "Dockerfile"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 1

[[vm]]
  size = "shared-cpu-1x"
  memory = "256mb"
```

### Docker (multi-stage build)
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

---

## 13. Task Breakdown with Hour Estimates

### Routing: Agent Assignments

| Phase | Scope | Hours | Builder |
|-------|-------|-------|---------|
| Phase 0: Foundation | DB schema + migrations + auth + deploy | ~60 hrs | Iron Man (3 agents) |
| Phase 1: Core Member & XP | Members API + XP engine + leaderboard | ~40 hrs | Wasp sprint |
| Phase 2: Events & Check-In | Events API + QR check-in | ~24 hrs | Wasp sprint |
| Phase 3: Dues & Billing | Dues API + Stripe | ~32 hrs | Wasp sprint |
| Phase 4: Engagement Features | Props, badges, quests, votes, intake, etc. | ~80 hrs | Iron Man (parallel) |
| Phase 5: PIA & Admin | Analytics, admin panel, sysadmin console | ~40 hrs | Wasp sprint |
| Phase 6: Polish & Launch | Email, monitoring, load test, deploy | ~24 hrs | Ant-Man tasks |

---

### TASK-001: Project Setup & Database Foundation
**Hours:** 16 hrs | **Agent:** Iron Man Agent 1

**Files to create:**
- `blue-ledger-api/go.mod` — module `github.com/blueledger/api`
- `blue-ledger-api/cmd/server/main.go` — Echo setup, route registration, graceful shutdown
- `blue-ledger-api/pkg/config/config.go` — env var loading
- `blue-ledger-api/pkg/db/db.go` — pgx pool, RLS session setter helper
- `blue-ledger-api/pkg/logger/logger.go` — zerolog setup
- `blue-ledger-api/migrations/001_initial_schema.up.sql` — all tables per spec §4
- `blue-ledger-api/migrations/001_initial_schema.down.sql`
- `blue-ledger-api/migrations/002_rls_policies.up.sql` — all RLS policies per spec §3
- `blue-ledger-api/migrations/002_rls_policies.down.sql`
- `blue-ledger-api/migrations/003_seed_badges_quests.up.sql` — default badge/quest definitions
- `blue-ledger-api/docker-compose.yml` — postgres:16, redis:7
- `blue-ledger-api/.env.example`
- `blue-ledger-api/Dockerfile`
- `blue-ledger-api/fly.toml`

**Acceptance criteria:**
- [ ] `docker-compose up` starts Postgres + Redis locally
- [ ] `go run ./cmd/migrate` runs all migrations cleanly
- [ ] `go run ./cmd/server` starts without errors
- [ ] GET `/health` returns `{ "status": "ok" }`
- [ ] All 20+ tables exist in DB with correct columns and constraints
- [ ] RLS policies are in place (verify with: login as chapter user, attempt cross-chapter query → should return empty)

---

### TASK-002: Authentication Service
**Hours:** 20 hrs | **Agent:** Iron Man Agent 1

**Files to create:**
- `internal/auth/service.go` — Login, Register, MagicLinkRequest, MagicLinkConsume, Refresh, Logout
- `internal/auth/handler.go` — HTTP handlers for all /auth endpoints
- `internal/auth/middleware.go` — JWTMiddleware, TenantSetterMiddleware, RoleGateMiddleware
- `internal/auth/tokens.go` — JWT sign/parse (RS256), refresh token generation/hash/store
- `internal/auth/service_test.go` — unit tests for auth service
- `internal/auth/handler_test.go` — integration tests for auth endpoints
- `pkg/email/resend.go` — SendMagicLink function

**Function signatures:**
```go
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error)
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResult, error)
func (s *AuthService) RequestMagicLink(ctx context.Context, email string) error
func (s *AuthService) ConsumeMagicLink(ctx context.Context, token string) (*AuthResult, error)
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*AuthResult, error)
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error

type AuthResult struct {
    AccessToken  string
    RefreshToken string
    User         *User
    Member       *Member  // nil for sysadmin
    Chapter      *Chapter // nil for sysadmin
}
```

**Validation rules:**
- email: required, valid email format, max 254 chars
- password: required, min 8 chars, max 72 chars (bcrypt limit)
- Login: lock account for 15 min after 10 failed attempts
- Magic link: rate-limit 3 requests per email per hour

**Acceptance criteria:**
- [ ] POST /auth/login returns 200 with access_token + refresh_token on valid credentials
- [ ] POST /auth/login returns 401 on bad password
- [ ] POST /auth/login returns 423 on locked account
- [ ] POST /auth/magic-link sends email and returns 202
- [ ] GET /auth/magic?token=<valid> returns 200 with tokens
- [ ] GET /auth/magic?token=<expired> returns 401
- [ ] GET /auth/magic?token=<used> returns 401
- [ ] POST /auth/refresh rotates refresh token
- [ ] Expired access_token on protected endpoint → 401
- [ ] Wrong chapter_id in JWT on protected endpoint → 403

---

### TASK-003: React Frontend Foundation
**Hours:** 16 hrs | **Agent:** Iron Man Agent 2

**Files to create:**
- `web/` — Vite + React + TypeScript project scaffold
- `web/src/styles/base.css` — all CSS from prototype's `<style>` block (verbatim)
- `web/tailwind.config.ts` — extend with all `--navy`, `--gold`, `--cream`, etc. tokens
- `web/src/api/client.ts` — Axios instance, request/response interceptors, token refresh logic
- `web/src/stores/authStore.ts` — Zustand store: currentUser, accessToken, actions
- `web/src/stores/uiStore.ts` — currentView, sidebarOpen, toast queue
- `web/src/types/index.ts` — all TypeScript interfaces (Member, Event, XPEntry, Badge, etc.)
- `web/src/components/` — Button, Card, Modal, Toast, Badge, Avatar, Sidebar, Topbar
- `web/src/features/auth/` — LoginPage, MagicLinkPage
- `web/src/App.tsx` — route definitions for all 47 screens (stub pages initially)

**Acceptance criteria:**
- [ ] `npm run dev` starts Vite on :3001
- [ ] `npm run build` produces a production build with no TypeScript errors
- [ ] Login screen matches prototype design exactly (navy background, gold shield, email/password form)
- [ ] POST /auth/login wired and functional from login form
- [ ] Successful login navigates to /dashboard
- [ ] access_token stored in Zustand memory, refresh_token in localStorage
- [ ] All 47 routes exist (stub pages ok, real content in later tasks)
- [ ] Sidebar renders with role-appropriate nav items based on JWT role

---

### TASK-004: Members API & Screens
**Hours:** 20 hrs | **Agent:** Iron Man Agent 2

**API files:**
- `internal/members/service.go` — GetMember, ListMembers, UpdateMember, ChangeRole, etc.
- `internal/members/handler.go`
- `internal/members/repository.go` — sqlc queries

**Frontend files:**
- `web/src/features/members/MemberDirectory.tsx`
- `web/src/features/members/MemberProfile.tsx`
- `web/src/features/members/MemberCard.tsx`
- `web/src/features/members/EditProfile.tsx`
- `web/src/api/members.ts`
- `web/src/hooks/useMembers.ts`

**Acceptance criteria:**
- [ ] GET /members returns paginated member list for chapter
- [ ] Cross-chapter query (spoofed JWT with wrong chapter_id) returns empty, not cross-tenant data
- [ ] Member directory renders with search filtering
- [ ] Member profile shows XP, level badge, service hours, badges earned
- [ ] Edit profile saves to DB and reflects on reload

---

### TASK-005: XP Engine, Leaderboard & Engagement Log
**Hours:** 20 hrs | **Agent:** Iron Man Agent 3

**API files:**
- `internal/xp/service.go` — AwardXP, RecalcAllXP, GetLeaderboard, GetEngagementLog
- `internal/xp/handler.go`
- `internal/xp/badge_engine.go` — CheckAndAwardBadges (async goroutine)

**Key design notes:**
- AwardXP is transactional: INSERT engagement_log + UPDATE members.xp_total in one transaction
- RecalcAllXP runs in a background job, not in-request (too slow for large chapters)
- Leaderboard: `SELECT member_id, xp_total FROM members WHERE chapter_id = ? ORDER BY xp_total DESC LIMIT 100`
- Level upgrade detection: compare old level vs new level after XP award → if changed, create notification

**Frontend files:**
- `web/src/features/xp-leaderboard/Leaderboard.tsx`
- `web/src/features/dashboard/Dashboard.tsx` — full dashboard with XP summary

**Acceptance criteria:**
- [ ] POST /xp/award inserts engagement_log row and updates member xp_total atomically
- [ ] Level upgrade creates a notification
- [ ] Leaderboard returns members sorted by XP
- [ ] POST /xp/recalculate recalculates XP from engagement_log correctly (all sources)
- [ ] Topbar XP pill updates after XP award (TanStack Query invalidation)

---

### TASK-006: Events API & QR Check-In
**Hours:** 24 hrs | **Agent:** Wasp sprint

**API files:**
- `internal/events/service.go`
- `internal/events/handler.go`
- `internal/events/qr.go` — HMAC token for events, QR token validation
- `internal/events/checkin.go` — QR check-in flow per spec §6

**Frontend files:**
- `web/src/features/events/EventList.tsx`
- `web/src/features/events/EventDetail.tsx`
- `web/src/features/events/CreateEvent.tsx`
- `web/src/features/events/QRScanner.tsx` — uses jsQR or react-qr-reader library
- `web/src/features/events/DigitalID.tsx` — uses qrcode library to generate member QR

**Acceptance criteria:**
- [ ] GET /events returns events for chapter only
- [ ] POST /events creates event and returns with generated qr_code_token
- [ ] POST /events/:id/rsvp awards +10 XP and writes to engagement_log
- [ ] POST /events/:id/checkin/qr validates scanner_token and member QR, awards attendance XP
- [ ] Duplicate check-in returns 409 Conflict
- [ ] QR scanner screen opens camera, reads QR, calls API, shows member name + XP toast
- [ ] Digital ID screen shows member QR code (generated from member_display_id)

---

### TASK-007: Dues & Stripe Integration
**Hours:** 32 hrs | **Agent:** Wasp sprint

**API files:**
- `internal/dues/service.go`
- `internal/dues/handler.go`
- `internal/dues/stripe.go` — create checkout sessions, handle webhooks

**Frontend files:**
- `web/src/features/dues/DuesPage.tsx`
- `web/src/features/dues/DuesRecord.tsx`
- `web/src/features/dues/PayDuesButton.tsx` — redirects to Stripe Checkout

**Billing:**
- `internal/platform/billing_handler.go` — chapter subscription management
- `internal/platform/stripe_webhook.go` — subscription event handling

**Acceptance criteria:**
- [ ] Admin can create dues records for all members for a semester
- [ ] PUT /dues/:id/mark-paid marks as paid and awards +50 XP
- [ ] POST /dues/:id/stripe-session creates Stripe Checkout Session and returns URL
- [ ] Stripe webhook: `checkout.session.completed` marks dues paid + awards XP
- [ ] Stripe webhook: subscription events update chapters.subscription_status
- [ ] Past-due chapters get 402 on API calls after 7-day grace period

---

### TASK-008: Engagement Features (Announcements, Props, Notifications)
**Hours:** 24 hrs | **Agent:** Wasp sprint

Covers: announcements, props (peer recognition), and the notification system.

**Acceptance criteria:**
- [ ] POST /props awards +10 XP to recipient and creates notification
- [ ] GET /notifications returns unread-first list for current user
- [ ] Creating announcement appears in feed for all chapter members
- [ ] PUT /notifications/read-all marks all as read
- [ ] Unread notification count shown in sidebar badge

---

### TASK-009: Service Log, Badges & Quests
**Hours:** 32 hrs | **Agent:** Wasp sprint

Service log with verification flow, badge engine, and quest progress tracking.

**Acceptance criteria:**
- [ ] POST /service creates unverified service log entry
- [ ] PUT /service/:id/verify verifies entry, awards XP (30 per hour), runs badge engine
- [ ] Badge engine awards "25-Hour Builder" when total service hours >= 25
- [ ] Badge engine awards "Chapter Legend" when xp_total >= 2500
- [ ] Quests show progress bars computed from real engagement_log data
- [ ] Earning a badge creates a notification

---

### TASK-010: Phase 4 Remaining Features
**Hours:** 40 hrs | **Agent:** Iron Man (parallel)

Covers: Intake Pipeline, Votes, Mentorship, Meeting Minutes, Scholarships, PIA Builder.

These are largely CRUD features using the same patterns established in prior tasks.
Each gets a service + handler + React feature folder.

**Acceptance criteria:**
- [ ] Intake pipeline stages move correctly (inquiry → under_review → ... → accepted)
- [ ] Anonymous votes: option counts shown, member identity NOT exposed in results until vote closes
- [ ] Mentorship pairing: mentor can have at most one active mentee
- [ ] Minutes: draft → finalize workflow; finalized minutes are read-only
- [ ] Scholarship applications: status workflow (submitted → under_review → finalist → awarded)

---

### TASK-011: Admin Panel & Sysadmin Console
**Hours:** 24 hrs | **Agent:** Wasp sprint

**Acceptance criteria:**
- [ ] Admin panel: member management (add, edit, change role, deactivate)
- [ ] Admin panel: XP management (award, recalculate, view full log)
- [ ] Admin panel: settings (semester start, dues amounts, point economy config)
- [ ] Sysadmin console: chapter list with subscription status
- [ ] Sysadmin console: audit log (all platform actions, filterable by chapter/actor/date)
- [ ] Sysadmin console: override subscription status for any chapter
- [ ] Sysadmin panel only visible to users with `is_sysadmin = true`

---

### TASK-012: Infrastructure, Email & Production Deploy
**Hours:** 24 hrs | **Agent:** Ant-Man tasks

- Fly.io production deploy (API + Postgres + Redis)
- Cloudflare R2 bucket setup + avatar upload endpoint
- Resend transactional emails: magic link, dues reminder (7 days before due), badge earned, event reminder (24 hrs before)
- Sentry error tracking integration (Go API + React frontend)
- GitHub Actions CI: lint + test + build + deploy to Fly on merge to main
- Production seed script (import prototype data for founding chapter)

---

## 14. Error Catalog

| Error Code | HTTP Status | When | User Message |
|------------|-------------|------|--------------|
| INVALID_CREDENTIALS | 401 | Bad email/password | Invalid email or password |
| ACCOUNT_LOCKED | 423 | Too many failed logins | Account locked. Try again in 15 minutes |
| TOKEN_EXPIRED | 401 | Expired JWT | Session expired. Please sign in again |
| TOKEN_INVALID | 401 | Bad JWT signature | Invalid session |
| MAGIC_LINK_EXPIRED | 401 | Stale magic link | This link has expired. Request a new one |
| MAGIC_LINK_USED | 401 | Already consumed | This link has already been used |
| FORBIDDEN | 403 | Insufficient role | You don't have permission to do that |
| SUBSCRIPTION_REQUIRED | 402 | Subscription lapsed | Chapter subscription is inactive. Contact your admin |
| NOT_FOUND | 404 | Resource doesn't exist | Not found |
| DUPLICATE_RSVP | 409 | Already RSVP'd | You've already RSVP'd to this event |
| DUPLICATE_CHECKIN | 409 | Already checked in | This member is already checked in |
| DUPLICATE_VOTE | 409 | Already voted | You've already voted on this item |
| INSUFFICIENT_XP | 400 | Store purchase | Not enough XP for this item |
| VALIDATION_ERROR | 422 | Input validation fails | [Field-specific message from validator] |
| RATE_LIMITED | 429 | Too many requests | Too many requests. Slow down |
| INTERNAL_ERROR | 500 | Unexpected error | Something went wrong. Please try again |

---

## 15. Security Considerations

1. **SQL Injection:** Not possible — sqlc generates parameterized queries only. No string interpolation in SQL.
2. **XSS:** React escapes by default. Content from the API (announcements, notes) rendered with React, never `dangerouslySetInnerHTML`.
3. **CSRF:** Not needed for JWT Bearer auth (no cookies). If cookies are ever added, add CSRF token.
4. **Rate Limiting:** Redis-backed, per-IP and per-user-ID. Auth endpoints: 10 req/min per IP. API endpoints: 1000 req/min per user.
5. **RLS Defense-in-Depth:** All service layer functions also filter by `chapter_id` in their WHERE clauses. RLS is the last defense, not the only defense.
6. **Sysadmin Isolation:** Sysadmin sessions do NOT set `app.chapter_id`. The sysadmin service account uses `SET ROLE blue_ledger_admin` (a role with `BYPASSRLS`). This role is NOT accessible from the regular API connection pool.
7. **QR Token Security:** QR scanner tokens are HMAC-SHA256 signed with a server secret. They are event-scoped (include event_id in the payload). A scanner token for Event A cannot be used to check into Event B.
8. **Audit Log:** All destructive admin actions (delete member, recalculate XP, cancel dues, change role) write to `audit_log` with actor, target, and metadata. The audit log is append-only — no UPDATE or DELETE.
9. **Magic Link HMAC:** The raw magic link token is never stored. Only SHA-256(token) is stored. The raw token is only in the email.
10. **Stripe Webhook Validation:** All Stripe webhook payloads are verified using `stripe.ConstructEvent` with the webhook signing secret before processing.

---

## 16. Assumptions

1. The founding chapter (Tau Tau Sigma / ΤΤΣ) is the first tenant. The `member_display_id` prefix stays "ΤΤΣ" for them; new chapters will have their own prefix configured at registration.
2. The prototype's `@chapter.org` email domain is used for mock data only. Production members use their real email addresses.
3. Direct messages (the `messages` feature in the prototype) are deferred to a post-launch sprint. The DM screen can show a "Coming Soon" state at launch.
4. The `sysadmin` account (CES1231) maps to a platform user with `is_sysadmin = true`. The sysadmin has no `chapter_id` JWT claim — their JWT omits it, and the TenantSetter middleware skips RLS setup for sysadmin sessions.
5. The XP Store redemption flow does not involve real money — items are physical (the chapter admin fulfills orders). No Stripe checkout for store items in v1.
6. Alumni are modeled as members with `status = 'alumni'`. There is no separate alumni authentication tier in v1; alumni can sign in if they have an account.
7. Stripe dues payment is opt-in per chapter. Chapters can continue using Zelle/Cash and mark dues paid manually through the admin panel.

---

## 17. Open Questions

1. **Domain name:** Is `blueledger.io` owned? What domain is used for the production API and app?
2. **Chapter onboarding flow:** Does a new chapter sign up self-serve (creates account, starts trial) or does the sysadmin provision them manually? The spec assumes self-serve with a landing page form.
3. **National HQ relationship:** Should Phi Beta Sigma national HQ have any cross-chapter visibility (e.g., see all chapters' service hours for national reporting)? Not in scope for v1, but the schema supports it.
4. **Member invites:** How does a new member get access to their chapter's Blue Ledger? Options: (a) admin adds them + sends magic link, (b) invite link with chapter code. The spec currently assumes option (a). Option (b) adds an `invite_tokens` table.
5. **Semester rollover:** When does `xp_semester` reset, and does this happen automatically or via admin action? Spec assumes admin triggers it via `PUT /settings`.
6. **Quiz feature:** The prototype includes a quiz feature (fraternity history quiz). This was not included in the MVP scope above. Is it in scope for v1?

---

## 18. Notes for AI Agents

If this spec is being executed by an AI coding agent (Iron Man, Wasp, Ant-Man):

- **Follow the table schemas exactly as defined.** Do not add or remove columns without flagging it as a deviation.
- **Every XP-awarding operation MUST be wrapped in a database transaction:** INSERT engagement_log + UPDATE members.xp_total must be atomic.
- **RLS is required on every tenant table.** If you add a new tenant table, add RLS policy immediately in the same migration.
- **Handler functions must set the Postgres session variable before any query:**
  ```go
  if _, err := tx.Exec(ctx, "SET LOCAL app.chapter_id = $1", chapterID); err != nil { ... }
  ```
- **JWT role checking is done in RoleGate middleware, not in service functions.** Service functions receive an already-authenticated context.
- **The badge engine runs asynchronously** — it should not block the API response. Use a goroutine with a short timeout context, or push to a Redis job queue.
- **All amounts (dues, fundraising goals) are stored in integer cents**, never floating point.
- **The React migration preserves the existing UI exactly.** Do not change colors, fonts, or layout. The CSS from `app/index.html` is the source of truth for visual design.
- **Do not write to `audit_log` from tenant-scoped code.** Only sysadmin actions and billing events go there. Tenant-level audit trail is `engagement_log`.

### Agent Hints

| Signal | Value | Agents |
|--------|-------|--------|
| Auth-critical | yes | Hawkeye: deep auth review on TASK-002 |
| Financial data | dues_records, fundraising_campaigns | Hawkeye: data exposure. Falcon: migration safety |
| High-traffic endpoints | GET /leaderboard, GET /members, GET /notifications | Black Panther: benchmark these |
| Database writes | engagement_log (append-only, high volume) | Hulk: deadlock testing on concurrent XP awards |
| State machines | intake_prospects.stage, dues_records.status, scholarships.status | Hulk: invalid transitions |
| Migration | 20+ migrations — first deploy | Falcon: verify rollback scripts. Run in staging first |
| External dependencies | Stripe, Resend, Cloudflare R2 | Vision: health checks. Hulk: failure simulation |
| Multi-tenancy | RLS on all tenant tables | Hawkeye: cross-tenant isolation test required |
| Builder (Phase 0) | iron-man | 3 agents: DB+auth, React foundation, XP engine |
| Builder (Phases 1-5) | wasp | Sequential sprint per phase |
| Builder (Phase 6) | ant-man | Individual polish tasks |
