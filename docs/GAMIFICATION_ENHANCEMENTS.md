# Blue Ledger: Gamification Enhancement Guide

**Version:** 1.0.0 Enhancement Roadmap  
**Status:** Recommended Features for v1.1+  
**Priority:** High-Impact Engagement Features

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Tier 1: Quick Wins (1-2 weeks)](#tier-1-quick-wins-1-2-weeks)
- [Tier 2: High Impact (2-4 weeks)](#tier-2-high-impact-2-4-weeks)
- [Tier 3: Advanced Features (4-8 weeks)](#tier-3-advanced-features-4-8-weeks)
- [Tier 4: Platform Features (8+ weeks)](#tier-4-platform-features-8-weeks)
- [Implementation Checklist](#implementation-checklist)
- [Database Schema Changes](#database-schema-changes)

---

## Executive Summary

Blue Ledger has strong fundamentals (XP, levels, badges). To make it **feel like a real game**, add:

1. **Streaks** - Reward consistency (daily logins, consecutive meetings)
2. **Challenges** - Time-limited quests with rewards
3. **Seasons** - Leaderboard resets with themed competitions
4. **Notifications** - Real-time achievement pop-ups and alerts
5. **Cosmetics** - Avatar customization and rewards
6. **Daily Tasks** - Repeating activities with bonuses
7. **Social Features** - Friending, rivalry, team-ups
8. **Sound/Animations** - Audio and visual feedback

---

## Tier 1: Quick Wins (1-2 weeks)

### 1.1 Streak System

**What:** Track consecutive days/meetings without missing

**Implementation:**
```go
// Add to members table
current_streak INT DEFAULT 0
longest_streak INT DEFAULT 0
streak_updated_at TIMESTAMP

// In engagement tracking
if attendance_today && yesterday_attended {
    current_streak++
} else if !attendance_today {
    current_streak = 0
}
```

**UI Changes:**
```
┌─────────────────────┐
│  🔥 Streak: 12 Days│
│                    │
│  Next reward @ 21  │
│  days: 250 XP      │
└─────────────────────┘
```

**XP Bonus:**
- 7-day streak: 50 XP bonus
- 14-day streak: 100 XP bonus
- 30-day streak: 250 XP bonus

**Effort:** 3-4 hours

---

### 1.2 Achievement Pop-ups

**What:** Celebrate wins with animated notifications

**Implementation:**
```javascript
// Frontend component
<AchievementNotification
  title="🏆 First Streak!"
  subtitle="3 consecutive meetings"
  xpGained={50}
  animation="slideInDown"
/>

// Show for 5 seconds then fade
// Queue multiple notifications
```

**Triggers:**
- Badge unlocked
- Level up
- Streak milestone (7, 14, 21, 30 days)
- XP milestone (500, 1000, 1500, 2000)
- Top 3 leaderboard position
- Officer role assigned

**Effort:** 2-3 hours

---

### 1.3 Daily Login Bonus

**What:** XP reward for checking in each day

**Implementation:**
```go
// Add to users table
last_login_at TIMESTAMP
daily_login_streak INT DEFAULT 0

// On login
if today != last_login {
    daily_login_streak++
    awardXP(user, 25) // Daily bonus
    
    if daily_login_streak % 7 == 0 {
        awardXP(user, 50) // Weekly milestone
    }
}
```

**UI:**
```
📅 Daily Login Bonus
├─ Day 1: 25 XP ✓
├─ Day 2: 25 XP ✓
├─ Day 3: 25 XP ⏳
├─ Day 7: +50 XP 🎁
└─ Come back tomorrow!
```

**Effort:** 2 hours

---

### 1.4 Sound Effects & Visual Feedback

**What:** Audio + animations for actions

**Implementation:**
```javascript
// Audio library
import useSound from 'use-sound';

const [playSuccess] = useSound('/sounds/success.mp3');
const [playError] = useSound('/sounds/error.mp3');
const [playLevelUp] = useSound('/sounds/levelup.mp3');

// On XP gain
const handleXPGain = (xp) => {
  playSuccess();
  animateXPPop(xp); // Floating text animation
};
```

**Effects:**
- ✓ Success sound (attendance, XP earned)
- ✗ Error sound (failed quiz)
- 📈 Level-up sound (fanfare)
- 🏆 Achievement sound (badge unlock)
- ⚠️ Streak warning (1 day left)

**Sound Source:** Use royalty-free from freesound.org or game sound packs

**Effort:** 2-3 hours

---

## Tier 2: High Impact (2-4 weeks)

### 2.1 Challenge System

**What:** Time-limited quests with leaderboards

**Database:**
```sql
CREATE TABLE challenges (
    id UUID PRIMARY KEY,
    chapter_id UUID FOREIGN KEY,
    title VARCHAR,
    description TEXT,
    duration INT, -- days
    max_participants INT,
    xp_reward INT,
    badge_reward UUID NULLABLE,
    started_at TIMESTAMP,
    ends_at TIMESTAMP,
    status ENUM('active', 'completed', 'archived')
);

CREATE TABLE challenge_participants (
    challenge_id UUID FOREIGN KEY,
    member_id UUID FOREIGN KEY,
    progress INT,
    rank INT,
    PRIMARY KEY(challenge_id, member_id)
);
```

**Example Challenges:**
```
⚡ "Attendance Warrior"
├─ Attend all meetings this week
├─ Participants: 23/30
├─ Top Reward: 500 XP + Badge
├─ Your Progress: 2/4 meetings
└─ Time Left: 3 days

⚡ "Service Sprint"
├─ Log 10 service hours
├─ Participants: 15/30
├─ Top Reward: 300 XP
├─ Your Progress: 4/10 hours
└─ Time Left: 7 days

⚡ "Quiz Master"
├─ Pass 5 quizzes
├─ Participants: 8/30
├─ Top Reward: 250 XP
├─ Your Progress: 2/5 quizzes
└─ Time Left: 14 days
```

**Admin Interface:**
```
[Create New Challenge]

Title: "Attendance Warrior"
Description: "Attend all 4 meetings this week"
Duration: 7 days
Max Participants: Unlimited
XP Reward: 500
Badge Reward: [Select Badge]
[Create Challenge]
```

**Effort:** 8-10 hours (backend + frontend)

---

### 2.2 Seasonal Leaderboard

**What:** Monthly/semester resets with themes

**Implementation:**
```go
// Add to leaderboards table
season VARCHAR, -- "Spring 2026", "Fall 2026"
season_start_at TIMESTAMP
season_end_at TIMESTAMP
seasonal_ranking BOOLEAN

// Query seasonal XP (only count during season)
SELECT member_id, SUM(xp) as seasonal_xp
FROM engagement_log
WHERE created_at BETWEEN season_start_at AND season_end_at
GROUP BY member_id
ORDER BY seasonal_xp DESC
```

**UI Tabs:**
```
┌─────────────────────────────────┐
│ [All Time] [Season] [Monthly]   │
├─────────────────────────────────┤
│                                 │
│ 🏆 Spring 2026 Leaderboard     │
│ (Reset March 1 - May 31)       │
│                                 │
│ #1  Marcus     4,200 XP  👑    │
│ #2  DeShawn    3,950 XP  🥇   │
│ #3  Elijah     3,620 XP  🥈   │
│ ...                            │
└─────────────────────────────────┘
```

**Seasonal Rewards:**
- Top 3 get special badge
- Top 10 get XP bonus
- Reset leaderboard each season
- Archive previous season

**Effort:** 6-8 hours

---

### 2.3 Cosmetics Shop

**What:** Use XP to buy avatar customization

**Features:**
```
Avatar Customization
├─ Hats (beanie, cap, crown)
├─ Shirts (colors, patterns)
├─ Accessories (chains, rings, badges)
├─ Backgrounds (custom themes)
└─ Emotes (celebration animations)

XP Shop
├─ Blue Hat: 300 XP
├─ Gold Chain: 500 XP
├─ Championship Badge: 1000 XP
├─ Custom Background: 750 XP
└─ Victory Emote: 200 XP
```

**Database:**
```sql
CREATE TABLE cosmetics (
    id UUID PRIMARY KEY,
    name VARCHAR,
    type ENUM('hat', 'shirt', 'accessory', 'background', 'emote'),
    xp_cost INT,
    image_url VARCHAR
);

CREATE TABLE member_cosmetics (
    member_id UUID FOREIGN KEY,
    cosmetic_id UUID FOREIGN KEY,
    purchased_at TIMESTAMP
);
```

**Effort:** 6-8 hours

---

### 2.4 Notifications System

**What:** Real-time alerts for key events

**Events:**
- Member joins chapter
- You earned a badge
- Level up!
- XP milestone reached
- Leaderboard ranking changed
- Friend activity update
- Challenge started/ending
- Event reminder (1 day before)
- Your props were liked

**Implementation:**
```go
// In-app notifications (stored in DB)
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    recipient_id UUID,
    type ENUM('badge', 'level_up', 'ranking', 'event', 'social'),
    title VARCHAR,
    message TEXT,
    action_url VARCHAR NULLABLE,
    read_at TIMESTAMP NULLABLE,
    created_at TIMESTAMP
);

// Real-time via WebSocket
socket.emit('notification', {
    type: 'badge_unlocked',
    badge_name: 'Dedicated',
    xp_gained: 100
});
```

**UI:**
```
Notification Bell
├─ 🔴 3 unread
│  ├─ 🏆 You unlocked "Dedicated" badge!
│  ├─ 📈 You leveled up to Gold!
│  └─ 🎯 Quiz starts in 1 day
└─ [Mark all as read]
```

**Effort:** 8-10 hours (backend + frontend + WebSocket)

---

## Tier 3: Advanced Features (4-8 weeks)

### 3.1 Social Features

**Friending System:**
```sql
CREATE TABLE friendships (
    user_id UUID,
    friend_id UUID,
    status ENUM('pending', 'accepted', 'blocked'),
    created_at TIMESTAMP,
    PRIMARY KEY(user_id, friend_id)
);
```

**Friend Activity:**
```
👥 Friends
├─ Marcus Williams
│  ├─ Just leveled up to 👑 Icon!
│  ├─ +250 XP for service
│  └─ 5 hours ago
│
└─ DeShawn Carter
   ├─ Unlocked "Scholar" badge
   └─ 2 hours ago

[Add Friend] [View Friends' Activity]
```

**Rivalry System:**
```
🔥 Rivals
├─ Compare XP head-to-head
├─ Predict who will rank higher
├─ Show progress race bar
└─ Celebrate when you overtake them
```

**Effort:** 8-10 hours

---

### 3.2 Team/Guild System

**What:** Groups compete for XP bonuses

**Features:**
```sql
CREATE TABLE teams (
    id UUID PRIMARY KEY,
    chapter_id UUID,
    name VARCHAR,
    leader_id UUID,
    description TEXT,
    created_at TIMESTAMP
);

CREATE TABLE team_members (
    team_id UUID,
    member_id UUID,
    joined_at TIMESTAMP,
    role ENUM('leader', 'member')
);
```

**Team Mechanics:**
- Earn 10% XP bonus for team activities
- Team leaderboard (sum of member XP)
- Team challenges (compete against other teams)
- Team perks (unlock at milestones)

**UI:**
```
⚔️ Teams
├─ Tau Sigma Sigma A Team
│  ├─ Members: 15
│  ├─ Total XP: 45,000
│  ├─ Rank: #2
│  └─ Perks: +10% XP
│
└─ Tau Sigma Sigma B Team
   ├─ Members: 12
   ├─ Total XP: 38,000
   ├─ Rank: #3
   └─ Perks: +5% XP
```

**Effort:** 10-12 hours

---

### 3.3 Achievement System (Expanded)

**What:** More granular achievements than badges

```
Achievements (100 total)

Combat Zone
├─ "First Blood" - Attend first meeting (5 XP)
├─ "On a Roll" - Attend 5 consecutive meetings (50 XP)
├─ "Unstoppable" - 30-day attendance streak (250 XP)
└─ "Legend" - 365+ consecutive day login (500 XP)

Service Zone
├─ "Volunteer" - Log 5 service hours (50 XP)
├─ "Community Hero" - 100 service hours (300 XP)
├─ "Guardian Angel" - Highest service hours in semester (500 XP)
└─ "Lifesaver" - 1000+ service hours (1000 XP)

Social Zone
├─ "Making Friends" - Add 5 friends (25 XP)
├─ "Popular" - 25+ friends (100 XP)
├─ "Props Master" - Give 100 props (150 XP)
└─ "Loved One" - Receive 100 props (250 XP)

Knowledge Zone
├─ "Scholar" - Pass all quizzes (100 XP)
├─ "Genius" - Perfect score on quiz (75 XP)
├─ "Quiz Master" - 500+ quiz points (300 XP)
└─ "Omniscient" - Master all topics (500 XP)

Officer Zone
├─ "Promoted" - Become E-Board (100 XP)
├─ "Leader" - Serve as officer (150 XP)
├─ "President" - Serve as President (300 XP)
└─ "Legacy" - Multiple terms as officer (500 XP)
```

**Progression Tracking:**
```
Achievement Progress:
├─ "On a Roll" - 3/5 consecutive meetings ⏳
├─ "Making Friends" - 12/25 friends ⏳
├─ "Scholar" - 4/5 quizzes passed ⏳
└─ "Volunteer" - 8/10 service hours ⏳
```

**Effort:** 8-10 hours

---

### 3.4 Daily/Weekly Quests

**What:** Repeating tasks for consistent engagement

```
Daily Quests (Reset at midnight)
├─ Log in to app (25 XP) ✓
├─ View leaderboard (25 XP) ✓
├─ Update profile (50 XP) ⏳
└─ Reward for all 3: 25 XP bonus

Weekly Quests (Reset Sunday)
├─ Attend meeting (100 XP) ✓
├─ Give 3 props (75 XP) ✓
├─ Log service hours (50 XP) ⏳
├─ Complete quiz (100 XP) ⏳
└─ Reward for all 4: 150 XP bonus

Progress Tracker
[████░░] 75% Complete
Next Reward: 50 XP
```

**Effort:** 6-8 hours

---

## Tier 4: Platform Features (8+ weeks)

### 4.1 Rewards/Loot System

**What:** Randomized rewards for activities

```
Loot Tiers:
├─ Common (70%) - 25 XP
├─ Rare (20%) - 50 XP + cosmetic
├─ Epic (8%) - 100 XP + cosmetic + avatar frame
├─ Legendary (2%) - 250 XP + exclusive cosmetic

Loot Drop Animation:
[Chest Opening]
You found: 🌟 Rare Reward!
  ├─ 50 XP
  ├─ Blue Hat (cosmetic)
  └─ Share on feed?
```

**Implementation:**
```go
func generateLoot() LootBox {
    roll := rand.Float64() // 0.0-1.0
    
    switch {
    case roll < 0.70:
        return CommonLoot(25 * xpMultiplier)
    case roll < 0.90:
        return RareLoot(50 * xpMultiplier)
    case roll < 0.98:
        return EpicLoot(100 * xpMultiplier)
    default:
        return LegendaryLoot(250 * xpMultiplier)
    }
}
```

**Effort:** 8-10 hours

---

### 4.2 Combo System

**What:** Bonus XP for multiple activities

```
Active Combos:
├─ Attendance Combo x2
│  └─ Attended meeting + logged into app
│  └─ Bonus: +25 XP (25% increase)
│
├─ Service Combo x3
│  └─ Logged service + admin approved + gave props
│  └─ Bonus: +75 XP (50% increase)
│
└─ Social Combo x2
   └─ Gave props + friended someone
   └─ Bonus: +15 XP (20% increase)
```

**Effort:** 6-8 hours

---

### 4.3 Seasonal Events

**What:** Limited-time themed activities

```
🎃 Fall Festival (Oct 1-31)
├─ Attend "Pumpkin Carving" event
├─ Earn 2x XP for all activities
├─ Unlock "Spooky Badge"
├─ Compete for "Festival King" title
└─ Exclusive cosmetics: Pumpkin hat, ghost emote

🎄 Winter Celebration (Dec 1-31)
├─ Attend "Holiday Party"
├─ Earn holiday XP multiplier
├─ Unlock "Festive Badge"
├─ White Elephant gift exchange integration
└─ Exclusive cosmetics: Santa hat, snow background

🏀 Spring Games (Mar-Apr)
├─ Sports tournament bracket
├─ Team competitions
├─ Earn "Champion" badge
└─ Leaderboard reset

🎓 Graduation Season (May)
├─ Honor seniors
├─ Transfer achievements
├─ Alumni badge unlock
└─ Special farewell events
```

**Effort:** 10-15 hours per event

---

### 4.4 Tournaments/Ranked Seasons

**What:** Ranked competition system

```
🏆 Ranked Season 1: Spring 2026

Division System:
├─ Bronze (0-500 XP)
├─ Silver (500-1000 XP)
├─ Gold (1000-1500 XP)
├─ Platinum (1500-2000 XP)
└─ Diamond (2000+ XP)

Your Division: Gold II (1,250 XP)
Progress: ███░░░░░░ 25% to Platinum

Ranked Rewards (at Season End):
├─ Bronze: Avatar frame
├─ Silver: Avatar frame + 50 XP
├─ Gold: Avatar frame + 100 XP + Badge
├─ Platinum: Avatar frame + 200 XP + Badge
└─ Diamond: Avatar frame + 500 XP + Exclusive Badge
```

**Effort:** 12-15 hours

---

## Implementation Checklist

### Phase 1: Quick Wins (Week 1-2)
- [ ] Streaks system
- [ ] Achievement pop-ups
- [ ] Daily login bonus
- [ ] Sound effects & animations
- [ ] Deploy and test

### Phase 2: High Impact (Week 3-6)
- [ ] Challenge system
- [ ] Seasonal leaderboards
- [ ] Cosmetics shop
- [ ] Notifications system
- [ ] Deploy and gather feedback

### Phase 3: Advanced (Week 7-12)
- [ ] Social features (friending, rivalry)
- [ ] Team/Guild system
- [ ] Expanded achievements (100+)
- [ ] Daily/Weekly quests
- [ ] Deploy and iterate

### Phase 4: Platform (Week 13+)
- [ ] Rewards/loot system
- [ ] Combo system
- [ ] Seasonal events
- [ ] Tournaments/ranked
- [ ] Continuous updates

---

## Database Schema Changes

### New Tables to Add

```sql
-- Streaks
ALTER TABLE members ADD COLUMN current_streak INT DEFAULT 0;
ALTER TABLE members ADD COLUMN longest_streak INT DEFAULT 0;
ALTER TABLE members ADD COLUMN streak_updated_at TIMESTAMP;

-- Challenges
CREATE TABLE challenges (
    id UUID PRIMARY KEY,
    chapter_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    duration INT,
    max_participants INT,
    xp_reward INT,
    badge_reward UUID NULLABLE,
    started_at TIMESTAMP,
    ends_at TIMESTAMP,
    status ENUM('active', 'completed', 'archived'),
    created_at TIMESTAMP,
    FOREIGN KEY(chapter_id) REFERENCES chapters(id),
    FOREIGN KEY(badge_reward) REFERENCES badges(id)
);

CREATE TABLE challenge_participants (
    challenge_id UUID NOT NULL,
    member_id UUID NOT NULL,
    progress INT DEFAULT 0,
    rank INT NULLABLE,
    completed_at TIMESTAMP NULLABLE,
    PRIMARY KEY(challenge_id, member_id),
    FOREIGN KEY(challenge_id) REFERENCES challenges(id),
    FOREIGN KEY(member_id) REFERENCES members(id)
);

-- Cosmetics
CREATE TABLE cosmetics (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    type ENUM('hat', 'shirt', 'accessory', 'background', 'emote'),
    xp_cost INT,
    image_url VARCHAR(255),
    created_at TIMESTAMP
);

CREATE TABLE member_cosmetics (
    member_id UUID NOT NULL,
    cosmetic_id UUID NOT NULL,
    purchased_at TIMESTAMP,
    PRIMARY KEY(member_id, cosmetic_id),
    FOREIGN KEY(member_id) REFERENCES members(id),
    FOREIGN KEY(cosmetic_id) REFERENCES cosmetics(id)
);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    recipient_id UUID NOT NULL,
    type ENUM('badge', 'level_up', 'ranking', 'event', 'social', 'challenge'),
    title VARCHAR(255),
    message TEXT,
    action_url VARCHAR(255) NULLABLE,
    read_at TIMESTAMP NULLABLE,
    created_at TIMESTAMP,
    FOREIGN KEY(recipient_id) REFERENCES users(id)
);

-- Friendships
CREATE TABLE friendships (
    user_id UUID NOT NULL,
    friend_id UUID NOT NULL,
    status ENUM('pending', 'accepted', 'blocked'),
    created_at TIMESTAMP,
    PRIMARY KEY(user_id, friend_id),
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(friend_id) REFERENCES users(id)
);

-- Teams
CREATE TABLE teams (
    id UUID PRIMARY KEY,
    chapter_id UUID NOT NULL,
    name VARCHAR(255),
    leader_id UUID NOT NULL,
    description TEXT,
    created_at TIMESTAMP,
    FOREIGN KEY(chapter_id) REFERENCES chapters(id),
    FOREIGN KEY(leader_id) REFERENCES users(id)
);

CREATE TABLE team_members (
    team_id UUID NOT NULL,
    member_id UUID NOT NULL,
    role ENUM('leader', 'member'),
    joined_at TIMESTAMP,
    PRIMARY KEY(team_id, member_id),
    FOREIGN KEY(team_id) REFERENCES teams(id),
    FOREIGN KEY(member_id) REFERENCES members(id)
);

-- Quests
CREATE TABLE quests (
    id UUID PRIMARY KEY,
    chapter_id UUID NOT NULL,
    type ENUM('daily', 'weekly', 'seasonal'),
    title VARCHAR(255),
    description TEXT,
    xp_reward INT,
    reset_interval ENUM('daily', 'weekly', 'monthly'),
    created_at TIMESTAMP,
    FOREIGN KEY(chapter_id) REFERENCES chapters(id)
);

CREATE TABLE member_quests (
    member_id UUID NOT NULL,
    quest_id UUID NOT NULL,
    progress INT DEFAULT 0,
    completed_at TIMESTAMP NULLABLE,
    expires_at TIMESTAMP,
    PRIMARY KEY(member_id, quest_id),
    FOREIGN KEY(member_id) REFERENCES members(id),
    FOREIGN KEY(quest_id) REFERENCES quests(id)
);

-- Seasonal Data
CREATE TABLE seasons (
    id UUID PRIMARY KEY,
    chapter_id UUID NOT NULL,
    name VARCHAR(255),
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP,
    FOREIGN KEY(chapter_id) REFERENCES chapters(id)
);

-- Achievements (Extended)
CREATE TABLE achievements (
    id UUID PRIMARY KEY,
    chapter_id UUID NOT NULL,
    name VARCHAR(255),
    description TEXT,
    category ENUM('combat', 'service', 'social', 'knowledge', 'officer'),
    xp_reward INT,
    requirement_type ENUM('streak', 'total_xp', 'count', 'completion'),
    requirement_value INT,
    icon_url VARCHAR(255),
    created_at TIMESTAMP,
    FOREIGN KEY(chapter_id) REFERENCES chapters(id)
);

CREATE TABLE member_achievements (
    member_id UUID NOT NULL,
    achievement_id UUID NOT NULL,
    progress INT DEFAULT 0,
    unlocked_at TIMESTAMP NULLABLE,
    PRIMARY KEY(member_id, achievement_id),
    FOREIGN KEY(member_id) REFERENCES members(id),
    FOREIGN KEY(achievement_id) REFERENCES achievements(id)
);
```

---

## Migration Strategy

```sql
-- Run these migrations in order

-- Migration: 022_add_streaks.up.sql
ALTER TABLE members ADD COLUMN current_streak INT DEFAULT 0;
ALTER TABLE members ADD COLUMN longest_streak INT DEFAULT 0;
ALTER TABLE members ADD COLUMN streak_updated_at TIMESTAMP;

-- Migration: 023_add_challenges.up.sql
-- (Use table creation SQL from above)

-- Migration: 024_add_cosmetics.up.sql
-- (Use table creation SQL from above)

-- Continue for each feature...
```

---

## API Endpoints for New Features

```
# Streaks
GET    /v1/members/:id/streak
GET    /v1/members/:id/streaks/leaderboard

# Challenges
GET    /v1/challenges
POST   /v1/challenges                           (admin only)
GET    /v1/challenges/:id
POST   /v1/challenges/:id/join
PUT    /v1/challenges/:id/progress

# Cosmetics
GET    /v1/cosmetics
POST   /v1/cosmetics/:id/purchase
GET    /v1/members/:id/cosmetics

# Notifications
GET    /v1/notifications
PUT    /v1/notifications/:id/read
PUT    /v1/notifications/read-all

# Friends
GET    /v1/users/:id/friends
POST   /v1/users/:id/friends/:friend_id
DELETE /v1/users/:id/friends/:friend_id

# Teams
GET    /v1/teams
POST   /v1/teams                                (admin only)
POST   /v1/teams/:id/join
GET    /v1/teams/:id/leaderboard

# Quests
GET    /v1/quests/daily
GET    /v1/quests/weekly
PUT    /v1/quests/:id/progress

# Achievements
GET    /v1/achievements
GET    /v1/members/:id/achievements
```

---

## Success Metrics

Track these KPIs to measure gamification impact:

| Metric | Current | Goal | Measurement |
|---|---|---|---|
| Daily Active Users | TBD | +40% | Events dashboard |
| Session Duration | TBD | +2x | Analytics |
| Feature Adoption | TBD | >80% | Usage tracking |
| XP Earned (Avg) | TBD | +50% | Database query |
| Engagement Streaks | TBD | >70% | Streaks DB |
| Challenge Participation | 0% | >60% | Challenge signup |
| Cosmetic Purchases | 0% | >40% | Cosmetic DB |
| Leaderboard Views | TBD | +3x | Analytics |

---

## Timeline & Effort Estimate

**Total Effort:** 12-16 weeks (1 engineer full-time or 2 engineers part-time)

```
Week 1-2:   Quick Wins (streaks, pop-ups, sounds)
Week 3-4:   Challenges + Notifications
Week 5-6:   Cosmetics + Seasonal Leaderboards
Week 7-9:   Social features + Teams
Week 10-12: Achievements + Quests
Week 13-16: Events + Tournaments + Polish
```

---

## Getting Started

1. **Start with Tier 1** - Quick wins build momentum
2. **Gather feedback** - Survey users on what they want
3. **Iterate** - Not all features work for every community
4. **Mobile-first** - Ensure responsive design
5. **Celebrate wins** - Share updates with chapter

---

## Support & Questions

For implementation questions:
- **GitHub Issues:** https://github.com/Ces1231/blue-ledger/issues
- **Architecture:** See [ARCHITECTURE.md](ARCHITECTURE.md)
- **API:** See [API_DOCUMENTATION.md](API_DOCUMENTATION.md)
- **Deployment:** See [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)
