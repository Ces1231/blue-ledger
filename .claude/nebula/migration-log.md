# Nebula Migration Log
Generated: 2026-03-24
Migration tool: golang-migrate
Migration directory: blue-ledger-api/migrations/
Last existing migration: none (initial schema)
New migrations: 001 through 020 (40 files total — 20 up + 20 down)

---

## Verdict: SAFE TO RUN

All 20 migrations are additive-only. No existing data touched.
No DROP, no TRUNCATE, no TYPE changes. Every migration uses
CREATE TABLE IF NOT EXISTS and CREATE INDEX IF NOT EXISTS.

---

## Migration Suite Summary

| # | File | Tables Created | Safety | Rollback |
|---|------|----------------|--------|----------|
| 001 | users_and_chapters | chapters, users, auth_sessions, magic_link_tokens | SAFE | EASY |
| 002 | rls_policies | (roles + policies only) | SAFE | EASY |
| 003 | members | members | SAFE | EASY |
| 004 | engagement_log | engagement_log | SAFE | EASY |
| 005 | events_and_rsvps | events, rsvps | SAFE | EASY |
| 006 | dues | dues_records | SAFE | EASY |
| 007 | service_log | service_log | SAFE | EASY |
| 008 | badges_and_quests | badges, member_badges, quests, member_quest_progress | SAFE | EASY |
| 009 | announcements_and_props | announcements, props | SAFE | EASY |
| 010 | notifications | notifications | SAFE | EASY |
| 011 | mentorships_and_intake | mentorships, intake_prospects | SAFE | EASY |
| 012 | votes | votes, vote_responses | SAFE | EASY |
| 013 | minutes_and_scholarships | chapter_minutes, scholarships | SAFE | EASY |
| 014 | store_and_goals | store_items, store_orders, chapter_goals | SAFE | EASY |
| 015 | job_board_and_resources | job_board, resources | SAFE | EASY |
| 016 | point_economy_and_config | point_economy, chapter_config | SAFE | EASY |
| 017 | audit_log | audit_log | SAFE | EASY |
| 018 | messaging | message_threads, messages | SAFE | EASY |
| 019 | fundraising_and_committees | fundraising_campaigns, committees | SAFE | EASY |
| 020 | study_groups_and_milestones | study_groups, milestones | SAFE | EASY |

**Total tables created: 31**

---

## Design Decisions & Flags

### 1. RLS Architecture
- Every tenant table has ENABLE ROW LEVEL SECURITY + chapter_isolation policy inline.
- Policy pattern: `USING (chapter_id = current_setting('app.chapter_id')::uuid)`.
- auth_sessions and magic_link_tokens use `user_isolation` (user_id scoped, not chapter_id).
- audit_log is NOT RLS-protected — INSERT only for blue_ledger_app, full access for blue_ledger_admin.

### 2. FK Rules — No Cross-Database Foreign Keys
- All member_id / chapter_id references are within the same database (single-DB architecture).
- No REFERENCES violations. Standard FK with appropriate ON DELETE behavior throughout.
- ON DELETE CASCADE: chapter_id FKs (chapter deleted = all data purged).
- ON DELETE CASCADE: user_id on auth_sessions, magic_link_tokens (user deleted = sessions revoked).
- ON DELETE SET NULL: awarded_by, verified_by, recorder_id, reviewer_id (nullable actor references).
- ON DELETE RESTRICT: posted_by, created_by on immutable records (force explicit cleanup).

### 3. Append-Only Tables (no updated_at, no UPDATE grants)
- engagement_log: INSERT only. The XP audit trail must not be mutated.
- props: INSERT only. Props are immutable once given.
- vote_responses: INSERT only. Votes cast cannot be changed.
- messages: INSERT + UPDATE (for read_by array append), no DELETE via app role.

### 4. JSONB Columns
- chapters.settings — XP level thresholds, notification prefs, feature flags.
- badges.criteria — badge engine criteria; extensible without schema migration.
- quests.steps — step definitions array.
- member_quest_progress.progress — per-step completion counts.
- announcements: no JSONB (structured enough to be columns).
- audit_log.metadata — flexible per-action context.

### 5. UUID Arrays (TEXT[])
- message_threads.participant_ids — GIN indexed.
- messages.read_by — GIN indexed.
- committees.member_ids — GIN indexed.
- study_groups.member_ids — GIN indexed.
- chapter_minutes.attendee_ids — not GIN indexed (lookup pattern is array overlap less frequent).
- All UUID[] columns use DEFAULT '{}'.

### 6. Seed Function
- seed_default_point_economy(UUID) in migration 016 is idempotent (ON CONFLICT DO NOTHING).
- 12 default activities from prototype with XP values matching prototype POINT_ECONOMY.
- Called per chapter during onboarding, NOT run as a migration-time side effect.

### 7. check_in_window_minutes on events
- Added per spec (005). Default 120 minutes. The QR scanner validates:
  `NOW() < event.event_date + event_time + check_in_window_minutes`.
- Application layer enforces this; DB stores the config.

### 8. Scholarships — member_id nullable
- External applicants (non-members) can apply: member_id IS NULL.
- If a member applies, member_id links to their profile for display.

### 9. Grants Summary
- blue_ledger_app: SELECT/INSERT/UPDATE/DELETE on all tables except:
  - engagement_log, props, vote_responses: SELECT + INSERT only (immutable).
  - audit_log: INSERT only (no reads via app role).
  - chapters: SELECT + UPDATE only (cannot create/drop chapters via app role).
- blue_ledger_admin: ALL PRIVILEGES on all tables (BYPASSRLS on the role itself).

---

## Code Changes Required After Running

The following application code must be added to complete the schema integration:

- `pkg/db/context.go` — implement SetChapterContext() using `SET LOCAL app.chapter_id = $1`
- `pkg/db/context.go` — implement SetUserContext() using `SET LOCAL app.current_user_id = $1`
- `internal/auth/service.go` — call seed_default_point_economy() in chapter registration handler
- `internal/platform/service.go` — provision blue_ledger_app and blue_ledger_admin login users
  with actual DB credentials (roles created in 002 are NOLOGIN — attach to real DB users)

---

## Run Command

```bash
# Run all migrations up
migrate -path blue-ledger-api/migrations \
        -database "$DATABASE_URL" \
        up

# Run a specific migration
migrate -path blue-ledger-api/migrations \
        -database "$DATABASE_URL" \
        up 1

# Rollback one migration
migrate -path blue-ledger-api/migrations \
        -database "$DATABASE_URL" \
        down 1

# Rollback all (DESTRUCTIVE — dev only)
migrate -path blue-ledger-api/migrations \
        -database "$DATABASE_URL" \
        down
```

---

## Files Written

All files in:
`/Users/carnell.smithgotyto.com/Documents/blue-ledger/blue-ledger-api/migrations/`

```
001_users_and_chapters.up.sql
001_users_and_chapters.down.sql
002_rls_policies.up.sql
002_rls_policies.down.sql
003_members.up.sql
003_members.down.sql
004_engagement_log.up.sql
004_engagement_log.down.sql
005_events_and_rsvps.up.sql
005_events_and_rsvps.down.sql
006_dues.up.sql
006_dues.down.sql
007_service_log.up.sql
007_service_log.down.sql
008_badges_and_quests.up.sql
008_badges_and_quests.down.sql
009_announcements_and_props.up.sql
009_announcements_and_props.down.sql
010_notifications.up.sql
010_notifications.down.sql
011_mentorships_and_intake.up.sql
011_mentorships_and_intake.down.sql
012_votes.up.sql
012_votes.down.sql
013_minutes_and_scholarships.up.sql
013_minutes_and_scholarships.down.sql
014_store_and_goals.up.sql
014_store_and_goals.down.sql
015_job_board_and_resources.up.sql
015_job_board_and_resources.down.sql
016_point_economy_and_config.up.sql
016_point_economy_and_config.down.sql
017_audit_log.up.sql
017_audit_log.down.sql
018_messaging.up.sql
018_messaging.down.sql
019_fundraising_and_committees.up.sql
019_fundraising_and_committees.down.sql
020_study_groups_and_milestones.up.sql
020_study_groups_and_milestones.down.sql
```
