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

---

## TASK-013 · Migration — `023_challenges` table
**Gap:** GAP-NEW-002  
**Priority:** P1 — prerequisite for challenge engine  
**Status:** `[x] done 2026-03-27`  
**Effort:** 15 min  
**Agent:** @ant-man  

**What to do:**
Create `blue-ledger-api/migrations/023_challenges.up.sql` and `.down.sql`.

```sql
-- 023_challenges.up.sql
CREATE TABLE IF NOT EXISTS challenges (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  challenger_id   UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  challenged_id   UUID        NOT NULL REFERENCES members(id)  ON DELETE CASCADE,
  type            TEXT        NOT NULL CHECK (type IN ('xp_duel','service_race','trivia','streak_showdown')),
  status          TEXT        NOT NULL DEFAULT 'pending'
                              CHECK (status IN ('pending','accepted','declined','active','completed','expired')),
  xp_stake        INT         NOT NULL DEFAULT 50 CHECK (xp_stake >= 10 AND xp_stake <= 500),
  game_data       JSONB       NOT NULL DEFAULT '{}',
  winner_id       UUID        REFERENCES members(id) ON DELETE SET NULL,
  expires_at      TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
  accepted_at     TIMESTAMPTZ,
  completed_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_challenges_chapter    ON challenges(chapter_id);
CREATE INDEX idx_challenges_challenger ON challenges(challenger_id);
CREATE INDEX idx_challenges_challenged ON challenges(challenged_id);
CREATE INDEX idx_challenges_status     ON challenges(status);
```

---

## TASK-014 · Backend — Go `challenges` package
**Gap:** GAP-NEW-002  
**Priority:** P1  
**Status:** `[x] done 2026-03-27`  
**Effort:** 3h  
**Agent:** @ant-man  
**Depends on:** TASK-013  

**What to do:**
Create `blue-ledger-api/internal/challenges/` with 3 files:

**`repository.go`** — raw SQL via pgx/v5:
- `Create(ctx, params)` → challenge row
- `GetByID(ctx, id, chapterID)` → challenge
- `ListForMember(ctx, memberID, chapterID)` → []challenge (active + recent)
- `UpdateStatus(ctx, id, status, winnerID)` → challenge
- `ExpireStale(ctx)` → bulk expire `pending/active` past `expires_at`

**`service.go`** — business logic:
- `ChallengeService.Send(ctx, challengerID, challengedID, type, xpStake)` — validates both members in same chapter, deduplicates active challenges, inserts row, publishes WS event `CHALLENGE_INVITE`
- `ChallengeService.Accept(ctx, memberID, challengeID)` — sets status=`accepted`, publishes `CHALLENGE_ACCEPTED`
- `ChallengeService.Decline(ctx, memberID, challengeID)` — sets status=`declined`
- `ChallengeService.Complete(ctx, challengeID, winnerID)` — sets status=`completed`, awards XP to winner via XPService, deducts from loser if xp_stake > 0
- `ChallengeService.ResolveTrivia(ctx, challengeID, answers map[memberID][]int)` — scores quiz, calls Complete

**`handler.go`** — Echo routes:
```
POST   /v1/challenges          — Send challenge (auth)
POST   /v1/challenges/:id/accept   — Accept (auth, must be challenged_id)
POST   /v1/challenges/:id/decline  — Decline (auth, must be challenged_id)
POST   /v1/challenges/:id/submit   — Submit trivia answers (auth)
GET    /v1/challenges           — List my challenges (active + last 20)
GET    /v1/challenges/:id       — Get challenge detail
```

**Register in `cmd/server/main.go`:**
```go
challengesSvc := challenges.NewService(pool, xpSvc)
challenges.NewHandler(challengesSvc, tokenManager).Register(v1)
```

---

## TASK-015 · Backend — WebSocket Hub + Presence System
**Gap:** GAP-NEW-001  
**Priority:** P1  
**Status:** `[x] done 2026-03-27`  
**Effort:** 3h  
**Agent:** @ant-man  

**What to do:**
Create `blue-ledger-api/internal/platform/hub.go`:

**Hub design:**
```go
type Hub struct {
    // chapter_id → set of client connections
    rooms   map[string]map[*Client]bool
    mu      sync.RWMutex
    redis   *redis.Client

    register   chan *Client
    unregister chan *Client
    broadcast  chan BroadcastMsg
}

type Client struct {
    hub        *Hub
    conn       *websocket.Conn
    chapterID  string
    memberID   string
    send       chan []byte
}

type BroadcastMsg struct {
    ChapterID string
    Type      string   // PRESENCE_UPDATE | CHALLENGE_INVITE | CHALLENGE_ACCEPTED | CHALLENGE_RESULT | NOTIFICATION | PROPS_RECEIVED
    Payload   any
}
```

**Presence flow:**
1. Client connects to `GET /v1/ws` with Bearer token in `?token=` query param (WS can't set headers)
2. Hub registers client, sets Redis key `presence:{chapterID}:{memberID}` with TTL 90s
3. Hub broadcasts `PRESENCE_UPDATE` to all chapter members with current online list
4. Client sends ping every 30s → hub refreshes TTL
5. On disconnect → delete Redis key → broadcast updated presence list
6. `GET /v1/presence` REST endpoint returns current online member IDs (from Redis SCAN)

**Add to `go.mod`:**
```bash
go get github.com/gorilla/websocket@v1.5.3
```

**Register in `main.go`:**
```go
hub := platform.NewHub(redisClient)
go hub.Run()
e.GET("/v1/ws", hub.HandleWebSocket, authMiddleware)
e.GET("/v1/presence", hub.HandlePresenceList, authMiddleware)
```

---

## TASK-016 · Frontend — `useWebSocket` + `usePresence` hooks
**Gap:** GAP-NEW-001  
**Priority:** P1  
**Status:** `[x] done 2026-03-27`  
**Effort:** 2h  
**Agent:** @ant-man  
**Depends on:** TASK-015  

**What to do:**

**`web/src/hooks/useWebSocket.ts`:**
```ts
// Singleton WebSocket connection per session
// Auto-reconnect with exponential backoff (1s → 2s → 4s → 8s → max 30s)
// Typed message dispatch via EventEmitter pattern
export function useWebSocket(): {
  send: (type: string, payload: unknown) => void
  on: (type: string, handler: (payload: unknown) => void) => () => void
  isConnected: boolean
}
```

**`web/src/hooks/usePresence.ts`:**
```ts
// Subscribes to PRESENCE_UPDATE messages from WS
// Falls back to GET /v1/presence polling every 60s if WS disconnected
export function usePresence(): {
  onlineMembers: string[]  // array of member IDs currently online
  isOnline: (memberID: string) => boolean
}
```

**Integration in `App.tsx`:**
- Initialize `useWebSocket()` at AppShell level
- Pass `onlineMembers` through context (`PresenceContext`)

---

## TASK-017 · Frontend — `OnlineBadge` component + Directory/Leaderboard integration
**Gap:** GAP-NEW-005  
**Priority:** P2  
**Status:** `[x] done 2026-03-27`  
**Effort:** 1h  
**Agent:** @ant-man  
**Depends on:** TASK-016  

**What to do:**

**`web/src/components/OnlineBadge.tsx`:**
```tsx
// Green pulsing dot overlay on member avatar
// Props: memberID: string
// Uses usePresence() to determine if online
export function OnlineBadge({ memberID }: { memberID: string }) {
  const { isOnline } = usePresence()
  if (!isOnline(memberID)) return null
  return <span className="online-dot" aria-label="Online now" />
}
```

**CSS in `base.css`:**
```css
.online-dot {
  position: absolute;
  bottom: 2px; right: 2px;
  width: 10px; height: 10px;
  background: #00ff88;
  border-radius: 50%;
  border: 2px solid #060D1A;
  animation: pulse-green 2s infinite;
}
@keyframes pulse-green {
  0%, 100% { box-shadow: 0 0 0 0 rgba(0,255,136,0.4); }
  50% { box-shadow: 0 0 0 6px rgba(0,255,136,0); }
}
```

**Add `<OnlineBadge>` to:**
- `DirectoryPage.tsx` — member card avatar wrapper
- `LeaderboardPage.tsx` — leaderboard row avatar
- `MemberProfilePage.tsx` — profile header avatar

**Wire DMs:** In `MessagesPage.tsx`, subscribe to WS event `MESSAGE_NEW` → invalidate TanStack Query `messages` cache to trigger re-fetch in near-real-time.

---

## TASK-018 · Frontend — `ChallengesPage`
**Gap:** GAP-NEW-002  
**Priority:** P2  
**Status:** `[x] done 2026-03-27`  
**Effort:** 2.5h  
**Agent:** @ant-man  
**Depends on:** TASK-014, TASK-016  

**What to do:**
Create `web/src/features/challenges/ChallengesPage.tsx` and `web/src/api/challenges.ts`.

**`web/src/api/challenges.ts`:**
```ts
listChallenges()     → GET /v1/challenges
sendChallenge(data)  → POST /v1/challenges
acceptChallenge(id)  → POST /v1/challenges/:id/accept
declineChallenge(id) → POST /v1/challenges/:id/decline
submitTrivia(id, answers) → POST /v1/challenges/:id/submit
```

**ChallengesPage UI sections:**
1. **Active Challenges** — incoming pending invites with Accept/Decline buttons; accepted challenges with status
2. **My Challenges** — challenges I sent (pending / in progress)
3. **Challenge History** — completed challenges with win/loss badge and XP gained/lost
4. **Stats bar** — W/L record, XP won from challenges, win streak

**Game type badge colors:**
- `xp_duel` → 🗡️ neon blue
- `service_race` → 🏃 green
- `trivia` → 🧠 gold
- `streak_showdown` → 🔥 orange

**Route:** Add `{ path: '/challenges', element: <ChallengesPage /> }` to `App.tsx`  
**Sidebar:** Add ⚔️ Challenges link under the gamification section

---

## TASK-019 · Frontend — `ChallengeModal` + profile integration
**Gap:** GAP-NEW-002  
**Priority:** P2  
**Status:** `[x] done 2026-03-27`  
**Effort:** 1.5h  
**Agent:** @ant-man  
**Depends on:** TASK-018  

**What to do:**
Create `web/src/components/ChallengeModal.tsx`.

**Props:**
```ts
interface ChallengeModalProps {
  isOpen: boolean
  onClose: () => void
  targetMember: { id: string; name: string; xp_total: number; level_key: string }
}
```

**UI:**
- Target member display: avatar + name + level badge
- Challenge type selector (4 cards with icons and descriptions)
- XP Stake slider: 10 → 500 XP (increments of 10)
- "Challenge to Battle" submit button
- Success state: "Challenge sent! 🗡️ Waiting for response..."
- WS event `CHALLENGE_INVITE` received by challenged member → toast notification "⚔️ {name} challenges you!"

**Add "⚔️ Challenge" button to:**
- `MemberProfilePage.tsx` — action bar (visible when viewing another member's profile)
- `DirectoryPage.tsx` — hover action on member card
- `LeaderboardPage.tsx` — hover action on leaderboard row

---

## TASK-020 · Backend + Frontend — Trivia Game Engine
**Gap:** GAP-NEW-002  
**Priority:** P3  
**Status:** `[x] done 2026-03-27`  
**Effort:** 3h  
**Agent:** @ant-man  
**Depends on:** TASK-014, TASK-016  

**What to do:**

**Backend (`data/QuizBank.xlsx` → `internal/challenges/trivia.go`):**
- Embed 50 PBS/chapter history trivia questions as `[]TriviaQuestion` (type, question, options[4], correct_index)
- On `Accept` of a `trivia` challenge → generate random 5-question set, store in `game_data` JSONB
- Timer: 120s per question set (tracked via `expires_at` update)
- `Submit` endpoint: scores answers, determines winner, calls `Complete`

**Frontend `web/src/features/challenges/TriviaGame.tsx`:**
- Full-screen game overlay activated when challenge status = `active` and type = `trivia`
- Question + 4 answer choices (A/B/C/D buttons)
- Progress bar (question X of 5)
- 120s countdown timer (red pulse when < 30s)
- Score reveal screen: "You got 4/5! Brother got 3/5 — YOU WIN 🏆 +100 XP"
- Confetti burst on win (`canvas-confetti` package)

---

## TASK-021 · Frontend — PWA Service Worker
**Gap:** GAP-NEW-003  
**Priority:** P3  
**Status:** `[x] done 2026-03-27`  
**Effort:** 2h  
**Agent:** @ant-man  

**What to do:**
Add PWA support to Vite build:

```bash
npm install -D vite-plugin-pwa
```

**`web/vite.config.ts`** — add `VitePWA` plugin:
```ts
VitePWA({
  registerType: 'autoUpdate',
  includeAssets: ['favicon.ico', 'assets/avatars/*.png'],
  manifest: false, // uses existing manifest.json
  workbox: {
    globPatterns: ['**/*.{js,css,html,ico,png,svg}'],
    runtimeCaching: [{
      urlPattern: /^https:\/\/.*\/v1\/(?!ws)/,
      handler: 'NetworkFirst',
      options: { cacheName: 'api-cache', networkTimeoutSeconds: 10 }
    }]
  }
})
```

**Benefits:**
- Offline access to cached pages
- Install prompt on mobile (Add to Home Screen)
- Background sync for service hours submission when offline

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
| TASK-013 — `023_challenges` migration | GAP-NEW-002 | 🔴 P1 | `[x] done 2026-03-27` |
| TASK-014 — `challenges` Go package | GAP-NEW-002 | 🔴 P1 | `[x] done 2026-03-27` |
| TASK-015 — WebSocket hub + presence | GAP-NEW-001 | 🔴 P1 | `[x] done 2026-03-27` |
| TASK-016 — `useWebSocket` + `usePresence` hooks | GAP-NEW-001 | 🔴 P1 | `[x] done 2026-03-27` |
| TASK-017 — `OnlineBadge` + DM real-time | GAP-NEW-005 | 🟡 P2 | `[x] done 2026-03-27` |
| TASK-018 — `ChallengesPage` frontend | GAP-NEW-002 | 🟡 P2 | `[x] done 2026-03-27` |
| TASK-019 — `ChallengeModal` + profile integration | GAP-NEW-002 | 🟡 P2 | `[x] done 2026-03-27` |
| TASK-020 — Trivia game engine | GAP-NEW-002 | 🟢 P3 | `[x] done 2026-03-27` |
| TASK-021 — PWA service worker | GAP-NEW-003 | 🟢 P3 | `[x] done 2026-03-27` |

---

*HEIMDALL — All-Seeing Eye — Task file complete. Last updated 2026-03-27.*
