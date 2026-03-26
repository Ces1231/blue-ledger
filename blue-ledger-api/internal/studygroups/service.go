package studygroups

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("study group not found")

// StudyGroup is a chapter-organised collaborative study session.
type StudyGroup struct {
	ID        string    `json:"id"`
	ChapterID string    `json:"chapter_id"`
	Topic     string    `json:"topic"`
	HostID    string    `json:"host_id"`
	HostName  *string   `json:"host_name,omitempty"`
	Date      string    `json:"date"` // "YYYY-MM-DD"
	Location  *string   `json:"location,omitempty"`
	MemberIDs []string  `json:"member_ids"`
	XPReward  int       `json:"xp_reward"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateInput holds the data required to create a study group.
type CreateInput struct {
	Topic    string  `json:"topic"     validate:"required,min=2,max=200"`
	Date     string  `json:"date"      validate:"required"`
	Location *string `json:"location"`
	XPReward int     `json:"xp_reward"`
}

// Service defines the study-groups business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*StudyGroup, error)
	Create(ctx context.Context, chapterID, hostID string, input CreateInput) (*StudyGroup, error)
	Join(ctx context.Context, chapterID, id, memberID string) error
	Delete(ctx context.Context, chapterID, id string) error
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new study-groups service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string) ([]*StudyGroup, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sg.id, sg.chapter_id, sg.host_id,
		       mb.name AS host_name,
		       sg.topic,
		       TO_CHAR(sg.date, 'YYYY-MM-DD') AS date,
		       sg.location,
		       sg.member_ids::TEXT[] AS member_ids,
		       sg.xp_reward, sg.created_at
		FROM study_groups sg
		LEFT JOIN members mb ON mb.id = sg.host_id
		WHERE sg.chapter_id = $1
		ORDER BY sg.date DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list study groups: %w", err)
	}
	defer rows.Close()

	var result []*StudyGroup
	for rows.Next() {
		sg := &StudyGroup{}
		var rawMembers pgtype.FlatArray[string]
		if err := rows.Scan(
			&sg.ID, &sg.ChapterID, &sg.HostID,
			&sg.HostName,
			&sg.Topic,
			&sg.Date,
			&sg.Location,
			&rawMembers,
			&sg.XPReward, &sg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan study group: %w", err)
		}
		sg.MemberIDs = toStringSlice(rawMembers)
		result = append(result, sg)
	}
	return result, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID, hostID string, input CreateInput) (*StudyGroup, error) {
	sg := &StudyGroup{}
	var rawMembers pgtype.FlatArray[string]
	err := s.pool.QueryRow(ctx, `
		INSERT INTO study_groups (chapter_id, host_id, topic, date, location, xp_reward)
		VALUES ($1, $2, $3, $4::DATE, $5, $6)
		RETURNING id, chapter_id, host_id, topic,
		          TO_CHAR(date, 'YYYY-MM-DD'),
		          location, member_ids::TEXT[], xp_reward, created_at
	`, chapterID, hostID, input.Topic, input.Date, input.Location, input.XPReward).
		Scan(
			&sg.ID, &sg.ChapterID, &sg.HostID,
			&sg.Topic,
			&sg.Date,
			&sg.Location, &rawMembers,
			&sg.XPReward, &sg.CreatedAt,
		)
	if err != nil {
		return nil, fmt.Errorf("create study group: %w", err)
	}
	sg.MemberIDs = toStringSlice(rawMembers)
	return sg, nil
}

func (s *service) Join(ctx context.Context, chapterID, id, memberID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE study_groups
		SET    member_ids = array_append(member_ids, $1::uuid)
		WHERE  id = $2 AND chapter_id = $3
		  AND  NOT ($1::uuid = ANY(member_ids))
	`, memberID, id, chapterID)
	if err != nil {
		return fmt.Errorf("join study group: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Either already joined (idempotent — ok) or the row doesn't exist.
		// Distinguish by checking existence.
		var exists bool
		if err = s.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM study_groups WHERE id = $1 AND chapter_id = $2)`,
			id, chapterID,
		).Scan(&exists); err != nil {
			return fmt.Errorf("check study group existence: %w", err)
		}
		if !exists {
			return ErrNotFound
		}
		// Row exists but member is already in the array — idempotent, return nil.
	}
	return nil
}

func (s *service) Delete(ctx context.Context, chapterID, id string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM study_groups WHERE id = $1 AND chapter_id = $2`,
		id, chapterID,
	)
	if err != nil {
		return fmt.Errorf("delete study group: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// toStringSlice converts a pgtype.FlatArray[string] to a plain []string,
// ensuring the result is never nil (returns an empty slice instead).
func toStringSlice(arr pgtype.FlatArray[string]) []string {
	if len(arr) == 0 {
		return []string{}
	}
	out := make([]string, len(arr))
	copy(out, arr)
	return out
}

// Compile-time assertions — keep the unused import of pgx tidy.
var _ = pgx.ErrNoRows
