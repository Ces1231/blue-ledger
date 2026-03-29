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
	ID          string     `json:"id"`
	ChapterID   string     `json:"chapter_id"`
	PostedBy    string     `json:"posted_by"`
	Company     string     `json:"company"`
	JobTitle    string     `json:"job_title"`
	JobType     *string    `json:"job_type,omitempty"`
	Location    *string    `json:"location,omitempty"`
	Link        *string    `json:"link,omitempty"`
	Description *string    `json:"description,omitempty"`
	Fields      []string   `json:"fields"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateInput for creating a job posting.
type CreateInput struct {
	Company     string     `json:"company" validate:"required"`
	JobTitle    string     `json:"job_title" validate:"required,min=2,max=200"`
	JobType     *string    `json:"job_type"`
	Location    *string    `json:"location"`
	Link        *string    `json:"link"`
	Description *string    `json:"description"`
	Fields      []string   `json:"fields"`
	Deadline    *time.Time `json:"deadline"`
}

// UpdateInput for updating a job posting.
type UpdateInput struct {
	Company     *string `json:"company"`
	JobTitle    *string `json:"job_title"`
	JobType     *string `json:"job_type"`
	Location    *string `json:"location"`
	Link        *string `json:"link"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
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
		SELECT id, chapter_id, posted_by, company, job_title, job_type, location, link, description, fields, deadline, is_active, created_at
		FROM job_board WHERE chapter_id = $1 AND is_active = TRUE ORDER BY created_at DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()
	var result []*JobPosting
	for rows.Next() {
		j := &JobPosting{}
		if err := rows.Scan(&j.ID, &j.ChapterID, &j.PostedBy, &j.Company, &j.JobTitle, &j.JobType,
			&j.Location, &j.Link, &j.Description, &j.Fields, &j.Deadline, &j.IsActive, &j.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		result = append(result, j)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*JobPosting, error) {
	j := &JobPosting{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, posted_by, company, job_title, job_type, location, link, description, fields, deadline, is_active, created_at
		FROM job_board WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&j.ID, &j.ChapterID, &j.PostedBy, &j.Company, &j.JobTitle, &j.JobType,
		&j.Location, &j.Link, &j.Description, &j.Fields, &j.Deadline, &j.IsActive, &j.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return j, err
}

func (s *service) Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*JobPosting, error) {
	fields := input.Fields
	if fields == nil {
		fields = []string{}
	}
	j := &JobPosting{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO job_board (chapter_id, posted_by, company, job_title, job_type, location, link, description, fields, deadline, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE)
		RETURNING id, chapter_id, posted_by, company, job_title, job_type, location, link, description, fields, deadline, is_active, created_at
	`, chapterID, memberID, input.Company, input.JobTitle, input.JobType, input.Location, input.Link, input.Description, fields, input.Deadline).
		Scan(&j.ID, &j.ChapterID, &j.PostedBy, &j.Company, &j.JobTitle, &j.JobType,
			&j.Location, &j.Link, &j.Description, &j.Fields, &j.Deadline, &j.IsActive, &j.CreatedAt)
	return j, err
}

func (s *service) Update(ctx context.Context, chapterID, id, memberID string, input UpdateInput) (*JobPosting, error) {
	j := &JobPosting{}
	err := s.pool.QueryRow(ctx, `
		UPDATE job_board
		SET company     = COALESCE($4, company),
		    job_title   = COALESCE($5, job_title),
		    job_type    = COALESCE($6, job_type),
		    description = COALESCE($7, description),
		    is_active   = COALESCE($8, is_active),
		    updated_at  = NOW()
		WHERE id = $1 AND chapter_id = $2 AND (posted_by = $3 OR $3 IS NOT NULL)
		RETURNING id, chapter_id, posted_by, company, job_title, job_type, location, link, description, fields, deadline, is_active, created_at
	`, id, chapterID, memberID, input.Company, input.JobTitle, input.JobType, input.Description, input.IsActive).
		Scan(&j.ID, &j.ChapterID, &j.PostedBy, &j.Company, &j.JobTitle, &j.JobType,
			&j.Location, &j.Link, &j.Description, &j.Fields, &j.Deadline, &j.IsActive, &j.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return j, err
}

func (s *service) Delete(ctx context.Context, chapterID, id, memberID, role string) error {
	var query string
	var args []any
	if role == "admin" {
		query = `DELETE FROM job_board WHERE id = $1 AND chapter_id = $2`
		args = []any{id, chapterID}
	} else {
		query = `DELETE FROM job_board WHERE id = $1 AND chapter_id = $2 AND posted_by = $3`
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
