package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LoginBonusRepository handles daily login bonus logic
type LoginBonusRepository struct {
	db *pgxpool.Pool
}

// NewLoginBonusRepository creates a new login bonus repository
func NewLoginBonusRepository(db *pgxpool.Pool) *LoginBonusRepository {
	return &LoginBonusRepository{db: db}
}

// DailyBonusResult represents the result of a daily login bonus
type DailyBonusResult struct {
	Awarded          bool
	XPGranted        int
	CurrentStreak    int
	BonusType        string // "daily", "weekly"
	Message          string
}

// AwardDailyLoginBonus awards XP if user hasn't logged in today
func (r *LoginBonusRepository) AwardDailyLoginBonus(ctx context.Context, userID uuid.UUID) (*DailyBonusResult, error) {
	result := &DailyBonusResult{Awarded: false, XPGranted: 0}

	var lastLoginAt sql.NullTime
	var dailyStreak int

	// Get current login info
	err := r.db.QueryRow(ctx, `
		SELECT last_login_at, daily_login_streak
		FROM users
		WHERE id = $1
	`, userID).Scan(&lastLoginAt, &dailyStreak)

	if err != nil {
		// User not found - this shouldn't happen after login, but handle gracefully
		if err.Error() == "no rows in result set" {
			return result, nil
		}
		return nil, err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Check if already logged in today
	if lastLoginAt.Valid {
		lastDay := time.Date(
			lastLoginAt.Time.Year(),
			lastLoginAt.Time.Month(),
			lastLoginAt.Time.Day(),
			0, 0, 0, 0, time.UTC,
		)

		if lastDay.Equal(today) {
			// Already logged in today
			result.Message = "You've already received your daily bonus today"
			return result, nil
		}

		// Check if yesterday was last login (maintain streak)
		yesterday := today.AddDate(0, 0, -1)
		if lastDay.Equal(yesterday) {
			dailyStreak++
			result.BonusType = "streak_continuation"
		} else {
			// Streak broken - reset
			dailyStreak = 1
			result.BonusType = "streak_reset"
		}
	} else {
		// First login
		dailyStreak = 1
		result.BonusType = "first_login"
	}

	// Calculate XP bonus
	baseXP := 25
	result.XPGranted = baseXP

	// Weekly milestone bonus (every 7 days)
	if dailyStreak > 0 && dailyStreak%7 == 0 {
		result.XPGranted += 50
		result.BonusType = "weekly_milestone"
	}

	// Update last_login_at and daily_login_streak
	_, err = r.db.Exec(ctx, `
		UPDATE users
		SET last_login_at = $1, daily_login_streak = $2
		WHERE id = $3
	`, now, dailyStreak, userID)

	if err != nil {
		return nil, err
	}

	result.Awarded = true
	result.CurrentStreak = dailyStreak
	result.Message = "Daily login bonus awarded!"

	return result, nil
}

// GetLoginStreak returns current login streak for a user
func (r *LoginBonusRepository) GetLoginStreak(ctx context.Context, userID uuid.UUID) (int, error) {
	var streak int
	err := r.db.QueryRow(ctx, `
		SELECT daily_login_streak FROM users WHERE id = $1
	`, userID).Scan(&streak)
	return streak, err
}
