package streaks

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Service handles all streak-related business logic
type Service struct {
	db *sql.DB
}

// NewService creates a new streak service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// StreakInfo represents a member's current streak data
type StreakInfo struct {
	MemberID      uuid.UUID `json:"member_id"`
	CurrentStreak int       `json:"current_streak"`
	LongestStreak int       `json:"longest_streak"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpdateStreakForAttendance updates a member's streak based on attendance
func (s *Service) UpdateStreakForAttendance(ctx context.Context, memberID uuid.UUID) (*StreakInfo, error) {
	var info StreakInfo
	var lastUpdatedDate sql.NullTime

	// Get current streak info
	err := s.db.QueryRowContext(ctx, `
		SELECT id, current_streak, longest_streak, streak_updated_at
		FROM members
		WHERE id = $1
	`, memberID).Scan(&memberID, &info.CurrentStreak, &info.LongestStreak, &lastUpdatedDate)

	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	
	// Check if we already updated today
	if lastUpdatedDate.Valid {
		lastDay := time.Date(
			lastUpdatedDate.Time.Year(),
			lastUpdatedDate.Time.Month(),
			lastUpdatedDate.Time.Day(),
			0, 0, 0, 0, time.UTC,
		)
		
		if lastDay.Equal(today) {
			// Already updated today
			info.UpdatedAt = lastUpdatedDate.Time
			return &info, nil
		}
		
		// Check if yesterday was last update (consecutive)
		yesterday := today.AddDate(0, 0, -1)
		if lastDay.Equal(yesterday) {
			info.CurrentStreak++
		} else {
			// Streak broken - reset to 1
			info.CurrentStreak = 1
		}
	} else {
		// First attendance
		info.CurrentStreak = 1
	}

	// Update longest streak if current exceeds it
	if info.CurrentStreak > info.LongestStreak {
		info.LongestStreak = info.CurrentStreak
	}

	// Update database
	now = time.Now()
	_, err = s.db.ExecContext(ctx, `
		UPDATE members
		SET current_streak = $1, longest_streak = $2, streak_updated_at = $3
		WHERE id = $4
	`, info.CurrentStreak, info.LongestStreak, now, memberID)

	if err != nil {
		return nil, err
	}

	info.UpdatedAt = now
	return &info, nil
}

// GetStreakInfo returns a member's streak information
func (s *Service) GetStreakInfo(ctx context.Context, memberID uuid.UUID) (*StreakInfo, error) {
	var info StreakInfo

	err := s.db.QueryRowContext(ctx, `
		SELECT id, current_streak, longest_streak, streak_updated_at
		FROM members
		WHERE id = $1
	`, memberID).Scan(&memberID, &info.CurrentStreak, &info.LongestStreak, &info.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	info.MemberID = memberID
	return &info, nil
}

// LeaderboardEntry represents a streak leaderboard entry
type LeaderboardEntry struct {
	Rank          int       `json:"rank"`
	MemberID      uuid.UUID `json:"member_id"`
	Name          string    `json:"name"`
	CurrentStreak int       `json:"current_streak"`
	LongestStreak int       `json:"longest_streak"`
	PhotoURL      string    `json:"photo_url"`
}

// GetStreakLeaderboard returns top members by current streak
func (s *Service) GetStreakLeaderboard(ctx context.Context, chapterID uuid.UUID, limit int) ([]LeaderboardEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT 
			ROW_NUMBER() OVER (ORDER BY m.current_streak DESC) as rank,
			m.id,
			u.first_name || ' ' || u.last_name as name,
			m.current_streak,
			m.longest_streak,
			u.photo_url
		FROM members m
		JOIN users u ON m.user_id = u.id
		WHERE m.chapter_id = $1
		ORDER BY m.current_streak DESC, m.streak_updated_at DESC
		LIMIT $2
	`, chapterID, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		err := rows.Scan(
			&entry.Rank,
			&entry.MemberID,
			&entry.Name,
			&entry.CurrentStreak,
			&entry.LongestStreak,
			&entry.PhotoURL,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// CalculateStreakBonus calculates XP bonus based on streak
func (s *Service) CalculateStreakBonus(streak int) int {
	switch {
	case streak >= 30:
		return 250
	case streak >= 14:
		return 100
	case streak >= 7:
		return 50
	default:
		return 0
	}
}

// ResetStreakIfNeeded checks if a member missed a day and resets streak
func (s *Service) ResetStreakIfNeeded(ctx context.Context, memberID uuid.UUID) error {
	var lastUpdated sql.NullTime

	err := s.db.QueryRowContext(ctx, `
		SELECT streak_updated_at FROM members WHERE id = $1
	`, memberID).Scan(&lastUpdated)

	if err != nil {
		return err
	}

	if !lastUpdated.Valid {
		return nil
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	lastDay := time.Date(
		lastUpdated.Time.Year(),
		lastUpdated.Time.Month(),
		lastUpdated.Time.Day(),
		0, 0, 0, 0, time.UTC,
	)
	
	// If more than 1 day has passed, reset streak
	daysSince := today.Sub(lastDay).Hours() / 24
	if daysSince > 1 {
		_, err := s.db.ExecContext(ctx, `
			UPDATE members
			SET current_streak = 0, streak_updated_at = $1
			WHERE id = $2
		`, now, memberID)
		return err
	}

	return nil
}
