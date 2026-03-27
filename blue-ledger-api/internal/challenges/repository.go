package challenges

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound      = errors.New("challenge not found")
	ErrUnauthorized  = errors.New("not your challenge")
	ErrBadStatus     = errors.New("challenge is not in the expected status")
	ErrSelfChallenge = errors.New("cannot challenge yourself")
	ErrAlreadyActive = errors.New("an active challenge already exists between these members")
)

// ── Domain types ────────────────────────────────────────────────────────────────

type Challenge struct {
	ID           string          `json:"id"`
	ChapterID    string          `json:"chapter_id"`
	ChallengerID string          `json:"challenger_id"`
	ChallengedID string          `json:"challenged_id"`
	Type         string          `json:"type"`
	Status       string          `json:"status"`
	XPStake      int             `json:"xp_stake"`
	GameData     json.RawMessage `json:"game_data"`
	WinnerID     *string         `json:"winner_id,omitempty"`
	ExpiresAt    time.Time       `json:"expires_at"`
	AcceptedAt   *time.Time      `json:"accepted_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`

	// Joined display fields
	ChallengerName  string `json:"challenger_name,omitempty"`
	ChallengedName  string `json:"challenged_name,omitempty"`
	ChallengerLevel string `json:"challenger_level,omitempty"`
	ChallengedLevel string `json:"challenged_level,omitempty"`
}

type SendInput struct {
	ChallengedID string `json:"challenged_id" validate:"required"`
	Type         string `json:"type"          validate:"required,oneof=xp_duel service_race trivia streak_showdown"`
	XPStake      int    `json:"xp_stake"      validate:"required,min=10,max=500"`
}

type SubmitInput struct {
	Answers []int `json:"answers" validate:"required"`
}

// ── Repository ──────────────────────────────────────────────────────────────────

type repository struct {
	db *pgxpool.Pool
}

func newRepository(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func (r *repository) create(ctx context.Context, ch *Challenge) (*Challenge, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO challenges
		  (chapter_id, challenger_id, challenged_id, type, xp_stake, game_data, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, chapter_id, challenger_id, challenged_id, type, status,
		          xp_stake, game_data, winner_id, expires_at, accepted_at, completed_at, created_at`,
		ch.ChapterID, ch.ChallengerID, ch.ChallengedID, ch.Type, ch.XPStake,
		ch.GameData, ch.ExpiresAt,
	)
	return scanChallenge(row)
}

func (r *repository) getByID(ctx context.Context, id, chapterID string) (*Challenge, error) {
	row := r.db.QueryRow(ctx, `
		SELECT c.id, c.chapter_id, c.challenger_id, c.challenged_id, c.type, c.status,
		       c.xp_stake, c.game_data, c.winner_id, c.expires_at, c.accepted_at, c.completed_at, c.created_at,
		       COALESCE(cr.name,'') AS challenger_name, COALESCE(cd.name,'') AS challenged_name,
		       COALESCE(cr.level_key,'neo') AS challenger_level, COALESCE(cd.level_key,'neo') AS challenged_level
		FROM   challenges c
		LEFT JOIN members cr ON cr.id = c.challenger_id
		LEFT JOIN members cd ON cd.id = c.challenged_id
		WHERE  c.id = $1 AND c.chapter_id = $2`, id, chapterID)
	return scanChallengeWithNames(row)
}

func (r *repository) listForMember(ctx context.Context, memberID, chapterID string) ([]*Challenge, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.chapter_id, c.challenger_id, c.challenged_id, c.type, c.status,
		       c.xp_stake, c.game_data, c.winner_id, c.expires_at, c.accepted_at, c.completed_at, c.created_at,
		       COALESCE(cr.name,'') AS challenger_name, COALESCE(cd.name,'') AS challenged_name,
		       COALESCE(cr.level_key,'neo') AS challenger_level, COALESCE(cd.level_key,'neo') AS challenged_level
		FROM   challenges c
		LEFT JOIN members cr ON cr.id = c.challenger_id
		LEFT JOIN members cd ON cd.id = c.challenged_id
		WHERE  c.chapter_id = $1
		  AND  (c.challenger_id = $2 OR c.challenged_id = $2)
		ORDER BY c.created_at DESC
		LIMIT 50`, chapterID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Challenge
	for rows.Next() {
		ch, err := scanChallengeWithNamesRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func (r *repository) hasActiveChallenge(ctx context.Context, chapterID, aID, bID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM challenges
		WHERE chapter_id = $1
		  AND status IN ('pending','accepted','active')
		  AND (
		    (challenger_id = $2 AND challenged_id = $3) OR
		    (challenger_id = $3 AND challenged_id = $2)
		  )`, chapterID, aID, bID).Scan(&count)
	return count > 0, err
}

func (r *repository) updateStatus(ctx context.Context, id, status string, winnerID *string, gameData json.RawMessage) (*Challenge, error) {
	var acceptedSet, completedSet string
	if status == "accepted" {
		acceptedSet = ", accepted_at = NOW()"
	}
	if status == "completed" || status == "declined" || status == "expired" {
		completedSet = ", completed_at = NOW()"
	}

	query := fmt.Sprintf(`
		UPDATE challenges
		SET    status = $2, winner_id = $3 %s %s
		     , game_data = COALESCE($4, game_data)
		WHERE  id = $1
		RETURNING id, chapter_id, challenger_id, challenged_id, type, status,
		          xp_stake, game_data, winner_id, expires_at, accepted_at, completed_at, created_at`,
		acceptedSet, completedSet)

	row := r.db.QueryRow(ctx, query, id, status, winnerID, gameData)
	return scanChallenge(row)
}

func (r *repository) expireStale(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE challenges
		SET status = 'expired', completed_at = NOW()
		WHERE status IN ('pending','active')
		  AND expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// ── Scan helpers ───────────────────────────────────────────────────────────────

func scanChallenge(row pgx.Row) (*Challenge, error) {
	var ch Challenge
	err := row.Scan(
		&ch.ID, &ch.ChapterID, &ch.ChallengerID, &ch.ChallengedID,
		&ch.Type, &ch.Status, &ch.XPStake, &ch.GameData,
		&ch.WinnerID, &ch.ExpiresAt, &ch.AcceptedAt, &ch.CompletedAt, &ch.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &ch, err
}

func scanChallengeWithNames(row pgx.Row) (*Challenge, error) {
	var ch Challenge
	err := row.Scan(
		&ch.ID, &ch.ChapterID, &ch.ChallengerID, &ch.ChallengedID,
		&ch.Type, &ch.Status, &ch.XPStake, &ch.GameData,
		&ch.WinnerID, &ch.ExpiresAt, &ch.AcceptedAt, &ch.CompletedAt, &ch.CreatedAt,
		&ch.ChallengerName, &ch.ChallengedName, &ch.ChallengerLevel, &ch.ChallengedLevel,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &ch, err
}

func scanChallengeWithNamesRows(rows pgx.Rows) (*Challenge, error) {
	var ch Challenge
	err := rows.Scan(
		&ch.ID, &ch.ChapterID, &ch.ChallengerID, &ch.ChallengedID,
		&ch.Type, &ch.Status, &ch.XPStake, &ch.GameData,
		&ch.WinnerID, &ch.ExpiresAt, &ch.AcceptedAt, &ch.CompletedAt, &ch.CreatedAt,
		&ch.ChallengerName, &ch.ChallengedName, &ch.ChallengerLevel, &ch.ChallengedLevel,
	)
	return &ch, err
}
