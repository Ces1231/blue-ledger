# Blue Ledger — Phase 1 Build Review

## Meta
| Field | Value |
|-------|-------|
| Document Type | Phase 1 Build Review |
| Author | J.A.R.V.I.S. |
| Review Date | 2026-03-25 |
| Spec Reference | `.claude/specs/saas-architecture.md` |
| State File | `.claude/project-state.md` |
| Eitri Report | `.claude/eitri/build-report.md` |
| Scope | Full codebase scan: Go API, React frontend, migrations, infrastructure |

---

## Executive Summary

The Blue Ledger scaffold is substantially further along than a typical early-stage
project. All 20 Go internal packages exist with real handler + service implementations.
All 44 React routes are registered and wired to real feature pages — not stubs.
All 20 migration files (40 total up/down files) are in place. The infrastructure
foundation is solid.

The key gaps are: 8 missing Go backend packages (quests, badges, committees,
fundraising, resources, job-board, alumni, SBC), 10 missing API client files
in the React layer, 10 missing or misrouted frontend routes, zero tests, sqlc
queries not generated, no CI/CD pipeline, Sentry not wired, and rate limiting
middleware not configured. These gaps prevent a live demo but none are
architectural blockers.

**Overall completion: approximately 68% of the full spec.**

---

## 1. Completion Status — Phase by Phase

### Phase 0: Foundation
**Status: COMPLETE — 100%**

| Item | Status |
|------|--------|
| Go project structure (go.mod, cmd/server, pkg/) | Done |
| 20 migration files (40 up/down), 31 tables | Done |
| Auth service (JWT RS256, refresh, magic link, bcrypt) | Done |
| Dockerfile (multi-stage, non-root) | Done |
| docker-compose.yml (Postgres + Redis) | Done |
| fly.toml | Done |
| .env.example | Done |
| React Vite + TypeScript + Tailwind scaffold | Done |
| Login page wired to API | Done |
| Axios client with token refresh interceptor | Done |
| Zustand stores (auth, ui) | Done |
| Shared components (Sidebar, Topbar, Button, Card, Modal, Toast, Badge) | Done |

### Phase 1: Core Member & XP
**Status: COMPLETE — 100%**

| Item | Status |
|------|--------|
| Members API (handler, service, repository) | Done |
| XP engine (award, recalculate, leaderboard) | Done |
| Badge engine (async, JSONB criteria evaluation) | Done |
| Engagement log endpoint | Done |
| React: Directory, MemberProfile, Leaderboard pages | Done |
| React: Dashboard wired to real data | Done |

### Phase 2: Events & Check-In
**Status: COMPLETE — 100%**

| Item | Status |
|------|--------|
| Events API (CRUD, RSVP, QR check-in) | Done |
| QR token generation + validation (HMAC) | Done |
| Scanner page (role-gated to chair+) | Done |
| React: EventsPage, EventDetailPage, ScannerPage | Done |
| /events/create page | Partial — page file EXISTS, route NOT registered |

### Phase 3: Dues & Billing
**Status: ~75%**

| Item | Status |
|------|--------|
| Dues API (CRUD, mark-paid) | Done |
| Stripe checkout session | Done |
| Stripe subscription webhook | Done |
| Zeffy webhook (POST /webhooks/zeffy via Zapier) | Done |
| React: DuesPage | Done |
| React: AdminDuesPage, AdminBillingPage | Done |
| Subscription gate middleware | NOT DONE — spec requires 402 for past_due chapters |
| Chapter registration flow (new tenant sign-up) | NOT DONE — POST /auth/register creates user but no full onboarding flow |

### Phase 4: Engagement Features
**Status: ~80% (backend), ~60% (frontend missing API clients for some features)**

| Item | Status |
|------|--------|
| Announcements API + React page | Done |
| Props API + React page | Done |
| Notifications API + React page | Done |
| Service log API (submit, verify) + React pages | Done |
| Votes API + React page | Done |
| Mentorship API + React page | Done |
| Minutes API + React page | Done |
| Scholarships API + React page | Done |
| Intake (prospect pipeline) API + React page | Done |
| Quests API routes | NOT DONE — no quests Go package; QuestsPage uses inline apiClient calls |
| Badges admin API routes | NOT DONE — no badges Go package; badge_engine.go handles award but no /badges CRUD |
| Fundraising API | NOT DONE — no Go package; FundraisingPage uses inline apiClient |
| Committees API | NOT DONE — no Go package; CommitteesPage uses inline apiClient |
| Resources API | NOT DONE — no Go package; ResourcesPage uses inline apiClient |
| Job Board API | NOT DONE — no Go package; JobBoardPage uses inline apiClient |
| Alumni API | NOT DONE — no Go package; AlumniPage uses inline apiClient |
| SBC Log API | NOT DONE — no Go package |

### Phase 5: PIA, Admin & Sysadmin
**Status: ~65%**

| Item | Status |
|------|--------|
| Chapter health API (/health, /health/attendance, /health/xp, /health/service, /health/dues) | Done |
| Admin panel (members, events, dues, badges, settings, billing) | Done (React pages) |
| Engagement log admin view | Done |
| Point economy admin view | Done |
| Sysadmin console React page | Done |
| Platform/sys API (chapters, users, audit log, subscription override) | Done |
| PIA builder / chapter health React page (/pia) | NOT DONE — no /pia route, no pia feature folder |
| Sysadmin sub-routes (/sys/chapters, /sys/audit-log, /sys/billing) | NOT DONE — routes missing from App.tsx |
| Store API + React page | Done |
| Goals API + React page | Done |
| Messages API + React page | Done |
| Settings API + React page | Done |

### Phase 6: Polish & Launch
**Status: ~15%**

| Item | Status |
|------|--------|
| Email: RESEND_API_KEY wired in config | Done (env var exists) |
| Email: SendMagicLink function | Done |
| Email: SendDuesReminder, SendBadgeEarned | Unknown — pkg/email/resend.go exists |
| Sentry error monitoring | NOT DONE — DSN in .env.example, not wired into Go or React |
| Rate limiting middleware | NOT DONE — spec requires Redis-backed RateLimiter in middleware stack |
| CI/CD pipelines | NOT DONE — no .github/workflows |
| Load testing | NOT DONE |
| Security review | NOT DONE |
| Production deploy (live Fly.io instance) | NOT DONE — fly.toml present, no provisioned resources |
| Seed script (prototype data import) | NOT DONE — cmd/seed/main.go present, implementation unknown |

---

## 2. Gaps

### 2A. Missing Go Backend Packages

These domains have React UI pages but NO corresponding Go package:

| Domain | React Page | Go Package | API Endpoints Affected |
|--------|-----------|------------|----------------------|
| Quests | `web/src/features/quests/QuestsPage.tsx` | MISSING | GET /quests, GET /quests/:id/progress |
| Badges (CRUD) | `web/src/features/admin/AdminBadgesPage.tsx` | MISSING | GET /badges, POST /badges, POST /badges/:id/award |
| Committees | `web/src/features/committees/CommitteesPage.tsx` | MISSING | CRUD /committees |
| Fundraising | `web/src/features/fundraising/FundraisingPage.tsx` | MISSING | CRUD /fundraising |
| Resources | `web/src/features/resources/ResourcesPage.tsx` | MISSING | CRUD /resources |
| Job Board | `web/src/features/job-board/JobBoardPage.tsx` | MISSING | CRUD /job-board |
| Alumni | `web/src/features/alumni/AlumniPage.tsx` | MISSING | CRUD /alumni |
| SBC Log | None | MISSING | CRUD /sbc |

Note: `badge_engine.go` exists in `internal/xp` and handles automatic badge awarding.
What is missing is the `/badges` management API (admin CRUD for badge definitions).

### 2B. Missing React Routes

These routes are in the spec (`saas-architecture.md` § Route → Screen Mapping) but not
registered in `web/src/App.tsx`:

| Spec Route | Component File | Status |
|-----------|---------------|--------|
| `/profile/edit` | `features/profile/EditProfilePage.tsx` | Page exists, NOT routed |
| `/events/create` | `features/events/CreateEventPage.tsx` | Page exists, NOT routed |
| `/announcements/create` | `features/announcements/CreateAnnouncementPage.tsx` | Page exists, NOT routed |
| `/sbc` | No file | Page AND route both missing |
| `/history` | No file | Page AND route both missing |
| `/network` | No file | Page AND route both missing |
| `/pia` | No file | PIA folder missing entirely |
| `/sys/chapters` | `features/sysadmin/SysadminConsolePage.tsx` | Sub-routes not split out |
| `/sys/audit-log` | `features/sysadmin/SysadminConsolePage.tsx` | Sub-routes not split out |
| `/sys/billing` | `features/sysadmin/SysadminConsolePage.tsx` | Sub-routes not split out |

### 2C. Missing React API Client Files

The spec pattern calls for one `web/src/api/{domain}.ts` per feature. Several features
currently use inline `apiClient` calls inside the page component instead:

| API Client File | Status | Workaround in Use |
|----------------|--------|------------------|
| `api/notifications.ts` | MISSING | Inline apiClient in NotificationsPage.tsx |
| `api/quests.ts` | MISSING | Inline apiClient in QuestsPage.tsx |
| `api/badges.ts` | MISSING | Inline apiClient in QuestsPage.tsx |
| `api/committees.ts` | MISSING | Inline apiClient in CommitteesPage.tsx |
| `api/fundraising.ts` | MISSING | Inline apiClient in FundraisingPage.tsx |
| `api/resources.ts` | MISSING | Inline apiClient in ResourcesPage.tsx |
| `api/study-groups.ts` | MISSING | Inline apiClient in StudyGroupsPage.tsx |
| `api/milestones.ts` | MISSING | Inline apiClient in MilestonesPage.tsx |
| `api/job-board.ts` | MISSING | Inline apiClient in JobBoardPage.tsx |
| `api/alumni.ts` | MISSING | Inline apiClient in AlumniPage.tsx |

This is a minor technical debt — the app still works — but it creates inconsistency
and makes typed response shapes harder to enforce across components.

### 2D. Missing Middleware

| Middleware | Spec Location | Status |
|-----------|--------------|--------|
| Redis-backed RateLimiter (100 req/min per IP) | `saas-architecture.md` §5 Middleware Stack | NOT WIRED — `redisClient` is assigned to `_ = redisClient` in main.go |
| Subscription gate (402 for past_due/canceled chapters) | `saas-architecture.md` §8 | NOT IMPLEMENTED |
| TenantSetter (SET LOCAL app.chapter_id) | §5 | Implemented in `internal/auth/middleware.go` |

### 2E. sqlc Queries Not Generated

`sqlc.yaml` is configured and 4 query files exist (`auth.sql`, `members.sql`,
`events.sql`, `engagement_log.sql`), but:

- `internal/db/` directory does NOT exist — sqlc generate has never been run
- Only 4 of 20 packages have query files; the remaining 16 packages query the
  database directly without generated code
- Running `sqlc generate` is a prerequisite before `go build` will succeed if
  any package imports from `internal/db`

### 2F. Missing Infrastructure

| Item | Status |
|------|--------|
| CI/CD pipeline (GitHub Actions) | NOT DONE — no `.github/` directory |
| Sentry SDK wired (Go) | NOT DONE — env var present, SDK not imported |
| Sentry SDK wired (React) | NOT DONE |
| Rate limiting (Redis) | NOT DONE |
| web/public/ directory (PWA manifest, favicon) | MISSING |
| `src/utils/` directory (date formatting, XP level helpers) | MISSING |

### 2G. No Tests

Zero test files exist in either the Go API or React frontend:
- No `*_test.go` files anywhere in `blue-ledger-api/`
- No `*.test.ts` or `*.spec.tsx` files anywhere in `web/`

This is the largest quality risk for a production deployment.

---

## 3. Spec Deviations

### 3A. Module Path Deviation
**Spec:** `github.com/blueledger/api`
**Built:** `github.com/ces1231/blue-ledger-api`
**Impact:** Cosmetic only — module path works, naming is just personal vs org name.
**Recommendation:** Acceptable for pre-production. Change to org path before open-sourcing.

### 3B. Migration Numbering Expanded
**Spec:** 3 migration files (001, 002, 003) covering all tables
**Built:** 20 migration files (001–020), one per domain
**Impact:** Better practice — granular rollbacks possible. Zero issues.
**Recommendation:** Keep as-is. The expansion is a strict improvement over the spec.

### 3C. Migration Numbering Conflict
**Issue:** Two files named `001_*.up.sql` exist in the migrations directory:
- `001_initial_schema.up.sql` (Eitri's first pass)
- `001_users_and_chapters.up.sql` (Nebula's full suite)

golang-migrate will fail on startup because migration 1 is ambiguous.
**Impact:** Server will not start against a fresh database.
**Priority: CRITICAL — must resolve before any demo.**

### 3D. Redis Not Utilized
**Spec:** Redis used for rate limiting, sessions, and notification queues
**Built:** Redis client connected, then immediately shadowed (`_ = redisClient`).
No Redis-backed features are implemented.
**Impact:** Rate limiting is absent; no session invalidation on the Redis layer.

### 3E. sqlc Pattern Inconsistently Applied
**Spec:** sqlc-generated queries for all packages
**Built:** Only 4 packages have `.sql` query files. The remaining 16 packages query
Postgres directly using `pgx` without generated code.
**Impact:** Loses compile-time query verification for most of the codebase.
The approach still works at runtime, but defeats a key reason sqlc was chosen.

### 3F. ScannerPage Duplicated
**Built:** `features/events/ScannerPage.tsx` AND `features/scanner/ScannerPage.tsx`
both exist. App.tsx imports from `features/events/ScannerPage`.
**Impact:** Dead code in `features/scanner/`. Not a runtime bug but creates confusion.

### 3G. Zeffy Webhook Secret Validation
**Spec:** Zeffy uses a simple HMAC secret in the header for verification.
**Built:** Handler comment acknowledges this, but the actual secret comparison
should be verified to confirm it is implemented (not just described).
**Recommendation:** Review `ZeffyWebhook` handler for actual HMAC comparison before
accepting live Zapier payloads.

### 3H. Dues Payment Architecture (Zeffy vs Stripe)
**Spec:** Member dues use Zeffy (0% fees) embedded popup; Stripe only for SaaS
subscription billing.
**Built:** `dues/stripe.go` and `CreateStripeSession` are built for dues — which
contradicts the spec's Zeffy-first model. The Zeffy webhook IS implemented.
**Impact:** The Stripe dues checkout path should likely be removed or marked
as a fallback. The primary dues flow is Zeffy → Zapier → `/webhooks/zeffy`.

---

## 4. Critical Missing Items (Demo/Deploy Blockers)

These items must be resolved before a live demo or staging deployment:

### BLOCKER 1: Duplicate migration 001
**File:** `blue-ledger-api/migrations/`
Two files share the sequence number `001`. golang-migrate reads by number and will
panic or produce a "migration out of sequence" error on startup.
**Fix:** Delete `001_initial_schema.up.sql` and `001_initial_schema.down.sql` (Eitri's
first pass). The Nebula suite (001 through 020) is the canonical set.

### BLOCKER 2: No RSA keypair generated
**File:** JWT keys referenced in config (`JWT_PRIVATE_KEY_PATH`, `JWT_PUBLIC_KEY_PATH`)
Without key files, the server exits at startup: "failed to load JWT keys".
**Fix:** Run `openssl genrsa -out keys/private.pem 4096 && openssl rsa -in keys/private.pem -pubout -out keys/public.pem`

### BLOCKER 3: go.sum not populated
**Confirmed:** `go.sum` file exists but `go mod tidy` must be run to ensure all
dependencies are pinned. The Docker build will fail if go.sum is stale.
**Fix:** `cd blue-ledger-api && go mod tidy`

### BLOCKER 4: Missing /badges and /quests API routes
**Impact:** `QuestsPage.tsx` and `AdminBadgesPage.tsx` will receive 404s from the
API. These are visible in the nav on every demo.
**Fix:** Implement `internal/badges` and `internal/quests` Go packages with handler
and service, then register routes in main.go.

### BLOCKER 5: Rate limiter unused
**Impact:** Any public API endpoint is unprotected from brute-force or abuse.
The auth endpoints (`/auth/login`, `/auth/magic-link`) are especially exposed.
**Fix:** Wire `redisClient` into rate limiting middleware on the Echo instance.

---

## 5. Next Steps — Prioritized and Agent-Routed

### Priority 1: Fix Blockers (do immediately, Ant-Man tasks)

| # | Task | Agent | Est |
|---|------|-------|-----|
| 1 | Delete `001_initial_schema.up/down.sql` (Eitri first-pass) | Developer (1 min) | 5 min |
| 2 | Run `go mod tidy` and populate `go.sum` | Developer | 5 min |
| 3 | Generate RSA keypair into `keys/` | Developer | 5 min |
| 4 | Wire rate limiter middleware using redisClient | Ant-Man | 2 hrs |

### Priority 2: Missing Go Packages (Sprint for Wasp, ~32 hrs)

Build the 8 missing Go packages. These are all straightforward CRUD with RLS.

| Package | Tables | Est |
|---------|--------|-----|
| `internal/badges` | `badges`, `member_badges` | 4 hrs |
| `internal/quests` | `quests`, `member_quest_progress` | 4 hrs |
| `internal/committees` | `committees` | 2 hrs |
| `internal/fundraising` | `fundraising_campaigns` | 3 hrs |
| `internal/resources` | `resources` | 2 hrs |
| `internal/jobboard` | `job_board` | 2 hrs |
| `internal/alumni` | `members` (status='alumni' filter) | 2 hrs |
| `internal/sbc` | (new table needed) | 3 hrs |

Suggested invocation after this review spec is filed:
```
Use wasp. Build from sprint spec .claude/tasks/SPRINT-002-missing-packages.md
```

### Priority 3: Missing Routes + API Clients (~12 hrs, Ant-Man or Wasp)

| # | Task | Est |
|---|------|-----|
| Wire `/profile/edit` route to `EditProfilePage.tsx` | 0.5 hrs |
| Wire `/events/create` route to `CreateEventPage.tsx` | 0.5 hrs |
| Wire `/announcements/create` route to `CreateAnnouncementPage.tsx` | 0.5 hrs |
| Create `/pia` route and `features/pia/PIAPage.tsx` (chapter health view) | 4 hrs |
| Create `/sbc` route and `features/sbc/SBCPage.tsx` | 2 hrs |
| Create `/history` stub route | 0.5 hrs |
| Create `/network` stub route | 0.5 hrs |
| Split `/sys/chapters`, `/sys/audit-log`, `/sys/billing` routes in App.tsx | 1 hr |
| Extract 10 inline API calls into dedicated `api/{domain}.ts` files | 3 hrs |

### Priority 4: Subscription Gate Middleware (~4 hrs, Ant-Man)

Implement the subscription gate described in spec §8:
- Middleware reads `chapter.subscription_status` from JWT or a fast Redis lookup
- Returns `402 Payment Required` if status is `past_due` (> 7 day grace) or `canceled`
- Exempt routes: `/auth/*`, `/webhooks/*`, `/v1/healthz`, `/sys/*`

### Priority 5: Test Suite (Sprint for Wasp, ~40 hrs)

Zero tests is the biggest quality gap. Suggested sprint:
- Auth service unit tests (Login, MagicLink, Refresh, lockout)
- Auth handler integration tests (all 7 endpoints)
- Members service + handler tests
- XP engine + badge engine tests
- Events + QR check-in tests
- Dues webhook tests (Stripe + Zeffy)
- React component tests for Dashboard, LeaderboardPage, MemberProfile
- React hook tests for useAuth

### Priority 6: Observability (~8 hrs, Ant-Man)

| Task | Est |
|------|-----|
| Wire Sentry SDK in Go API (`github.com/getsentry/sentry-go/echo`) | 2 hrs |
| Wire Sentry SDK in React (`@sentry/react`) | 2 hrs |
| Add `src/utils/` with date formatting and XP level helpers | 1 hr |
| Add `web/public/` with PWA manifest and favicon | 1 hr |
| Verify `pkg/email/resend.go` sends all three email types | 2 hrs |

### Priority 7: CI/CD Pipeline (~8 hrs, Falcon)

| Task | Est |
|------|-----|
| `.github/workflows/ci.yml` — Go test + build on PR | 2 hrs |
| `.github/workflows/cd.yml` — Fly.io deploy on merge to main | 3 hrs |
| Coverage gate (minimum 60% on auth and xp packages) | 1 hr |
| `web/.github/workflows/ci.yml` — React build + type-check | 2 hrs |

---

## 6. AI Assistant Feature Spec

### Overview

Add a model-agnostic AI assistant to The Blue Ledger. Chapter admins configure
which AI provider and API key to use. Members and admins interact with the assistant
from within the app. The assistant has chapter-aware context (member count, recent
activity, PIA metrics) and can help with specific fraternity chapter management tasks.

### Design Decisions

- **Model-agnostic:** Support Claude (Anthropic) and ChatGPT (OpenAI) with more
  models addable via a provider enum. The chapter admin selects the model.
- **API key per chapter:** Each chapter provides their own API key. Blue Ledger never
  stores it in plaintext — it is encrypted at rest using AES-256 with a per-chapter
  key derivation.
- **Context injection:** Every request includes a system prompt with chapter name,
  current semester, member count, recent XP activity summary, and dues collection rate.
  This makes the assistant relevant without requiring members to re-explain context.
- **Three built-in use cases at launch:**
  1. PIA report drafting (structured prompt for PIA reports)
  2. Announcement drafting (given a topic, draft a chapter announcement)
  3. General chapter Q&A (fraternity bylaws questions, Robert's Rules, etc.)

### Where It Lives

**Primary location: `/assistant` — new top-level route**
- Accessible from the Sidebar under a new "AI Assistant" nav item
- Visible to roles: `member`, `chair`, `pia`, `admin` (all chapter members)
- Admin configuration: `Admin Panel → Settings → AI Assistant tab`

The assistant is a full-page chat interface. It is NOT a floating widget — keep it
as a dedicated screen to avoid cluttering the core app UI.

### What the Assistant Can Do

| Capability | Prompt Template |
|-----------|----------------|
| Draft PIA report | System: chapter stats. User: "Help me write a PIA report for [semester]" |
| Draft announcement | User: "Draft an announcement about [topic]" |
| Summarize meeting minutes | User pastes minutes text, asks for summary |
| Answer bylaws / Roberts Rules questions | General Q&A against model training data |
| Interpret chapter health data | System: health metrics injected. User: "What areas need attention?" |
| Suggest engagement ideas | "Our attendance has dropped 20%. What are some ideas to re-engage members?" |

### API Endpoint Design

#### New table: `ai_assistant_config`
```sql
CREATE TABLE ai_assistant_config (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id        UUID NOT NULL UNIQUE REFERENCES chapters(id),
  provider          TEXT NOT NULL CHECK (provider IN ('anthropic', 'openai')),
  model             TEXT NOT NULL,                        -- 'claude-3-5-haiku-latest', 'gpt-4o-mini', etc.
  encrypted_api_key TEXT NOT NULL,                        -- AES-256 encrypted
  key_hint          TEXT,                                 -- last 4 chars of key for display
  is_enabled        BOOLEAN NOT NULL DEFAULT FALSE,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE ai_assistant_config ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON ai_assistant_config
  USING (chapter_id = current_setting('app.chapter_id')::uuid);
```

#### New migration: `021_ai_assistant_config.up.sql`

#### New Go package: `internal/assistant`

Files:
- `internal/assistant/service.go`
- `internal/assistant/handler.go`
- `internal/assistant/providers.go` — provider interface + Anthropic + OpenAI adapters

##### Service Interface

```go
type AssistantService interface {
    GetConfig(ctx context.Context, chapterID string) (*AssistantConfig, error)
    SaveConfig(ctx context.Context, chapterID string, req SaveConfigRequest) (*AssistantConfig, error)
    Chat(ctx context.Context, chapterID string, memberID string, req ChatRequest) (<-chan string, error) // SSE stream
    BuildSystemPrompt(ctx context.Context, chapterID string) (string, error)
}

type SaveConfigRequest struct {
    Provider string `json:"provider" validate:"required,oneof=anthropic openai"`
    Model    string `json:"model"    validate:"required"`
    APIKey   string `json:"api_key"  validate:"required,min=10"`
}

type ChatRequest struct {
    Messages []ChatMessage `json:"messages" validate:"required,min=1,max=50"`
    Mode     string        `json:"mode"     validate:"oneof=pia announcement general"`
}

type ChatMessage struct {
    Role    string `json:"role"    validate:"required,oneof=user assistant"`
    Content string `json:"content" validate:"required,max=8000"`
}
```

##### Provider Interface

```go
type Provider interface {
    // Stream returns a channel that emits text chunks as they arrive from the model.
    Stream(ctx context.Context, systemPrompt string, messages []ChatMessage, model string) (<-chan string, error)
    // AvailableModels returns the list of models this provider supports.
    AvailableModels() []ModelInfo
}

type ModelInfo struct {
    ID          string `json:"id"`           // "claude-3-5-haiku-latest"
    DisplayName string `json:"display_name"` // "Claude 3.5 Haiku"
    Provider    string `json:"provider"`     // "anthropic"
    IsFast      bool   `json:"is_fast"`      // true for cheap/fast models
}
```

#### API Endpoints

| Method | Path | Auth | Roles | Description |
|--------|------|------|-------|-------------|
| `GET`  | `/assistant/config` | JWT | admin | Get current AI config (api_key redacted, shows key_hint) |
| `PUT`  | `/assistant/config` | JWT | admin | Save provider + model + API key (encrypts key, validates by test call) |
| `DELETE` | `/assistant/config` | JWT | admin | Remove AI config (disables assistant for chapter) |
| `GET`  | `/assistant/models` | JWT | admin | List available models per provider |
| `POST` | `/assistant/chat` | JWT | Any | Send message, receive SSE stream response |

##### POST /assistant/chat
Request:
```json
{
  "messages": [
    { "role": "user", "content": "Help me draft an announcement about our spring cookout" }
  ],
  "mode": "announcement"
}
```

Response: `Content-Type: text/event-stream`
```
data: {"delta": "Here"}
data: {"delta": " is"}
data: {"delta": " a draft"}
data: [DONE]
```

Error when not configured:
```json
{ "error": { "code": "ASSISTANT_NOT_CONFIGURED", "message": "AI assistant not configured for this chapter", "status": 422 } }
```

##### Context Injection (System Prompt Template)

```
You are the AI assistant for the {chapter_name} chapter of Phi Beta Sigma Fraternity.

Chapter context:
- Chapter: {chapter_name} ({greek_letters}), {city}, {state}
- University: {university}
- Current semester: {semester}
- Active members: {active_member_count}
- Average XP this semester: {avg_xp_semester}
- Dues collection rate: {dues_collection_rate}%
- Upcoming events this month: {upcoming_event_count}
- Service hours this semester: {total_service_hours}

You help chapter officers and members with:
1. Drafting PIA (Programs, Initiatives, Activities) reports
2. Drafting chapter announcements
3. Answering questions about chapter operations, fraternity bylaws, and Robert's Rules
4. Interpreting chapter health and engagement data

When asked to draft content, provide a complete draft in a professional tone
appropriate for Phi Beta Sigma Fraternity. Do not make up names, dates, or
statistics that are not provided in the context above.
```

The system prompt is assembled by `BuildSystemPrompt()`, which queries the
chapter's current semester stats from the database before each conversation.

#### React Implementation

New files:
- `web/src/features/assistant/AssistantPage.tsx` — full-page chat UI
- `web/src/features/assistant/AssistantConfigPage.tsx` — admin config (provider, model, API key)
- `web/src/api/assistant.ts` — API client

New route in App.tsx:
```tsx
<Route path="/assistant" element={<AssistantPage />} />
<Route path="/admin/assistant" element={<AssistantConfigPage />} />
```

The chat UI should:
- Display a conversation thread (user messages right-aligned, assistant left-aligned)
- Stream the assistant's response word-by-word (SSE → React state append)
- Include a "Mode" selector: General, PIA Report, Announcement Draft
- Show a "Not configured" state with a link to Admin → Settings when no API key is set
- Keep conversation in local component state (no persistence — reset on page refresh)

### Security Considerations

- API keys are encrypted using AES-256-GCM before database storage
- The encryption key is derived from `AI_ASSISTANT_ENCRYPTION_KEY` env var (32 bytes)
- API keys are NEVER returned in API responses — only `key_hint` (last 4 chars)
- The `PUT /assistant/config` endpoint validates the API key by making a minimal
  test call to the provider before saving (ensures key works before commitment)
- Rate limit: 10 chat requests per member per hour (Redis counter per member_id)
- Max messages per conversation: 50 (enforced server-side)
- Max message length: 8000 characters (enforced server-side)
- Sysadmins cannot read chapter API keys — they are chapter-scoped and encrypted

### Hour Estimate

| Task | Est |
|------|-----|
| Migration (`021_ai_assistant_config`) | 0.5 hr |
| `internal/assistant/service.go` + `handler.go` | 6 hrs |
| `internal/assistant/providers.go` (Anthropic + OpenAI adapters + SSE) | 6 hrs |
| API key encryption helper (`pkg/crypto/aes.go`) | 2 hrs |
| Route registration + env var | 0.5 hr |
| React: `AssistantPage.tsx` (chat UI + SSE stream) | 5 hrs |
| React: `AssistantConfigPage.tsx` (provider/model/key form) | 3 hrs |
| React: `api/assistant.ts` | 1 hr |
| Sidebar nav item | 0.5 hr |
| Tests (provider mock, handler, service) | 4 hrs |
| **Total** | **~28 hrs** |

**Recommended builder:** Wasp (sprint scope, ~28 hrs, single feature)

Suggested invocation:
```
Use wasp. Build from sprint spec .claude/tasks/SPRINT-003-ai-assistant.md
```

---

## 7. Summary Scorecard

| Category | Score | Notes |
|----------|-------|-------|
| Database migrations | 95% | 20 migrations, 31 tables, RLS — only blocker is duplicate 001 |
| Go API — core packages | 100% | All 20 packages built with real implementations |
| Go API — engagement packages | 70% | 8 missing packages (badges/quests/committees/etc.) |
| Go API — middleware | 40% | Rate limiting not wired, subscription gate missing |
| React routes | 78% | 44 of ~56 spec routes present; 10 missing or misrouted |
| React feature pages | 85% | Real implementations across all major features |
| React API clients | 67% | 10 of ~20 needed client files using inline calls |
| Infrastructure | 60% | Dockerfile/fly.toml/compose done; CI/CD, Sentry, rate limit missing |
| Tests | 0% | Zero test files in Go or React |
| **Overall** | **~68%** | Solid scaffold, key gaps in secondary features and quality |

---

## 8. Notes for AI Agents

### For Wasp (sprint builder):
The highest-value next sprint is the missing Go packages (Section 5, Priority 2).
All 8 packages follow the exact same pattern: pgx queries, service interface,
Echo handler with RLS via `auth.GetChapterID(c)`, registered in main.go.
Read `internal/announcements/` as the canonical pattern example — it is the
simplest complete package and represents the target pattern for all 8.

### For Ant-Man (Blocker fixes):
Fix the duplicate migration 001 first (delete `001_initial_schema.up.sql` and
`001_initial_schema.down.sql`). Then wire the rate limiter into main.go using
the existing `redisClient` variable and Echo's built-in rate limiter middleware.

### For Falcon (CI/CD):
The project uses Go 1.23, Echo v4, and Vite + React. Target Fly.io via
`flyctl deploy`. The `fly.toml` is already in `blue-ledger-api/`. Wire
`FLY_API_TOKEN` as a GitHub secret.

### For Iron Man (full feature sprint):
The AI Assistant (Section 6) is the next major feature. It requires a new
Go package, SSE streaming, AES encryption helper, and a React chat UI.
Spec is fully defined in Section 6 above — no assumptions needed.
Builder recommendation: Wasp (28 hrs, sequential build).

---

— J.A.R.V.I.S.
