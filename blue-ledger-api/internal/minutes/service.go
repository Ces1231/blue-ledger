package minutes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("minutes not found")

// Minutes is the domain model for meeting minutes.
type Minutes struct {
	ID          string     `json:"id"`
	ChapterID   string     `json:"chapter_id"`
	Title       string     `json:"title"`
	MeetingDate string     `json:"meeting_date"`
	Body        string     `json:"body"`
	CreatedBy   string     `json:"created_by"`
	IsFinalized bool       `json:"is_finalized"`
	FinalizedAt *time.Time `json:"finalized_at,omitempty"`
	FinalizedBy *string    `json:"finalized_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	// Joined
	AuthorFirstName *string `json:"author_first_name,omitempty"`
	AuthorLastName  *string `json:"author_last_name,omitempty"`
}

// CreateInput holds data for creating minutes.
type CreateInput struct {
	Title       string `json:"title" validate:"required,min=2,max=200"`
	MeetingDate string `json:"meeting_date" validate:"required"`
	Body        string `json:"body" validate:"required"`
}

// UpdateInput holds data for updating minutes.
type UpdateInput struct {
	Title *string `json:"title"`
	Body  *string `json:"body"`
}

// Service defines the minutes business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string, page, perPage int) ([]*Minutes, int, error)
	Create(ctx context.Context, chapterID, createdByMemberID string, input CreateInput) (*Minutes, error)
	Get(ctx context.Context, chapterID, minutesID string) (*Minutes, error)
	Update(ctx context.Context, chapterID, minutesID string, input UpdateInput) (*Minutes, error)
	Finalize(ctx context.Context, chapterID, minutesID, memberID string) (*Minutes, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new minutes service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string, page, perPage int) ([]*Minutes, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM meeting_minutes WHERE chapter_id = $1`, chapterID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count minutes: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			mm.id, mm.chapter_id, mm.title, mm.meeting_date, mm.body,
			mm.created_by, mm.is_finalized, mm.finalized_at, mm.finalized_by,
			mm.created_at, mm.updated_at,
			m.first_name, m.last_name
		FROM meeting_minutes mm
		LEFT JOIN members m ON m.id = mm.created_by
		WHERE mm.chapter_id = $1
		ORDER BY mm.meeting_date DESC
		LIMIT $2 OFFSET $3
	`, chapterID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list minutes: %w", err)
	}
	defer rows.Close()

	var result []*Minutes
	for rows.Next() {
		mn := &Minutes{}
		if err := rows.Scan(
			&mn.ID, &mn.ChapterID, &mn.Title, &mn.MeetingDate, &mn.Body,
			&mn.CreatedBy, &mn.IsFinalized, &mn.FinalizedAt, &mn.FinalizedBy,
			&mn.CreatedAt, &mn.UpdatedAt,
			&mn.AuthorFirstName, &mn.AuthorLastName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan minutes: %w", err)
		}
		result = append(result, mn)
	}
	return result, total, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID, createdByMemberID string, input CreateInput) (*Minutes, error) {
	mn := &Minutes{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO meeting_minutes (chapter_id, title, meeting_date, body, created_by, is_finalized)
		VALUES ($1, $2, $3, $4, $5, FALSE)
		RETURNING id, chapter_id, title, meeting_date, body, created_by, is_finalized, finalized_at, finalized_by, created_at, updated_at
	`, chapterID, input.Title, input.MeetingDate, input.Body, createdByMemberID).
		Scan(&mn.ID, &mn.ChapterID, &mn.Title, &mn.MeetingDate, &mn.Body,
			&mn.CreatedBy, &mn.IsFinalized, &mn.FinalizedAt, &mn.FinalizedBy, &mn.CreatedAt, &mn.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create minutes: %w", err)
	}
	return mn, nil
}

func (s *service) Get(ctx context.Context, chapterID, minutesID string) (*Minutes, error) {
	mn := &Minutes{}
	err := s.pool.QueryRow(ctx, `
		SELECT
			mm.id, mm.chapter_id, mm.title, mm.meeting_date, mm.body,
			mm.created_by, mm.is_finalized, mm.finalized_at, mm.finalized_by,
			mm.created_at, mm.updated_at,
			m.first_name, m.last_name
		FROM meeting_minutes mm
		LEFT JOIN members m ON m.id = mm.created_by
		WHERE mm.id = $1 AND mm.chapter_id = $2
	`, minutesID, chapterID).
		Scan(&mn.ID, &mn.ChapterID, &mn.Title, &mn.MeetingDate, &mn.Body,
			&mn.CreatedBy, &mn.IsFinalized, &mn.FinalizedAt, &mn.FinalizedBy,
			&mn.CreatedAt, &mn.UpdatedAt,
			&mn.AuthorFirstName, &mn.AuthorLastName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get minutes: %w", err)
	}
	return mn, nil
}

func (s *service) Update(ctx context.Context, chapterID, minutesID string, input UpdateInput) (*Minutes, error) {
	mn := &Minutes{}
	err := s.pool.QueryRow(ctx, `
		UPDATE meeting_minutes
		SET
			title      = COALESCE($3, title),
			body       = COALESCE($4, body),
			updated_at = NOW()
		WHERE id = $1 AND chapter_id = $2 AND is_finalized = FALSE
		RETURNING id, chapter_id, title, meeting_date, body, created_by, is_finalized, finalized_at, finalized_by, created_at, updated_at
	`, minutesID, chapterID, input.Title, input.Body).
		Scan(&mn.ID, &mn.ChapterID, &mn.Title, &mn.MeetingDate, &mn.Body,
			&mn.CreatedBy, &mn.IsFinalized, &mn.FinalizedAt, &mn.FinalizedBy, &mn.CreatedAt, &mn.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update minutes: %w", err)
	}
	return mn, nil
}

func (s *service) Finalize(ctx context.Context, chapterID, minutesID, memberID string) (*Minutes, error) {
	mn := &Minutes{}
	err := s.pool.QueryRow(ctx, `
		UPDATE meeting_minutes
		SET is_finalized = TRUE, finalized_at = NOW(), finalized_by = $3, updated_at = NOW()
		WHERE id = $1 AND chapter_id = $2 AND is_finalized = FALSE
		RETURNING id, chapter_id, title, meeting_date, body, created_by, is_finalized, finalized_at, finalized_by, created_at, updated_at
	`, minutesID, chapterID, memberID).
		Scan(&mn.ID, &mn.ChapterID, &mn.Title, &mn.MeetingDate, &mn.Body,
			&mn.CreatedBy, &mn.IsFinalized, &mn.FinalizedAt, &mn.FinalizedBy, &mn.CreatedAt, &mn.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finalize minutes: %w", err)
	}
	return mn, nil
}
