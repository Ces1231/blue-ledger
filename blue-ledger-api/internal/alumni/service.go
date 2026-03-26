package alumni

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("alumni record not found")

// Alumnus is a chapter alumni record.
type Alumnus struct {
	ID             string    `json:"id"`
	ChapterID      string    `json:"chapter_id"`
	Name           string    `json:"name"`
	GraduationYear *int      `json:"graduation_year,omitempty"`
	Employer       *string   `json:"employer,omitempty"`
	Title          *string   `json:"title,omitempty"`
	City           *string   `json:"city,omitempty"`
	LinkedInURL    *string   `json:"linkedin_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateInput for creating an alumni record.
type CreateInput struct {
	Name           string  `json:"name" validate:"required,min=2,max=100"`
	GraduationYear *int    `json:"graduation_year"`
	Employer       *string `json:"employer"`
	Title          *string `json:"title"`
	City           *string `json:"city"`
	LinkedInURL    *string `json:"linkedin_url"`
}

// UpdateInput for updating an alumni record.
type UpdateInput struct {
	Name           *string `json:"name"`
	GraduationYear *int    `json:"graduation_year"`
	Employer       *string `json:"employer"`
	Title          *string `json:"title"`
	City           *string `json:"city"`
	LinkedInURL    *string `json:"linkedin_url"`
}

// Service defines alumni business logic.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Alumnus, error)
	GetByID(ctx context.Context, chapterID, id string) (*Alumnus, error)
	Create(ctx context.Context, chapterID string, input CreateInput) (*Alumnus, error)
	Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Alumnus, error)
	Delete(ctx context.Context, chapterID, id string) error
}

type service struct{ pool *pgxpool.Pool }

// NewService creates a new alumni service.
func NewService(pool *pgxpool.Pool) Service { return &service{pool: pool} }

func (s *service) List(ctx context.Context, chapterID string) ([]*Alumnus, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, name, graduation_year, employer, title, city, linkedin_url, created_at
		FROM alumni WHERE chapter_id = $1 ORDER BY graduation_year DESC NULLS LAST, name ASC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list alumni: %w", err)
	}
	defer rows.Close()
	var result []*Alumnus
	for rows.Next() {
		a := &Alumnus{}
		if err := rows.Scan(&a.ID, &a.ChapterID, &a.Name, &a.GraduationYear, &a.Employer, &a.Title, &a.City, &a.LinkedInURL, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan alumnus: %w", err)
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*Alumnus, error) {
	a := &Alumnus{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, name, graduation_year, employer, title, city, linkedin_url, created_at
		FROM alumni WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&a.ID, &a.ChapterID, &a.Name, &a.GraduationYear, &a.Employer, &a.Title, &a.City, &a.LinkedInURL, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Alumnus, error) {
	a := &Alumnus{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO alumni (chapter_id, name, graduation_year, employer, title, city, linkedin_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, chapter_id, name, graduation_year, employer, title, city, linkedin_url, created_at
	`, chapterID, input.Name, input.GraduationYear, input.Employer, input.Title, input.City, input.LinkedInURL).
		Scan(&a.ID, &a.ChapterID, &a.Name, &a.GraduationYear, &a.Employer, &a.Title, &a.City, &a.LinkedInURL, &a.CreatedAt)
	return a, err
}

func (s *service) Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Alumnus, error) {
	a := &Alumnus{}
	err := s.pool.QueryRow(ctx, `
		UPDATE alumni
		SET name            = COALESCE($3, name),
		    graduation_year = COALESCE($4, graduation_year),
		    employer        = COALESCE($5, employer),
		    title           = COALESCE($6, title),
		    city            = COALESCE($7, city),
		    linkedin_url    = COALESCE($8, linkedin_url)
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, name, graduation_year, employer, title, city, linkedin_url, created_at
	`, id, chapterID, input.Name, input.GraduationYear, input.Employer, input.Title, input.City, input.LinkedInURL).
		Scan(&a.ID, &a.ChapterID, &a.Name, &a.GraduationYear, &a.Employer, &a.Title, &a.City, &a.LinkedInURL, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (s *service) Delete(ctx context.Context, chapterID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM alumni WHERE id = $1 AND chapter_id = $2`, id, chapterID)
	if err != nil {
		return fmt.Errorf("delete alumnus: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
