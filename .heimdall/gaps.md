# HEIMDALL — Gap Analysis
**Generated:** 2026-03-26
**Source:** HEIMDALL reindex + static analysis

---

## GAP-001 · 🔴 CRASH — Missing `sbc_log` table migration

**Severity:** P0 — Runtime crash on first request to any `/v1/sbc` route  
**Layer:** Backend — database  
**Discovered via:** `internal/sbc/service.go` queries `sbc_log` table; no migration creates it

### What exists
- `blue-ledger-api/internal/sbc/service.go` — full CRUD service, queries `sbc_log`
- `blue-ledger-api/internal/sbc/handler.go` — full handler, registered at `/v1/sbc`
- Route mounted in `cmd/server/main.go` line 245

### What is missing
- Migration `022_sbc_log.up.sql` — `CREATE TABLE sbc_log`
- Migration `022_sbc_log.down.sql` — `DROP TABLE sbc_log`

### Table spec (reverse-engineered from `service.go`)
```sql
CREATE TABLE IF NOT EXISTS sbc_log (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id   UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id    UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  category     TEXT        NOT NULL CHECK (category IN ('service','brotherhood','conduct')),
  note         TEXT        NOT NULL,
  recorded_by  UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Fix
Create `blue-ledger-api/migrations/022_sbc_log.up.sql` and `022_sbc_log.down.sql`.

### Files to create
- `blue-ledger-api/migrations/022_sbc_log.up.sql`
- `blue-ledger-api/migrations/022_sbc_log.down.sql`

---

## GAP-002 · 🔴 404 — Missing `milestones` backend package

**Severity:** P1 — All frontend `/milestones` requests return 404  
**Layer:** Backend — missing Go package  
**Discovered via:** `web/src/features/milestones/MilestonesPage.tsx` is a `StubPage` ("Coming soon"), no API module, no backend package

### What exists
- Migration `020_study_groups_and_milestones.up.sql` — `milestones` table ✅
- Frontend page `web/src/features/milestones/MilestonesPage.tsx` — renders `StubPage`
- Route `/milestones` registered in `web/src/App.tsx`

### What is missing
- `blue-ledger-api/internal/milestones/` package (handler, service, repository)
- Route registration in `cmd/server/main.go`
- API client module `web/src/api/milestones.ts`
- Full frontend page (currently `StubPage`)

### DB Schema (from migration 020)
```sql
milestones (
  id UUID PK,
  chapter_id UUID → chapters,
  member_id  UUID → members,
  type TEXT CHECK ('birthday','graduation','new_job','engagement','other'),
  title TEXT NOT NULL,
  date DATE,
  description TEXT,
  created_at TIMESTAMPTZ
)
-- RLS: chapter_isolation on chapter_id
-- No updated_at (immutable records)
```

### Required API endpoints
| Method | Path | Role | Description |
|---|---|---|---|
| GET | `/v1/milestones` | auth | List chapter milestones |
| POST | `/v1/milestones` | member | Create milestone |
| DELETE | `/v1/milestones/:id` | admin | Delete milestone |

### Files to create
- `blue-ledger-api/internal/milestones/service.go`
- `blue-ledger-api/internal/milestones/repository.go`
- `blue-ledger-api/internal/milestones/handler.go`
- `web/src/api/milestones.ts`
- `web/src/features/milestones/MilestonesPage.tsx` (replace StubPage)

### Files to modify
- `blue-ledger-api/cmd/server/main.go` — register handler

---

## GAP-003 · 🔴 404 — Missing `study-groups` backend package

**Severity:** P1 — All frontend `/study-groups` requests return 404  
**Layer:** Backend — missing Go package  
**Discovered via:** `web/src/features/study-groups/StudyGroupsPage.tsx` is a `StubPage`, no API module, no backend package

### What exists
- Migration `020_study_groups_and_milestones.up.sql` — `study_groups` table ✅
- Frontend page `web/src/features/study-groups/StudyGroupsPage.tsx` — renders `StubPage`
- Route `/study-groups` registered in `web/src/App.tsx`

### What is missing
- `blue-ledger-api/internal/studygroups/` package (handler, service, repository)
- Route registration in `cmd/server/main.go`
- API client module `web/src/api/study-groups.ts` (file exists in `web/src/api/` list — **needs verification**)
- Full frontend page (currently `StubPage`)

### DB Schema (from migration 020)
```sql
study_groups (
  id         UUID PK,
  chapter_id UUID → chapters,
  topic      TEXT NOT NULL,
  host_id    UUID → members,
  date       DATE NOT NULL,
  location   TEXT,
  member_ids UUID[] DEFAULT '{}',   -- GIN indexed
  xp_reward  INT DEFAULT 0,
  created_at TIMESTAMPTZ
)
-- RLS: chapter_isolation on chapter_id
-- GIN index on member_ids for "groups I'm in" queries
```

### Required API endpoints
| Method | Path | Role | Description |
|---|---|---|---|
| GET | `/v1/study-groups` | auth | List chapter study groups |
| POST | `/v1/study-groups` | member | Create study group |
| POST | `/v1/study-groups/:id/join` | member | Join study group (append member_id) |
| DELETE | `/v1/study-groups/:id` | admin/host | Delete study group |

### Files to create
- `blue-ledger-api/internal/studygroups/service.go`
- `blue-ledger-api/internal/studygroups/repository.go`
- `blue-ledger-api/internal/studygroups/handler.go`
- `web/src/api/study-groups.ts` (verify if stub exists)
- `web/src/features/study-groups/StudyGroupsPage.tsx` (replace StubPage)

### Files to modify
- `blue-ledger-api/cmd/server/main.go` — register handler

---

## GAP-004 · 🟡 STUB — Multiple frontend pages are `StubPage`

**Severity:** P2 — Features visible in nav but non-functional; user-facing dead ends  
**Layer:** Frontend  

| Route | StubPage file | Backend exists? |
|---|---|---|
| `/milestones` | `features/milestones/MilestonesPage.tsx` | ❌ (see GAP-002) |
| `/study-groups` | `features/study-groups/StudyGroupsPage.tsx` | ❌ (see GAP-003) |
| `/resources` | `features/resources/ResourcesPage.tsx` | ✅ `/v1/resources` |
| `/job-board` | `features/job-board/JobBoardPage.tsx` | ✅ `/v1/job-board` |
| `/fundraising` | `features/fundraising/FundraisingPage.tsx` | ✅ `/v1/fundraising` |
| `/committees` | `features/committees/CommitteesPage.tsx` | ✅ `/v1/committees` |
| `/alumni` | `features/alumni/AlumniPage.tsx` | ✅ `/v1/alumni` |

**Note:** Resources, Job Board, Fundraising, Committees, Alumni — backend packages fully exist with handlers registered. Only the frontend pages are stubs. These are P2 UI work.

---

## GAP-005 · 🟡 CONFIG — RSA keys not generated; `.env` not present

**Severity:** P2 — Server will not start without JWT keys  
**Layer:** Infrastructure / local dev  

### What is missing
- `blue-ledger-api/keys/private.pem`
- `blue-ledger-api/keys/public.pem`
- `blue-ledger-api/.env` (optional for dev, required for external integrations)

### Fix
```bash
cd blue-ledger-api
make keys
cp .env.example .env   # if .env.example exists, else create manually
```

### Minimum `.env` for local dev
```env
MAGIC_LINK_HMAC_SECRET=changeme-local-dev
RESEND_API_KEY=re_placeholder
ALLOWED_ORIGINS=http://localhost:3001
```

---

## GAP-006 · 🟡 SCHEMA — `sbc_log` has no migration (see GAP-001) but `sbc` migration file numbering

**Severity:** Note  
The last migration is `021_ai_assistant`. The next migration should be `022`. No migration exists for `sbc_log`. The `sbc` package was likely written assuming the table would be created — but the migration was never authored.

Verify: search for any alternate table name used by `sbc` service (e.g. `sbc_entries`, `service_brotherhood_conduct`) before writing 022.

---

## GAP-007 · 🟢 COSMETIC — `study-groups` API client file exists but may be empty

**Severity:** P3  
`web/src/api/study-groups.ts` appears in the `src/api/` directory listing. Content unknown — may be a stub or fully empty. Verify before building the frontend page.

---

## Remediation Order

| Priority | Gap | Effort | Risk |
|---|---|---|---|
| 1 | GAP-001 — Write `022_sbc_log` migration | 15 min | 🔴 App crashes without this |
| 2 | GAP-002 — Build `milestones` backend + frontend | 2h | 🟡 404s until done |
| 3 | GAP-003 — Build `study-groups` backend + frontend | 2h | 🟡 404s until done |
| 4 | GAP-005 — Generate keys + `.env` | 5 min | 🔴 Server won't start without keys |
| 5 | GAP-004 — Build out stub frontend pages (resources, job-board, etc.) | 1h each | 🟢 Backend ready, UI only |
| 6 | GAP-007 — Verify `study-groups.ts` API client | 5 min | 🟢 Cosmetic |

---

*HEIMDALL — All-Seeing Eye — Gap file complete.*
