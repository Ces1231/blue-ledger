package milestones

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("milestone not found")

// Milestone is a life-event milestone for brotherhood recognition.
type Milestone struct {
	ID          string     `json:"id"`
	ChapterID   string     `json:"chapter_id"`
	MemberID    string     `json:"member_id"`
	MemberName  *string    `json:"member_name,omitempty"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Date        *string    `json:"date,omitempty"` // "YYYY-MM-DD"
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateInput holds the data required to create a milestone.
type CreateInput struct {
	MemberID    string  `json:"member_id"   validate:"required"`
	Type        string  `json:"type"        validate:"required,oneof=birthday graduation new_job engagement other"`
	Title       string  `json:"title"       validate:"required,min=2,max=200"`
	Date        *string `json:"date"`
	Description *string `json:"description"`
}

// Service defines the milestones business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string, memberID string) ([]*Milestone, error)
	Create(ctx context.Context, chapterID string, input CreateInput) (*Milestone, error)
	Delete(ctx context.Context, chapterID, id string) error
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new milestones service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string, memberID string) ([]*Milestone, error) {
	query := `
		SELECT m.id, m.chapter_id, m.member_id,
		       mb.name AS member_name,
		       m.type, m.title,
		       TO_CHAR(m.date, 'YYYY-MM-DD') AS date,
		       m.description, m.created_at
		FROM milestones m
		LEFT JOIN members mb ON mb.id = m.member_id
		WHERE m.chapter_id = $1
	`
	args := []any{chapterID}

	if memberID != "" {
		query += ` AND m.member_id = $2`
		args = append(args, memberID)
	}
	query += ` ORDER BY m.created_at DESC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list milestones: %w", err)
	}
	defer rows.Close()

	var result []*Milestone
	for rows.Next() {
		ms := &Milestone{}
		if err := rows.Scan(
			&ms.ID, &ms.ChapterID, &ms.MemberID,
			&ms.MemberName,
			&ms.Type, &ms.Title,
			&ms.Date,
			&ms.Description, &ms.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan milestone: %w", err)
		}
		result = append(result, ms)
	}
	return result, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Milestone, error) {
	ms := &Milestone{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO milestones (chapter_id, member_id, type, title, date, description)
		VALUES ($1, $2, $3, $4, $5::DATE, $6)
		RETURNING id, chapter_id, member_id, type, title,
		          TO_CHAR(date, 'YYYY-MM-DD'), description, created_at
	`, chapterID, input.MemberID, input.Type, input.Title, input.Date, input.Description).
		Scan(
			&ms.ID, &ms.ChapterID, &ms.MemberID,
			&ms.Type, &ms.Title,
			&ms.Date,
			&ms.Description, &ms.CreatedAt,
		)
	if err != nil {
		return nil, fmt.Errorf("create milestone: %w", err)
	}
	return ms, nil
}

func (s *service) Delete(ctx context.Context, chapterID, id string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM milestones WHERE id = $1 AND chapter_id = $2`,
		id, chapterID,
	)
	if err != nil {
		return fmt.Errorf("delete milestone: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Ensure pgx.ErrNoRows is imported (used for consistency with other packages).
var _ = pgx.ErrNoRows
