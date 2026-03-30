# Blue Ledger Gamification - Quick Reference & Checklists

**For:** @ant-man  
**Date:** March 30, 2026  
**Status:** Ready to Begin Implementation

---

## 📋 QUICK START CHECKLIST

### Pre-Implementation
- [ ] Review `/docs/TIER_1_IMPLEMENTATION_STEPS.md` (step-by-step guide)
- [ ] Review `/docs/STATUS_REPORT_TIER_1_AND_2.md` (complete status)
- [ ] Review `/docs/DEVELOPER_TASK_BREAKDOWN.md` (92 tasks overview)
- [ ] Run `git pull origin main` to get latest code
- [ ] Create feature branch: `git checkout -b feat/tier-1-integration`

### Database Setup
- [ ] Verify PostgreSQL 16 running locally
- [ ] Verify migrations folder has 023 files (streaks, notifications, login)
- [ ] Run migrations: `make migrate-up` (or equivalent)
- [ ] Verify columns added to `members` table (streak columns)
- [ ] Verify `notifications` table created with correct schema

### Backend Integration (6-8 hours)
- [ ] **T1.1: Daily Login Bonus** (30 min)
  - [ ] Add LoginBonusRepository to auth handler
  - [ ] Call AwardDailyLoginBonus in Login method
  - [ ] Test login bonus is awarded

- [ ] **T1.2: Streak Updates on Attendance** (1 hour)
  - [ ] Import streaks service in events handler
  - [ ] Add streaks service to handler struct
  - [ ] Call UpdateStreakForAttendance in CheckInQR
  - [ ] Verify streak updates on event check-in

- [ ] **T1.3: Streak Bonuses to XP** (30 min)
  - [ ] Import streaks service in xp service
  - [ ] Add streak bonus calculation to AwardXP
  - [ ] Test XP bonus on streak milestones

- [ ] **T1.4: Register Endpoints** (15 min)
  - [ ] Create streaks handler in main.go
  - [ ] Register GET /v1/members/:id/streak
  - [ ] Register GET /v1/chapters/:id/streaks/leaderboard
  - [ ] Test endpoints with curl/Postman

- [ ] **Backend Testing** (1-2 hours)
  - [ ] Write unit tests for streak service
  - [ ] Write unit tests for login bonus
  - [ ] Run full backend test suite
  - [ ] Verify no regressions

### Frontend Integration (2.5-3 hours)
- [ ] **T1.5: Add StreakWidget to Dashboard** (1 hour)
  - [ ] Import StreakWidget component
  - [ ] Add to dashboard JSX
  - [ ] Verify widget displays current streak
  - [ ] Test with real member data

- [ ] **T1.6: Notification System** (1 hour)
  - [ ] Create NotificationProvider (or use context)
  - [ ] Wrap App component with provider
  - [ ] Add notification render location (bottom-right)
  - [ ] Export notifications for components to use

- [ ] **T1.7: Sound Effects** (30 min)
  - [ ] Import soundEffects utility
  - [ ] Add sound on XP gain events
  - [ ] Add sound on level up
  - [ ] Add sound on streak milestone
  - [ ] Add volume control toggle to settings

- [ ] **Frontend Testing** (1-2 hours)
  - [ ] Test components render without errors
  - [ ] Test notifications appear and dismiss
  - [ ] Test sounds play (mute working)
  - [ ] Test StreakWidget with different data

### Integration Testing (2 hours)
- [ ] **T1.8: End-to-End Testing**
  - [ ] [ ] Full login flow (bonus awarded)
  - [ ] [ ] Event attendance → streak update
  - [ ] [ ] XP gain → streak bonus awarded
  - [ ] [ ] Notification appears → auto-dismiss
  - [ ] [ ] Sound plays → mute works
  - [ ] [ ] Leaderboard shows correct rankings
  - [ ] [ ] Dashboard displays streak
  - [ ] [ ] No database errors in logs
  - [ ] [ ] Performance: <100ms response time
  - [ ] [ ] No regression in existing features

### Final Verification
- [ ] All 8 tasks implemented
- [ ] All tests passing
- [ ] No console errors
- [ ] Database migrations rollback works
- [ ] Code review passed
- [ ] Ready for v1.1.0 release

---

## 📁 FILE REFERENCE GUIDE

### Tier 1 Integration Files

**Backend Files to Modify:**

| File | Task | Change |
|------|------|--------|
| `internal/auth/handler.go` | T1.1 | Add LoginBonusRepository, call AwardDailyLoginBonus |
| `internal/events/handler.go` | T1.2 | Add streaks service, call UpdateStreakForAttendance |
| `internal/xp/service.go` | T1.3 | Add streak bonus calculation to AwardXP |
| `cmd/server/main.go` | T1.4 | Register streak endpoints |
| (No new files needed) | T1.1-T1.4 | All existing files, just modifications |

**Backend Files Already Created:**

| File | Purpose | Status |
|------|---------|--------|
| `internal/streaks/service.go` | Streak logic | ✅ Ready |
| `internal/streaks/handler.go` | Streak endpoints | ✅ Ready |
| `internal/auth/loginbonus.go` | Login bonus logic | ✅ Ready |
| `migrations/022_streaks.up.sql` | Streak columns | ✅ Ready |
| `migrations/022_streaks.down.sql` | Rollback | ✅ Ready |
| `migrations/023_notifications_and_login.up.sql` | Notifications | ✅ Ready |
| `migrations/023_notifications_and_login.down.sql` | Rollback | ✅ Ready |

**Frontend Files to Modify:**

| File | Task | Change |
|------|------|--------|
| `src/pages/Dashboard.tsx` | T1.5 | Add StreakWidget component |
| `src/App.tsx` | T1.6 | Add NotificationProvider |
| Various components | T1.7 | Add soundEffects calls |

**Frontend Files Already Created:**

| File | Purpose | Status |
|------|---------|--------|
| `src/components/AchievementNotification.tsx` | Notification UI | ✅ Ready |
| `src/components/StreakWidget.tsx` | Streak display | ✅ Ready |
| `src/utils/soundEffects.ts` | Sound manager | ✅ Ready |
| `src/hooks/useAchievementNotifications.ts` | Notification hook | ✅ Ready |

### Documentation Files

| Document | Purpose | Read Time |
|----------|---------|-----------|
| `TIER_1_IMPLEMENTATION_STEPS.md` | **Step-by-step guide** | 20 min |
| `STATUS_REPORT_TIER_1_AND_2.md` | Overall status & metrics | 15 min |
| `DEVELOPER_TASK_BREAKDOWN.md` | All 92 tasks across 4 tiers | 30 min |
| `TIER_2_CHALLENGE_SYSTEM.md` | Tier 2 detailed specs | 30 min |
| `GAMIFICATION_ENHANCEMENTS.md` | Full 1000-line roadmap | 60 min |

---

## ⏱️ TIME ESTIMATES

### Tier 1 Complete Implementation

| Phase | Tasks | Hours | Details |
|-------|-------|-------|---------|
| **Backend Integration** | T1.1-T1.4 | 2.5 | Auth (30m), Events (1h), XP (30m), Routes (15m) |
| **Frontend Integration** | T1.5-T1.7 | 2.5 | Dashboard (1h), Provider (1h), Sounds (30m) |
| **Testing** | T1.8 | 2 | Unit, integration, E2E tests |
| **Code Review & Polish** | - | 1 | Reviews, minor fixes |
| **Total** | 1-8 tasks | **8 hours** | Can be done in 1 day of focused work |

### Tier 2 Challenge System

| Phase | Tasks | Hours | Details |
|-------|-------|-------|---------|
| **Database** | T2.1.1 | 1 | Migrations setup |
| **Service Layer** | T2.1.2-T2.1.4 | 7 | Service, repository, handlers |
| **Advanced Backend** | T2.1.5-T2.1.8 | 4 | Templates, rewards, search, admin |
| **Frontend Components** | T2.2.1-T2.2.6 | 12 | 6 React components |
| **Integration & Testing** | - | 3 | E2E, performance testing |
| **Total** | 14 tasks | **27 hours** | 1 week of focused work |

---

## 🧪 TESTING COMMANDS

### Backend Testing

```bash
# Run all backend tests
make test

# Run tests for specific package
make test PKG=./internal/streaks

# Run with coverage
make test-coverage

# Run migrations
make migrate-up

# Rollback migrations (test rollback)
make migrate-down

# Check migration status
make migrate-status
```

### Frontend Testing

```bash
# Start dev server
npm run dev

# Run frontend tests
npm run test

# Run with coverage
npm run test:coverage

# Build for production
npm run build

# Preview production build
npm run preview
```

### Manual Testing (Postman/curl)

```bash
# Login
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Get streak info
curl -X GET http://localhost:8080/v1/members/{memberId}/streak \
  -H "Authorization: Bearer {token}"

# Get leaderboard
curl -X GET http://localhost:8080/v1/chapters/{chapterId}/streaks/leaderboard \
  -H "Authorization: Bearer {token}"

# Join challenge
curl -X POST http://localhost:8080/v1/challenges/{challengeId}/join \
  -H "Authorization: Bearer {token}"
```

---

## 🔍 CODE LOCATIONS - QUICK REFERENCE

### Where to Add Code

**Daily Login Bonus Integration:**
```
File: blue-ledger-api/internal/auth/handler.go
Location: In Login method, after line ~110 (after h.svc.Login())
Action: Call loginBonusService.AwardDailyLoginBonus()
```

**Streak Update on Attendance:**
```
File: blue-ledger-api/internal/events/handler.go
Location: In CheckInQR method, line ~210-240
Action: Call streaksService.UpdateStreakForAttendance()
```

**Streak Bonus in XP:**
```
File: blue-ledger-api/internal/xp/service.go
Location: In AwardXP method, line ~117, after award XP
Action: Get streak, add bonus XP if streak milestone
```

**Register Routes:**
```
File: blue-ledger-api/cmd/server/main.go
Location: In main(), after event handlers
Action: Create streaksHandler, register 2 routes
```

**Add StreakWidget:**
```
File: web/src/pages/Dashboard.tsx
Location: Add to dashboard grid, line ~50-100
Action: Import and render <StreakWidget memberId={user.member.id} />
```

**Add Notification Provider:**
```
File: web/src/App.tsx
Location: Around root render, line ~50-100
Action: Wrap app with <NotificationProvider>
```

---

## 🚨 COMMON ISSUES & SOLUTIONS

### Issue: Migration fails to run
**Solution:** 
- Check PostgreSQL is running: `psql -U postgres`
- Check migrations folder exists with .sql files
- Verify migrations table exists: `SELECT * FROM public.migrations;`

### Issue: Streaks not updating on attendance
**Solution:**
- Verify migration 022 ran successfully
- Check streaks service is properly injected into events handler
- Add logging to UpdateStreakForAttendance to debug
- Verify member_id is correct UUID format

### Issue: Notifications not appearing
**Solution:**
- Check NotificationProvider is wrapping app in App.tsx
- Verify useAchievementNotifications hook is imported
- Check browser console for JavaScript errors
- Verify notification context is being used

### Issue: Sounds not playing
**Solution:**
- Check browser allows audio playback
- Verify sound files exist in public/sounds/ folder
- Check audio permissions in settings
- Test on different browser (Chrome, Firefox, Safari)

### Issue: StreakWidget shows "Loading forever"
**Solution:**
- Check API endpoint is registered in routes
- Verify JWT token is valid and includes member_id
- Check browser network tab for API response
- Ensure memberId prop is correct UUID

### Issue: Leaderboard performance slow
**Solution:**
- Verify indexes are created on challenge_participants table
- Limit query to top 100: `LIMIT 100`
- Add caching with Redis
- Use pagination for large leaderboards

---

## 📊 SUCCESS METRICS

### After Tier 1 Completion

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Daily Logins | +30% | Check analytics |
| Daily Active Users | +20-25% | Check user sessions |
| Session Duration | +10-15% | Check average session time |
| XP Engagement | +40% | Check XP earned per user |
| Feature Adoption | >70% | Check StreakWidget views |

### Performance Targets

| Operation | Target | Tools |
|-----------|--------|-------|
| Login → Bonus Award | <500ms | Application Performance Monitoring |
| Check-in → Streak Update | <1000ms | Database query logs |
| Get Streak Info | <100ms | Response time logging |
| Get Leaderboard | <500ms | Database query analysis |
| Notification Appear | <1000ms | Frontend performance metrics |

---

## 📞 GETTING HELP

### If Stuck On:

**Backend Integration Issues**
→ Check `/docs/TIER_1_IMPLEMENTATION_STEPS.md` (lines 100-300)

**Database Questions**
→ Check migrations at `blue-ledger-api/migrations/02*`

**Frontend Component Issues**
→ Check existing components in `web/src/components/` for patterns

**API Integration**
→ Check existing handlers (auth, events, xp) for patterns

**Testing Questions**
→ Check Makefile for test commands

---

## ✅ FINAL CHECKLIST BEFORE COMMIT

Before pushing code, ensure:

- [ ] All 8 Tier 1 tasks complete
- [ ] All tests passing (backend + frontend)
- [ ] No console errors in browser
- [ ] No lint warnings
- [ ] Database migrations tested (up and down)
- [ ] Code reviewed by peer
- [ ] Commit message: `feat: Complete Tier 1 quick wins integration`
- [ ] Push to GitHub: `git push origin feat/tier-1-integration`
- [ ] Create Pull Request with description
- [ ] Get code review approval
- [ ] Merge to main
- [ ] Tag version: `v1.1.0`

---

## 🚀 READY TO START!

**Your next steps:**

1. **Read** `TIER_1_IMPLEMENTATION_STEPS.md` (20 minutes)
2. **Review** existing code structure (30 minutes)
3. **Start** with T1.1 (Daily Login Bonus)
4. **Test** each task as you complete it
5. **Commit** when all 8 tasks are done

**Total Time: 6-8 focused hours** 

**You can have v1.1.0 ready TODAY!** 🎯

---

**Last Updated:** March 30, 2026  
**For:** @ant-man (Developer)  
**Status:** Ready for Implementation ✅

