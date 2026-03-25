package committees

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("committee not found")

// Committee is the domain model.
type Committee struct {
	ID              string    `json:"id"`
	ChapterID       string    `json:"chapter_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ChairID         *string   `json:"chair_id,omitempty"`
	MeetingSchedule *string   `json:"meeting_schedule,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// CommitteeMember links a member to a committee.
type CommitteeMember struct {
	CommitteeID string    `json:"committee_id"`
	MemberID    string    `json:"member_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

// CreateInput for creating a committee.
type CreateInput struct {
	Name            string  `json:"name" validate:"required,min=2,max=100"`
	Description     string  `json:"description"`
	ChairID         *string `json:"chair_id"`
	MeetingSchedule *string `json:"meeting_schedule"`
}

// UpdateInput for updating a committee.
type UpdateInput struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	ChairID         *string `json:"chair_id"`
	MeetingSchedule *string `json:"meeting_schedule"`
}

// AddMemberInput for adding a member to a committee.
type AddMemberInput struct {
	MemberID string `json:"member_id" validate:"required"`
	Role     string `json:"role"`
}

// Service defines committee business logic.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Committee, error)
	GetByID(ctx context.Context, chapterID, id string) (*Committee, error)
	Create(ctx context.Context, chapterID string, input CreateInput) (*Committee, error)
	Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Committee, error)
	Delete(ctx context.Context, chapterID, id string) error
	AddMember(ctx context.Context, committeeID string, input AddMemberInput) error
	RemoveMember(ctx context.Context, committeeID, memberID string) error
}

type service struct{ pool *pgxpool.Pool }

// NewService creates a new committees service.
func NewService(pool *pgxpool.Pool) Service { return &service{pool: pool} }

func (s *service) List(ctx context.Context, chapterID string) ([]*Committee, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, name, description, chair_id, meeting_schedule, created_at
		FROM committees WHERE chapter_id = $1 ORDER BY name ASC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list committees: %w", err)
	}
	defer rows.Close()
	var result []*Committee
	for rows.Next() {
		c := &Committee{}
		if err := rows.Scan(&c.ID, &c.ChapterID, &c.Name, &c.Description, &c.ChairID, &c.MeetingSchedule, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan committee: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*Committee, error) {
	c := &Committee{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, name, description, chair_id, meeting_schedule, created_at
		FROM committees WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&c.ID, &c.ChapterID, &c.Name, &c.Description, &c.ChairID, &c.MeetingSchedule, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Committee, error) {
	c := &Committee{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO committees (chapter_id, name, description, chair_id, meeting_schedule)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, chapter_id, name, description, chair_id, meeting_schedule, created_at
	`, chapterID, input.Name, input.Description, input.ChairID, input.MeetingSchedule).
		Scan(&c.ID, &c.ChapterID, &c.Name, &c.Description, &c.ChairID, &c.MeetingSchedule, &c.CreatedAt)
	return c, err
}

func (s *service) Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Committee, error) {
	c := &Committee{}
	err := s.pool.QueryRow(ctx, `
		UPDATE committees
		SET name             = COALESCE($3, name),
		    description      = COALESCE($4, description),
		    chair_id         = COALESCE($5, chair_id),
		    meeting_schedule = COALESCE($6, meeting_schedule)
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, name, description, chair_id, meeting_schedule, created_at
	`, id, chapterID, input.Name, input.Description, input.ChairID, input.MeetingSchedule).
		Scan(&c.ID, &c.ChapterID, &c.Name, &c.Description, &c.ChairID, &c.MeetingSchedule, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *service) Delete(ctx context.Context, chapterID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM committees WHERE id = $1 AND chapter_id = $2`, id, chapterID)
	if err != nil {
		return fmt.Errorf("delete committee: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *service) AddMember(ctx context.Context, committeeID string, input AddMemberInput) error {
	role := input.Role
	if role == "" {
		role = "member"
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO committee_members (committee_id, member_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (committee_id, member_id) DO UPDATE SET role = EXCLUDED.role
	`, committeeID, input.MemberID, role)
	return err
}

func (s *service) RemoveMember(ctx context.Context, committeeID, memberID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM committee_members WHERE committee_id = $1 AND member_id = $2`, committeeID, memberID)
	return err
}
