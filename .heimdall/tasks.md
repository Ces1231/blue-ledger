# HEIMDALL — Task List
**Generated:** 2026-03-26
**Source:** gaps.md

---

## TASK-001 · Write `sbc_log` migration
**Gap:** GAP-001  
**Priority:** P0 — CRASH  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~15 min  

**What to do:**
1. Create `blue-ledger-api/migrations/022_sbc_log.up.sql`
   - `CREATE TABLE sbc_log` with RLS, indexes, grants matching all other domain tables
   - Columns: `id`, `chapter_id` → chapters, `member_id` → members, `category CHECK ('service','brotherhood','conduct')`, `note`, `recorded_by` → members, `created_at`
2. Create `blue-ledger-api/migrations/022_sbc_log.down.sql`
   - `DROP TABLE IF EXISTS sbc_log`

**Acceptance criteria:**
- `make migrate-up` runs cleanly (no error)
- `GET /v1/sbc` returns `{"data":[]}` not a 500
- `POST /v1/sbc` creates a record

---

## TASK-002 · Generate local dev RSA keys + create `.env`
**Gap:** GAP-005  
**Priority:** P0 — Server will not start  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~5 min  

**What to do:**
1. Run from `blue-ledger-api/`:
   ```
   make keys
   ```
2. Create `blue-ledger-api/.env` with at minimum:
   ```
   MAGIC_LINK_HMAC_SECRET=changeme-local-dev
   RESEND_API_KEY=re_placeholder
   ALLOWED_ORIGINS=http://localhost:3001,http://localhost:80
   ```

**Acceptance criteria:**
- `keys/private.pem` and `keys/public.pem` exist
- `make dev` (or `make up`) starts without JWT key errors

---

## TASK-003 · Build `milestones` Go package
**Gap:** GAP-002  
**Priority:** P1 — 404  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~1.5h  

**What to do:**
1. Create `blue-ledger-api/internal/milestones/service.go`
   - `Milestone` struct matching migration 020 schema
   - `Service` interface: `List`, `Create`, `Delete`
   - `NewService(pool)` implementation
2. Create `blue-ledger-api/internal/milestones/repository.go`
   - Raw pgx queries: list by chapter (ordered by `date DESC`), insert, soft-delete or hard-delete
3. Create `blue-ledger-api/internal/milestones/handler.go`
   - `RegisterRoutes(g, jwtMW)`: `GET /`, `POST /`, `DELETE /:id`
   - Use `auth.GetChapterID`, `auth.GetMemberID` (pattern from `sbc/handler.go`)
4. Register in `cmd/server/main.go`:
   - Import, `milestonesSvc := milestones.NewService(pool)`, `milestonesHandler := milestones.NewHandler(milestonesSvc)`
   - `milestonesHandler.RegisterRoutes(v1.Group("/milestones"), jwtMW)`

**Reference pattern:** `internal/sbc/` (simplest matching domain — no XP, no sub-resources)

**Acceptance criteria:**
- `GET /v1/milestones` → `200 {"data":[]}`
- `POST /v1/milestones` with `{type, title, member_id, date}` → `201 {"data":{...}}`
- `DELETE /v1/milestones/:id` (admin only) → `200`
- `go build ./...` passes

---

## TASK-004 · Build `study-groups` Go package
**Gap:** GAP-003  
**Priority:** P1 — 404  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~2h  

**What to do:**
1. Create `blue-ledger-api/internal/studygroups/service.go`
   - `StudyGroup` struct: `id`, `chapter_id`, `topic`, `host_id`, `date`, `location`, `member_ids []string`, `xp_reward`, `created_at`
   - `Service` interface: `List`, `Create`, `Join`, `Delete`
   - `NewService(pool)` implementation
2. Create `blue-ledger-api/internal/studygroups/repository.go`
   - List: `SELECT ... WHERE chapter_id = $1 ORDER BY date DESC`
   - Create: `INSERT INTO study_groups (...)`
   - Join: `UPDATE study_groups SET member_ids = array_append(member_ids, $1) WHERE id = $2 AND chapter_id = $3 AND NOT ($1 = ANY(member_ids))`
   - Delete: `DELETE FROM study_groups WHERE id = $1 AND chapter_id = $2`
3. Create `blue-ledger-api/internal/studygroups/handler.go`
   - `RegisterRoutes`: `GET /`, `POST /`, `POST /:id/join`, `DELETE /:id`
4. Register in `cmd/server/main.go`

**Acceptance criteria:**
- `GET /v1/study-groups` → `200 {"data":[]}`
- `POST /v1/study-groups` with `{topic, date, location, xp_reward}` → `201`
- `POST /v1/study-groups/:id/join` appends caller's `member_id` to `member_ids` array; idempotent
- `go build ./...` passes

---

## TASK-005 · Build `milestones` frontend page
**Gap:** GAP-002  
**Priority:** P1 — depends on TASK-003  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~1h  

**What to do:**
1. Verify / create `web/src/api/milestones.ts`
   - `listMilestones()` → `GET /v1/milestones`
   - `createMilestone(data)` → `POST /v1/milestones`
   - `deleteMilestone(id)` → `DELETE /v1/milestones/:id`
2. Replace `web/src/features/milestones/MilestonesPage.tsx` — remove `StubPage`, build real UI:
   - Use `useQuery` to fetch milestones
   - Card list: each milestone shows type emoji, title, member name, date
   - `+` button → modal with `react-hook-form` (type select, title, member picker, date)
   - Admin delete icon per item
   - Use existing `Card`, `Modal`, `Button`, `Topbar` components from `web/src/components/`

**Milestone type → emoji map:**
```
birthday → 🎂  graduation → 🎓  new_job → 💼  engagement → 💍  other → 🏅
```

**Acceptance criteria:**
- Page loads without error
- Empty state renders when no milestones
- Creating a milestone optimistically updates the list
- Admin sees delete controls; members do not

---

## TASK-006 · Build `study-groups` frontend page
**Gap:** GAP-003  
**Priority:** P1 — depends on TASK-004  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~1h  

**What to do:**
1. Verify / create `web/src/api/study-groups.ts`
   - `listStudyGroups()` → `GET /v1/study-groups`
   - `createStudyGroup(data)` → `POST /v1/study-groups`
   - `joinStudyGroup(id)` → `POST /v1/study-groups/:id/join`
   - `deleteStudyGroup(id)` → `DELETE /v1/study-groups/:id`
2. Replace `web/src/features/study-groups/StudyGroupsPage.tsx`:
   - List all study groups: topic, host name, date, location, attendee count
   - "Join" button on groups the current member hasn't joined
   - "Joined" badge if `member_ids` includes current member
   - "Host" tag if `host_id` === current member
   - Create group button → modal form (topic, date, location, xp_reward)
   - Admin delete; host can delete their own group

**Acceptance criteria:**
- Page loads without error
- Join button calls API and refreshes list
- Member can join once (idempotent — button disables after joining)
- Create group opens modal and submits correctly

---

## TASK-007 · Build `resources` frontend page
**Gap:** GAP-004  
**Priority:** P2 — backend ready  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~1h  

**What to do:**
- Replace `web/src/features/resources/ResourcesPage.tsx` (currently StubPage)
- API: `GET /v1/resources`, `POST /v1/resources`, `DELETE /v1/resources/:id`
- Verify `web/src/api/resources.ts` is not a stub

**UI spec:**
- Card list: title, category badge, URL link, posted-by, date
- Filter bar by category
- "Add Resource" button (admin/chair) → modal (title, url, category, description)
- Admin delete icon

---

## TASK-008 · Build `job-board` frontend page
**Gap:** GAP-004  
**Priority:** P2 — backend ready  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~1h  

**What to do:**
- Replace `web/src/features/job-board/JobBoardPage.tsx` (currently StubPage)
- API: `GET /v1/job-board`, `POST /v1/job-board`, `DELETE /v1/job-board/:id`
- Verify `web/src/api/job-board.ts` is not a stub

**UI spec:**
- Card list: title, company, location, "Apply" link to URL, expires badge
- "Post Job" button → modal (title, company, location, url, description, expires_at)
- Admin delete

---

## TASK-009 · Build `fundraising` frontend page
**Gap:** GAP-004  
**Priority:** P2 — backend ready  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~1.5h  

**What to do:**
- Replace `web/src/features/fundraising/FundraisingPage.tsx` (currently StubPage)
- API: `GET /v1/fundraising`, `POST /v1/fundraising`, `POST /v1/fundraising/:id/donate`
- Verify `web/src/api/fundraising.ts` is not a stub

**UI spec:**
- Campaign cards: name, progress bar (`raised_cents / goal_cents`), deadline, status badge
- "Donate" button → modal (amount, donor_name, message)
- Admin: "New Campaign" button

---

## TASK-010 · Build `committees` frontend page
**Gap:** GAP-004  
**Priority:** P2 — backend ready  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~45 min  

**What to do:**
- Replace `web/src/features/committees/CommitteesPage.tsx` (currently StubPage)
- API: `GET /v1/committees`
- Verify `web/src/api/committees.ts` is not a stub

**UI spec:**
- List: committee name, chair name, description, member count
- Read-only for members; admin can add/edit

---

## TASK-011 · Build `alumni` frontend page
**Gap:** GAP-004  
**Priority:** P2 — backend ready  
**Status:** `[x] done 2026-03-26`  
**Effort:** ~45 min  

**What to do:**
- Replace `web/src/features/alumni/AlumniPage.tsx` (currently StubPage)
- API: `GET /v1/alumni`
- Verify `web/src/api/alumni.ts` is not a stub

**UI spec:**
- Card grid: avatar, name, graduation year, employer, city, LinkedIn link
- Search/filter by graduation year or city

---

## TASK-012 · Verify `study-groups.ts` API client content
**Gap:** GAP-007  
**Priority:** P3  
**Status:** `[x] done 2026-03-26`  
**Effort:** 5 min  

**What to do:**
- Open `web/src/api/study-groups.ts`
- If stub/empty → implement as part of TASK-006
- If already has functions → verify they match the actual endpoint paths

---

## Progress Tracker

| Task | Gap | Priority | Status |
|---|---|---|---|
| TASK-001 — `sbc_log` migration | GAP-001 | 🔴 P0 | `[x] done 2026-03-26` |
| TASK-002 — RSA keys + `.env` | GAP-005 | 🔴 P0 | `[x] done 2026-03-26` |
| TASK-003 — `milestones` Go package | GAP-002 | 🔴 P1 | `[x] done 2026-03-26` |
| TASK-004 — `study-groups` Go package | GAP-003 | 🔴 P1 | `[x] done 2026-03-26` |
| TASK-005 — `milestones` frontend | GAP-002 | 🔴 P1 | `[x] done 2026-03-26` |
| TASK-006 — `study-groups` frontend | GAP-003 | 🔴 P1 | `[x] done 2026-03-26` |
| TASK-007 — `resources` frontend | GAP-004 | 🟡 P2 | `[x] done 2026-03-26` |
| TASK-008 — `job-board` frontend | GAP-004 | 🟡 P2 | `[x] done 2026-03-26` |
| TASK-009 — `fundraising` frontend | GAP-004 | 🟡 P2 | `[x] done 2026-03-26` |
| TASK-010 — `committees` frontend | GAP-004 | 🟡 P2 | `[x] done 2026-03-26` |
| TASK-011 — `alumni` frontend | GAP-004 | 🟡 P2 | `[x] done 2026-03-26` |
| TASK-012 — verify `study-groups.ts` | GAP-007 | 🟢 P3 | `[x] done 2026-03-26` |

---

*HEIMDALL — All-Seeing Eye — Task file complete.*
