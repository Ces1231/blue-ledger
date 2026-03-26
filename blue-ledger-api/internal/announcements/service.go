package announcements

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("announcement not found")

// Announcement is the domain model for a chapter announcement.
type Announcement struct {
	ID        string     `json:"id"`
	ChapterID string     `json:"chapter_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Category  *string    `json:"category,omitempty"`
	PostedBy  string     `json:"posted_by"`
	IsPinned  bool       `json:"is_pinned"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	// Joined fields
	PosterFirstName *string `json:"poster_first_name,omitempty"`
	PosterLastName  *string `json:"poster_last_name,omitempty"`
}

// CreateInput holds the data for creating an announcement.
type CreateInput struct {
	Title    string  `json:"title" validate:"required,min=2,max=200"`
	Body     string  `json:"body" validate:"required"`
	Category *string `json:"category"`
	IsPinned bool    `json:"is_pinned"`
}

// UpdateInput holds the data for updating an announcement.
type UpdateInput struct {
	Title    *string `json:"title"`
	Body     *string `json:"body"`
	Category *string `json:"category"`
	IsPinned *bool   `json:"is_pinned"`
}

// Service defines the announcements business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Announcement, error)
	Create(ctx context.Context, chapterID, posterMemberID string, input CreateInput) (*Announcement, error)
	Update(ctx context.Context, chapterID, announcementID string, input UpdateInput) (*Announcement, error)
	Delete(ctx context.Context, chapterID, announcementID string) error
	Pin(ctx context.Context, chapterID, announcementID string, pinned bool) (*Announcement, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new announcements service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string) ([]*Announcement, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			a.id, a.chapter_id, a.title, a.body, a.category, a.posted_by,
			a.is_pinned, a.is_active, a.created_at, a.updated_at,
			m.first_name AS poster_first_name, m.last_name AS poster_last_name
		FROM announcements a
		LEFT JOIN members m ON m.id = a.posted_by
		WHERE a.chapter_id = $1 AND a.is_active = TRUE
		ORDER BY a.is_pinned DESC, a.created_at DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}
	defer rows.Close()

	var result []*Announcement
	for rows.Next() {
		a := &Announcement{}
		if err := rows.Scan(
			&a.ID, &a.ChapterID, &a.Title, &a.Body, &a.Category,
			&a.PostedBy, &a.IsPinned, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
			&a.PosterFirstName, &a.PosterLastName,
		); err != nil {
			return nil, fmt.Errorf("scan announcement: %w", err)
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID, posterMemberID string, input CreateInput) (*Announcement, error) {
	a := &Announcement{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO announcements (chapter_id, title, body, category, posted_by, is_pinned, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE)
		RETURNING id, chapter_id, title, body, category, posted_by, is_pinned, is_active, created_at, updated_at
	`, chapterID, input.Title, input.Body, input.Category, posterMemberID, input.IsPinned).
		Scan(&a.ID, &a.ChapterID, &a.Title, &a.Body, &a.Category, &a.PostedBy,
			&a.IsPinned, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create announcement: %w", err)
	}
	return a, nil
}

func (s *service) Update(ctx context.Context, chapterID, announcementID string, input UpdateInput) (*Announcement, error) {
	a := &Announcement{}
	err := s.pool.QueryRow(ctx, `
		UPDATE announcements
		SET
			title     = COALESCE($3, title),
			body      = COALESCE($4, body),
			category  = COALESCE($5, category),
			is_pinned = COALESCE($6, is_pinned),
			updated_at = NOW()
		WHERE id = $1 AND chapter_id = $2 AND is_active = TRUE
		RETURNING id, chapter_id, title, body, category, posted_by, is_pinned, is_active, created_at, updated_at
	`, announcementID, chapterID, input.Title, input.Body, input.Category, input.IsPinned).
		Scan(&a.ID, &a.ChapterID, &a.Title, &a.Body, &a.Category, &a.PostedBy,
			&a.IsPinned, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update announcement: %w", err)
	}
	return a, nil
}

func (s *service) Delete(ctx context.Context, chapterID, announcementID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE announcements SET is_active = FALSE, updated_at = NOW()
		WHERE id = $1 AND chapter_id = $2
	`, announcementID, chapterID)
	if err != nil {
		return fmt.Errorf("delete announcement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *service) Pin(ctx context.Context, chapterID, announcementID string, pinned bool) (*Announcement, error) {
	a := &Announcement{}
	err := s.pool.QueryRow(ctx, `
		UPDATE announcements SET is_pinned = $3, updated_at = NOW()
		WHERE id = $1 AND chapter_id = $2 AND is_active = TRUE
		RETURNING id, chapter_id, title, body, category, posted_by, is_pinned, is_active, created_at, updated_at
	`, announcementID, chapterID, pinned).
		Scan(&a.ID, &a.ChapterID, &a.Title, &a.Body, &a.Category, &a.PostedBy,
			&a.IsPinned, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("pin announcement: %w", err)
	}
	return a, nil
}
