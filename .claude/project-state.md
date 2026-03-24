# Blue Ledger — Project State

meta:
  last_updated: 2026-03-24
  last_updated_by: eitri
  project: The Blue Ledger
  version: 1.0.0
  project_stage: pre-production

---

## Stack

### Current (Prototype)
- Language: HTML + vanilla JavaScript (single-file SPA)
- Architecture: Static browser app + Glide (no-code) + Google Sheets + Make.com
- Deployment: Netlify (static) for demo; Glide app for production
- Auth: Glide allowlist (production); in-app email lookup (demo prototype)

### Target (SaaS — per saas-architecture.md)
- Backend: Go 1.23 + Echo v4 + sqlc + pgx/v5 + golang-migrate
- Frontend: React 18 + TypeScript + Vite 5 + TanStack Query + Zustand + Tailwind CSS
- Database: PostgreSQL 16 (RLS multi-tenant) + Redis 7
- Auth: JWT RS256 + refresh tokens + magic links (Resend)
- Hosting: Fly.io (API + DB + Redis) + Cloudflare R2 (storage) + Cloudflare CDN
- Payments: Stripe (chapter subscriptions + dues)
- Monitoring: Sentry + Fly.io built-in metrics

---

## Packages (Target Architecture)

| Package | Purpose | Key Functions |
|---------|---------|---------------|
| `internal/auth` | Authentication, JWT, magic links, middleware | Login, Register, RequestMagicLink, ConsumeMagicLink, RefreshTokens, Logout |
| `internal/members` | Member CRUD, profiles, roles | GetMember, ListMembers, UpdateMember, ChangeRole, ChangeStatus |
| `internal/xp` | XP engine, leaderboard, badge engine | AwardXP, RecalcAllXP, GetLeaderboard, CheckAndAwardBadges |
| `internal/events` | Event CRUD, RSVP, QR check-in | CreateEvent, ListEvents, RSVP, QRCheckin, GenerateQRToken |
| `internal/dues` | Dues records, Stripe checkout, webhook | CreateDues, MarkPaid, CreateStripeSession, HandleStripeWebhook |
| `internal/notifications` | Notification creation and management | CreateNotification, MarkRead, MarkAllRead |
| `internal/announcements` | Announcements CRUD | CreateAnnouncement, ListAnnouncements, Pin |
| `internal/props` | Peer recognition + XP | GiveProps, ListProps |
| `internal/service` | Service log + verification | SubmitService, VerifyService |
| `internal/intake` | Prospect pipeline | AddProspect, UpdateStage, ListProspects |
| `internal/votes` | Chapter voting | CreateVote, SubmitResponse, GetResults |
| `internal/mentorship` | Mentorship pairings | RegisterMentor, RequestPairing, UpdatePairing |
| `internal/minutes` | Meeting minutes | CreateMinutes, FinalizeMinutes |
| `internal/scholarships` | Scholarship applications | AddApplication, UpdateStatus |
| `internal/platform` | Sysadmin + billing | ListChapters, OverrideSubscription, GetAuditLog |
| `pkg/db` | pgx pool, RLS session setter | NewPool, SetChapterContext |
| `pkg/email` | Resend transactional email | SendMagicLink, SendDuesReminder, SendBadgeEarned |
| `pkg/storage` | Cloudflare R2 uploads | UploadAvatar, GetPresignedURL |
| `pkg/config` | Environment configuration | Load |
| `pkg/logger` | Zerolog structured logging | New |

---

## Handler Map

| Handler File | Feature Package(s) |
|-------------|-------------------|
| `internal/auth/handler.go` | `internal/auth` |
| `internal/members/handler.go` | `internal/members` |
| `internal/xp/handler.go` | `internal/xp` |
| `internal/events/handler.go` | `internal/events` |
| `internal/dues/handler.go` | `internal/dues` |
| `internal/notifications/handler.go` | `internal/notifications` |
| `internal/announcements/handler.go` | `internal/announcements` |
| `internal/props/handler.go` | `internal/props` |
| `internal/service/handler.go` | `internal/service` |
| `internal/intake/handler.go` | `internal/intake` |
| `internal/votes/handler.go` | `internal/votes` |
| `internal/mentorship/handler.go` | `internal/mentorship` |
| `internal/minutes/handler.go` | `internal/minutes` |
| `internal/scholarships/handler.go` | `internal/scholarships` |
| `internal/platform/sys_handler.go` | `internal/platform` |
| `internal/platform/billing_handler.go` | `internal/platform` |

---

## Database Schema (Target)

Latest migration: `020_study_groups_and_milestones`
Migration tool: golang-migrate
Migration directory: blue-ledger-api/migrations/
Total migrations: 20 (001–020), 40 files (up + down)
Migration status: WRITTEN — not yet applied to a live database
Last authored: 2026-03-24 by nebula
Migration log: .claude/nebula/migration-log.md

| Table | Tenant-Scoped | Key Columns | RLS |
|-------|--------------|-------------|-----|
| `chapters` | No (root) | id, name, stripe_customer_id, subscription_status, plan_tier, settings | No |
| `users` | No (cross-chapter identity) | id, email, password_hash, is_sysadmin | No |
| `members` | Yes | id, chapter_id, user_id, role, xp_total, dues_status, status | Yes |
| `auth_sessions` | No | id, user_id, refresh_token (hashed), expires_at | No |
| `magic_link_tokens` | No | id, user_id, token_hash, expires_at, used_at | No |
| `events` | Yes | id, chapter_id, name, event_type, xp_attend, qr_code_token | Yes |
| `rsvps` | Yes | id, chapter_id, event_id, member_id, status, attended | Yes |
| `engagement_log` | Yes | id, chapter_id, member_id, activity, xp_awarded, source (append-only) | Yes |
| `service_log` | Yes | id, chapter_id, member_id, hours, verified, verified_by | Yes |
| `dues_records` | Yes | id, chapter_id, member_id, semester, amount_cents, status | Yes |
| `badges` | Yes | id, chapter_id, name, rarity, criteria (JSONB) | Yes |
| `member_badges` | Yes | id, chapter_id, member_id, badge_id, awarded_at | Yes |
| `announcements` | Yes | id, chapter_id, title, body, is_pinned | Yes |
| `props` | Yes | id, chapter_id, from_id, to_id, category, xp_awarded | Yes |
| `quests` | Yes | id, chapter_id, title, steps (JSONB) | Yes |
| `member_quest_progress` | Yes | id, chapter_id, member_id, quest_id, progress (JSONB) | Yes |
| `notifications` | Yes | id, chapter_id, user_id, type, is_read | Yes |
| `mentorships` | Yes | id, chapter_id, mentor_id, mentee_id | Yes |
| `intake_prospects` | Yes | id, chapter_id, name, stage, interest_level | Yes |
| `votes` | Yes | id, chapter_id, title, vote_type, options (TEXT[]) | Yes |
| `vote_responses` | Yes | id, chapter_id, vote_id, member_id, option_index | Yes |
| `chapter_minutes` | Yes | id, chapter_id, meeting_date, body, status | Yes |
| `scholarships` | Yes | id, chapter_id, applicant_name, status | Yes |
| `fundraising_campaigns` | Yes | id, chapter_id, name, goal_cents, raised_cents | Yes |
| `store_items` | Yes | id, chapter_id, name, xp_cost | Yes |
| `store_orders` | Yes | id, chapter_id, member_id, item_id, xp_spent | Yes |
| `chapter_goals` | Yes | id, chapter_id, title, target, current, status | Yes |
| `job_board` | Yes | id, chapter_id, company, job_title | Yes |
| `point_economy` | Yes | id, chapter_id, activity, xp | Yes |
| `audit_log` | No (platform-wide) | id, actor_id, chapter_id, action, metadata | No |
| `message_threads` | Yes | id, chapter_id, participant_ids UUID[] | Yes |
| `messages` | Yes | id, thread_id, chapter_id, sender_id, body, read_by UUID[] | Yes |
| `fundraising_campaigns` | Yes | id, chapter_id, title, goal_cents, current_cents | Yes |
| `committees` | Yes | id, chapter_id, name, chair_id, member_ids UUID[] | Yes |
| `study_groups` | Yes | id, chapter_id, topic, host_id, date, member_ids UUID[] | Yes |
| `milestones` | Yes | id, chapter_id, member_id, type, title, date | Yes |
| `resources` | Yes | id, chapter_id, title, type, url, file_key, tags TEXT[] | Yes |
| `chapter_config` | Yes | id, chapter_id, key, value | Yes |
| `point_economy` | Yes | id, chapter_id, activity, xp, is_active | Yes |

---

## Existing Prototype Data Model (app/index.html)

Key JavaScript constants (in-memory):
- `MEMBERS` — 7 active members + 1 sysadmin (CES1231)
- `ENGAGEMENT_LOG` — XP audit trail (currently partial — scanner/RSVP XP not always logged)
- `SERVICE_LOG` — service hours with verification state
- `EVENTS` + `RSVP_RESPONSES` — 4 events
- `PROPS_DATA` — peer recognition (10 XP per prop)
- `ANNOUNCEMENTS` — 4 announcements
- `BADGES` — 10 badge definitions
- `POINT_ECONOMY` — 10 activity types with XP values
- `MENTORSHIP` — 4 mentorship pairs
- `INTAKE_DATA` — 4 prospects in pipeline
- `VOTES` — 3 vote items
- `STORE_ITEMS` — 8 items (XP redemption)
- `SCHOLARSHIPS` — 4 applications
- `FUNDRAISING` — 3 campaigns
- `COMMITTEES` — 4 committees
- `JOB_BOARD` — 3 job postings
- `CHAPTER_GOALS` — 6 goals
- `SIGMA_BETA` — SBC log
- `ALUMNI_DATA` — 4 alumni profiles
- `NOTIFICATIONS_DATA` — 8 notification types
- `CHAPTER_HISTORY` — 7 historical milestones
- `MINUTES_DATA` — 2 meeting minutes records

---

## Auth & Middleware (Target)

- JWT type: RS256, 15-minute access tokens
- Refresh tokens: opaque, 30-day, stored hashed in `auth_sessions`
- Magic links: HMAC-SHA256 signed, 15-minute expiry, single-use
- Password hashing: bcrypt cost 12
- Account lockout: 15 min after 10 failed attempts
- Middleware stack: RateLimiter → RequestLogger → CORS → JWTMiddleware → TenantSetter → RoleGate → Handler
- RLS session variable: `SET LOCAL app.chapter_id = '<uuid>'` on every request
- Sysadmin: `is_sysadmin = true` on `users` table; `role = 'sysadmin'` in JWT; bypasses RLS via service account with BYPASSRLS

---

## External Dependencies (Target)

| Service | Purpose | Health Check |
|---------|---------|--------------|
| Fly.io Postgres | Primary database | `SELECT 1` |
| Fly.io Redis | Sessions, rate limiting, job queue | `PING` |
| Resend | Transactional email (magic links, reminders) | API status endpoint |
| Stripe | Chapter subscriptions + dues payments | Webhook event handler |
| Cloudflare R2 | Avatar and resource file storage | S3 HeadBucket |
| Sentry | Error monitoring (Go API + React) | SDK ping |
| Cloudflare CDN | Static assets + DDoS protection | — |

### Retained from Prototype (to be replaced)
| Service | Status |
|---------|--------|
| Glide | Replacing with React frontend |
| Google Sheets | Replacing with PostgreSQL |
| Make.com | Replacing with Go background jobs |
| Cloudinary | Replacing with Cloudflare R2 |
| Stripe | Retained, expanded to subscriptions |

---

## Architectural Decisions

| Decision | Choice | Rationale |
|---------|--------|-----------|
| Multi-tenancy model | RLS + chapter_id FK on all tenant tables | Single schema, simple migrations, proven pattern |
| API language | Go + Echo | Single binary deploy, low memory, strong stdlib |
| Frontend | React 18 + TypeScript + Vite | Preserves existing UI; TypeScript prevents API contract drift |
| Token storage | access_token in memory (Zustand), refresh_token in localStorage | Balances XSS risk vs UX (no refresh on tab close) |
| XP storage | Denormalized cache on members.xp_total + source-of-truth in engagement_log | Fast leaderboard queries; RecalcXP for drift correction |
| Amount storage | Integer cents everywhere | No floating point rounding errors |
| Badge criteria | JSONB on badges.criteria | No schema migration needed to add badge types |
| Chapter settings | JSONB on chapters.settings | XP level thresholds, notification prefs, etc. without migrations |

---

## Roles

| Role | Key Access |
|------|-----------|
| member | Own profile, events, dues, quests, props, notifications |
| chair | + QR scanner, service verification, intake pipeline, event creation |
| pia | + PIA builder, scholarship management, chapter health analytics |
| admin | Everything within chapter (member management, XP, settings, billing) |
| sysadmin | Platform-level (all chapters, audit log, billing override) — no chapter isolation |

---

## Task History

| Task ID | Date | Title | Packages | Status |
|---------|------|-------|----------|--------|
| saas-architecture | 2026-03-24 | SaaS Architecture Spec | All | Specified |
| TASK-001 | 2026-03-24 | Project Setup & DB Foundation | pkg/db, migrations | Specified |
| TASK-002 | 2026-03-24 | Authentication Service | internal/auth | Specified |
| TASK-003 | 2026-03-24 | React Frontend Foundation | web/src | Specified |
| TASK-004 | 2026-03-24 | Members API & Screens | internal/members | Specified |
| TASK-005 | 2026-03-24 | XP Engine & Leaderboard | internal/xp | Specified |
| TASK-006 | 2026-03-24 | Events API & QR Check-In | internal/events | Specified |
| TASK-007 | 2026-03-24 | Dues & Stripe Integration | internal/dues, internal/platform | Specified |
| TASK-008 | 2026-03-24 | Engagement Features | internal/announcements, internal/props, internal/notifications | Specified |
| TASK-009 | 2026-03-24 | Service Log, Badges & Quests | internal/service, internal/xp (badge engine) | Specified |
| TASK-010 | 2026-03-24 | Phase 4 Remaining Features | internal/intake, internal/votes, internal/mentorship, internal/minutes, internal/scholarships | Specified |
| TASK-011 | 2026-03-24 | Admin Panel & Sysadmin Console | internal/platform | Specified |
| TASK-012 | 2026-03-24 | Infrastructure & Production Deploy | Fly.io, R2, Sentry, Resend, CI | Specified |

---

## Observability Status (owned by Vision)

Last scan: 2026-03-24
Verdict: NOT READY for production as instrumented

### Logging
- No structured logging framework
- No console.error calls anywhere in codebase
- SYS_AUDIT_LOG exists but is in-memory only (destroyed on reload)
- Audit log timestamps all read "Just now" — no real timestamps

### Error Handling
- No global error boundary (window.onerror / unhandledrejection)
- Silent fallbacks throughout (optional chaining, ||MEMBERS[0])
- Login silently falls back to admin account on unknown email

### Metrics
- No metrics framework
- No analytics (no screen view tracking, no event tracking)
- No performance observer

### Health Checks
- All integration status in System Console is mocked/static
- "Test Connection" buttons are UI stubs — no real connectivity probe

### Critical Gaps
1. No global JS error boundary — exceptions crash silently
2. Login falls back to admin on bad email — security risk if exposed
3. In-memory audit log — destroyed on reload, no accountability
4. Danger zone ops have no server-side verification
5. XP mutations (scanner, RSVP, props, mentorship) do not write to engLog
6. Client-side-only role enforcement — bypassable via browser console

### Report
See: .claude/vision/observability-report.md

---

## Infrastructure Status (owned by Eitri)

Last built: 2026-03-24

### Containers

| Service | Dockerfile | Base Image (build) | Base Image (runtime) | Health Check |
|---------|-----------|-------------------|---------------------|--------------|
| blue-ledger-api | blue-ledger-api/Dockerfile | golang:1.23-alpine | alpine:3.20 | GET /v1/healthz |

### Orchestration

| Tool | File | Services |
|------|------|----------|
| Docker Compose (local dev) | blue-ledger-api/docker-compose.yml | postgres:16-alpine, redis:7-alpine |
| Fly.io (production) | blue-ledger-api/fly.toml | shared-cpu-1x, 256MB, region: iad |

### Cloud Resources

| Resource | Provider | Status |
|----------|----------|--------|
| Postgres | Fly.io Postgres | Config in fly.toml — not provisioned |
| Redis | Fly.io Redis | Config in fly.toml — not provisioned |
| Object storage | Cloudflare R2 | Client in pkg/storage/r2.go — bucket not created |
| Email | Resend | Client in pkg/email/resend.go — API key required |
| Stripe | Stripe Billing | Client in internal/dues/stripe.go — keys required |

### Networking

| Concern | Status |
|---------|--------|
| TLS | Fly.io force_https = true |
| CORS | Configurable via ALLOWED_ORIGINS env var |
| Rate limiting | Echo RateLimiter middleware, config via RATE_LIMIT_RPS/BURST |
| Reverse proxy | Fly.io built-in (no nginx needed) |

### Frontend

| Concern | Status |
|---------|--------|
| Dev server | Vite on :3001, proxy /api → localhost:8080/v1 |
| API client | Axios singleton + 401 refresh queue (src/api/client.ts) |
| State | Zustand (auth + UI), TanStack Query (server state) |
| Routes | 47 routes defined in src/App.tsx |
| CSS | Full prototype CSS ported to src/styles/base.css |

### Secrets

| Secret | Dev | Staging/Prod |
|--------|-----|-------------|
| JWT RSA keypair | Generate locally → keys/ (gitignored) | Fly.io secrets |
| DATABASE_URL | .env (local docker-compose) | Fly.io secrets |
| REDIS_URL | .env (local docker-compose) | Fly.io secrets |
| STRIPE_SECRET_KEY | .env | Fly.io secrets |
| RESEND_API_KEY | .env | Fly.io secrets |
| R2 credentials | .env | Fly.io secrets |

### Health Checks

| Service | Endpoint | Checks |
|---------|----------|--------|
| Go API | GET /v1/healthz | DB ping, Redis ping |
| Postgres (compose) | pg_isready -U blue_ledger_app | — |
| Redis (compose) | redis-cli ping | — |

### Environment Parity

See: .env.example (root), web/.env.example

Key differences dev vs prod:
- APP_ENV: development → production
- DATABASE_URL: local docker → Fly.io Postgres URL
- ALLOWED_ORIGINS: localhost:3001 → production frontend domain
- JWT keys: local file paths → Fly.io secrets (base64 encoded)
- MAGIC_LINK_BASE_URL: localhost:3001/magic → production URL

---

## Drift Log

| Date | Agent | Note |
|------|-------|------|
| 2026-03-24 | vision | Initial scan. No prior state file existed. All sections populated by Vision from direct codebase analysis. |
| 2026-03-24 | jarvis | SaaS architecture spec generated. Target stack defined: Go + Echo + PostgreSQL (RLS) + React + Fly.io. 12 tasks specified across 6 phases. ~280-340 hrs estimated. See .claude/specs/saas-architecture.md |
| 2026-03-24 | nebula | Full migration suite written: 20 migrations (001-020), 40 files. 31 tables. golang-migrate format. RLS enabled on all tenant tables. Roles: blue_ledger_app + blue_ledger_admin. seed_default_point_economy() function included. See .claude/nebula/migration-log.md |
| 2026-03-24 | eitri | Full application scaffold built. Go API: all core packages (auth, members, xp, events, dues, notifications), migrations, Dockerfile, docker-compose, fly.toml. React SPA: 47 routes, 4 feature pages (dashboard, directory, member profile, leaderboard), auth flow, full CSS port from prototype, .env.example files. See .claude/eitri/build-report.md |
