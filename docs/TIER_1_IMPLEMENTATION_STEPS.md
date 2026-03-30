// TIER 1 INTEGRATION GUIDE - Implementation Steps
// ================================================
// Follow these steps in order to complete TIER 1 quick wins integration

/*
FILE: blue-ledger-api/internal/auth/handler.go
LOCATION: In the Handler struct - Add login bonus repository field
TASK: T1.1 - Integrate daily login bonus in auth handler
TIME: 30 minutes

STEP 1: Add field to Handler struct
----- Find this struct around line 10-14:
type Handler struct {
	svc Service
}

ADD THIS FIELD:
type Handler struct {
	svc Service
	loginBonusRepo *LoginBonusRepository  // ADD THIS LINE
}

----- Find the NewHandler function around line 16:
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

MODIFY TO:
func NewHandler(svc Service, db *sql.DB) *Handler {
	return &Handler{
		svc: svc,
		loginBonusRepo: NewLoginBonusRepository(db),  // ADD THIS LINE
	}
}

STEP 2: Modify Login method
Find the Login method around line 97-120
After the line: resp, err := h.svc.Login(...)
Add this code:

	if err == nil && resp != nil {
		// Award daily login bonus
		if bonus, bonusErr := h.loginBonusRepo.AwardDailyLoginBonus(c.Request().Context(), resp.UserID); bonusErr == nil && bonus.Awarded {
			// TODO: Send notification and award XP
			// This will be done in T1.6-T1.7
		}
	}

STEP 3: Update handler registration
Find where this handler is created in main.go and pass db parameter
*/

/*
FILE: blue-ledger-api/internal/events/service.go
LOCATION: CheckInQR method or similar event attendance recording
TASK: T1.2 - Integrate streak updates in attendance handler
TIME: 1 hour

STEP 1: Import streaks package
ADD TO IMPORTS:
import (
	...
	"blue-ledger/internal/streaks"
	...
)

STEP 2: Add streaks service to Handler
In events/handler.go, modify Handler struct:
type Handler struct {
	svc Service
	qrSecret string
	streaksService *streaks.Service  // ADD THIS LINE
}

STEP 3: Update NewHandler
func NewHandler(svc Service, qrSecret string, db *sql.DB) *Handler {
	return &Handler{
		svc: svc,
		qrSecret: qrSecret,
		streaksService: streaks.NewService(db),  // ADD THIS LINE
	}
}

STEP 4: In CheckInQR method, after successful check-in
Find the CheckInQR method around line 210-240
After the line: result, err := h.svc.CheckInQR(...)

ADD THIS CODE:
if err == nil {
	// Update member's attendance streak
	memberID := result.MemberID // Assuming result has MemberID
	if streakInfo, streakErr := h.streaksService.UpdateStreakForAttendance(c.Request().Context(), memberID); streakErr == nil {
		// TODO: Calculate streak bonus XP and notify
		// This will be done in T1.3
	}
}
*/

/*
FILE: blue-ledger-api/internal/xp/service.go
LOCATION: AwardXP method
TASK: T1.3 - Add streak bonuses to XP calculation
TIME: 30 minutes

STEP 1: Import streaks package
ADD TO IMPORTS:
import (
	...
	"blue-ledger/internal/streaks"
	...
)

STEP 2: Add streaksService to xpService struct
Find xpService struct and ADD:
type xpService struct {
	db *sql.DB
	streaksService *streaks.Service  // ADD THIS LINE
}

STEP 3: Update NewXPService
Find NewXPService function and ADD streaks initialization:
func NewXPService(db *sql.DB) Service {
	return &xpService{
		db: db,
		streaksService: streaks.NewService(db),  // ADD THIS LINE
	}
}

STEP 4: Modify AwardXP method
Find AwardXP method around line 117
AFTER the line where XP is awarded to database:

ADD THIS CODE:
// Check for streak bonus
if streakInfo, err := x.streaksService.GetStreakInfo(ctx, input.MemberID); err == nil {
	streakBonus := x.streaksService.CalculateStreakBonus(streakInfo.CurrentStreak)
	if streakBonus > 0 {
		input.Amount += streakBonus
		// Update XP again with bonus
		_, _ = x.db.ExecContext(ctx, `
			UPDATE engagement_log 
			SET xp_amount = xp_amount + $1 
			WHERE member_id = $2 ORDER BY created_at DESC LIMIT 1
		`, streakBonus, input.MemberID)
	}
}
*/

/*
FILE: blue-ledger-api/cmd/server/main.go
LOCATION: Routes registration
TASK: T1.4 - Register streak endpoints in router
TIME: 15 minutes

STEP 1: Import streaks package
ADD TO IMPORTS:
import (
	...
	"blue-ledger/internal/streaks"
	...
)

STEP 2: Create streaks handler and register routes
Find where routes are registered (around line where auth routes are registered)

ADD THIS CODE:
// Streaks routes
streaksHandler := streaks.NewHandler(streaksService)
v1Routes.GET("/members/:id/streak", streaksHandler.GetStreakInfo)
v1Routes.GET("/chapters/:chapterId/streaks/leaderboard", streaksHandler.GetStreakLeaderboard)
*/

/*
FILE: web/src/pages/Dashboard.tsx (or similar)
LOCATION: Dashboard page component
TASK: T1.5 - Add StreakWidget to dashboard
TIME: 1 hour

STEP 1: Import StreakWidget
ADD TO IMPORTS:
import { StreakWidget } from '@/components/StreakWidget';
import { useAuth } from '@/hooks/useAuth'; // If not already imported

STEP 2: Use in Dashboard render
Find the dashboard JSX and ADD:

<div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
  <StreakWidget 
    memberId={user?.member?.id || ''} 
    className="md:col-span-1"
    showAnimation={true}
  />
  
  {/* Other dashboard widgets... */}
</div>

STEP 3: Make sure memberId is available
Verify that user.member.id is accessible in the component
If using a different context, adjust accordingly
*/

/*
FILE: web/src/App.tsx or web/src/pages/_layout.tsx
LOCATION: Root layout component
TASK: T1.6 - Integrate notification hook in app
TIME: 1 hour

STEP 1: Import hook
ADD TO IMPORTS:
import { useAchievementNotifications } from '@/hooks/useAchievementNotifications';
import { AchievementNotification } from '@/components/AchievementNotification';

STEP 2: Add to layout/app component
IN YOUR COMPONENT:

const { notifications, showXPGained, showLevelUp, showStreakMilestone } = useAchievementNotifications();

STEP 3: Create context provider (optional but recommended)
Create file: web/src/context/NotificationContext.tsx

import React, { createContext, useContext } from 'react';
import { useAchievementNotifications } from '@/hooks/useAchievementNotifications';

const NotificationContext = createContext(null);

export const NotificationProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const notificationSystem = useAchievementNotifications();
  
  return (
    <NotificationContext.Provider value={notificationSystem}>
      {children}
    </NotificationContext.Provider>
  );
};

export const useNotifications = () => {
  const ctx = useContext(NotificationContext);
  if (!ctx) throw new Error('useNotifications must be used within NotificationProvider');
  return ctx;
};

STEP 4: Wrap app with provider
In your App.tsx or main layout:

<NotificationProvider>
  <YourAppComponents />
  {/* Render notifications */}
  <div className="fixed top-0 left-0 right-0 z-50">
    {notifications.map((notif) => (
      <AchievementNotification
        key={notif.id}
        {...notif}
        onDismiss={() => removeNotification(notif.id)}
      />
    ))}
  </div>
</NotificationProvider>

STEP 5: Use notifications throughout app
Any component can now import and use:
import { useNotifications } from '@/context/NotificationContext';

const MyComponent = () => {
  const { showXPGained, showLevelUp } = useNotifications();
  
  // Use these methods anywhere in your app
  showXPGained(100, 'Event attended');
};
*/

/*
FILE: web/src/pages/XPReward.tsx (or component that shows XP)
LOCATION: XP notification trigger
TASK: T1.7 - Add sounds to XP gains
TIME: 30 minutes

STEP 1: Import sound system
ADD TO IMPORTS:
import { soundEffects } from '@/utils/soundEffects';
import { useNotifications } from '@/context/NotificationContext';

STEP 2: In XP reward display
WHEN XP IS EARNED, ADD:

const handleXPGained = (xpAmount: number, reason: string) => {
  // Play sound
  soundEffects.playSuccess();
  
  // Show notification
  const { showXPGained } = useNotifications();
  showXPGained(xpAmount, reason);
};

STEP 3: For level up
WHEN LEVEL UP OCCURS, ADD:

const handleLevelUp = (newLevel: number, tierName: string) => {
  soundEffects.playLevelUp();
  const { showLevelUp } = useNotifications();
  showLevelUp(newLevel, tierName, xpGained);
};

STEP 4: For streak milestones
WHEN STREAK MILESTONE REACHED, ADD:

const handleStreakMilestone = (days: number) => {
  soundEffects.playStreak();
  const { showStreakMilestone } = useNotifications();
  showStreakMilestone(days, streakBonus);
};

STEP 5: Add volume control to settings
In Settings page, ADD:

import { soundEffects } from '@/utils/soundEffects';

<div className="space-y-4">
  <label>
    Sound Effects Volume
    <input
      type="range"
      min="0"
      max="1"
      step="0.1"
      value={soundEffects.getVolume()}
      onChange={(e) => soundEffects.setVolume(parseFloat(e.target.value))}
    />
  </label>
  
  <button 
    onClick={() => soundEffects.isMuted() ? soundEffects.unmute() : soundEffects.mute()}
  >
    {soundEffects.isMuted() ? 'Unmute' : 'Mute'} Sound Effects
  </button>
</div>
*/

/*
FILE: (Testing)
TASK: T1.8 - End-to-end testing
TIME: 2 hours

TESTING CHECKLIST:

Backend Tests:
[ ] Run migrations: go run cmd/migrate/main.go up
[ ] Verify tables exist: 
    - SELECT * FROM members LIMIT 1 (check for current_streak, longest_streak)
    - SELECT * FROM users LIMIT 1 (check for last_login_at, daily_login_streak)
    - SELECT * FROM notifications LIMIT 1
[ ] Unit test streaks: go test ./internal/streaks/...
[ ] Unit test login bonus: go test ./internal/auth/... -run TestLoginBonus
[ ] Start server: make dev

Frontend Tests:
[ ] Start dev server: cd web && npm run dev
[ ] Test notification component appears when prop changes
[ ] Test sound plays on XP gain
[ ] Test StreakWidget displays correctly
[ ] Test volume control works
[ ] Test mute button works

Integration Tests:
[ ] Login as user
  ✓ Verify daily login bonus is awarded (25 XP)
  ✓ Check notification appears
  ✓ Check sound plays
[ ] Second login same day
  ✓ Verify NO additional bonus
[ ] Attend event
  ✓ Verify streak updates
  ✓ Verify streak bonus XP awarded
  ✓ Check StreakWidget updates
[ ] Reach 7-day streak
  ✓ Verify milestone notification
  ✓ Verify streak bonus sound
  ✓ Check StreakWidget highlights milestone
[ ] Test leaderboard endpoint: GET /v1/chapters/{id}/streaks/leaderboard

Manual Testing:
[ ] Check database: SELECT * FROM engagement_log ORDER BY created_at DESC;
[ ] Verify XP amounts include streakbonus
[ ] Verify notifications table is populated
[ ] Check localStorage for soundEffectsConfig

Regression Tests:
[ ] Existing login still works
[ ] Existing event attendance still works
[ ] XP awarding still works
[ ] No errors in server logs
[ ] No errors in browser console
*/

// END OF TIER 1 IMPLEMENTATION GUIDE
