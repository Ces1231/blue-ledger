package sbc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("sbc log entry not found")

// Entry is a service/brotherhood/conduct log entry.
type Entry struct {
	ID         string    `json:"id"`
	ChapterID  string    `json:"chapter_id"`
	MemberID   string    `json:"member_id"`
	Category   string    `json:"category"`
	Note       string    `json:"note"`
	RecordedBy string    `json:"recorded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateInput for creating an SBC log entry.
type CreateInput struct {
	MemberID string `json:"member_id" validate:"required"`
	Category string `json:"category" validate:"required,oneof=service brotherhood conduct"`
	Note     string `json:"note" validate:"required"`
}

// UpdateInput for updating an SBC log entry.
type UpdateInput struct {
	Category *string `json:"category"`
	Note     *string `json:"note"`
}

// Service defines SBC log business logic.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Entry, error)
	Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*Entry, error)
	Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Entry, error)
	Delete(ctx context.Context, chapterID, id string) error
}

type service struct{ pool *pgxpool.Pool }

// NewService creates a new SBC log service.
func NewService(pool *pgxpool.Pool) Service { return &service{pool: pool} }

func (s *service) List(ctx context.Context, chapterID string) ([]*Entry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, member_id, category, note, recorded_by, created_at
		FROM sbc_log WHERE chapter_id = $1 ORDER BY created_at DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list sbc log: %w", err)
	}
	defer rows.Close()
	var result []*Entry
	for rows.Next() {
		e := &Entry{}
		if err := rows.Scan(&e.ID, &e.ChapterID, &e.MemberID, &e.Category, &e.Note, &e.RecordedBy, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan sbc entry: %w", err)
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID, recordedBy string, input CreateInput) (*Entry, error) {
	e := &Entry{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO sbc_log (chapter_id, member_id, category, note, recorded_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, chapter_id, member_id, category, note, recorded_by, created_at
	`, chapterID, input.MemberID, input.Category, input.Note, recordedBy).
		Scan(&e.ID, &e.ChapterID, &e.MemberID, &e.Category, &e.Note, &e.RecordedBy, &e.CreatedAt)
	return e, err
}

func (s *service) Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Entry, error) {
	e := &Entry{}
	err := s.pool.QueryRow(ctx, `
		UPDATE sbc_log
		SET category = COALESCE($3, category),
		    note     = COALESCE($4, note)
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, member_id, category, note, recorded_by, created_at
	`, id, chapterID, input.Category, input.Note).
		Scan(&e.ID, &e.ChapterID, &e.MemberID, &e.Category, &e.Note, &e.RecordedBy, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return e, err
}

func (s *service) Delete(ctx context.Context, chapterID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM sbc_log WHERE id = $1 AND chapter_id = $2`, id, chapterID)
	if err != nil {
		return fmt.Errorf("delete sbc entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
