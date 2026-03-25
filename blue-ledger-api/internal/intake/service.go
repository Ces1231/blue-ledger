package intake

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("intake prospect not found")

// Prospect is the domain model for an intake/prospect record.
type Prospect struct {
	ID          string     `json:"id"`
	ChapterID   string     `json:"chapter_id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       *string    `json:"email,omitempty"`
	Phone       *string    `json:"phone,omitempty"`
	University  *string    `json:"university,omitempty"`
	GradYear    *int       `json:"grad_year,omitempty"`
	Stage       string     `json:"stage"`
	Notes       *string    `json:"notes,omitempty"`
	AddedBy     string     `json:"added_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
}

// CreateInput holds data for adding a new prospect.
type CreateInput struct {
	FirstName  string  `json:"first_name" validate:"required,min=1,max=100"`
	LastName   string  `json:"last_name" validate:"required,min=1,max=100"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	University *string `json:"university"`
	GradYear   *int    `json:"grad_year"`
	Notes      *string `json:"notes"`
}

// UpdateInput holds data for updating a prospect.
type UpdateInput struct {
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	University *string `json:"university"`
	GradYear   *int    `json:"grad_year"`
	Notes      *string `json:"notes"`
}

// Service defines the intake business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Prospect, error)
	Create(ctx context.Context, chapterID, addedByMemberID string, input CreateInput) (*Prospect, error)
	Update(ctx context.Context, chapterID, prospectID string, input UpdateInput) (*Prospect, error)
	UpdateStage(ctx context.Context, chapterID, prospectID, stage string) (*Prospect, error)
	Archive(ctx context.Context, chapterID, prospectID string) error
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new intake service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string) ([]*Prospect, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, first_name, last_name, email, phone, university,
			grad_year, stage, notes, added_by, created_at, updated_at, archived_at
		FROM intake_prospects
		WHERE chapter_id = $1 AND archived_at IS NULL
		ORDER BY stage, created_at DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list prospects: %w", err)
	}
	defer rows.Close()

	var result []*Prospect
	for rows.Next() {
		p := &Prospect{}
		if err := rows.Scan(
			&p.ID, &p.ChapterID, &p.FirstName, &p.LastName, &p.Email, &p.Phone,
			&p.University, &p.GradYear, &p.Stage, &p.Notes, &p.AddedBy,
			&p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt,
		); err != nil {
			return nil, fmt.Errorf("scan prospect: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID, addedByMemberID string, input CreateInput) (*Prospect, error) {
	p := &Prospect{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO intake_prospects
			(chapter_id, first_name, last_name, email, phone, university, grad_year, stage, notes, added_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'prospect', $8, $9)
		RETURNING id, chapter_id, first_name, last_name, email, phone, university,
			grad_year, stage, notes, added_by, created_at, updated_at, archived_at
	`, chapterID, input.FirstName, input.LastName, input.Email, input.Phone,
		input.University, input.GradYear, input.Notes, addedByMemberID).
		Scan(&p.ID, &p.ChapterID, &p.FirstName, &p.LastName, &p.Email, &p.Phone,
			&p.University, &p.GradYear, &p.Stage, &p.Notes, &p.AddedBy,
			&p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt)
	if err != nil {
		return nil, fmt.Errorf("create prospect: %w", err)
	}
	return p, nil
}

func (s *service) Update(ctx context.Context, chapterID, prospectID string, input UpdateInput) (*Prospect, error) {
	p := &Prospect{}
	err := s.pool.QueryRow(ctx, `
		UPDATE intake_prospects
		SET
			first_name  = COALESCE($3, first_name),
			last_name   = COALESCE($4, last_name),
			email       = COALESCE($5, email),
			phone       = COALESCE($6, phone),
			university  = COALESCE($7, university),
			grad_year   = COALESCE($8, grad_year),
			notes       = COALESCE($9, notes),
			updated_at  = NOW()
		WHERE id = $1 AND chapter_id = $2 AND archived_at IS NULL
		RETURNING id, chapter_id, first_name, last_name, email, phone, university,
			grad_year, stage, notes, added_by, created_at, updated_at, archived_at
	`, prospectID, chapterID,
		input.FirstName, input.LastName, input.Email, input.Phone,
		input.University, input.GradYear, input.Notes).
		Scan(&p.ID, &p.ChapterID, &p.FirstName, &p.LastName, &p.Email, &p.Phone,
			&p.University, &p.GradYear, &p.Stage, &p.Notes, &p.AddedBy,
			&p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update prospect: %w", err)
	}
	return p, nil
}

func (s *service) UpdateStage(ctx context.Context, chapterID, prospectID, stage string) (*Prospect, error) {
	p := &Prospect{}
	err := s.pool.QueryRow(ctx, `
		UPDATE intake_prospects
		SET stage = $3, updated_at = NOW()
		WHERE id = $1 AND chapter_id = $2 AND archived_at IS NULL
		RETURNING id, chapter_id, first_name, last_name, email, phone, university,
			grad_year, stage, notes, added_by, created_at, updated_at, archived_at
	`, prospectID, chapterID, stage).
		Scan(&p.ID, &p.ChapterID, &p.FirstName, &p.LastName, &p.Email, &p.Phone,
			&p.University, &p.GradYear, &p.Stage, &p.Notes, &p.AddedBy,
			&p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update prospect stage: %w", err)
	}
	return p, nil
}

func (s *service) Archive(ctx context.Context, chapterID, prospectID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE intake_prospects SET archived_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND chapter_id = $2 AND archived_at IS NULL
	`, prospectID, chapterID)
	if err != nil {
		return fmt.Errorf("archive prospect: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
