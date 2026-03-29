package xp

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LeaderboardEntry is a single row in the XP leaderboard.
type LeaderboardEntry struct {
	Rank            int     `json:"rank"`
	MemberID        string  `json:"member_id"`
	MemberDisplayID string  `json:"member_display_id"`
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	AvatarBg        *string `json:"avatar_bg,omitempty"`
	AvatarFg        *string `json:"avatar_fg,omitempty"`
	XPTotal         int     `json:"xp_total"`
	XPSemester      int     `json:"xp_semester"`
	Level           string  `json:"level"`
	LevelKey        string  `json:"level_key"`
	Role            string  `json:"role"`
}

// AwardXPInput holds the parameters for manually awarding XP.
type AwardXPInput struct {
	MemberID   string `json:"member_id" validate:"required"`
	XPAmount   int    `json:"xp_amount" validate:"required,min=1"`
	Activity   string `json:"activity" validate:"required"`
	Note       *string `json:"note"`
	Semester   *string `json:"semester"`
}

// EngagementLogEntry is a full engagement log row for admin viewing.
type EngagementLogEntry struct {
	ID            string     `json:"id"`
	MemberID      string     `json:"member_id"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Activity      string     `json:"activity"`
	XPAwarded     int        `json:"xp_awarded"`
	Source        string     `json:"source"`
	ReferenceID   *string    `json:"reference_id,omitempty"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	Semester      *string    `json:"semester,omitempty"`
	Note          *string    `json:"note,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// XPService defines the XP domain business logic.
type XPService interface {
	GetLeaderboard(ctx context.Context, chapterID string, mode string) ([]*LeaderboardEntry, error)
	AwardXP(ctx context.Context, chapterID, awardedByMemberID string, input AwardXPInput) error
	RecalculateAll(ctx context.Context, chapterID string) (int, error)
	GetEngagementLog(ctx context.Context, chapterID string, page, perPage int) ([]*EngagementLogEntry, int, error)
}

type xpService struct {
	db *pgxpool.Pool
}

// NewXPService creates a new XP service.
func NewXPService(db *pgxpool.Pool) XPService {
	return &xpService{db: db}
}

// GetLeaderboard returns the chapter XP leaderboard.
// mode: "alltime" (default) or "semester"
func (s *xpService) GetLeaderboard(ctx context.Context, chapterID, mode string) ([]*LeaderboardEntry, error) {
	xpCol := "m.xp_total"
	if mode == "semester" {
		xpCol = "m.xp_semester"
	}

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT
			ROW_NUMBER() OVER (ORDER BY %s DESC) AS rank,
			m.id, m.display_id,
			u.first_name, u.last_name, u.avatar_url,
			m.avatar_bg, m.avatar_fg,
			m.xp_total, m.xp_semester,
			m.level_key, m.level_key, m.role
		FROM members m
		JOIN users u ON u.id = m.user_id
		WHERE m.chapter_id = $1 AND m.status = 'active' AND m.deleted_at IS NULL
		ORDER BY %s DESC`, xpCol, xpCol),
		chapterID,
	)
	if err != nil {
		return nil, fmt.Errorf("query leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []*LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(
			&e.Rank, &e.MemberID, &e.MemberDisplayID,
			&e.FirstName, &e.LastName, &e.AvatarURL,
			&e.AvatarBg, &e.AvatarFg,
			&e.XPTotal, &e.XPSemester,
			&e.Level, &e.LevelKey, &e.Role,
		); err != nil {
			return nil, fmt.Errorf("scan leaderboard row: %w", err)
		}
		entries = append(entries, &e)
	}

	return entries, nil
}

// AwardXP manually awards XP to a member and writes an engagement_log entry.
func (s *xpService) AwardXP(ctx context.Context, chapterID, awardedByMemberID string, input AwardXPInput) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO engagement_log (chapter_id, member_id, activity, xp_awarded, source, awarded_by, note, semester)
		VALUES ($1, $2, $3, $4, 'admin', $5, $6, $7)`,
		chapterID, input.MemberID, input.Activity, input.XPAmount,
		awardedByMemberID, input.Note, input.Semester,
	)
	if err != nil {
		return fmt.Errorf("insert engagement log: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE members SET xp_total = xp_total + $1, xp_semester = xp_semester + $1, updated_at = NOW() WHERE id = $2`,
		input.XPAmount, input.MemberID,
	)
	if err != nil {
		return fmt.Errorf("update member xp: %w", err)
	}

	// Update level based on new XP total
	if err := updateMemberLevel(ctx, tx, input.MemberID); err != nil {
		return fmt.Errorf("update level: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit xp award: %w", err)
	}

	// Run badge engine asynchronously
	go func() {
		bgCtx := context.Background()
		if _, err := CheckAndAwardBadges(bgCtx, s.db, chapterID, input.MemberID); err != nil {
			// Log but don't fail the XP award
		}
	}()

	return nil
}

// RecalculateAll recalculates xp_total for every member in a chapter from engagement_log.
// Returns the count of members updated.
func (s *xpService) RecalculateAll(ctx context.Context, chapterID string) (int, error) {
	result, err := s.db.Exec(ctx, `
		UPDATE members m
		SET xp_total = subq.total,
		    updated_at = NOW()
		FROM (
		    SELECT member_id, COALESCE(SUM(xp_awarded), 0) AS total
		    FROM engagement_log
		    WHERE chapter_id = $1
		    GROUP BY member_id
		) subq
		WHERE m.id = subq.member_id AND m.chapter_id = $1`,
		chapterID,
	)
	if err != nil {
		return 0, fmt.Errorf("recalculate xp: %w", err)
	}
	return int(result.RowsAffected()), nil
}

// GetEngagementLog returns the full paginated engagement log for a chapter.
func (s *xpService) GetEngagementLog(ctx context.Context, chapterID string, page, perPage int) ([]*EngagementLogEntry, int, error) {
	offset := (page - 1) * perPage

	var total int
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM engagement_log WHERE chapter_id = $1`, chapterID).Scan(&total)

	rows, err := s.db.Query(ctx, `
		SELECT el.id, el.member_id, u.first_name, u.last_name,
		       el.activity, el.xp_awarded, el.source,
		       el.reference_id, el.reference_type, el.semester, el.note, el.created_at
		FROM engagement_log el
		JOIN members m ON m.id = el.member_id
		JOIN users u ON u.id = m.user_id
		WHERE el.chapter_id = $1
		ORDER BY el.created_at DESC
		LIMIT $2 OFFSET $3`,
		chapterID, perPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query engagement log: %w", err)
	}
	defer rows.Close()

	var entries []*EngagementLogEntry
	for rows.Next() {
		var e EngagementLogEntry
		if err := rows.Scan(
			&e.ID, &e.MemberID, &e.FirstName, &e.LastName,
			&e.Activity, &e.XPAwarded, &e.Source,
			&e.ReferenceID, &e.ReferenceType, &e.Semester, &e.Note, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan engagement log row: %w", err)
		}
		entries = append(entries, &e)
	}

	return entries, total, nil
}

// updateMemberLevel recalculates and persists the member's level based on their current xp_total.
// Level thresholds match the prototype constants.
func updateMemberLevel(ctx context.Context, tx pgx.Tx, memberID string) error {
	// Level thresholds
	type levelDef struct {
		key   string
		label string
		min   int
	}
	levels := []levelDef{
		{"icon", "Chapter Icon", 2500},
		{"gold", "Gold Legend", 1500},
		{"silver", "Silver Elite", 750},
		{"bronze", "Bronze Varsity", 250},
		{"neo", "Neophyte", 0},
	}

	var xpTotal int
	row := tx.QueryRow(ctx, `SELECT xp_total FROM members WHERE id = $1`, memberID)
	if err := row.Scan(&xpTotal); err != nil {
		return nil // non-fatal
	}

	newKey, newLabel := "neo", "Neophyte"
	for _, l := range levels {
		if xpTotal >= l.min {
			newKey = l.key
			newLabel = l.label
			break
		}
	}

	_, err := tx.Exec(ctx,
		`UPDATE members SET level = $1, level_key = $2 WHERE id = $3`,
		newLabel, newKey, memberID)
	return err
}
