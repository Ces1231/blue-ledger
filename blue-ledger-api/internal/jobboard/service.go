package jobboard

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("job posting not found")

// JobPosting is a job listing posted by a member.
type JobPosting struct {
	ID          string    `json:"id"`
	ChapterID   string    `json:"chapter_id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    *string   `json:"location,omitempty"`
	Description string    `json:"description"`
	URL         *string   `json:"url,omitempty"`
	PostedBy    string    `json:"posted_by"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateInput for creating a job posting.
type CreateInput struct {
	Title       string  `json:"title" validate:"required,min=2,max=200"`
	Company     string  `json:"company" validate:"required"`
	Location    *string `json:"location"`
	Description string  `json:"description" validate:"required"`
	URL         *string `json:"url"`
}

// UpdateInput for updating a job posting.
type UpdateInput struct {
	Title       *string `json:"title"`
	Company     *string `json:"company"`
	Location    *string `json:"location"`
	Description *string `json:"description"`
	URL         *string `json:"url"`
	Active      *bool   `json:"active"`
}

// Service defines job board business logic.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*JobPosting, error)
	GetByID(ctx context.Context, chapterID, id string) (*JobPosting, error)
	Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*JobPosting, error)
	Update(ctx context.Context, chapterID, id, memberID string, input UpdateInput) (*JobPosting, error)
	Delete(ctx context.Context, chapterID, id, memberID, role string) error
}

type service struct{ pool *pgxpool.Pool }

// NewService creates a new job board service.
func NewService(pool *pgxpool.Pool) Service { return &service{pool: pool} }

func (s *service) List(ctx context.Context, chapterID string) ([]*JobPosting, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, title, company, location, description, url, posted_by, active, created_at
		FROM job_postings WHERE chapter_id = $1 AND active = TRUE ORDER BY created_at DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()
	var result []*JobPosting
	for rows.Next() {
		j := &JobPosting{}
		if err := rows.Scan(&j.ID, &j.ChapterID, &j.Title, &j.Company, &j.Location, &j.Description, &j.URL, &j.PostedBy, &j.Active, &j.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		result = append(result, j)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*JobPosting, error) {
	j := &JobPosting{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, title, company, location, description, url, posted_by, active, created_at
		FROM job_postings WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&j.ID, &j.ChapterID, &j.Title, &j.Company, &j.Location, &j.Description, &j.URL, &j.PostedBy, &j.Active, &j.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return j, err
}

func (s *service) Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*JobPosting, error) {
	j := &JobPosting{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO job_postings (chapter_id, title, company, location, description, url, posted_by, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
		RETURNING id, chapter_id, title, company, location, description, url, posted_by, active, created_at
	`, chapterID, input.Title, input.Company, input.Location, input.Description, input.URL, memberID).
		Scan(&j.ID, &j.ChapterID, &j.Title, &j.Company, &j.Location, &j.Description, &j.URL, &j.PostedBy, &j.Active, &j.CreatedAt)
	return j, err
}

func (s *service) Update(ctx context.Context, chapterID, id, memberID string, input UpdateInput) (*JobPosting, error) {
	j := &JobPosting{}
	err := s.pool.QueryRow(ctx, `
		UPDATE job_postings
		SET title       = COALESCE($4, title),
		    company     = COALESCE($5, company),
		    description = COALESCE($6, description),
		    active      = COALESCE($7, active)
		WHERE id = $1 AND chapter_id = $2 AND (posted_by = $3 OR $3 IS NOT NULL)
		RETURNING id, chapter_id, title, company, location, description, url, posted_by, active, created_at
	`, id, chapterID, memberID, input.Title, input.Company, input.Description, input.Active).
		Scan(&j.ID, &j.ChapterID, &j.Title, &j.Company, &j.Location, &j.Description, &j.URL, &j.PostedBy, &j.Active, &j.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return j, err
}

func (s *service) Delete(ctx context.Context, chapterID, id, memberID, role string) error {
	var query string
	var args []any
	if role == "admin" {
		query = `DELETE FROM job_postings WHERE id = $1 AND chapter_id = $2`
		args = []any{id, chapterID}
	} else {
		query = `DELETE FROM job_postings WHERE id = $1 AND chapter_id = $2 AND posted_by = $3`
		args = []any{id, chapterID, memberID}
	}
	tag, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
