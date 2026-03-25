package scholarships

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("scholarship application not found")

// Application is the domain model for a scholarship application.
type Application struct {
	ID          string     `json:"id"`
	ChapterID   string     `json:"chapter_id"`
	MemberID    string     `json:"member_id"`
	Title       string     `json:"title"`
	Amount      int        `json:"amount_cents"`
	Provider    *string    `json:"provider,omitempty"`
	Deadline    *string    `json:"deadline,omitempty"`
	Status      string     `json:"status"`
	Notes       *string    `json:"notes,omitempty"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
	AwardedAt   *time.Time `json:"awarded_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	// Joined
	MemberFirstName *string `json:"member_first_name,omitempty"`
	MemberLastName  *string `json:"member_last_name,omitempty"`
}

// CreateInput holds data for creating a scholarship application.
type CreateInput struct {
	Title    string  `json:"title" validate:"required,min=2,max=200"`
	Amount   int     `json:"amount_cents" validate:"min=0"`
	Provider *string `json:"provider"`
	Deadline *string `json:"deadline"`
	Notes    *string `json:"notes"`
}

// UpdateInput holds data for updating a scholarship application.
type UpdateInput struct {
	Title     *string `json:"title"`
	Amount    *int    `json:"amount_cents"`
	Provider  *string `json:"provider"`
	Deadline  *string `json:"deadline"`
	Status    *string `json:"status"`
	Notes     *string `json:"notes"`
}

// Service defines the scholarships business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string, page, perPage int) ([]*Application, int, error)
	Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*Application, error)
	Get(ctx context.Context, chapterID, appID string) (*Application, error)
	Update(ctx context.Context, chapterID, appID string, input UpdateInput) (*Application, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new scholarships service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string, page, perPage int) ([]*Application, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM scholarship_applications WHERE chapter_id = $1`, chapterID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count scholarship applications: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			sa.id, sa.chapter_id, sa.member_id, sa.title, sa.amount_cents, sa.provider,
			sa.deadline, sa.status, sa.notes, sa.applied_at, sa.awarded_at,
			sa.created_at, sa.updated_at,
			m.first_name, m.last_name
		FROM scholarship_applications sa
		LEFT JOIN members m ON m.id = sa.member_id
		WHERE sa.chapter_id = $1
		ORDER BY sa.created_at DESC
		LIMIT $2 OFFSET $3
	`, chapterID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list scholarship applications: %w", err)
	}
	defer rows.Close()

	var result []*Application
	for rows.Next() {
		a := &Application{}
		if err := rows.Scan(
			&a.ID, &a.ChapterID, &a.MemberID, &a.Title, &a.Amount, &a.Provider,
			&a.Deadline, &a.Status, &a.Notes, &a.AppliedAt, &a.AwardedAt,
			&a.CreatedAt, &a.UpdatedAt,
			&a.MemberFirstName, &a.MemberLastName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan scholarship application: %w", err)
		}
		result = append(result, a)
	}
	return result, total, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*Application, error) {
	a := &Application{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO scholarship_applications
			(chapter_id, member_id, title, amount_cents, provider, deadline, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, 'tracking', $7)
		RETURNING id, chapter_id, member_id, title, amount_cents, provider,
			deadline, status, notes, applied_at, awarded_at, created_at, updated_at
	`, chapterID, memberID, input.Title, input.Amount, input.Provider, input.Deadline, input.Notes).
		Scan(&a.ID, &a.ChapterID, &a.MemberID, &a.Title, &a.Amount, &a.Provider,
			&a.Deadline, &a.Status, &a.Notes, &a.AppliedAt, &a.AwardedAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create scholarship application: %w", err)
	}
	return a, nil
}

func (s *service) Get(ctx context.Context, chapterID, appID string) (*Application, error) {
	a := &Application{}
	err := s.pool.QueryRow(ctx, `
		SELECT
			sa.id, sa.chapter_id, sa.member_id, sa.title, sa.amount_cents, sa.provider,
			sa.deadline, sa.status, sa.notes, sa.applied_at, sa.awarded_at,
			sa.created_at, sa.updated_at,
			m.first_name, m.last_name
		FROM scholarship_applications sa
		LEFT JOIN members m ON m.id = sa.member_id
		WHERE sa.id = $1 AND sa.chapter_id = $2
	`, appID, chapterID).
		Scan(&a.ID, &a.ChapterID, &a.MemberID, &a.Title, &a.Amount, &a.Provider,
			&a.Deadline, &a.Status, &a.Notes, &a.AppliedAt, &a.AwardedAt,
			&a.CreatedAt, &a.UpdatedAt, &a.MemberFirstName, &a.MemberLastName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get scholarship application: %w", err)
	}
	return a, nil
}

func (s *service) Update(ctx context.Context, chapterID, appID string, input UpdateInput) (*Application, error) {
	a := &Application{}
	err := s.pool.QueryRow(ctx, `
		UPDATE scholarship_applications
		SET
			title       = COALESCE($3, title),
			amount_cents = COALESCE($4, amount_cents),
			provider    = COALESCE($5, provider),
			deadline    = COALESCE($6, deadline),
			status      = COALESCE($7, status),
			notes       = COALESCE($8, notes),
			updated_at  = NOW()
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, member_id, title, amount_cents, provider,
			deadline, status, notes, applied_at, awarded_at, created_at, updated_at
	`, appID, chapterID, input.Title, input.Amount, input.Provider, input.Deadline, input.Status, input.Notes).
		Scan(&a.ID, &a.ChapterID, &a.MemberID, &a.Title, &a.Amount, &a.Provider,
			&a.Deadline, &a.Status, &a.Notes, &a.AppliedAt, &a.AwardedAt, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update scholarship application: %w", err)
	}
	return a, nil
}
