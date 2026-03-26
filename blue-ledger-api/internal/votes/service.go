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
	Options     []string   `json:"options"`
	IsOpen      bool       `json:"is_open"`
	CreatedBy   string     `json:"created_by"`
	ClosesAt    *time.Time `json:"closes_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	// Viewer state
	MyResponse *string `json:"my_response,omitempty"`
}

// VoteResult holds the tally for a vote.
type VoteResult struct {
	VoteID  string            `json:"vote_id"`
	Total   int               `json:"total"`
	Options map[string]int    `json:"options"`
}

// CreateInput holds data for creating a new vote.
type CreateInput struct {
	Title       string     `json:"title" validate:"required,min=2,max=200"`
	Description *string    `json:"description"`
	Options     []string   `json:"options" validate:"required,min=2"`
	ClosesAt    *time.Time `json:"closes_at"`
}

// RespondInput holds a member's vote choice.
type RespondInput struct {
	Choice string `json:"choice" validate:"required"`
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
			v.id, v.chapter_id, v.title, v.description, v.options,
			v.is_open, v.created_by, v.closes_at, v.created_at,
			vr.choice AS my_response
		FROM votes v
		LEFT JOIN vote_responses vr ON vr.vote_id = v.id AND vr.member_id = $2
		WHERE v.chapter_id = $1
		ORDER BY v.is_open DESC, v.created_at DESC
	`, chapterID, memberID)
	if err != nil {
		return nil, fmt.Errorf("list votes: %w", err)
	}
	defer rows.Close()

	var result []*Vote
	for rows.Next() {
		v := &Vote{}
		if err := rows.Scan(
			&v.ID, &v.ChapterID, &v.Title, &v.Description, &v.Options,
			&v.IsOpen, &v.CreatedBy, &v.ClosesAt, &v.CreatedAt,
			&v.MyResponse,
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
		INSERT INTO votes (chapter_id, title, description, options, is_open, created_by, closes_at)
		VALUES ($1, $2, $3, $4, TRUE, $5, $6)
		RETURNING id, chapter_id, title, description, options, is_open, created_by, closes_at, created_at
	`, chapterID, input.Title, input.Description, input.Options, createdByMemberID, input.ClosesAt).
		Scan(&v.ID, &v.ChapterID, &v.Title, &v.Description, &v.Options,
			&v.IsOpen, &v.CreatedBy, &v.ClosesAt, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create vote: %w", err)
	}
	return v, nil
}

func (s *service) Vote(ctx context.Context, chapterID, voteID, memberID string, input RespondInput) error {
	// Check vote is open
	var isOpen bool
	if err := s.pool.QueryRow(ctx,
		`SELECT is_open FROM votes WHERE id = $1 AND chapter_id = $2`, voteID, chapterID,
	).Scan(&isOpen); errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("check vote: %w", err)
	}
	if !isOpen {
		return ErrVoteClosed
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO vote_responses (vote_id, chapter_id, member_id, choice)
		VALUES ($1, $2, $3, $4)
	`, voteID, chapterID, memberID, input.Choice)
	if err != nil {
		// Unique constraint violation — already voted
		if isUniqueViolation(err) {
			return ErrAlreadyVoted
		}
		return fmt.Errorf("insert vote response: %w", err)
	}
	return nil
}

func (s *service) GetResults(ctx context.Context, chapterID, voteID string) (*VoteResult, error) {
	// Verify vote exists
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
		SELECT choice, COUNT(*) FROM vote_responses
		WHERE vote_id = $1
		GROUP BY choice
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
		var choice string
		var count int
		if err := rows.Scan(&choice, &count); err != nil {
			return nil, fmt.Errorf("scan vote result: %w", err)
		}
		result.Options[choice] = count
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
