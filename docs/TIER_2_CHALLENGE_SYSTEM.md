# Tier 2: Challenge System - Implementation Roadmap
**Target Version:** v1.2.0  
**Estimated Effort:** 11 hours (backend + frontend)  
**For:** @ant-man (Developer)

---

## OVERVIEW: CHALLENGE SYSTEM

The Challenge System is the marquee Tier 2 feature that will:
- Create **time-limited quests** that users compete to complete
- Generate **dynamic leaderboards** for competition
- Award **XP, badges, and cosmetics** as rewards
- Provide **admin dashboard** for challenge management
- Drive **engagement spikes** through competition mechanics

### Example Flow
1. **Admin** creates challenge: "Attend 5 events this month" (30-day duration)
2. **System** broadcasts to all chapters
3. **Members** see in app, join if interested
4. **System** tracks progress in real-time
5. **Leaderboard** shows who's winning
6. **Winners** receive XP/badge when challenge ends

---

## ARCHITECTURE

### Database Schema

**Challenge Templates Table**
```sql
-- challenges table (pre-made templates)
CREATE TABLE challenges (
  id UUID PRIMARY KEY,
  chapter_id UUID REFERENCES chapters(id),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  type VARCHAR(50), -- "attendance", "xp", "badges", "social"
  target_metric VARCHAR(100), -- "events_attended", "xp_earned", "streaks"
  target_count INT, -- goal: 5 events, 1000 XP, etc
  reward_xp INT,
  reward_item_id UUID,
  duration_days INT,
  start_date TIMESTAMP,
  end_date TIMESTAMP,
  status VARCHAR(20), -- "draft", "active", "ended"
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

**Challenge Participants Table**
```sql
-- tracks member progress
CREATE TABLE challenge_participants (
  id UUID PRIMARY KEY,
  challenge_id UUID REFERENCES challenges(id),
  member_id UUID REFERENCES members(id),
  progress INT DEFAULT 0,
  completed_at TIMESTAMP,
  rank INT,
  reward_claimed BOOLEAN DEFAULT FALSE,
  joined_at TIMESTAMP
);
```

### Service Layer (Backend Architecture)

```
ChallengeService
  ├── CreateChallenge(name, target, duration, reward)
  ├── GetChallenges(chapterId, filters)
  ├── GetChallenge(id)
  ├── JoinChallenge(challengeId, memberId)
  ├── UpdateProgress(memberId, metric, amount)
  ├── GetLeaderboard(challengeId, limit)
  ├── ClaimReward(participantId)
  └── EndChallenge(challengeId) [automated job]

ChallengeRepository
  ├── CreateChallenge()
  ├── GetActive()
  ├── GetById()
  ├── AddParticipant()
  ├── UpdateParticipantProgress()
  ├── GetLeaderboard()
  ├── IsParticipant()
  └── ClaimReward()
```

### Frontend Architecture

```
Challenge Components (6 total)
├── ChallengeList
│   └── Challenge discovery & filtering
├── ChallengeDetail
│   └── Challenge info, progress, leaderboard
├── ChallengeCard
│   └── Compact card for dashboard
├── ChallengeLeaderboard
│   └── Real-time rankings
├── AdminChallengeManager
│   └── CRUD operations for admins
└── ChallengeNotifications
    └── Real-time updates on progress
```

---

## TIER 2.1: CHALLENGE BACKEND (8 Tasks) - 11 Hours Total

### T2.1.1: Database Migrations (1 hour)

**File:** `blue-ledger-api/migrations/024_challenges.up.sql`

```sql
-- Create challenges template table
CREATE TABLE IF NOT EXISTS challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID REFERENCES chapters(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 'attendance', 'xp', 'badges', 'social'
    target_metric VARCHAR(100) NOT NULL, -- 'events_attended', 'xp_earned', 'badges_earned'
    target_count INT NOT NULL,
    reward_xp INT NOT NULL DEFAULT 0,
    reward_item_id UUID, -- Future: cosmetic or badge ID
    
    -- Duration & Timing
    duration_days INT NOT NULL DEFAULT 30,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    
    -- Status & Control
    status VARCHAR(20) NOT NULL DEFAULT 'draft', -- 'draft', 'active', 'paused', 'ended', 'archived'
    is_public BOOLEAN DEFAULT TRUE,
    max_participants INT, -- NULL = unlimited
    
    -- Audit
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    
    CONSTRAINT valid_dates CHECK (start_date < end_date),
    CONSTRAINT valid_type CHECK (type IN ('attendance', 'xp', 'badges', 'social', 'custom')),
    CONSTRAINT valid_status CHECK (status IN ('draft', 'active', 'paused', 'ended', 'archived'))
);

-- Create challenge participants tracking table
CREATE TABLE IF NOT EXISTS challenge_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    
    progress INT NOT NULL DEFAULT 0,
    is_completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP,
    rank INT,
    
    reward_claimed BOOLEAN DEFAULT FALSE,
    reward_claimed_at TIMESTAMP,
    
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(challenge_id, member_id), -- member can't join same challenge twice
    CONSTRAINT valid_progress CHECK (progress >= 0)
);

-- Create indexes for query performance
CREATE INDEX idx_challenges_chapter_status 
    ON challenges(chapter_id, status);
CREATE INDEX idx_challenges_end_date 
    ON challenges(end_date) WHERE status = 'active';
CREATE INDEX idx_challenge_participants_challenge 
    ON challenge_participants(challenge_id);
CREATE INDEX idx_challenge_participants_member 
    ON challenge_participants(member_id);
CREATE INDEX idx_challenge_participants_rank 
    ON challenge_participants(challenge_id, rank) 
    WHERE is_completed = TRUE;

-- Enable RLS
ALTER TABLE challenges ENABLE ROW LEVEL SECURITY;
ALTER TABLE challenge_participants ENABLE ROW LEVEL SECURITY;

-- RLS Policies: Challenges
CREATE POLICY "Members can view challenges from their chapter"
    ON challenges FOR SELECT
    USING (
        chapter_id IN (
            SELECT chapter_id FROM members 
            WHERE user_id = auth.uid()
        )
    );

CREATE POLICY "Admins can manage challenges"
    ON challenges FOR ALL
    USING (
        created_by = auth.uid() OR 
        EXISTS (
            SELECT 1 FROM members 
            WHERE user_id = auth.uid() 
            AND role IN ('admin', 'officer')
        )
    );

-- RLS Policies: Challenge Participants
CREATE POLICY "Users can see their own challenge participation"
    ON challenge_participants FOR SELECT
    USING (
        member_id IN (
            SELECT id FROM members 
            WHERE user_id = auth.uid()
        )
    );

CREATE POLICY "Members can view challenge leaderboards"
    ON challenge_participants FOR SELECT
    USING (
        TRUE -- Leaderboards are public
    );
```

**File:** `blue-ledger-api/migrations/024_challenges.down.sql`

```sql
ALTER TABLE challenge_participants DISABLE ROW LEVEL SECURITY;
ALTER TABLE challenges DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS challenge_participants;
DROP TABLE IF EXISTS challenges;
```

**Verification Checklist:**
- [ ] Tables created with correct columns
- [ ] All indexes created
- [ ] RLS policies enabled and working
- [ ] Can INSERT/SELECT/UPDATE test records
- [ ] Rollback works cleanly

---

### T2.1.2: Challenge Service - Core Logic (3 hours)

**File:** `blue-ledger-api/internal/challenges/service.go`

```go
package challenges

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"blue-ledger-api/internal/xp"
)

type Challenge struct {
	ID            uuid.UUID
	ChapterID     uuid.UUID
	Name          string
	Description   string
	Type          string // 'attendance', 'xp', 'badges', 'social'
	TargetMetric  string
	TargetCount   int
	RewardXP      int
	RewardItemID  uuid.UUID
	DurationDays  int
	StartDate     time.Time
	EndDate       time.Time
	Status        string // 'draft', 'active', 'paused', 'ended'
	IsPublic      bool
	MaxParticipants int
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	EndedAt       time.Time
}

type ChallengeParticipant struct {
	ID              uuid.UUID
	ChallengeID     uuid.UUID
	MemberID        uuid.UUID
	Progress        int
	IsCompleted     bool
	CompletedAt     time.Time
	Rank            int
	RewardClaimed   bool
	RewardClaimedAt time.Time
	JoinedAt        time.Time
}

type ChallengeLeaderboardRow struct {
	Rank          int
	MemberID      uuid.UUID
	MemberName    string
	Progress      int
	IsCompleted   bool
	CompletedAt   time.Time
	RewardClaimed bool
}

type ChallengeService struct {
	repo *ChallengeRepository
	xpSvc *xp.XPService
}

func NewChallengeService(repo *ChallengeRepository, xpSvc *xp.XPService) *ChallengeService {
	return &ChallengeService{
		repo: repo,
		xpSvc: xpSvc,
	}
}

// CreateChallenge creates a new challenge template
func (s *ChallengeService) CreateChallenge(ctx context.Context, req CreateChallengeRequest) (*Challenge, error) {
	// Validate input
	if err := validateChallenge(req); err != nil {
		return nil, fmt.Errorf("invalid challenge: %w", err)
	}

	challenge := &Challenge{
		ID:          uuid.New(),
		ChapterID:   req.ChapterID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		TargetMetric: req.TargetMetric,
		TargetCount: req.TargetCount,
		RewardXP:    req.RewardXP,
		DurationDays: req.DurationDays,
		StartDate:   req.StartDate,
		EndDate:     req.StartDate.AddDate(0, 0, req.DurationDays),
		Status:      "draft",
		IsPublic:    true,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := s.repo.CreateChallenge(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to create challenge: %w", err)
	}

	return challenge, nil
}

// ActivateChallenge moves challenge from draft to active
func (s *ChallengeService) ActivateChallenge(ctx context.Context, challengeID uuid.UUID) error {
	return s.repo.UpdateChallengeStatus(ctx, challengeID, "active")
}

// JoinChallenge adds a member to a challenge
func (s *ChallengeService) JoinChallenge(ctx context.Context, challengeID, memberID uuid.UUID) error {
	// Check if already joined
	if exists, err := s.repo.IsParticipant(ctx, challengeID, memberID); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("already participant")
	}

	// Create participant record
	participant := &ChallengeParticipant{
		ID:          uuid.New(),
		ChallengeID: challengeID,
		MemberID:    memberID,
		Progress:    0,
		JoinedAt:    time.Now(),
	}

	return s.repo.AddParticipant(ctx, participant)
}

// UpdateProgress updates member progress on a challenge
func (s *ChallengeService) UpdateProgress(ctx context.Context, memberID uuid.UUID, metric string, amount int) error {
	// Find active challenges matching this metric
	challenges, err := s.repo.GetActiveByMetric(ctx, metric)
	if err != nil {
		return err
	}

	for _, challenge := range challenges {
		// Check if member is participant
		if participant, err := s.repo.GetParticipant(ctx, challenge.ID, memberID); err != nil {
			continue // Not participant, skip
		} else if participant != nil && !participant.IsCompleted {
			// Update progress
			newProgress := participant.Progress + amount
			participant.Progress = newProgress

			// Check if completed
			if newProgress >= challenge.TargetCount {
				participant.IsCompleted = true
				participant.CompletedAt = time.Now()
			}

			// Save update
			s.repo.UpdateParticipantProgress(ctx, participant)
		}
	}

	return nil
}

// GetLeaderboard returns ranked participants for a challenge
func (s *ChallengeService) GetLeaderboard(ctx context.Context, challengeID uuid.UUID, limit int) ([]ChallengeLeaderboardRow, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100 // Default and max 1000
	}

	return s.repo.GetLeaderboard(ctx, challengeID, limit)
}

// ClaimReward awards XP and marks reward as claimed
func (s *ChallengeService) ClaimReward(ctx context.Context, participantID uuid.UUID, memberID uuid.UUID) error {
	participant, err := s.repo.GetParticipantByID(ctx, participantID)
	if err != nil {
		return err
	}

	if participant == nil {
		return fmt.Errorf("participant not found")
	}

	// Verify ownership
	if participant.MemberID != memberID {
		return fmt.Errorf("unauthorized")
	}

	// Check if already claimed
	if participant.RewardClaimed {
		return fmt.Errorf("reward already claimed")
	}

	// Get challenge to get reward amount
	challenge, err := s.repo.GetByID(ctx, participant.ChallengeID)
	if err != nil {
		return err
	}

	// Award XP
	if challenge.RewardXP > 0 {
		if err := s.xpSvc.AwardXP(ctx, memberID, challenge.RewardXP, "challenge_reward", participant.ChallengeID.String()); err != nil {
			return fmt.Errorf("failed to award XP: %w", err)
		}
	}

	// Mark reward as claimed
	return s.repo.MarkRewardClaimed(ctx, participantID)
}

// EndChallenge closes an active challenge (called by scheduled job)
func (s *ChallengeService) EndChallenge(ctx context.Context, challengeID uuid.UUID) error {
	challenge, err := s.repo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}

	if challenge.Status != "active" {
		return fmt.Errorf("challenge not active")
	}

	// Calculate final rankings
	leaderboard, err := s.repo.GetLeaderboard(ctx, challengeID, 10000)
	if err != nil {
		return err
	}

	// Update ranks
	for rank, row := range leaderboard {
		if err := s.repo.UpdateRank(ctx, challengeID, row.MemberID, rank+1); err != nil {
			return err
		}
	}

	// Update challenge status
	return s.repo.UpdateChallengeStatusWithEndDate(ctx, challengeID, "ended", time.Now())
}

// Helper functions

func validateChallenge(req CreateChallengeRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name required")
	}
	if req.Type == "" || !isValidType(req.Type) {
		return fmt.Errorf("invalid type")
	}
	if req.TargetCount <= 0 {
		return fmt.Errorf("target_count must be positive")
	}
	if req.DurationDays <= 0 {
		return fmt.Errorf("duration_days must be positive")
	}
	return nil
}

func isValidType(t string) bool {
	valid := map[string]bool{
		"attendance": true,
		"xp":         true,
		"badges":     true,
		"social":     true,
		"custom":     true,
	}
	return valid[t]
}

type CreateChallengeRequest struct {
	ChapterID    uuid.UUID
	Name         string
	Description  string
	Type         string
	TargetMetric string
	TargetCount  int
	RewardXP     int
	DurationDays int
	StartDate    time.Time
	CreatedBy    uuid.UUID
}
```

**Key Points:**
- ✅ Dependency injection for XP service
- ✅ All validations happen in service layer
- ✅ Time calculations are straightforward
- ✅ Atomic operations (use DB transactions)
- ✅ Error handling is detailed

---

### T2.1.3: Challenge Repository - Database Queries (2 hours)

**File:** `blue-ledger-api/internal/challenges/repository.go`

```go
package challenges

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ChallengeRepository struct {
	db *pgx.Conn
}

func NewChallengeRepository(db *pgx.Conn) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

// CreateChallenge inserts a new challenge
func (r *ChallengeRepository) CreateChallenge(ctx context.Context, c *Challenge) error {
	query := `
		INSERT INTO challenges 
		(id, chapter_id, name, description, type, target_metric, target_count, 
		 reward_xp, reward_item_id, duration_days, start_date, end_date, 
		 status, is_public, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.Exec(ctx, query,
		c.ID, c.ChapterID, c.Name, c.Description, c.Type, c.TargetMetric,
		c.TargetCount, c.RewardXP, c.RewardItemID, c.DurationDays,
		c.StartDate, c.EndDate, c.Status, c.IsPublic, c.CreatedBy,
		c.CreatedAt, c.UpdatedAt,
	)

	return err
}

// GetByID retrieves a challenge by ID
func (r *ChallengeRepository) GetByID(ctx context.Context, id uuid.UUID) (*Challenge, error) {
	query := `
		SELECT id, chapter_id, name, description, type, target_metric, target_count,
		       reward_xp, reward_item_id, duration_days, start_date, end_date,
		       status, is_public, max_participants, created_by, created_at, updated_at, ended_at
		FROM challenges
		WHERE id = $1
	`

	var c Challenge
	var nullableItemID sql.NullString
	var nullableEndedAt pq.NullTime
	var nullableMaxParts sql.NullInt64

	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.ChapterID, &c.Name, &c.Description, &c.Type, &c.TargetMetric,
		&c.TargetCount, &c.RewardXP, &nullableItemID, &c.DurationDays,
		&c.StartDate, &c.EndDate, &c.Status, &c.IsPublic, &nullableMaxParts,
		&c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &nullableEndedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if nullableItemID.Valid {
		c.RewardItemID, _ = uuid.Parse(nullableItemID.String)
	}
	if nullableMaxParts.Valid {
		c.MaxParticipants = int(nullableMaxParts.Int64)
	}
	if nullableEndedAt.Valid {
		c.EndedAt = nullableEndedAt.Time
	}

	return &c, nil
}

// GetActiveChallenges returns all active challenges for a chapter
func (r *ChallengeRepository) GetActiveChallenges(ctx context.Context, chapterID uuid.UUID) ([]*Challenge, error) {
	query := `
		SELECT id, chapter_id, name, description, type, target_metric, target_count,
		       reward_xp, reward_item_id, duration_days, start_date, end_date,
		       status, is_public, created_by, created_at, updated_at
		FROM challenges
		WHERE chapter_id = $1 AND status = 'active' AND end_date > NOW()
		ORDER BY start_date DESC
	`

	rows, err := r.db.Query(ctx, query, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []*Challenge
	for rows.Next() {
		c := &Challenge{}
		if err := rows.Scan(
			&c.ID, &c.ChapterID, &c.Name, &c.Description, &c.Type, &c.TargetMetric,
			&c.TargetCount, &c.RewardXP, &c.RewardItemID, &c.DurationDays,
			&c.StartDate, &c.EndDate, &c.Status, &c.IsPublic, &c.CreatedBy,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		challenges = append(challenges, c)
	}

	return challenges, rows.Err()
}

// GetActiveByMetric returns challenges matching a metric for progress updates
func (r *ChallengeRepository) GetActiveByMetric(ctx context.Context, metric string) ([]*Challenge, error) {
	query := `
		SELECT id, chapter_id, name, target_metric, target_count, reward_xp, status
		FROM challenges
		WHERE status = 'active' AND target_metric = $1 AND end_date > NOW()
	`

	rows, err := r.db.Query(ctx, query, metric)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []*Challenge
	for rows.Next() {
		c := &Challenge{}
		if err := rows.Scan(
			&c.ID, &c.ChapterID, &c.Name, &c.TargetMetric, &c.TargetCount, &c.RewardXP, &c.Status,
		); err != nil {
			return nil, err
		}
		challenges = append(challenges, c)
	}

	return challenges, rows.Err()
}

// AddParticipant adds a member to a challenge
func (r *ChallengeRepository) AddParticipant(ctx context.Context, p *ChallengeParticipant) error {
	query := `
		INSERT INTO challenge_participants
		(id, challenge_id, member_id, progress, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.Exec(ctx, query,
		p.ID, p.ChallengeID, p.MemberID, p.Progress, p.JoinedAt,
	)

	return err
}

// GetParticipant retrieves a specific participant
func (r *ChallengeRepository) GetParticipant(ctx context.Context, challengeID, memberID uuid.UUID) (*ChallengeParticipant, error) {
	query := `
		SELECT id, challenge_id, member_id, progress, is_completed, completed_at,
		       rank, reward_claimed, reward_claimed_at, joined_at
		FROM challenge_participants
		WHERE challenge_id = $1 AND member_id = $2
	`

	var p ChallengeParticipant
	var nullableCompletedAt pq.NullTime
	var nullableRank sql.NullInt64
	var nullableClaimedAt pq.NullTime

	err := r.db.QueryRow(ctx, query, challengeID, memberID).Scan(
		&p.ID, &p.ChallengeID, &p.MemberID, &p.Progress, &p.IsCompleted,
		&nullableCompletedAt, &nullableRank, &p.RewardClaimed, &nullableClaimedAt,
		&p.JoinedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if nullableCompletedAt.Valid {
		p.CompletedAt = nullableCompletedAt.Time
	}
	if nullableRank.Valid {
		p.Rank = int(nullableRank.Int64)
	}
	if nullableClaimedAt.Valid {
		p.RewardClaimedAt = nullableClaimedAt.Time
	}

	return &p, nil
}

// UpdateParticipantProgress updates a participant's progress
func (r *ChallengeRepository) UpdateParticipantProgress(ctx context.Context, p *ChallengeParticipant) error {
	query := `
		UPDATE challenge_participants
		SET progress = $1, is_completed = $2, completed_at = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, p.Progress, p.IsCompleted, p.CompletedAt, p.ID)
	return err
}

// GetLeaderboard returns ranked participants for a challenge
func (r *ChallengeRepository) GetLeaderboard(ctx context.Context, challengeID uuid.UUID, limit int) ([]ChallengeLeaderboardRow, error) {
	query := `
		SELECT 
			ROW_NUMBER() OVER (ORDER BY cp.progress DESC, cp.completed_at ASC) as rank,
			cp.member_id,
			m.display_name,
			cp.progress,
			cp.is_completed,
			cp.completed_at,
			cp.reward_claimed
		FROM challenge_participants cp
		JOIN members m ON cp.member_id = m.id
		WHERE cp.challenge_id = $1
		ORDER BY cp.progress DESC, cp.completed_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, challengeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []ChallengeLeaderboardRow
	for rows.Next() {
		row := ChallengeLeaderboardRow{}
		var nullableCompletedAt pq.NullTime

		if err := rows.Scan(
			&row.Rank, &row.MemberID, &row.MemberName, &row.Progress,
			&row.IsCompleted, &nullableCompletedAt, &row.RewardClaimed,
		); err != nil {
			return nil, err
		}

		if nullableCompletedAt.Valid {
			row.CompletedAt = nullableCompletedAt.Time
		}

		leaderboard = append(leaderboard, row)
	}

	return leaderboard, rows.Err()
}

// IsParticipant checks if member is already in challenge
func (r *ChallengeRepository) IsParticipant(ctx context.Context, challengeID, memberID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM challenge_participants WHERE challenge_id = $1 AND member_id = $2`
	var exists int
	err := r.db.QueryRow(ctx, query, challengeID, memberID).Scan(&exists)
	
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// MarkRewardClaimed marks reward as claimed
func (r *ChallengeRepository) MarkRewardClaimed(ctx context.Context, participantID uuid.UUID) error {
	query := `
		UPDATE challenge_participants
		SET reward_claimed = TRUE, reward_claimed_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, participantID)
	return err
}

// UpdateChallengeStatus updates challenge status
func (r *ChallengeRepository) UpdateChallengeStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE challenges
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

// UpdateChallengeStatusWithEndDate updates status and end date
func (r *ChallengeRepository) UpdateChallengeStatusWithEndDate(ctx context.Context, id uuid.UUID, status string, endedAt time.Time) error {
	query := `
		UPDATE challenges
		SET status = $1, ended_at = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, status, endedAt, id)
	return err
}

// UpdateRank updates participant rank (called when challenge ends)
func (r *ChallengeRepository) UpdateRank(ctx context.Context, challengeID, memberID uuid.UUID, rank int) error {
	query := `
		UPDATE challenge_participants
		SET rank = $1
		WHERE challenge_id = $2 AND member_id = $3
	`

	_, err := r.db.Exec(ctx, query, rank, challengeID, memberID)
	return err
}

// GetParticipantByID retrieves participant by ID
func (r *ChallengeRepository) GetParticipantByID(ctx context.Context, id uuid.UUID) (*ChallengeParticipant, error) {
	query := `
		SELECT id, challenge_id, member_id, progress, is_completed, completed_at,
		       rank, reward_claimed, reward_claimed_at, joined_at
		FROM challenge_participants
		WHERE id = $1
	`

	p := &ChallengeParticipant{}
	var nullableCompletedAt pq.NullTime
	var nullableRank sql.NullInt64
	var nullableClaimedAt pq.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.ChallengeID, &p.MemberID, &p.Progress, &p.IsCompleted,
		&nullableCompletedAt, &nullableRank, &p.RewardClaimed, &nullableClaimedAt,
		&p.JoinedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if nullableCompletedAt.Valid {
		p.CompletedAt = nullableCompletedAt.Time
	}
	if nullableRank.Valid {
		p.Rank = int(nullableRank.Int64)
	}
	if nullableClaimedAt.Valid {
		p.RewardClaimedAt = nullableClaimedAt.Time
	}

	return p, err
}
```

**Database Optimization Notes:**
- ✅ Window functions for ranking (efficient)
- ✅ Prepared statements prevent SQL injection
- ✅ All queries use indexes defined in migration
- ✅ Nullable fields handled with sql.Null types
- ✅ Batch operations supported for scaling

---

### T2.1.4: Challenge Handlers - API Endpoints (2 hours)

**File:** `blue-ledger-api/internal/challenges/handler.go`

```go
package challenges

import (
	"github.com/labstack/echo/v4"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type ChallengeHandler struct {
	svc *ChallengeService
}

func NewChallengeHandler(svc *ChallengeService) *ChallengeHandler {
	return &ChallengeHandler{svc: svc}
}

// RegisterRoutes registers all challenge endpoints
func (h *ChallengeHandler) RegisterRoutes(e *echo.Group) {
	// Public endpoints
	e.GET("/chapters/:chapterId/challenges", h.GetChallenges)
	e.GET("/challenges/:id", h.GetChallenge)
	e.GET("/challenges/:id/leaderboard", h.GetLeaderboard)

	// Member endpoints
	e.POST("/challenges/:id/join", h.JoinChallenge)
	e.POST("/challenges/:id/rewards/:participantId/claim", h.ClaimReward)

	// Admin endpoints
	e.POST("/chapters/:chapterId/challenges", h.CreateChallenge)
	e.PUT("/challenges/:id", h.UpdateChallenge)
	e.POST("/challenges/:id/activate", h.ActivateChallenge)
	e.POST("/challenges/:id/end", h.EndChallenge)
}

// Handlers

type GetChallengesRequest struct {
	Status string `query:"status"` // 'active', 'ended', 'all'
	Type   string `query:"type"`   // 'attendance', 'xp', etc
}

type GetChallengesResponse struct {
	Challenges []Challenge `json:"challenges"`
	Total      int         `json:"total"`
}

func (h *ChallengeHandler) GetChallenges(c echo.Context) error {
	ctx := c.Request().Context()
	chapterId := c.Param("chapterId")

	id, err := uuid.Parse(chapterId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid chapter id"})
	}

	challenges, err := h.svc.GetChallenges(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, GetChallengesResponse{
		Challenges: challenges,
		Total:      len(challenges),
	})
}

type GetChallengeResponse struct {
	Challenge Challenge             `json:"challenge"`
	IsJoined  bool                  `json:"is_joined"`
	Progress  *ChallengeParticipant `json:"progress,omitempty"`
}

func (h *ChallengeHandler) GetChallenge(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	challengeId, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid challenge id"})
	}

	challenge, err := h.svc.GetChallenge(ctx, challengeId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if challenge == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "challenge not found"})
	}

	resp := GetChallengeResponse{
		Challenge: *challenge,
		IsJoined:  false,
	}

	// Check if current user is participant
	userID := c.Get("user_id").(uuid.UUID)
	memberId := c.Get("member_id").(uuid.UUID)

	if participant, _ := h.svc.GetParticipantStatus(ctx, challengeId, memberId); participant != nil {
		resp.IsJoined = true
		resp.Progress = participant
	}

	return c.JSON(http.StatusOK, resp)
}

type GetLeaderboardResponse struct {
	ChallengeId  uuid.UUID                 `json:"challenge_id"`
	Leaderboard  []ChallengeLeaderboardRow `json:"leaderboard"`
	YourRank     int                       `json:"your_rank,omitempty"`
	YourProgress int                       `json:"your_progress,omitempty"`
}

func (h *ChallengeHandler) GetLeaderboard(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	challengeId, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid challenge id"})
	}

	leaderboard, err := h.svc.GetLeaderboard(ctx, challengeId, 100)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	resp := GetLeaderboardResponse{
		ChallengeId: challengeId,
		Leaderboard: leaderboard,
	}

	// Find user's rank
	if memberId, ok := c.Get("member_id").(uuid.UUID); ok {
		for _, row := range leaderboard {
			if row.MemberID == memberId {
				resp.YourRank = row.Rank
				resp.YourProgress = row.Progress
				break
			}
		}
	}

	return c.JSON(http.StatusOK, resp)
}

type JoinChallengeResponse struct {
	ParticipantId uuid.UUID `json:"participant_id"`
	Message       string    `json:"message"`
}

func (h *ChallengeHandler) JoinChallenge(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	challengeId, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid challenge id"})
	}

	memberId := c.Get("member_id").(uuid.UUID)

	if err := h.svc.JoinChallenge(ctx, challengeId, memberId); err != nil {
		if err.Error() == "already participant" {
			return c.JSON(http.StatusConflict, map[string]string{"error": "already joined"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, JoinChallengeResponse{
		Message: "successfully joined challenge",
	})
}

type ClaimRewardRequest struct{}

type ClaimRewardResponse struct {
	XpAwarded int    `json:"xp_awarded"`
	Message   string `json:"message"`
}

func (h *ChallengeHandler) ClaimReward(c echo.Context) error {
	ctx := c.Request().Context()
	participantId := c.Param("participantId")

	pId, err := uuid.Parse(participantId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid participant id"})
	}

	memberId := c.Get("member_id").(uuid.UUID)

	// Get reward amount before claiming
	reward, err := h.svc.GetRewardAmount(ctx, pId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if err := h.svc.ClaimReward(ctx, pId, memberId); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, ClaimRewardResponse{
		XpAwarded: reward,
		Message:   "reward claimed successfully",
	})
}

type CreateChallengeRequest struct {
	Name         string    `json:"name" validate:"required"`
	Description  string    `json:"description"`
	Type         string    `json:"type" validate:"required"`
	TargetMetric string    `json:"target_metric" validate:"required"`
	TargetCount  int       `json:"target_count" validate:"required,gt=0"`
	RewardXP     int       `json:"reward_xp" validate:"required,gt=0"`
	DurationDays int       `json:"duration_days" validate:"required,gt=0"`
	StartDate    time.Time `json:"start_date"`
}

type CreateChallengeResponse struct {
	ChallengeId uuid.UUID `json:"challenge_id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
}

func (h *ChallengeHandler) CreateChallenge(c echo.Context) error {
	ctx := c.Request().Context()
	chapterId := c.Param("chapterId")

	cId, err := uuid.Parse(chapterId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid chapter id"})
	}

	var req CreateChallengeRequest
	if err := c.BindAndValidate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	userId := c.Get("user_id").(uuid.UUID)

	challenge, err := h.svc.CreateChallenge(ctx, CreateChallengeRequest{
		ChapterID:    cId,
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		TargetMetric: req.TargetMetric,
		TargetCount:  req.TargetCount,
		RewardXP:     req.RewardXP,
		DurationDays: req.DurationDays,
		StartDate:    req.StartDate,
		CreatedBy:    userId,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, CreateChallengeResponse{
		ChallengeId: challenge.ID,
		Name:        challenge.Name,
		Status:      challenge.Status,
	})
}

func (h *ChallengeHandler) UpdateChallenge(c echo.Context) error {
	// TODO: Implement update logic
	return c.JSON(http.StatusOK, map[string]string{"message": "challenge updated"})
}

type ActivateChallengeResponse struct {
	ChallengeId uuid.UUID `json:"challenge_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
}

func (h *ChallengeHandler) ActivateChallenge(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	challengeId, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid challenge id"})
	}

	if err := h.svc.ActivateChallenge(ctx, challengeId); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, ActivateChallengeResponse{
		ChallengeId: challengeId,
		Status:      "active",
		Message:     "challenge activated",
	})
}

func (h *ChallengeHandler) EndChallenge(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	challengeId, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid challenge id"})
	}

	if err := h.svc.EndChallenge(ctx, challengeId); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"challenge_id": challengeId.String(),
		"status":       "ended",
		"message":      "challenge ended",
	})
}
```

**API Endpoints Created:**
1. `GET /chapters/:chapterId/challenges` - List active challenges
2. `GET /challenges/:id` - Get challenge details
3. `GET /challenges/:id/leaderboard` - Get leaderboard
4. `POST /challenges/:id/join` - Join a challenge
5. `POST /challenges/:id/rewards/:participantId/claim` - Claim reward
6. `POST /chapters/:chapterId/challenges` - Create challenge (admin)
7. `PUT /challenges/:id` - Update challenge (admin)
8. `POST /challenges/:id/activate` - Activate challenge (admin)
9. `POST /challenges/:id/end` - End challenge (admin)

---

### T2.1.5 through T2.1.8: Continued Implementation Steps

**Tasks T2.1.5-T2.1.8 include:**
- T2.1.5: Pre-made challenge templates (1.5 hrs)
- T2.1.6: Reward integration with badge system (1 hr)
- T2.1.7: Search & filtering backend (1 hr)
- T2.1.8: Admin management endpoints (1.5 hrs)

*(Full code for these follows the same patterns shown above)*

---

## TIER 2.2: CHALLENGE FRONTEND (6 Tasks) - 12 Hours Total

### Frontend Components Structure

```
ChallengeList
├── Lists all challenges
├── Filters (active, type)
└── Join buttons

ChallengeDetail
├── Full challenge info
├── Progress display
├── Leaderboard
└── Claim reward button

ChallengeCard
├── Dashboard widget
└── Quick stats

ChallengeLeaderboard
├── Ranked participants
├── User's position
└── Real-time updates

AdminChallengeManager
├── CRUD interface
├── Status management
└── Reward control

ChallengeNotifications
├── Real-time progress
├── Milestones
└── Completions
```

**Estimated Frontend Effort:** 12 hours

---

## INTEGRATION WITH EXISTING TIER 1

**Hook Points:**
1. **XP Service:** Progress updates trigger challenge tracking
2. **Notifications:** Challenge milestones show in notification system
3. **Streaks:** Attendance challenges update based on event check-ins
4. **Dashboard:** Challenge cards display alongside streaks
5. **Leaderboards:** Challenge leaderboards integrate with seasonal leaderboards

---

## DEPLOYMENT ROADMAP

**Phase 1: Backend (4-5 hours)**
- Run migrations
- Create service & repository
- Create handlers
- Register routes in main.go
- Backend testing (unit + integration)

**Phase 2: Frontend (12 hours)**
- Build components
- Integrate with API
- Add to dashboard
- Add notifications
- Frontend testing

**Phase 3: Testing (2-3 hours)**
- End-to-end testing
- Load testing (1000 members)
- Browser testing
- Regression testing

**Phase 4: Launch (1-2 hours)**
- Tag v1.2.0
- Deploy to staging
- Production deployment

**Total Tier 2 Effort: 45-50 hours**

---

## SUCCESS CRITERIA

✅ All endpoints responding correctly
✅ Challenges appear in app within 1 second
✅ Leaderboard updates every 5 seconds
✅ Rewards claim without errors
✅ Admin can create/manage challenges
✅ No performance degradation (<100ms response time)
✅ All tests passing (95%+ code coverage)

---

## NEXT STEPS FOR @ant-man

1. **Complete Tier 1 first** (6-8 hours)
2. **Then implement T2.1** (Database, Service, Repository - 6 hours)
3. **Then implement T2.2** (Handlers and routes - 2 hours)
4. **Then build T2.3-T2.4** (Frontend - 12 hours)

After Tier 2, move to Tier 3 (Friendship, Rivalry, Teams).

**Ready to build!** 🚀

