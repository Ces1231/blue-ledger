package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("resource not found")

// Resource is a chapter document/link resource.
type Resource struct {
	ID          string    `json:"id"`
	ChapterID   string    `json:"chapter_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Type        string    `json:"type"`
	URL         *string   `json:"url,omitempty"`
	FileKey     *string   `json:"file_key,omitempty"`
	Category    *string   `json:"category,omitempty"`
	Tags        []string  `json:"tags"`
	PostedBy    string    `json:"posted_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateInput for creating a resource.
type CreateInput struct {
	Title       string   `json:"title" validate:"required,min=2,max=200"`
	Description *string  `json:"description"`
	Type        string   `json:"type" validate:"required,oneof=link pdf video image document other"`
	URL         *string  `json:"url"`
	FileKey     *string  `json:"file_key"`
	Category    *string  `json:"category"`
	Tags        []string `json:"tags"`
}

// UpdateInput for updating a resource.
type UpdateInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	URL         *string `json:"url"`
	Category    *string `json:"category"`
}

// Service defines resource business logic.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Resource, error)
	GetByID(ctx context.Context, chapterID, id string) (*Resource, error)
	Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*Resource, error)
	Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Resource, error)
	Delete(ctx context.Context, chapterID, id string) error
}

type service struct{ pool *pgxpool.Pool }

// NewService creates a new resources service.
func NewService(pool *pgxpool.Pool) Service { return &service{pool: pool} }

func (s *service) List(ctx context.Context, chapterID string) ([]*Resource, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, title, description, type, url, file_key, category, tags, posted_by, created_at
		FROM resources WHERE chapter_id = $1 ORDER BY title ASC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}
	defer rows.Close()
	var result []*Resource
	for rows.Next() {
		r := &Resource{}
		if err := rows.Scan(&r.ID, &r.ChapterID, &r.Title, &r.Description, &r.Type,
			&r.URL, &r.FileKey, &r.Category, &r.Tags, &r.PostedBy, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan resource: %w", err)
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*Resource, error) {
	r := &Resource{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, title, description, type, url, file_key, category, tags, posted_by, created_at
		FROM resources WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&r.ID, &r.ChapterID, &r.Title, &r.Description, &r.Type,
		&r.URL, &r.FileKey, &r.Category, &r.Tags, &r.PostedBy, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return r, err
}

func (s *service) Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*Resource, error) {
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}
	r := &Resource{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO resources (chapter_id, title, description, type, url, file_key, category, tags, posted_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, chapter_id, title, description, type, url, file_key, category, tags, posted_by, created_at
	`, chapterID, input.Title, input.Description, input.Type, input.URL, input.FileKey, input.Category, tags, memberID).
		Scan(&r.ID, &r.ChapterID, &r.Title, &r.Description, &r.Type,
			&r.URL, &r.FileKey, &r.Category, &r.Tags, &r.PostedBy, &r.CreatedAt)
	return r, err
}

func (s *service) Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Resource, error) {
	r := &Resource{}
	err := s.pool.QueryRow(ctx, `
		UPDATE resources
		SET title       = COALESCE($3, title),
		    description = COALESCE($4, description),
		    url         = COALESCE($5, url),
		    category    = COALESCE($6, category)
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, title, description, type, url, file_key, category, tags, posted_by, created_at
	`, id, chapterID, input.Title, input.Description, input.URL, input.Category).
		Scan(&r.ID, &r.ChapterID, &r.Title, &r.Description, &r.Type,
			&r.URL, &r.FileKey, &r.Category, &r.Tags, &r.PostedBy, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return r, err
}

func (s *service) Delete(ctx context.Context, chapterID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM resources WHERE id = $1 AND chapter_id = $2`, id, chapterID)
	if err != nil {
		return fmt.Errorf("delete resource: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
