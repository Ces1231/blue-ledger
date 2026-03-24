package xp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// BadgeCriteria represents the JSONB criteria stored on a badge definition.
type BadgeCriteria struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

// BadgeDefinition is a badge row from the database.
type BadgeDefinition struct {
	ID          string
	Name        string
	Icon        string
	Description string
	XPReward    int
	Criteria    BadgeCriteria
}

// CheckAndAwardBadges evaluates all unearned badges for a given member and awards
// any whose criteria are now met. This runs asynchronously after XP events.
// Returns the list of newly awarded badge names (for notification content).
func CheckAndAwardBadges(ctx context.Context, pool *pgxpool.Pool, chapterID, memberID string) ([]string, error) {
	// Load all active badge definitions for this chapter that the member hasn't earned yet
	rows, err := pool.Query(ctx, `
		SELECT b.id, b.name, COALESCE(b.icon, '🏆'), COALESCE(b.description, ''), b.xp_reward, b.criteria
		FROM badges b
		WHERE b.chapter_id = $1
		  AND b.is_active = true
		  AND NOT EXISTS (
		    SELECT 1 FROM member_badges mb
		    WHERE mb.badge_id = b.id AND mb.member_id = $2
		  )`,
		chapterID, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("load badges: %w", err)
	}
	defer rows.Close()

	var badges []BadgeDefinition
	for rows.Next() {
		var b BadgeDefinition
		var criteriaJSON []byte
		if err := rows.Scan(&b.ID, &b.Name, &b.Icon, &b.Description, &b.XPReward, &criteriaJSON); err != nil {
			continue
		}
		if err := json.Unmarshal(criteriaJSON, &b.Criteria); err != nil {
			continue
		}
		badges = append(badges, b)
	}

	if len(badges) == 0 {
		return nil, nil
	}

	// Load member stats needed for criteria evaluation
	var xpTotal int
	pool.QueryRow(ctx, `SELECT xp_total FROM members WHERE id = $1`, memberID).Scan(&xpTotal)

	var serviceHours float64
	pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(hours), 0) FROM service_log
		WHERE member_id = $1 AND verified = true`, memberID).Scan(&serviceHours)

	var onTimeMeetings int
	pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM engagement_log
		WHERE member_id = $1 AND source = 'checkin' AND activity ILIKE '%on time%'`, memberID).Scan(&onTimeMeetings)

	var paidSemesters int
	pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM dues_records
		WHERE member_id = $1 AND status = 'paid'`, memberID).Scan(&paidSemesters)

	var awarded []string

	for _, badge := range badges {
		met := false

		switch badge.Criteria.Type {
		case "xp_gte":
			met = float64(xpTotal) >= badge.Criteria.Value
		case "service_hours_gte":
			met = serviceHours >= badge.Criteria.Value
		case "on_time_meetings_gte":
			met = float64(onTimeMeetings) >= badge.Criteria.Value
		case "dues_on_time_consecutive_semesters_gte":
			met = float64(paidSemesters) >= badge.Criteria.Value
		default:
			// Unknown criteria type — skip without error
			log.Warn().Str("criteria_type", badge.Criteria.Type).Str("badge_id", badge.ID).Msg("unknown badge criteria type")
			continue
		}

		if !met {
			continue
		}

		// Award the badge
		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Error().Err(err).Str("badge_id", badge.ID).Msg("begin badge award tx")
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO member_badges (chapter_id, member_id, badge_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (member_id, badge_id) DO NOTHING`,
			chapterID, memberID, badge.ID,
		)
		if err != nil {
			tx.Rollback(ctx)
			log.Error().Err(err).Str("badge_id", badge.ID).Msg("insert member_badge")
			continue
		}

		if badge.XPReward > 0 {
			_, err = tx.Exec(ctx, `
				INSERT INTO engagement_log (chapter_id, member_id, activity, xp_awarded, source, reference_id, reference_type)
				VALUES ($1, $2, $3, $4, 'badge', $5, 'badge')`,
				chapterID, memberID, fmt.Sprintf("Badge earned: %s", badge.Name), badge.XPReward, badge.ID,
			)
			if err != nil {
				tx.Rollback(ctx)
				continue
			}

			_, err = tx.Exec(ctx,
				`UPDATE members SET xp_total = xp_total + $1 WHERE id = $2`,
				badge.XPReward, memberID)
			if err != nil {
				tx.Rollback(ctx)
				continue
			}
		}

		// Create notification
		_, _ = tx.Exec(ctx, `
			INSERT INTO notifications (chapter_id, user_id, type, title, body, link)
			SELECT chapter_id, user_id, 'badge', $2, $3, '/quests'
			FROM members WHERE id = $1`,
			memberID,
			fmt.Sprintf("Badge Earned: %s %s", badge.Icon, badge.Name),
			badge.Description,
		)

		if err := tx.Commit(ctx); err != nil {
			log.Error().Err(err).Str("badge_id", badge.ID).Msg("commit badge award")
			continue
		}

		awarded = append(awarded, badge.Name)
		log.Info().Str("member_id", memberID).Str("badge", badge.Name).Msg("badge awarded")
	}

	return awarded, nil
}
