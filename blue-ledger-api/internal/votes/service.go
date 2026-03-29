package votes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound     = errors.New("vote not found")
	ErrAlreadyVoted = errors.New("member has already voted on this item")
	ErrVoteClosed   = errors.New("voting is closed")
)

// Vote is the domain model for a chapter vote/poll.
type Vote struct {
	ID          string     `json:"id"`
	ChapterID   string     `json:"chapter_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	VoteType    string     `json:"vote_type"`
	Options     []string   `json:"options"`
	IsActive    bool       `json:"is_active"`
	IsAnonymous bool       `json:"is_anonymous"`
	CreatedBy   string     `json:"created_by"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	// Viewer state
	MyVoteIndex *int `json:"my_vote_index,omitempty"`
}

// VoteResult holds the tally for a vote.
type VoteResult struct {
	VoteID  string         `json:"vote_id"`
	Total   int            `json:"total"`
	Options map[string]int `json:"options"` // key = stringified option index
}

// CreateInput holds data for creating a new vote.
type CreateInput struct {
	Title       string     `json:"title" validate:"required,min=2,max=200"`
	Description *string    `json:"description"`
	VoteType    string     `json:"vote_type" validate:"required,oneof=election referendum motion"`
	Options     []string   `json:"options" validate:"required,min=2"`
	Deadline    *time.Time `json:"deadline"`
	IsAnonymous bool       `json:"is_anonymous"`
}

// RespondInput holds a member's vote choice (0-based option index).
type RespondInput struct {
	OptionIndex int `json:"option_index" validate:"min=0"`
}

// Service defines the votes business logic interface.
type Service interface {
	List(ctx context.Context, chapterID, memberID string) ([]*Vote, error)
	Create(ctx context.Context, chapterID, createdByMemberID string, input CreateInput) (*Vote, error)
	Vote(ctx context.Context, chapterID, voteID, memberID string, input RespondInput) error
	GetResults(ctx context.Context, chapterID, voteID string) (*VoteResult, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new votes service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID, memberID string) ([]*Vote, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			v.id, v.chapter_id, v.title, v.description, v.vote_type, v.options,
			v.is_active, v.is_anonymous, v.created_by, v.deadline, v.created_at,
			vr.option_index AS my_vote_index
		FROM votes v
		LEFT JOIN vote_responses vr ON vr.vote_id = v.id AND vr.member_id = $2
		WHERE v.chapter_id = $1
		ORDER BY v.is_active DESC, v.created_at DESC
	`, chapterID, memberID)
	if err != nil {
		return nil, fmt.Errorf("list votes: %w", err)
	}
	defer rows.Close()

	var result []*Vote
	for rows.Next() {
		v := &Vote{}
		if err := rows.Scan(
			&v.ID, &v.ChapterID, &v.Title, &v.Description, &v.VoteType, &v.Options,
			&v.IsActive, &v.IsAnonymous, &v.CreatedBy, &v.Deadline, &v.CreatedAt,
			&v.MyVoteIndex,
		); err != nil {
			return nil, fmt.Errorf("scan vote: %w", err)
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID, createdByMemberID string, input CreateInput) (*Vote, error) {
	v := &Vote{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO votes (chapter_id, title, description, vote_type, options, is_active, is_anonymous, created_by, deadline)
		VALUES ($1, $2, $3, $4, $5, TRUE, $6, $7, $8)
		RETURNING id, chapter_id, title, description, vote_type, options, is_active, is_anonymous, created_by, deadline, created_at
	`, chapterID, input.Title, input.Description, input.VoteType, input.Options, input.IsAnonymous, createdByMemberID, input.Deadline).
		Scan(&v.ID, &v.ChapterID, &v.Title, &v.Description, &v.VoteType, &v.Options,
			&v.IsActive, &v.IsAnonymous, &v.CreatedBy, &v.Deadline, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create vote: %w", err)
	}
	return v, nil
}

func (s *service) Vote(ctx context.Context, chapterID, voteID, memberID string, input RespondInput) error {
	var isActive bool
	if err := s.pool.QueryRow(ctx,
		`SELECT is_active FROM votes WHERE id = $1 AND chapter_id = $2`, voteID, chapterID,
	).Scan(&isActive); errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("check vote: %w", err)
	}
	if !isActive {
		return ErrVoteClosed
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO vote_responses (vote_id, chapter_id, member_id, option_index)
		VALUES ($1, $2, $3, $4)
	`, voteID, chapterID, memberID, input.OptionIndex)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyVoted
		}
		return fmt.Errorf("insert vote response: %w", err)
	}
	return nil
}

func (s *service) GetResults(ctx context.Context, chapterID, voteID string) (*VoteResult, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM votes WHERE id = $1 AND chapter_id = $2)`, voteID, chapterID,
	).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check vote exists: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	rows, err := s.pool.Query(ctx, `
		SELECT option_index, COUNT(*) FROM vote_responses
		WHERE vote_id = $1
		GROUP BY option_index
		ORDER BY option_index
	`, voteID)
	if err != nil {
		return nil, fmt.Errorf("get vote results: %w", err)
	}
	defer rows.Close()

	result := &VoteResult{
		VoteID:  voteID,
		Options: make(map[string]int),
	}
	for rows.Next() {
		var optIdx int
		var count int
		if err := rows.Scan(&optIdx, &count); err != nil {
			return nil, fmt.Errorf("scan vote result: %w", err)
		}
		result.Options[fmt.Sprintf("%d", optIdx)] = count
		result.Total += count
	}
	return result, rows.Err()
}

// isUniqueViolation checks for PostgreSQL unique constraint error code 23505.
func isUniqueViolation(err error) bool {
	return err != nil && (func() bool {
		type pgErr interface{ SQLState() string }
		if e, ok := err.(pgErr); ok {
			return e.SQLState() == "23505"
		}
		return false
	})()
}
