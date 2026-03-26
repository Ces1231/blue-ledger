package badges

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("badge not found")

// Badge is the domain model for a chapter badge definition.
type Badge struct {
	ID          string         `json:"id"`
	ChapterID   string         `json:"chapter_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Criteria    map[string]any `json:"criteria"`
	XPBonus     int            `json:"xp_bonus"`
	ImageURL    *string        `json:"image_url,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// MemberBadge records a badge earned by a member.
type MemberBadge struct {
	ID         string    `json:"id"`
	MemberID   string    `json:"member_id"`
	BadgeID    string    `json:"badge_id"`
	EarnedAt   time.Time `json:"earned_at"`
	AwardedBy  *string   `json:"awarded_by,omitempty"`
	BadgeName  *string   `json:"badge_name,omitempty"`
}

// CreateInput holds the data for creating a badge.
type CreateInput struct {
	Name        string         `json:"name" validate:"required,min=2,max=100"`
	Description string         `json:"description" validate:"required"`
	Criteria    map[string]any `json:"criteria"`
	XPBonus     int            `json:"xp_bonus"`
	ImageURL    *string        `json:"image_url"`
}

// UpdateInput holds the data for updating a badge.
type UpdateInput struct {
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	Criteria    map[string]any `json:"criteria"`
	XPBonus     *int           `json:"xp_bonus"`
	ImageURL    *string        `json:"image_url"`
}

// AwardInput holds the data for manually awarding a badge.
type AwardInput struct {
	MemberID string `json:"member_id" validate:"required"`
}

// Service defines the badges business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Badge, error)
	GetByID(ctx context.Context, chapterID, id string) (*Badge, error)
	Create(ctx context.Context, chapterID string, input CreateInput) (*Badge, error)
	Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Badge, error)
	Delete(ctx context.Context, chapterID, id string) error
	Award(ctx context.Context, chapterID, badgeID, memberID, awardedBy string) (*MemberBadge, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new badges service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string) ([]*Badge, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, name, description, criteria, xp_bonus, image_url, created_at
		FROM badges
		WHERE chapter_id = $1
		ORDER BY name ASC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list badges: %w", err)
	}
	defer rows.Close()

	var result []*Badge
	for rows.Next() {
		b := &Badge{}
		if err := rows.Scan(&b.ID, &b.ChapterID, &b.Name, &b.Description, &b.Criteria, &b.XPBonus, &b.ImageURL, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan badge: %w", err)
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*Badge, error) {
	b := &Badge{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, name, description, criteria, xp_bonus, image_url, created_at
		FROM badges
		WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&b.ID, &b.ChapterID, &b.Name, &b.Description, &b.Criteria, &b.XPBonus, &b.ImageURL, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get badge: %w", err)
	}
	return b, nil
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Badge, error) {
	b := &Badge{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO badges (chapter_id, name, description, criteria, xp_bonus, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, chapter_id, name, description, criteria, xp_bonus, image_url, created_at
	`, chapterID, input.Name, input.Description, input.Criteria, input.XPBonus, input.ImageURL).
		Scan(&b.ID, &b.ChapterID, &b.Name, &b.Description, &b.Criteria, &b.XPBonus, &b.ImageURL, &b.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create badge: %w", err)
	}
	return b, nil
}

func (s *service) Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Badge, error) {
	b := &Badge{}
	err := s.pool.QueryRow(ctx, `
		UPDATE badges
		SET
			name        = COALESCE($3, name),
			description = COALESCE($4, description),
			xp_bonus    = COALESCE($5, xp_bonus),
			image_url   = COALESCE($6, image_url)
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, name, description, criteria, xp_bonus, image_url, created_at
	`, id, chapterID, input.Name, input.Description, input.XPBonus, input.ImageURL).
		Scan(&b.ID, &b.ChapterID, &b.Name, &b.Description, &b.Criteria, &b.XPBonus, &b.ImageURL, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update badge: %w", err)
	}
	return b, nil
}

func (s *service) Delete(ctx context.Context, chapterID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM badges WHERE id = $1 AND chapter_id = $2
	`, id, chapterID)
	if err != nil {
		return fmt.Errorf("delete badge: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *service) Award(ctx context.Context, chapterID, badgeID, memberID, awardedBy string) (*MemberBadge, error) {
	mb := &MemberBadge{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO member_badges (member_id, badge_id, awarded_by)
		VALUES ($1, $2, $3)
		RETURNING id, member_id, badge_id, earned_at, awarded_by
	`, memberID, badgeID, awardedBy).
		Scan(&mb.ID, &mb.MemberID, &mb.BadgeID, &mb.EarnedAt, &mb.AwardedBy)
	if err != nil {
		return nil, fmt.Errorf("award badge: %w", err)
	}
	return mb, nil
}
