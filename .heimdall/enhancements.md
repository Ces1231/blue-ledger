# HEIMDALL — Enhancement Plan
**Feature:** Real-Time Presence + Member Challenge Engine  
**Requested:** 2026-03-27  
**Status:** PLANNED — Not yet implemented  
**Author:** @jarvis  

---

## Executive Summary

**Yes — this is fully feasible** with the current stack. Redis 7 is already running, Echo v4 natively supports WebSocket upgrades, and the gamification engine already awards XP — we simply need to add a competitive layer on top.

The enhancement transforms Blue Ledger from a *passive tracking app* into a **live social arena** where members can see who's online, challenge brothers to competitions, and earn/lose XP in real-time battles.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    BROWSER (React)                       │
│                                                          │
│  useWebSocket() ──────────────────────────────────────  │
│       │          Singleton WS conn (auto-reconnect)      │
│       ▼                                                  │
│  usePresence()    ChallengeModal    TriviaGame           │
│       │                │                 │               │
│  OnlineBadge     ChallengesPage    DirectoryPage         │
└───────┬─────────────────────────────────────────────────┘
        │  ws://localhost:8081/v1/ws?token=<JWT>
        │
┌───────▼─────────────────────────────────────────────────┐
│                 GO API (Echo v4)                          │
│                                                          │
│  platform/hub.go ──── Hub.Run() goroutine               │
│       │                   │                              │
│  Per-client goroutine    Redis Pub/Sub                   │
│  (read pump / write pump)  channel: presence:{chapterID} │
│       │                                                  │
│  challenges/handler.go                                   │
│  challenges/service.go ── XPService (award/deduct)       │
│  challenges/repository.go ── pgx/v5 → challenges table  │
└───────┬──────────────────────────────────────────────────┘
        │
┌───────▼──────────────────────────────────────────────────┐
│  REDIS 7                                                  │
│  Keys:  presence:{chapterID}:{memberID}  TTL=90s         │
│  Sets:  (none — use SCAN pattern)                        │
│  Pub/Sub: presence:{chapterID}                           │
│           challenges:{chapterID}                         │
└───────────────────────────────────────────────────────────┘
```

---

## Part 1 — Real-Time Presence System

### How It Works

1. **Connect** — When a member loads the app, `useWebSocket()` opens `ws://.../v1/ws?token=<JWT>` (JWT passed as query param since browsers can't set WS headers)
2. **Register** — Go Hub receives the connection, validates the JWT, extracts `memberID` + `chapterID`, adds client to the chapter's room
3. **Heartbeat** — Client sends `{"type":"PING"}` every 30s; Hub refreshes Redis key `presence:{chapterID}:{memberID}` (TTL 90s)
4. **Presence broadcast** — On any join/leave, Hub re-scans Redis for `presence:{chapterID}:*` keys, builds online member list, broadcasts `PRESENCE_UPDATE` to all chapter clients
5. **Disconnect** — On WS close, Hub removes client, deletes Redis key, broadcasts updated list
6. **REST fallback** — `GET /v1/presence` returns current online IDs from Redis SCAN (for initial page load before WS connects)

### WebSocket Message Types

```typescript
// Server → Client
type WSMessage =
  | { type: 'PRESENCE_UPDATE';   payload: { onlineMembers: string[] } }
  | { type: 'CHALLENGE_INVITE';  payload: Challenge }
  | { type: 'CHALLENGE_ACCEPTED';payload: { challengeID: string; challengerName: string } }
  | { type: 'CHALLENGE_RESULT';  payload: { challengeID: string; winnerID: string; xpDelta: number } }
  | { type: 'NOTIFICATION';      payload: Notification }
  | { type: 'PROPS_RECEIVED';    payload: { from: string; message: string; xp: number } }
  | { type: 'MESSAGE_NEW';       payload: { conversationID: string; senderName: string } }
  | { type: 'PONG';              payload: null }

// Client → Server
type WSClientMsg =
  | { type: 'PING' }
  | { type: 'CHALLENGE_INVITE'; payload: { challengedID: string; type: ChallengeType; xpStake: number } }
```

### UI — Online Indicators

```
Directory Page                  Leaderboard Page
┌──────────────────────────┐    ┌───────────────────────────────────┐
│ 👤  Marcus Williams  🟢  │    │  1   👤🟢  Marcus J.   4,200 XP  │
│     Chair • 2,400 XP     │    │  2   👤    DeShawn C.  3,800 XP  │
│    [⚔️ Challenge]        │    │  3   👤🟢  Jordan H.   2,100 XP  │
└──────────────────────────┘    └───────────────────────────────────┘
       🟢 = online now                🟢 = online indicator (OnlineBadge)
```

---

## Part 2 — Member Challenge Engine

### Challenge Flow State Machine

```
                    ┌─────────┐
                    │ PENDING │  ← Member A sends challenge to Member B
                    └────┬────┘
          WS: CHALLENGE_INVITE → B
                         │
              ┌──────────┴──────────┐
              ▼                     ▼
        ┌──────────┐         ┌──────────┐
        │ ACCEPTED │         │ DECLINED │  ← challenge closed
        └────┬─────┘         └──────────┘
   WS: CHALLENGE_ACCEPTED → A
             │
             ▼
        ┌─────────┐
        │  ACTIVE │  ← both members enter game session
        └────┬────┘
             │
    ┌────────┴────────┐
    │ time runs out   │ game complete
    ▼                 ▼
┌─────────┐     ┌───────────┐
│ EXPIRED │     │ COMPLETED │  ← winner determined, XP transferred
└─────────┘     └───────────┘
             WS: CHALLENGE_RESULT → both members
```

### Challenge Types

| Type | Icon | Duration | How Winner Is Determined |
|------|------|----------|--------------------------|
| **XP Duel** | ⚡ | 24h | First to earn 100 XP from any activity |
| **Service Race** | 🏃 | 7 days | Most approved service hours logged |
| **Trivia Battle** | 🧠 | 10 min | 5 PBS/chapter trivia questions, best score |
| **Streak Showdown** | 🔥 | Next 3 events | First to check in to 3 consecutive events |

### XP Stakes

- Minimum stake: **10 XP**
- Maximum stake: **500 XP**
- Winner receives: `+xp_stake` XP (awarded via `XPService.AwardXP`)
- Loser loses: `xp_stake / 2` XP (deducted — never below 0)
- If declined or expired: no XP change

### Database Schema

```sql
challenges (
  id              UUID PK
  chapter_id      UUID → chapters (RLS)
  challenger_id   UUID → members
  challenged_id   UUID → members
  type            TEXT CHECK (xp_duel|service_race|trivia|streak_showdown)
  status          TEXT DEFAULT 'pending' CHECK (pending|accepted|declined|active|completed|expired)
  xp_stake        INT  DEFAULT 50
  game_data       JSONB DEFAULT '{}'   -- stores: trivia questions, scores, progress
  winner_id       UUID → members NULL
  expires_at      TIMESTAMPTZ DEFAULT NOW() + '24h'
  accepted_at     TIMESTAMPTZ NULL
  completed_at    TIMESTAMPTZ NULL
  created_at      TIMESTAMPTZ
)
```

### `game_data` JSONB by type

**Trivia:**
```json
{
  "questions": [
    { "id": 12, "question": "What year was Phi Beta Sigma founded?", "options": ["1910","1914","1918","1920"], "correct": 1 }
  ],
  "answers": {
    "challenger": [1, 0, 2, 1, 3],
    "challenged": [1, 1, 2, 0, 3]
  },
  "scores": { "challenger": 4, "challenged": 3 }
}
```

**XP Duel:**
```json
{
  "target_xp": 100,
  "challenger_start_xp": 2450,
  "challenged_start_xp": 1200,
  "challenger_current_xp": 2480,
  "challenged_current_xp": 1260
}
```

---

## Part 3 — UI Screens

### Challenges Page (`/challenges`)

```
⚔️ CHALLENGES                                    [+ Challenge a Brother]

─── INCOMING ────────────────────────────────────────────────────────
  🧠 TRIVIA BATTLE                                      🟡 PENDING
  DeShawn Carter challenges you to Trivia Battle
  Stake: 75 XP  •  Expires in 23h
  [✅ Accept]  [❌ Decline]

─── ACTIVE ─────────────────────────────────────────────────────────
  ⚡ XP DUEL                                            🔵 ACTIVE
  You vs. Jordan Hayes
  Stake: 50 XP  •  Goal: earn 100 XP  •  3d remaining
  You: +30 XP  ██████░░░░  Jordan: +20 XP  ████░░░░░░

─── HISTORY ─────────────────────────────────────────────────────────
  🏃 SERVICE RACE     vs. Marcus Williams    ✅ YOU WON  +100 XP
  🧠 TRIVIA BATTLE    vs. Jordan Hayes       ❌ YOU LOST  -25 XP
  ⚡ XP DUEL          vs. DeShawn Carter     ✅ YOU WON  +50 XP

─── YOUR STATS ──────────────────────────────────────────────────────
  🏆 3 Wins  •  💀 1 Loss  •  ⚡ +125 XP earned from challenges
```

### Challenge Modal

```
┌─────────────────────────────────────────────────────┐
│  ⚔️ CHALLENGE DeShawn Carter                    [×] │
│  Gold Legend  •  3,800 XP                           │
│                                                     │
│  SELECT CHALLENGE TYPE                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────┐  │
│  │  ⚡ XP   │  │ 🏃 SVC  │  │ 🧠 QUIZ │  │🔥  │  │
│  │  DUEL   │  │  RACE   │  │ BATTLE  │  │STK │  │
│  │  24h    │  │  7 days │  │  10min  │  │3ev │  │
│  └──────────┘  └──────────┘  └──────────┘  └────┘  │
│                                                     │
│  XP STAKE  ─────────────────●─────────────── 500   │
│  10                        100                     │
│                                                     │
│  You risk: 50 XP  •  You could win: +50 XP         │
│                                                     │
│         [⚔️ SEND CHALLENGE]                         │
└─────────────────────────────────────────────────────┘
```

### Trivia Game Screen

```
┌────────────────────────────────────────────────────────────────┐
│  🧠 TRIVIA BATTLE  vs. Marcus Williams          2:04 ⏱️ 🔴    │
│  Question 3 of 5              ███████████░░░░░░░░░░ 60%       │
│                                                                │
│  "In what city was Phi Beta Sigma Fraternity founded?"         │
│                                                                │
│  [A] Philadelphia, PA       [B] Washington, D.C.              │
│  [C] Atlanta, GA             [D] New York, NY                  │
│                                                                │
│  Score  You: 2/2 ✅✅    Marcus: 1/2 ✅❌                       │
└────────────────────────────────────────────────────────────────┘
```

---

## Part 4 — Implementation Sequence

Execute in this order to avoid blocked tasks:

```
Week 1 — Infrastructure
  TASK-013  Migration 023_challenges  (15 min)
  TASK-015  WebSocket hub + presence  (3h)
  TASK-016  useWebSocket + usePresence hooks (2h)

Week 2 — Challenge Engine
  TASK-014  Go challenges package  (3h)
  TASK-017  OnlineBadge + DM real-time  (1h)
  TASK-018  ChallengesPage frontend  (2.5h)
  TASK-019  ChallengeModal + profile  (1.5h)

Week 3 — Polish
  TASK-020  Trivia game engine  (3h)
  TASK-021  PWA service worker  (2h)

Total effort: ~18h
```

---

## Part 5 — Agent Assignments

| Task | Agent | Rationale |
|------|-------|-----------|
| TASK-013 migration | @ant-man | SQL migration specialist |
| TASK-014 Go challenges pkg | @ant-man | Go backend service writer |
| TASK-015 WebSocket hub | @ant-man | Go concurrency + Redis |
| TASK-016 WS/Presence hooks | @ant-man | React/TS specialist |
| TASK-017 OnlineBadge | @ant-man | React component builder |
| TASK-018 ChallengesPage | @ant-man | Feature page builder |
| TASK-019 ChallengeModal | @ant-man | UI component |
| TASK-020 Trivia engine | @ant-man | Full-stack feature |
| TASK-021 PWA service worker | @ant-man | Vite/PWA config |

---

## Part 6 — Future Enhancements (Post-MVP)

| Enhancement | Description | Effort |
|-------------|-------------|--------|
| **Team Challenges** | Pledge class vs. pledge class service hour race | 4h |
| **Tournament Mode** | Bracket-style chapter tournament, top 8 → single elimination | 8h |
| **Live Chat in Game** | In-game WS chat during trivia battles | 2h |
| **Challenge Notifications** | Push notification via PWA when challenged (requires TASK-021) | 2h |
| **Leaderboard integration** | "Challenge #1 Ranked Member" button on Leaderboard | 1h |
| **Spectator Mode** | Chapter members can watch active trivia battle in read-only mode | 4h |
| **AI-Generated Questions** | Use Claude to generate new chapter-specific trivia questions | 2h |

---

*HEIMDALL — All-Seeing Eye — Enhancement plan complete. Last updated 2026-03-27.*
