# Quick Wins Implementation Guide

**Version:** 1.1.0 Gamification Phase 1  
**Status:** Implementation Complete  
**Effort:** ~20-30 hours development + testing

---

## Overview

This document outlines the implementation of 4 "quick win" gamification features for Blue Ledger that can be deployed in 1-2 weeks. These features provide immediate engagement boost and lay groundwork for advanced gamification.

---

## Features Implemented

### 1. Streak System 🔥

**Purpose:** Reward consistency and daily engagement

**Database Changes:**
```sql
-- Migration: 022_streaks.up.sql
ALTER TABLE members ADD COLUMN current_streak INT DEFAULT 0;
ALTER TABLE members ADD COLUMN longest_streak INT DEFAULT 0;
ALTER TABLE members ADD COLUMN streak_updated_at TIMESTAMP DEFAULT NOW();

-- Indexes for performance
CREATE INDEX idx_members_current_streak ON members(current_streak DESC);
CREATE INDEX idx_members_longest_streak ON members(longest_streak DESC);
```

**Backend Implementation:**

Location: `blue-ledger-api/internal/streaks/`

**Service** (`service.go`):
- `UpdateStreakForAttendance()` - Updates streak when member attends event
- `GetStreakInfo()` - Returns current/longest streak for a member
- `GetStreakLeaderboard()` - Returns top members by current streak
- `CalculateStreakBonus()` - Calculates XP bonus based on streak length

**Handler** (`handler.go`):
- `GetStreakInfo()` - GET /v1/members/:id/streak
- `GetStreakLeaderboard()` - GET /v1/chapters/:chapterId/streaks/leaderboard

**XP Bonuses:**
- 7-day streak: +50 XP
- 14-day streak: +100 XP
- 30-day streak: +250 XP

**Integration Points:**

1. **Attendance Recording:**
```go
// In attendance handler
if err == nil {
    streakService.UpdateStreakForAttendance(ctx, memberID)
    streakBonus := streakService.CalculateStreakBonus(streakInfo.CurrentStreak)
    xpService.AwardXP(memberID, 100 + streakBonus) // Base + streak bonus
}
```

2. **Dashboard Display:**
```go
// Add to member profile handler
streak, _ := streakService.GetStreakInfo(ctx, memberID)
response["streak"] = streak
response["next_milestone"] = calculateNextMilestone(streak.CurrentStreak)
```

---

### 2. Achievement Notifications (Pop-ups) 🏆

**Purpose:** Celebrate wins with animated UI notifications

**Database Changes:**
```sql
-- Migration: 023_notifications_and_login.up.sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    action_url VARCHAR(255),
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY(recipient_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_notifications_recipient_unread ON notifications(recipient_id, read_at)
WHERE read_at IS NULL;
```

**Frontend Implementation:**

Location: `web/src/components/`

**Component** (`AchievementNotification.tsx`):
- Animated slide-in from top
- Auto-dismiss after 5 seconds
- Type-specific colors and icons
- Progress bar indicating auto-dismiss time
- Manual close button

**Properties:**
```typescript
interface AchievementNotificationProps {
  title: string;              // "🏆 Badge Unlocked!"
  subtitle?: string;          // "Achievement description"
  xpGained?: number;          // 100
  badgeName?: string;         // "Dedication"
  type?: 'badge' | 'level_up' | 'streak' | 'daily_login';
  icon?: string;              // Emoji override
  autoDismiss?: number;       // 5000ms default
  onDismiss?: () => void;
}
```

**Types & Colors:**
```
badge       → yellow/amber gradient (🏆)
level_up    → purple/pink gradient (📈)
streak      → orange/red gradient (🔥)
daily_login → blue/cyan gradient (📅)
```

**Animations:**
- Entry: Spring bounce (stiffness: 300, damping: 30)
- Exit: Scale down + fade
- Icon: Bounce + rotate on appearance
- Progress bar: Linear shrink (scaleX)

**Usage in React:**

```typescript
import AchievementNotification from '@/components/AchievementNotification';

// In component
const [achievements, setAchievements] = useState<Achievement[]>([]);

return (
  <>
    {achievements.map((ach) => (
      <AchievementNotification
        key={ach.id}
        title={ach.title}
        subtitle={ach.subtitle}
        xpGained={ach.xp}
        type={ach.type}
        onDismiss={() => removeAchievement(ach.id)}
      />
    ))}
  </>
);
```

---

### 3. Daily Login Bonus 📅

**Purpose:** Reward daily engagement with consistent XP

**Database Changes:**
```sql
-- Migration: 023_notifications_and_login.up.sql
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP;
ALTER TABLE users ADD COLUMN daily_login_streak INT DEFAULT 0;
CREATE INDEX idx_users_last_login ON users(last_login_at);
```

**Backend Implementation:**

Location: `blue-ledger-api/internal/auth/`

**Repository** (`loginbonus.go`):
- `AwardDailyLoginBonus()` - Awards XP if first login of day
- `GetLoginStreak()` - Returns current login streak

**Logic:**
```
IF last_login_at is NULL:
  daily_login_streak = 1
  xp_awarded = 25
ELSE:
  last_login_day = last_login_at.Date()
  today_day = now.Date()
  
  IF last_login_day == today_day:
    already_rewarded_today()
  ELSE IF last_login_day == yesterday:
    daily_login_streak++
    xp_awarded = 25
    IF daily_login_streak % 7 == 0:
      xp_awarded += 50  // Weekly milestone bonus
  ELSE:
    daily_login_streak = 1
    xp_awarded = 25
```

**Bonus Structure:**
- Daily: +25 XP
- Weekly (every 7 days): +50 bonus XP (total: 75 XP)

**Integration:**

```go
// In auth/handler.go login handler
loginBonus, err := loginBonusRepo.AwardDailyLoginBonus(ctx, userID)
if err == nil && loginBonus.Awarded {
    xpService.AwardXP(userID, loginBonus.XPGranted)
    notifService.CreateNotification(ctx, userID, "daily_login", 
        "Daily Login Bonus", 
        fmt.Sprintf("Streak: %d days", loginBonus.CurrentStreak))
}
```

---

### 4. Sound Effects 🎵

**Purpose:** Provide audio feedback for achievements and actions

**Frontend Implementation:**

Location: `web/src/utils/soundEffects.ts`

**Sound Manager Class:**
```typescript
class SoundEffectsManager {
  play(type: SoundType): void
  playSuccess(): void      // Attendance, XP earned
  playError(): void         // Failed action
  playLevelUp(): void       // Level up fanfare
  playAchievement(): void   // Badge unlocked
  playStreak(): void        // Streak milestone
  playClick(): void         // UI click
  
  setVolume(volume: number): void  // 0-1
  mute(): void
  unmute(): void
  isMuted(): boolean
}
```

**Sound Types:**
| Type | Usage | Example |
|------|-------|---------|
| success | XP gained, attendance logged | +100 XP popup |
| error | Invalid action, failure | Can't join event |
| levelup | Level milestone reached | Leveled up! |
| achievement | Badge/achievement unlocked | Badge earned! |
| streak | Streak milestone reached | 7-day streak! |
| click | UI interactions | Button click |

**User Preferences:**
- Stored in localStorage: `soundEffectsConfig`
- Survives page refresh
- Volume (0-1) and muted state persisted

**Integration:**

```typescript
import { soundEffects } from '@/utils/soundEffects';

// Award XP
soundEffects.playSuccess();
showNotification('100 XP earned!');

// Level up
soundEffects.playLevelUp();
showNotification('Level Up!');

// Settings
<VolumeControl 
  value={soundEffects.getVolume()}
  onChange={soundEffects.setVolume}
/>
<button onClick={() => soundEffects.mute()}>
  Mute Sound
</button>
```

---

## React Hook: `useAchievementNotifications`

Centralized hook for managing achievement notifications with sound:

```typescript
Location: web/src/hooks/useAchievementNotifications.ts

const {
  notifications,              // Array of active notifications
  showAchievement,           // Generic notification
  showBadgeUnlocked,         // Badge: 🏆
  showLevelUp,               // Level up: 📈
  showStreakMilestone,       // Streak: 🔥
  showDailyLoginBonus,       // Daily login: 📅
  showXPGained,              // XP gain: ⭐
  showError,                 // Error: ❌
  removeNotification,
  clearAll,
} = useAchievementNotifications();

// Usage
showBadgeUnlocked('Dedication', 100);
showLevelUp(5, 'Silver');
showStreakMilestone(7, 50);
showDailyLoginBonus(25, 1);
```

---

## React Component: `StreakWidget`

Displays member streak information with animations:

```typescript
Location: web/src/components/StreakWidget.tsx

<StreakWidget 
  memberId={memberId}
  className="mb-4"
  showAnimation={true}
/>
```

**Display:**
- Current streak (large, animated flame)
- Personal best
- Next milestone and bonus XP
- Progress bar to next milestone
- Color gradient based on streak length

---

## API Endpoints

### Streaks

```
GET /v1/members/{id}/streak
Returns:
{
  "member_id": "uuid",
  "current_streak": 7,
  "longest_streak": 21,
  "updated_at": "2026-03-30T10:00:00Z"
}

GET /v1/chapters/{chapterId}/streaks/leaderboard?limit=100
Returns:
{
  "data": [
    {
      "rank": 1,
      "member_id": "uuid",
      "name": "Marcus Williams",
      "current_streak": 30,
      "longest_streak": 45,
      "photo_url": "..."
    },
    ...
  ],
  "total": 87
}
```

### Notifications

```
GET /v1/notifications
GET /v1/notifications/unread
PUT /v1/notifications/{id}/read
PUT /v1/notifications/read-all
```

### Login Bonus

```
POST /v1/auth/login-bonus
Returns:
{
  "awarded": true,
  "xp_granted": 25,
  "current_streak": 1,
  "bonus_type": "daily",
  "message": "Daily login bonus awarded!"
}
```

---

## Integration Checklist

### Backend
- [x] Database migrations (streaks, notifications, login)
- [x] Streaks service (tracking, calculations, leaderboard)
- [x] Notifications service (CRUD operations)
- [x] Login bonus repository (tracking, awarding)
- [x] API handlers for streaks
- [ ] Update auth handler to call login bonus
- [ ] Update attendance handler to update streaks
- [ ] Update XP service to award streak bonuses
- [ ] Register handlers in routes
- [ ] Run migrations on database

### Frontend
- [x] Achievement notification component (animated)
- [x] Sound effects utility (manager class)
- [x] useAchievementNotifications hook
- [x] StreakWidget component
- [ ] Integrate notification hook in Layout/App
- [ ] Show notifications on XP events
- [ ] Show notifications on level up
- [ ] Show notifications on badge unlock
- [ ] Add StreakWidget to dashboard
- [ ] Add StreakWidget to member profile
- [ ] Add volume control to settings
- [ ] Test all sounds and animations

### Testing
- [ ] Unit tests: Streak calculations
- [ ] Unit tests: Login bonus logic
- [ ] Unit tests: Notification creation
- [ ] Integration tests: Login bonus flow
- [ ] Integration tests: Streak update on attendance
- [ ] E2E: Achievement notification appears
- [ ] E2E: Sound plays (muted in CI)
- [ ] E2E: StreakWidget displays correctly
- [ ] Manual: Test all notification types
- [ ] Manual: Verify streak calculations

---

## Deployment Steps

### 1. Database

```bash
# In blue-ledger-api/
go run cmd/migrate/main.go up
# Runs: 022_streaks.up.sql and 023_notifications_and_login.up.sql
```

### 2. Backend

```bash
# Update handlers to integrate new features
# Build and test
make build
make test

# Deploy
make deploy
```

### 3. Frontend

```bash
# Install any new dependencies (if needed)
cd web
npm install

# Build
npm run build

# Deploy
npm run deploy
```

---

## Performance Considerations

### Database
- **Streaks**: Indexed on `current_streak` and `longest_streak` for fast leaderboard queries
- **Notifications**: Indexed on `recipient_id` and `read_at` for fast unread queries
- **Batch operations**: Use prepared statements for bulk notification creation

### Frontend
- **Sound preloading**: Preload sounds on app startup for instant playback
- **Animation optimization**: Use `transform` and `opacity` for GPU acceleration
- **Memory**: Remove old notifications to prevent memory leaks

### API
- **Leaderboard**: Cache for 5 minutes (updates after each attendance)
- **Notification list**: Paginated (limit 20 per page)
- **Streak info**: Fetch on demand (not in every API call)

---

## Future Enhancements

### Phase 2 (4-6 weeks):
- Challenge system with leaderboards
- Cosmetics shop (avatar customization)
- Seasonal resets

### Phase 3 (8-12 weeks):
- Social features (friending, rivalry)
- Team/Guild system
- 100+ achievements

### Phase 4 (12+ weeks):
- Rewards/loot system
- Ranked seasons with divisions
- Seasonal events

---

## Troubleshooting

### Sounds not playing
- Check browser autoplay policy
- Verify `localStorage` is accessible
- Check browser console for errors
- Ensure `soundEffects` is initialized before use

### Streaks not updating
- Verify migration ran: `SELECT * FROM members LIMIT 1`
- Check `streak_updated_at` is being updated
- Verify attendance event is triggering streak update
- Check database transaction commits

### Notifications not appearing
- Verify notification table exists
- Check `recipient_id` is correct
- Verify notification component is mounted in Layout
- Check browser console for component errors

### Animations laggy
- Check for expensive re-renders
- Use React.memo for notification components
- Profile with Chrome DevTools
- Reduce animation complexity for slower devices

---

## File Manifest

### Backend
- `blue-ledger-api/migrations/022_streaks.up.sql` - Streak columns
- `blue-ledger-api/migrations/022_streaks.down.sql` - Rollback
- `blue-ledger-api/migrations/023_notifications_and_login.up.sql` - Notifications + login
- `blue-ledger-api/migrations/023_notifications_and_login.down.sql` - Rollback
- `blue-ledger-api/internal/streaks/service.go` - Streak logic
- `blue-ledger-api/internal/streaks/handler.go` - Streak endpoints
- `blue-ledger-api/internal/auth/loginbonus.go` - Login bonus logic

### Frontend
- `web/src/components/AchievementNotification.tsx` - Pop-up component
- `web/src/components/StreakWidget.tsx` - Streak display
- `web/src/utils/soundEffects.ts` - Sound manager
- `web/src/hooks/useAchievementNotifications.ts` - Notification hook

---

## Timeline

| Week | Task | Status |
|------|------|--------|
| 1 | Database migrations | ✅ Complete |
| 1 | Backend services | ✅ Complete |
| 1 | Frontend components | ✅ Complete |
| 2 | Integration & testing | 🔄 In Progress |
| 2 | Bug fixes & polish | ⏳ Pending |
| 2 | Production deployment | ⏳ Pending |

---

## Success Metrics

Track these after deployment:

| Metric | Target | Measurement |
|--------|--------|-------------|
| Daily Active Users | +25% | Analytics |
| Session Duration | +15% | Analytics |
| Daily Logins | +40% | Login events |
| Engagement Events | +50% | Event tracking |
| Notification Adoption | >80% | User feedback |
| Sound Adoption | >60% | Settings |

---

## Support

- **Questions?** File GitHub issue
- **Bug?** Report with screenshots/logs
- **Enhancement?** Submit PR with tests

---
