package scholarships

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("scholarship not found")

// Application is the domain model for a scholarship record.
type Application struct {
	ID            string     `json:"id"`
	ChapterID     string     `json:"chapter_id"`
	MemberID      *string    `json:"member_id,omitempty"`
	ApplicantName string     `json:"applicant_name"`
	School        *string    `json:"school,omitempty"`
	Gpa           *string    `json:"gpa,omitempty"`
	Major         *string    `json:"major,omitempty"`
	AcademicYear  *int       `json:"academic_year,omitempty"`
	City          *string    `json:"city,omitempty"`
	Semester      *string    `json:"semester,omitempty"`
	Essay         *string    `json:"essay,omitempty"`
	SubmittedAt   *string    `json:"submitted_at,omitempty"`
	AmountCents   *int       `json:"amount_cents,omitempty"`
	Status        string     `json:"status"`
	Notes         *string    `json:"notes,omitempty"`
	ReviewerID    *string    `json:"reviewer_id,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	// Joined
	MemberFirstName *string `json:"member_first_name,omitempty"`
	MemberLastName  *string `json:"member_last_name,omitempty"`
}

// CreateInput holds data for creating a scholarship record.
type CreateInput struct {
	ApplicantName string  `json:"applicant_name" validate:"required,min=2,max=200"`
	School        *string `json:"school"`
	Gpa           *string `json:"gpa"`
	Major         *string `json:"major"`
	AcademicYear  *int    `json:"academic_year"`
	City          *string `json:"city"`
	Semester      *string `json:"semester"`
	Essay         *string `json:"essay"`
	SubmittedAt   *string `json:"submitted_at"`
	AmountCents   *int    `json:"amount_cents"`
	Notes         *string `json:"notes"`
}

// UpdateInput holds data for updating a scholarship record.
type UpdateInput struct {
	ApplicantName *string `json:"applicant_name"`
	School        *string `json:"school"`
	Gpa           *string `json:"gpa"`
	Major         *string `json:"major"`
	AmountCents   *int    `json:"amount_cents"`
	Status        *string `json:"status"`
	Notes         *string `json:"notes"`
	ReviewerID    *string `json:"reviewer_id"`
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
		`SELECT COUNT(*) FROM scholarships WHERE chapter_id = $1`, chapterID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count scholarships: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			sc.id, sc.chapter_id, sc.member_id, sc.applicant_name,
			sc.school, sc.gpa, sc.major, sc.academic_year, sc.city,
			sc.semester, sc.essay, sc.submitted_at::text,
			sc.amount_cents, sc.status, sc.notes,
			sc.reviewer_id, sc.reviewed_at,
			sc.created_at, sc.updated_at,
				u.first_name, u.last_name
			FROM scholarships sc
			LEFT JOIN members m ON m.id = sc.member_id
			LEFT JOIN users u ON u.id = m.user_id
		WHERE sc.chapter_id = $1
		ORDER BY sc.created_at DESC
		LIMIT $2 OFFSET $3
	`, chapterID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list scholarships: %w", err)
	}
	defer rows.Close()

	var result []*Application
	for rows.Next() {
		a := &Application{}
		if err := rows.Scan(
			&a.ID, &a.ChapterID, &a.MemberID, &a.ApplicantName,
			&a.School, &a.Gpa, &a.Major, &a.AcademicYear, &a.City,
			&a.Semester, &a.Essay, &a.SubmittedAt,
			&a.AmountCents, &a.Status, &a.Notes,
			&a.ReviewerID, &a.ReviewedAt,
			&a.CreatedAt, &a.UpdatedAt,
			&a.MemberFirstName, &a.MemberLastName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan scholarship: %w", err)
		}
		result = append(result, a)
	}
	return result, total, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*Application, error) {
	a := &Application{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO scholarships
			(chapter_id, member_id, applicant_name, school, gpa, major, academic_year,
			 city, semester, essay, submitted_at, amount_cents, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'submitted', $13)
		RETURNING id, chapter_id, member_id, applicant_name,
			school, gpa, major, academic_year, city, semester, essay, submitted_at::text,
			amount_cents, status, notes, reviewer_id, reviewed_at, created_at, updated_at
	`, chapterID, memberID, input.ApplicantName, input.School, input.Gpa, input.Major,
		input.AcademicYear, input.City, input.Semester, input.Essay, input.SubmittedAt,
		input.AmountCents, input.Notes).
		Scan(&a.ID, &a.ChapterID, &a.MemberID, &a.ApplicantName,
			&a.School, &a.Gpa, &a.Major, &a.AcademicYear, &a.City,
			&a.Semester, &a.Essay, &a.SubmittedAt,
			&a.AmountCents, &a.Status, &a.Notes, &a.ReviewerID, &a.ReviewedAt,
			&a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create scholarship: %w", err)
	}
	return a, nil
}

func (s *service) Get(ctx context.Context, chapterID, appID string) (*Application, error) {
	a := &Application{}
	err := s.pool.QueryRow(ctx, `
		SELECT
			sc.id, sc.chapter_id, sc.member_id, sc.applicant_name,
			sc.school, sc.gpa, sc.major, sc.academic_year, sc.city,
			sc.semester, sc.essay, sc.submitted_at::text,
			sc.amount_cents, sc.status, sc.notes,
			sc.reviewer_id, sc.reviewed_at,
			sc.created_at, sc.updated_at,
				u.first_name, u.last_name
			FROM scholarships sc
			LEFT JOIN members m ON m.id = sc.member_id
			LEFT JOIN users u ON u.id = m.user_id
		WHERE sc.id = $1 AND sc.chapter_id = $2
	`, appID, chapterID).
		Scan(&a.ID, &a.ChapterID, &a.MemberID, &a.ApplicantName,
			&a.School, &a.Gpa, &a.Major, &a.AcademicYear, &a.City,
			&a.Semester, &a.Essay, &a.SubmittedAt,
			&a.AmountCents, &a.Status, &a.Notes, &a.ReviewerID, &a.ReviewedAt,
			&a.CreatedAt, &a.UpdatedAt, &a.MemberFirstName, &a.MemberLastName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get scholarship: %w", err)
	}
	return a, nil
}

func (s *service) Update(ctx context.Context, chapterID, appID string, input UpdateInput) (*Application, error) {
	a := &Application{}
	err := s.pool.QueryRow(ctx, `
		UPDATE scholarships
		SET
			applicant_name = COALESCE($3, applicant_name),
			school         = COALESCE($4, school),
			gpa            = COALESCE($5, gpa),
			major          = COALESCE($6, major),
			amount_cents   = COALESCE($7, amount_cents),
			status         = COALESCE($8, status),
			notes          = COALESCE($9, notes),
			reviewer_id    = COALESCE($10, reviewer_id),
			updated_at     = NOW()
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, member_id, applicant_name,
			school, gpa, major, academic_year, city, semester, essay, submitted_at::text,
			amount_cents, status, notes, reviewer_id, reviewed_at, created_at, updated_at
	`, appID, chapterID, input.ApplicantName, input.School, input.Gpa, input.Major,
		input.AmountCents, input.Status, input.Notes, input.ReviewerID).
		Scan(&a.ID, &a.ChapterID, &a.MemberID, &a.ApplicantName,
			&a.School, &a.Gpa, &a.Major, &a.AcademicYear, &a.City,
			&a.Semester, &a.Essay, &a.SubmittedAt,
			&a.AmountCents, &a.Status, &a.Notes, &a.ReviewerID, &a.ReviewedAt,
			&a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update scholarship: %w", err)
	}
	return a, nil
}
