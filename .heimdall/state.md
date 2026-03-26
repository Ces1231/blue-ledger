# HEIMDALL STATE FILE
**Project:** The Blue Ledger
**Indexed:** 2026-03-26
**Status:** ONLINE

---

## 1. PROJECT IDENTITY

| Field | Value |
|---|---|
| Name | The Blue Ledger |
| Short Name | Blue Ledger |
| Purpose | Chapter Engagement & Gamification Platform for Phi Beta Sigma |
| Version | 1.0.0 (released 2025-03-20) |
| Owner / System Identity | CES1231 |
| License | MIT |
| Theme Color | `#001A4D` (navy) · `#C9A84C` (gold) · `#1A6B3A` (green) |
| PWA Manifest | `manifest.json` — standalone display, portrait |

---

## 2. REPOSITORY STRUCTURE

```
blue-ledger-master/
├── app/index.html              — Legacy static prototype (47 screens, 290KB)
├── assets/avatars/             — 5 avatar PNGs (neo → icon)
├── data/                       — MasterSheet.xlsx (15-tab Google Sheet), QuizBank.xlsx
├── docs/MasterPlan.html        — 12-week execution plan
├── guides/                     — Week 3–12 guides + Year1 enhancements (HTML)
├── web/                        — React/TypeScript frontend (Vite + Tailwind)
├── blue-ledger-api/            — Go REST API (Echo framework)
├── manifest.json               — PWA manifest
├── CHANGELOG.md                — v1.0.0 release notes
├── DEPLOYMENT.md               — Netlify drop / Glide / full-stack deploy options
├── README.md                   — Full project overview (383 lines)
├── setup-github.sh             — Helper push script
└── .heimdall/state.md          — This file
```

---

## 3. TECH STACK

### Frontend — `web/`
| Layer | Technology |
|---|---|
| Framework | React 18.3 + TypeScript 5.5 |
| Build | Vite 5.3 |
| Router | React Router DOM v6 |
| Data fetching | TanStack React Query v5 |
| HTTP client | Axios (with auto token-refresh interceptor) |
| State management | Zustand 4 (`authStore`, `uiStore`) |
| Forms | React Hook Form v7 |
| Styling | Tailwind CSS 3.4 + PostCSS |
| QR Codes | qrcode.react v4 |
| Testing | Vitest + Testing Library + jsdom |
| Containerization | Nginx (Dockerfile in `web/`) |

### Backend — `blue-ledger-api/`
| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP Framework | Echo v4 |
| Database | PostgreSQL 16 (pgx/v5 driver, pgxpool) |
| Cache / Rate limiting | Redis 7 (go-redis/v9) |
| Auth | RS256 JWT (golang-jwt/v5) + Magic Links (HMAC-SHA256) |
| Migrations | golang-migrate/v4 (file source) |
| Email | Resend (resendlabs/resend-go) |
| Storage | Cloudflare R2 (aws-sdk-go-v2 / S3-compatible) |
| Payments | Stripe (stripe-go/v78) |
| AI | Anthropic Claude + OpenAI (streaming SSE) |
| Validation | go-playground/validator v10 |
| Logging | zerolog (structured JSON) |
| Environment | godotenv (.env support) |
| Containerization | Docker (multi-stage Dockerfile) |

### Infrastructure
| Layer | Technology |
|---|---|
| API hosting | Fly.io (`blue-ledger-api`, region `iad`) |
| Database | Fly Postgres or external |
| Redis | Fly Redis or external |
| Storage | Cloudflare R2 |
| Email | Resend |
| Payments | Stripe (SaaS subscription billing) |
| Demo / static | Netlify drop |

---

## 4. DATABASE SCHEMA — 21 Migrations

All tables use `gen_random_uuid()` PKs. RLS (`app.chapter_id` setting) enforces tenant isolation on all chapter-scoped tables. `blue_ledger_app` = app role; `blue_ledger_admin` = sysadmin / bypass-RLS role.

### Core / Auth (001–002)
| Table | Key Columns | Notes |
|---|---|---|
| `chapters` | id, name, greek_letters, city, state_code, university, member_id_prefix, semester_start, zeffy_form_id, stripe_customer_id, stripe_subscription_id, subscription_status, plan_tier, settings (JSONB) | Tenant root. Plan tiers: `starter`, `growth`, `chapter_pro`. Status: `trialing/active/past_due/canceled/paused`. |
| `users` | id, email, email_verified, password_hash, first_name, last_name, is_sysadmin, failed_login_count, locked_until | Cross-chapter identity. No chapter_id. No RLS. |
| `auth_sessions` | id, user_id, token_hash, chapter_id, member_id, device_hint, ip_address, expires_at, revoked_at | Refresh token store. One active session per device. |
| `magic_link_tokens` | id, user_id, token_hash, chapter_id, used_at, expires_at | HMAC-SHA256 signed, single-use. |

### Members (003)
| Table | Key Columns | Notes |
|---|---|---|
| `members` | id, chapter_id, user_id, display_id, name, role (`member/chair/pia/admin/sysadmin`), status (`active/inactive/alumni/suspended/pledging`), xp_total, xp_semester, level_key (`neo/scholar/leader/sage/legend/icon`), streak, dues_status, privacy_settings (JSONB) | Per-chapter profile. UNIQUE(chapter_id, user_id). |

### XP / Engagement (004)
| Table | Key Columns | Notes |
|---|---|---|
| `engagement_log` | id, chapter_id, member_id, activity, xp_awarded, source (`checkin/rsvp/admin/system/quiz/service/dues/badge/props/mentorship`), reference_id, reference_type, awarded_by, semester | Append-only XP audit trail. Source of truth for XP calculations. |

### Events / RSVPs (005)
| Table | Key Columns | Notes |
|---|---|---|
| `events` | id, chapter_id, name, event_type (`Meeting/Service/Social/Conference/Committee/SBC/Other`), event_date, xp_attend (default 50), xp_rsvp (default 10), qr_code_token, check_in_window_minutes (120) | QR check-in via signed token. |
| `rsvps` | id, chapter_id, event_id, member_id, status (`yes/no/maybe`), checked_in (bool), checked_in_at, scanned_by | UNIQUE(event_id, member_id). |

### Dues (006)
| Table | Key Columns | Notes |
|---|---|---|
| `dues_records` | id, chapter_id, member_id, semester, amount_cents, due_date, paid_at, payment_method (`zeffy/cash/zelle/venmo/check/waived/admin_override`), zeffy_form_id, zeffy_transaction_id, status, xp_awarded | UNIQUE(chapter_id, member_id, semester). Zeffy = 0% fee dues collection. |

### Service Log (007)
| Table | Key Columns | Notes |
|---|---|---|
| `service_log` | id, chapter_id, member_id, description, hours, date, approved_by, approved_at, xp_awarded | XP awarded on approval. |

### Badges & Quests (008)
| Table | Key Columns | Notes |
|---|---|---|
| `badges` | id, chapter_id, name, icon, category, description, requirement, xp_reward, rarity (`common/uncommon/rare/legendary`), criteria (JSONB) | Flexible badge engine criteria. |
| `member_badges` | id, chapter_id, member_id, badge_id, awarded_at, awarded_by | UNIQUE(member_id, badge_id). |
| `quests` | id, chapter_id, name, description, steps (JSONB array), xp_reward, is_active | Multi-step quests. |
| `member_quest_progress` | id, chapter_id, member_id, quest_id, completed_steps (JSONB), completed_at | Per-member progress tracking. |

### Announcements & Props (009)
| Table | Key Columns | Notes |
|---|---|---|
| `announcements` | id, chapter_id, member_id (author), title, body, pinned | Chapter-wide announcements. |
| `props` | id, chapter_id, from_member_id, to_member_id, message | Peer recognition; triggers XP. |

### Notifications (010)
| Table | Key Columns | Notes |
|---|---|---|
| `notifications` | id, chapter_id, member_id, type, title, body, read_at, action_url | Push-style in-app notifications. |

### Mentorship & Intake (011)
| Table | Key Columns | Notes |
|---|---|---|
| `mentorships` | id, chapter_id, mentor_id, mentee_id, status, started_at | Paired mentorship tracking. |
| `intake_prospects` | id, chapter_id, first_name, last_name, email, phone, status, assigned_to, notes | Prospect pipeline for new member intake. |

### Votes (012)
| Table | Key Columns | Notes |
|---|---|---|
| `votes` | id, chapter_id, title, description, type (`yea_nay/ranked/multiple_choice`), status, closes_at | Chapter voting. |
| `vote_responses` | id, chapter_id, vote_id, member_id, choice, ranked_choices | UNIQUE(vote_id, member_id). |

### Minutes & Scholarships (013)
| Table | Key Columns | Notes |
|---|---|---|
| `meeting_minutes` | id, chapter_id, event_id, recorded_by, content, approved_at | Meeting minutes. |
| `scholarships` | id, chapter_id, name, amount_cents, deadline, status | Scholarship definitions + applications. |
| `scholarship_applications` | id, scholarship_id, chapter_id, member_id, status, submitted_at | Application tracking. |

### Store & Goals (014)
| Table | Key Columns | Notes |
|---|---|---|
| `store_items` | id, chapter_id, name, description, xp_cost, quantity, category | XP store catalog. |
| `store_redemptions` | id, chapter_id, member_id, item_id, redeemed_at | Tracks XP spends. |
| `chapter_goals` | id, chapter_id, title, description, target_value, current_value, unit, due_date, status | Chapter-wide goals. |

### Job Board & Resources (015)
| Table | Key Columns | Notes |
|---|---|---|
| `job_postings` | id, chapter_id, posted_by, title, company, location, url, description, expires_at | Brother-shared job opportunities. |
| `resources` | id, chapter_id, posted_by, title, url, category, description | Shared chapter resources. |

### Point Economy & Config (016)
| Table | Key Columns | Notes |
|---|---|---|
| `point_economy` | id, chapter_id, activity, xp, category, is_active | Per-chapter configurable XP per activity. UNIQUE(chapter_id, activity). |
| `chapter_config` | id, chapter_id, key, value, description | Key/value feature flags and settings. UNIQUE(chapter_id, key). |

**Function:** `seed_default_point_economy(p_chapter_id UUID)` — idempotent, seeds 12 default XP activities on chapter creation.

### Audit Log (017)
| Table | Key Columns | Notes |
|---|---|---|
| `audit_log` | id, actor_id, actor_type (`user/system/webhook`), chapter_id, action, target_type, target_id, metadata (JSONB), ip_address | Platform-wide. No RLS. Sysadmin-readable only. App role can INSERT only. |

### Messaging (018)
| Table | Key Columns | Notes |
|---|---|---|
| `conversations` | id, chapter_id, participants (UUID[]) | Direct message threads. |
| `messages` | id, chapter_id, conversation_id, sender_id, content, read_by (UUID[]) | DM message store. |

### Fundraising & Committees (019)
| Table | Key Columns | Notes |
|---|---|---|
| `fundraising_campaigns` | id, chapter_id, name, goal_cents, raised_cents, status, deadline | Campaign tracking. |
| `fundraising_donations` | id, campaign_id, chapter_id, member_id, amount_cents, donor_name, message | Donation records. |
| `committees` | id, chapter_id, name, chair_id, description, members (UUID[]) | Chapter committees. |

### Study Groups & Milestones (020)
| Table | Key Columns | Notes |
|---|---|---|
| `study_groups` | id, chapter_id, name, subject, organizer_id, members (UUID[]), meeting_schedule | Study group tracking. |
| `milestones` | id, chapter_id, title, description, achieved_at, celebrated | Chapter milestone history. |

### AI Assistant (021)
| Table | Key Columns | Notes |
|---|---|---|
| `ai_configs` | id, chapter_id, provider (`claude/openai`), model (default `claude-sonnet-4-6`), system_prompt, enabled | One row per chapter. UNIQUE(chapter_id). |
| `ai_messages` | id, chapter_id, member_id, role (`user/assistant`), content, provider, model, created_at | Append-only chat history. |

---

## 5. BACKEND — API ARCHITECTURE

### Module
`github.com/ces1231/blue-ledger-api`

### Entry Point
`blue-ledger-api/cmd/server/main.go` — loads config, connects DB + Redis, runs migrations, registers all handlers, starts Echo with graceful shutdown.

### Pattern — 3-Layer per Domain
Each domain under `internal/` follows:
```
handler.go      — Echo HTTP binding, request/response types, route registration
service.go      — Business logic interface + implementation
repository.go   — Raw SQL queries (pgx/v5 directly, no ORM)
```

### Auth Architecture
- **RS256 JWT** — 2048-bit RSA key pair (`keys/private.pem` / `keys/public.pem`; generated via `make keys`)
- **Access token:** 15 min expiry (configurable)
- **Refresh token:** 720h (30 days) expiry; stored hashed in `auth_sessions`
- **Magic links:** HMAC-SHA256 signed token, single-use, time-limited
- **Account lockout:** `failed_login_count` + `locked_until` on `users`
- **JWT Claims:** `user_id`, `chapter_id`, `member_id`, `role`, `email`, `is_sysadmin`
- **Middleware:** `JWTMiddleware(tokenManager)` — validates Bearer token, stores claims in Echo context

### Domain Services (all registered in main.go)
| Package | Route Prefix | Key Operations |
|---|---|---|
| `auth` | `/v1/auth` | register, login, magic-link, refresh, logout, me |
| `members` | `/v1/members` | CRUD, XP history |
| `xp` | `/v1` | leaderboard, award XP, recalculate, engagement log |
| `events` | `/v1/events` | CRUD, RSVP, QR token generation, QR check-in |
| `dues` | `/v1/dues` + `/v1/webhooks` | CRUD, mark paid, Zeffy webhook, Stripe subscription webhook |
| `notifications` | `/v1/notifications` | list, mark read |
| `badges` | `/v1/badges` | CRUD, manual award |
| `quests` | `/v1/quests` | CRUD, progress tracking |
| `announcements` | `/v1/announcements` | CRUD |
| `props` | `/v1/props` | give props (triggers XP via XPService) |
| `servicelog` | `/v1/service` | log hours, approve (triggers XP) |
| `intake` | `/v1/intake` | prospect pipeline CRUD |
| `votes` | `/v1/votes` | create votes, submit responses |
| `mentorship` | `/v1/mentorship` | pair mentor/mentee (triggers XP) |
| `minutes` | `/v1/minutes` | CRUD meeting minutes |
| `scholarships` | `/v1/scholarships` | manage scholarships + applications |
| `store` | `/v1/store` | catalog, redeem items |
| `goals` | `/v1/goals` | chapter goals CRUD |
| `messages` | `/v1/messages` | conversations + messages |
| `health` | `/v1/health` | chapter health dashboard |
| `settings` | `/v1/settings` | chapter + user settings |
| `committees` | `/v1/committees` | CRUD |
| `fundraising` | `/v1/fundraising` | campaigns + donations |
| `resources` | `/v1/resources` | CRUD |
| `jobboard` | `/v1/job-board` | CRUD |
| `alumni` | `/v1/alumni` | alumni network |
| `sbc` | `/v1/sbc` | SBC-specific features |
| `ai` | `/v1/ai` | chat (streaming SSE), config CRUD, history |
| `platform` (sys) | `/v1/sys` | sysadmin: chapters, users, audit log, XP recalc |
| `platform` (billing) | `/v1/billing` | Stripe billing portal, subscription management |

### Middleware Stack
1. `RequestID` — attach X-Request-ID
2. `RecoverWithConfig` — panic recovery + zerolog
3. `CORSWithConfig` — configurable allowed origins
4. `LoggerWithConfig` — JSON request logs
5. `RateLimiter` — 100 req/min per IP (Redis-backed; no-op if Redis unavailable)
6. `JWTMiddleware` — per-route auth guard

### Config (`pkg/config/config.go`)
Loaded from env vars with defaults:
- `APP_ENV`, `PORT` (8080), `API_BASE_URL`
- `DATABASE_URL`, `REDIS_URL`
- `JWT_PRIVATE_KEY_PATH`, `JWT_PUBLIC_KEY_PATH`, `JWT_ACCESS_EXPIRY` (15m), `JWT_REFRESH_EXPIRY` (720h)
- `MAGIC_LINK_HMAC_SECRET`, `MAGIC_LINK_EXPIRY`
- `RESEND_API_KEY`, `EMAIL_FROM`, `EMAIL_FROM_NAME`
- `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`, `R2_PUBLIC_URL`
- `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_STARTER_PRICE_ID`, `STRIPE_GROWTH_PRICE_ID`, `STRIPE_PRO_PRICE_ID`
- `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`
- `SENTRY_DSN`, `ALLOWED_ORIGINS`
- `RATE_LIMIT_PER_MINUTE_IP`, `RATE_LIMIT_PER_MINUTE_USER`

### AI Service
- Dual provider: **Claude** (Anthropic) + **OpenAI** — streaming SSE via `StreamClaude` / `StreamOpenAI`
- Default model: `claude-sonnet-4-6`
- Per-chapter config stored in `ai_configs` (provider, model, system_prompt, enabled)
- `BuildContext(chapterID)` — assembles chapter context string for injection into system prompt
- Chat history persisted in `ai_messages`

### XP Engine
- `XPService.AwardXP()` — writes to `engagement_log`, updates `members.xp_total` + `xp_semester`
- `XPService.RecalculateAll()` — full recalculation from `engagement_log` source of truth
- Level keys: `neo` (0–499) → `scholar` (500–999) → `leader` (1000–1499) → `sage` (1500–2499) → `legend` (2500+) → `icon`
- Used cross-domain by: `props`, `servicelog`, `mentorship` (all inject XPService)

---

## 6. FRONTEND — REACT APP ARCHITECTURE

### Entry
`web/src/main.tsx` → `web/src/App.tsx`

### Directory Structure
```
src/
├── api/           — Axios API modules (one file per domain)
│   └── client.ts  — apiClient with auto token-refresh interceptor
├── components/    — Shared UI: Badge, Button, Card, Modal, ProtectedRoute,
│                    Sidebar, Toast, Topbar
├── features/      — Page-level feature folders (35+ modules)
├── hooks/         — useAuth, useCurrentMember
├── stores/        — authStore (Zustand + persist), uiStore (Zustand)
├── styles/        — Global CSS
├── types/         — Shared TypeScript interfaces (index.ts, 331 lines)
└── test/          — Test utilities
```

### Auth Flow
1. `authStore` (Zustand + `persist` middleware) — persists `refreshToken` + `user` to `localStorage` (`bl-auth` key); `accessToken` kept **in-memory only** (XSS defense)
2. `apiClient` request interceptor — attaches `Authorization: Bearer <accessToken>`
3. `apiClient` response interceptor — on 401, silently refreshes via `POST /auth/refresh`, retries original request; queues concurrent requests during refresh
4. `useAuth()` hook — convenience wrapper exposing `user`, `role`, `isAdmin`, `isChair`, `isSysadmin`, `memberID`, `chapterID`, `login`, `logout`
5. `ProtectedRoute` — wraps routes with auth check; `requiredRole` prop enforces role gating

### Route Map
| Path | Component | Role Guard |
|---|---|---|
| `/login` | LoginPage | Public |
| `/magic` | MagicLinkPage | Public |
| `/dashboard` | DashboardPage | Auth |
| `/profile` | MemberProfileSelf → `/members/:id` | Auth |
| `/digital-id` | DigitalIDPage | Auth |
| `/notifications` | NotificationsPage | Auth |
| `/members` | DirectoryPage | Auth |
| `/members/:id` | MemberProfilePage | Auth |
| `/events` | EventsPage | Auth |
| `/events/:id` | EventDetailPage | Auth |
| `/scanner` | ScannerPage | `chair+` |
| `/leaderboard` | LeaderboardPage | Auth |
| `/service` | ServicePage | Auth |
| `/dues` | DuesPage | Auth |
| `/quests` | QuestsPage | Auth |
| `/announcements` | AnnouncementsPage | Auth |
| `/props` | PropsPage | Auth |
| `/votes` | VotesPage | Auth |
| `/mentorship` | MentorshipPage | Auth |
| `/minutes` | MinutesPage | Auth |
| `/scholarships` | ScholarshipsPage | Auth |
| `/fundraising` | FundraisingPage | Auth |
| `/store` | StorePage | Auth |
| `/goals` | ChapterGoalsPage | Auth |
| `/job-board` | JobBoardPage | Auth |
| `/committees` | CommitteesPage | Auth |
| `/study-groups` | StudyGroupsPage | Auth |
| `/messages` | MessagesPage | Auth |
| `/resources` | ResourcesPage | Auth |
| `/milestones` | MilestonesPage | Auth |
| `/alumni` | AlumniPage | Auth |
| `/assistant` | AssistantPage | Auth |
| `/intake` | IntakePage | `chair+` |
| `/admin` | AdminPage | `admin` |
| `/admin/members` | AdminMembersPage | `admin` |
| `/admin/events` | AdminEventsPage | `admin` |
| `/admin/dues` | AdminDuesPage | `admin` |
| `/admin/badges` | AdminBadgesPage | `admin` |
| `/admin/engagement-log` | EngagementLogPage | `admin` |
| `/admin/point-economy` | PointEconomyPage | `admin` |
| `/admin/settings` | AdminSettingsPage | `admin` |
| `/admin/ai-config` | AIConfigPage | `admin` |
| `/admin/billing` | AdminBillingPage | `admin` |
| `/sys` | SysadminConsolePage | `sysadmin` |

### Lazy Loading
All routes except Login, Dashboard, Directory, MemberProfile, Leaderboard are `React.lazy()` loaded.

### API Client Modules (`src/api/`)
One module per domain mirroring the backend: `auth`, `members`, `xp`, `events`, `dues`, `notifications`, `badges`, `quests`, `announcements`, `props`, `service`, `intake`, `votes`, `mentorship`, `minutes`, `scholarships`, `store`, `goals`, `messages`, `fundraising`, `committees`, `resources`, `job-board`, `alumni`, `platform`, `settings`, `health`, `ai`

---

## 7. USER ROLES & PERMISSIONS

| Role | Access Level |
|---|---|
| `member` | Standard — own profile, directory, events, XP, store, quests, props, etc. |
| `chair` | Member + scanner access, intake pipeline, check-in |
| `pia` | Member + PIA analytics reports |
| `admin` | Chapter admin — all member data, dues admin, badge/quest mgmt, settings, billing, AI config |
| `sysadmin` | Platform-level — all chapters, users, audit log, global XP recalc; dark-mode console at `/sys` |

**CES1231** — special system identity; `is_sysadmin = true` on `users` table.

---

## 8. GAMIFICATION SYSTEM

### XP Level Thresholds (defaults)
| Level Key | Label | Min XP |
|---|---|---|
| `neo` | Neophyte | 0 |
| `scholar` | Bronze Varsity | 500 |
| `leader` | Silver Elite | 1,000 |
| `sage` | Gold Legend | 1,500 |
| `legend` | Chapter Icon | 2,500 |
| `icon` | Chapter Icon+ | — |

### Avatar Assets
- `avatar-base.png` — Neophyte (0–499 XP)
- `avatar-bronze.png` — Bronze Varsity (500–999 XP)
- `avatar-silver.png` — Silver Elite (1,000–1,499 XP)
- `avatar-gold.png` — Gold Legend (1,500–2,499 XP)
- `avatar-icon.png` — Chapter Icon (2,500+ XP)

### XP Sources (configurable via `point_economy`)
checkin, rsvp, admin-award, system, quiz, service-hours, dues-payment, badge-award, props-received, mentorship

### Streak System
`members.streak` — consecutive meetings attended (incremented on check-in, reset on no-show).

### Badge Criteria Engine
JSONB `criteria` field supports types:
- `{"type":"xp_threshold","threshold":N}`
- `{"type":"meeting_streak","count":N}`
- `{"type":"service_hours","min_hours":N}`
- Extensible without schema changes

---

## 9. INTEGRATIONS

| Integration | Purpose | Implementation |
|---|---|---|
| **Stripe** | SaaS subscription billing (starter/growth/chapter_pro) | `stripe-go/v78`; webhook at `POST /webhooks/stripe` |
| **Zeffy** | 0% fee dues collection | Webhook via Zapier → `POST /webhooks/zeffy`; matches by email |
| **Resend** | Transactional email (magic links, notifications) | `resendlabs/resend-go` |
| **Cloudflare R2** | Avatar + file storage | `aws-sdk-go-v2` (S3-compatible) |
| **Anthropic Claude** | AI assistant (default: `claude-sonnet-4-6`) | Streaming SSE |
| **OpenAI** | AI assistant (alternative provider) | Streaming SSE |

---

## 10. DEPLOYMENT

### Local Dev
```bash
# Full stack (Postgres + Redis + API + Web)
cd blue-ledger-api && make up

# API only (against docker Postgres/Redis)
make dev

# Generate RSA keys (first time)
make keys
```

### Docker Services (docker-compose.yml)
- `postgres` — postgres:16-alpine, port 5432, volume `postgres_data`
- `redis` — redis:7-alpine, port 6379, append-only, 128MB LRU
- `api` — Go API, port 8080
- `web` — Nginx serving React build, port 3001 (inferred)

### Fly.io (Production)
- App: `blue-ledger-api`
- Region: `iad` (Washington DC)
- VM: `shared-cpu-1x` / 256MB RAM
- Health check: `GET /v1/healthz` every 15s
- Auto stop/start machines; min 1 running
- Concurrency: soft 200 / hard 250 req

### Env Vars Required for Production
`DATABASE_URL`, `REDIS_URL`, `JWT_PRIVATE_KEY_PATH`, `JWT_PUBLIC_KEY_PATH`, `MAGIC_LINK_HMAC_SECRET`, `RESEND_API_KEY`, `EMAIL_FROM`, `R2_*`, `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_*_PRICE_ID`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `ALLOWED_ORIGINS`

### Frontend Env Vars
`VITE_API_BASE_URL` — defaults to `/api`

---

## 11. TESTING

### Backend
- `go test ./...` — unit tests per package
- Notable: `blue-ledger-api/internal/auth/tokens_test.go`

### Frontend
- `vitest run` — unit + component tests
- Coverage: `vitest run --coverage`
- Notable: `web/src/api/ai.test.ts`, `web/src/components/Button.test.tsx`

---

## 12. KNOWN GAPS / DEBT

- Static prototype (`app/index.html`) is an in-memory demo; data resets on refresh
- QR scanner in prototype is simulated (Glide provides real camera)
- Make.com / Zapier automations require manual setup (documented in guides)
- Stripe payment UI requires additional config (Enhancement E3)
- No seed data script beyond `cmd/seed/` placeholder
- `sbc` package is sparse — SBC-specific features partially stubbed
- `milestones` feature has no dedicated API handler registered in main.go (frontend feature exists, backend route may be missing)
- `study-groups` same — present in frontend but not confirmed in backend route table

---

## 13. PROTOTYPE DEMO ACCOUNTS

| Account | Email | Role |
|---|---|---|
| Marcus J. Williams | m.williams@chapter.org | Admin / President |
| DeShawn A. Carter | d.carter@chapter.org | Chair |
| Elijah T. Brooks | e.brooks@chapter.org | PIA Analyst |
| Jordan M. Hayes | j.hayes@chapter.org | Member |
| **CES1231** | ces1231@blueledger.sys | **Sysadmin** |

---

## 14. QUICK REFERENCE — KEY FILE PATHS

| What | Where |
|---|---|
| API entry point | `blue-ledger-api/cmd/server/main.go` |
| Config loader | `blue-ledger-api/pkg/config/config.go` |
| JWT / tokens | `blue-ledger-api/internal/auth/tokens.go` |
| Auth middleware | `blue-ledger-api/internal/auth/middleware.go` |
| XP engine | `blue-ledger-api/internal/xp/service.go` |
| AI service | `blue-ledger-api/internal/ai/service.go` |
| All migrations | `blue-ledger-api/migrations/` (001–021) |
| React app root | `web/src/App.tsx` |
| Auth store | `web/src/stores/authStore.ts` |
| Axios client | `web/src/api/client.ts` |
| useAuth hook | `web/src/hooks/useAuth.ts` |
| TypeScript types | `web/src/types/index.ts` |
| Tailwind config | `web/tailwind.config.ts` |
| Vite config | `web/vite.config.ts` |
| Makefile | `blue-ledger-api/Makefile` |
| Docker Compose | `blue-ledger-api/docker-compose.yml` |
| Fly.io config | `blue-ledger-api/fly.toml` |
| PWA manifest | `manifest.json` |

---

*HEIMDALL — All-Seeing Eye — State file complete.*
