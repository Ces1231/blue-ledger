# 🚀 Local Testing Guide - Tier 1 Gamification

## ✅ Docker Compose Stack is Running

Your complete Blue Ledger environment with Tier 1 gamification features is now running locally!

---

## 📊 Services Status

| Service | URL | Port | Status |
|---------|-----|------|--------|
| **Frontend** | http://localhost:3001 | 3001 | ✅ Running |
| **Backend API** | http://localhost:8081 | 8081 | ✅ Running |
| **PostgreSQL** | localhost | 5432 | ✅ Running |
| **Redis** | localhost | 6379 | ✅ Running |

---

## 🧪 Testing Tier 1 Features

### Feature 1: Daily Login Bonus (T1.1)

**What it does:** Awards 25 XP on first login of the day, +50 XP every 7 days

**How to test:**
1. Navigate to http://localhost:3001
2. Click "Login" and use magic link or test credentials
3. On successful login:
   - ✅ Dashboard should show +25 XP (base bonus)
   - ✅ Login streak counter increments to 1
   - ✅ Notification appears (if available)

**Expected API Calls:**
```bash
# Check daily login award
curl -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

---

### Feature 2: Streak Updates on Attendance (T1.2)

**What it does:** Automatically tracks and increments streaks when members attend events

**How to test:**
1. Navigate to Events section
2. Create or select an event
3. Check in (scan QR or manual ID entry)
4. Observe:
   - ✅ Member's streak increments by 1
   - ✅ Attendance is recorded
   - ✅ Notification shows event check-in

**Expected behavior:**
- Same day multiple check-ins: Streak stays same
- Next day check-in: Streak increments
- Missed day: Streak resets

---

### Feature 3: Streak Bonuses in XP Calculation (T1.3)

**What it does:** Automatically adds XP bonuses based on current streak level

**How to test:**
1. Maintain a streak:
   - Day 1-6: No bonus
   - Day 7+: +50 XP bonus per activity
   - Day 14+: +100 XP bonus
   - Day 30+: +250 XP bonus

2. Check attendance after streak milestones:
   - ✅ XP total increases by activity_xp + streak_bonus
   - ✅ Dashboard shows updated XP total
   - ✅ Notifications show bonus breakdown

**Expected XP Calculation:**
```
Event Base XP: 100
Streak: 7 days
Total Awarded: 100 + 50 = 150 XP
```

---

### Feature 4: Streak Endpoints (T1.4)

**What it does:** API endpoints to retrieve member and chapter streak data

**Test with curl:**

```bash
# Get member streak info
curl -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:8081/v1/streaks/members/{member-id}

# Response:
{
  "member_id": "uuid",
  "current_streak": 7,
  "longest_streak": 15,
  "updated_at": "2026-03-30T16:00:00Z"
}

# Get chapter leaderboard
curl -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:8081/v1/streaks/chapters/{chapter-id}/leaderboard

# Response:
{
  "data": [
    {
      "rank": 1,
      "member_name": "John Doe",
      "current_streak": 30,
      "longest_streak": 45
    },
    ...
  ],
  "total": 35
}
```

---

### Feature 5: StreakWidget on Dashboard (T1.5)

**What it does:** Interactive dashboard widget showing streak progress

**How to test:**
1. Login and navigate to Dashboard
2. Look for the animated streak widget with:
   - 🔥 Flame icon (animated)
   - Current streak display (e.g., "5 days")
   - Personal best display (e.g., "15 days")
   - Progress bar to next milestone
   - Color coding:
     - Gray: 0-6 days
     - Yellow: 7-13 days  
     - Orange: 14-29 days
     - Red: 30+ days

3. **Verify animations:**
   - ✅ Flame icon scales and rotates smoothly
   - ✅ Progress bar animates to current value
   - ✅ Colors update based on streak level

---

### Feature 6: Notification Provider (T1.6)

**What it does:** Real-time notification system with auto-polling

**How to test:**
1. Open browser DevTools console (F12)
2. Check for notification API calls every 30 seconds
3. Perform an XP-earning action:
   - Login → notification appears
   - Attend event → notification appears
   - Level up → notification appears
4. Verify:
   - ✅ Notification count updates
   - ✅ Unread badge shows
   - ✅ Notifications clear when clicked
   - ✅ Auto-poll continues every 30 seconds

**API Test:**
```bash
# Fetch current user's notifications
curl -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:8081/v1/notifications

# Response:
{
  "data": [
    {
      "id": "uuid",
      "title": "Daily Login Bonus",
      "body": "+25 XP awarded",
      "type": "login_bonus",
      "read": false,
      "created_at": "2026-03-30T16:30:00Z"
    }
  ]
}
```

---

### Feature 7: Sound Effects (T1.7)

**What it does:** Audio feedback for gamification events

**How to test:**
1. Ensure browser allows audio
2. Perform XP-earning actions:
   - ✅ Login → success sound
   - ✅ Level up → fanfare sound
   - ✅ Streak milestone → special sound
   - ✅ Achievement → achievement sound

3. Test sound control:
   - Mute/unmute toggle in settings
   - Volume slider
   - Sound preferences save to localStorage

**DevTools Check:**
```javascript
// In browser console:
localStorage.getItem('soundEffectsConfig')
// Should show: {"volume":0.7,"muted":false}
```

---

### Feature 8: End-to-End Test (T1.8)

**Full workflow test:**

1. **Reset and prepare:**
   ```bash
   # Clear existing notifications (optional)
   curl -X DELETE http://localhost:8081/v1/notifications \
     -H "Authorization: Bearer $JWT_TOKEN"
   ```

2. **Execute full flow:**
   - [ ] Login → observe daily bonus
   - [ ] Navigate to Dashboard → see StreakWidget
   - [ ] Go to Events → create/join event
   - [ ] Check in → trigger streak update
   - [ ] Check XP total → verify streak bonus applied
   - [ ] View notifications → see all events logged
   - [ ] Hear sounds → confirm audio working
   - [ ] Return to Dashboard → see updated widget

3. **Verify all data:**
   ```bash
   # Check member XP total
   curl -H "Authorization: Bearer $JWT_TOKEN" \
     http://localhost:8081/v1/members/me | jq '.xp_total'
   
   # Check streak info
   curl -H "Authorization: Bearer $JWT_TOKEN" \
     http://localhost:8081/v1/streaks/members/$(id) | jq '.'
   
   # Check notifications
   curl -H "Authorization: Bearer $JWT_TOKEN" \
     http://localhost:8081/v1/notifications | jq '.data | length'
   ```

---

## 🔧 Useful Commands

### View Logs

```bash
# All services
docker-compose logs -f

# Just API
docker-compose logs -f api

# Just Web
docker-compose logs -f web

# Just Database
docker-compose logs -f postgres

# Last 50 lines
docker-compose logs --tail=50
```

### Stop Services

```bash
# Stop all (keep data)
docker-compose stop

# Stop and remove containers (keep data)
docker-compose down

# Stop and remove everything (DELETE DATA!)
docker-compose down -v
```

### Restart

```bash
# Restart all services
docker-compose restart

# Restart just API
docker-compose restart api

# Full rebuild and start
docker-compose down -v && docker-compose up -d
```

### Database Access

```bash
# Access PostgreSQL directly
docker exec -it blue-ledger-api-postgres-1 psql -U blue_ledger -d blue_ledger

# View all tables
\dt

# Check streaks table
SELECT * FROM streaks LIMIT 10;

# Check notifications table
SELECT * FROM notifications LIMIT 10;

# Exit
\q
```

### Redis Access

```bash
# Connect to Redis
docker exec -it blue-ledger-api-redis-1 redis-cli

# View keys
KEYS *

# Check specific key
GET user:{user-id}:streak

# Exit
exit
```

---

## 📱 Frontend Testing

### Dashboard URL
- **http://localhost:3001/dashboard**
- Shows StreakWidget, XP progress, and notifications

### Events URL
- **http://localhost:3001/events**
- Check in to events, trigger streak updates

### Member Directory
- **http://localhost:3001/members**
- View all members and their stats

### Leaderboard
- **http://localhost:3001/leaderboard**
- See XP and streak rankings

---

## 🛠️ API Testing Examples

### Get JWT Token

```bash
# Login to get token
curl -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@tausigmasigma.org"}' \
  | jq '.access_token' -r

# Save as environment variable
TOKEN=$(curl -s -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@tausigmasigma.org"}' \
  | jq '.access_token' -r)

echo $TOKEN
```

### Test Each Tier 1 Endpoint

```bash
# 1. Daily Login Bonus
curl -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# 2. Get Member Streak
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8081/v1/streaks/members/{member-id}

# 3. Get Chapter Leaderboard
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8081/v1/streaks/chapters/{chapter-id}/leaderboard

# 4. Get Notifications
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8081/v1/notifications

# 5. Award XP (with streak bonus)
curl -X POST http://localhost:8081/v1/xp/award \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"member_id":"uuid","xp_amount":100,"reason":"event_attendance"}'
```

---

## ⚠️ Troubleshooting

### API Won't Start

```bash
# Check logs
docker-compose logs api

# Common issues:
# 1. Database not ready
# 2. Redis connection failed
# 3. Migration error

# Fix: Restart with fresh database
docker-compose down -v
docker-compose up -d
```

### Frontend Shows Blank Page

```bash
# Check frontend logs
docker-compose logs web

# Verify API is responding
curl http://localhost:8081/v1/healthz

# Restart web service
docker-compose restart web
```

### Can't Connect to Database

```bash
# Check database is running
docker-compose ps postgres

# Check connection
docker exec -it blue-ledger-api-postgres-1 pg_isready

# Verify credentials in docker-compose.yml
# Username: blue_ledger
# Password: localpassword
# Database: blue_ledger
```

### No Sound Effects

```bash
# Check browser console for errors (F12)
# Verify soundEffects are loaded:
localStorage.getItem('soundEffectsConfig')

# Check browser permissions for audio
# Some browsers require user interaction before audio plays
```

---

## 📊 Expected Test Results

| Feature | Status | Notes |
|---------|--------|-------|
| T1.1 Daily Bonus | ✅ | +25 XP on first login |
| T1.2 Streak Update | ✅ | Increment on attendance |
| T1.3 Streak Bonus | ✅ | Auto-calculated on XP |
| T1.4 Endpoints | ✅ | JWT-protected endpoints |
| T1.5 Dashboard Widget | ✅ | Animated with colors |
| T1.6 Notifications | ✅ | Polling every 30 sec |
| T1.7 Sound Effects | ✅ | Audio on all events |
| T1.8 E2E Integration | ✅ | All features working |

---

## 📝 Testing Checklist

- [ ] Frontend loads at http://localhost:3001
- [ ] API responds to health check
- [ ] Can login with test user
- [ ] Daily bonus awarded on login
- [ ] StreakWidget displays on dashboard
- [ ] Can check in to event
- [ ] Streak updates after check-in
- [ ] XP increases with streak bonus
- [ ] Notifications appear
- [ ] Sound plays on XP event
- [ ] All endpoints return proper responses
- [ ] No errors in console logs

---

## 🎯 Next Steps

1. **Manual Testing:** Follow the testing checklist above
2. **API Testing:** Use provided curl commands
3. **Performance Test:** Monitor resource usage with `docker stats`
4. **Data Verification:** Check database with SQL queries
5. **Ready for Staging:** When satisfied, prepare for staging deployment

---

**Status: ✅ LOCAL TESTING ENVIRONMENT READY**

All services running. Begin testing Tier 1 gamification features!
