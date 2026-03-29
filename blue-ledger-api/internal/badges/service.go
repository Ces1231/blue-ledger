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
	Icon        *string        `json:"icon,omitempty"`
	Category    *string        `json:"category,omitempty"`
	Description *string        `json:"description,omitempty"`
	Requirement *string        `json:"requirement,omitempty"`
	XPReward    int            `json:"xp_reward"`
	Rarity      string         `json:"rarity"`
	IsActive    bool           `json:"is_active"`
	Criteria    map[string]any `json:"criteria"`
	CreatedAt   time.Time      `json:"created_at"`
}

// MemberBadge records a badge earned by a member.
type MemberBadge struct {
	ID        string    `json:"id"`
	ChapterID string    `json:"chapter_id"`
	MemberID  string    `json:"member_id"`
	BadgeID   string    `json:"badge_id"`
	AwardedAt time.Time `json:"awarded_at"`
	AwardedBy *string   `json:"awarded_by,omitempty"`
}

// CreateInput holds the data for creating a badge.
type CreateInput struct {
	Name        string         `json:"name" validate:"required,min=2,max=100"`
	Icon        *string        `json:"icon"`
	Category    *string        `json:"category"`
	Description *string        `json:"description"`
	Requirement *string        `json:"requirement"`
	XPReward    int            `json:"xp_reward"`
	Rarity      string         `json:"rarity"`
	Criteria    map[string]any `json:"criteria"`
}

// UpdateInput holds the data for updating a badge.
type UpdateInput struct {
	Name        *string        `json:"name"`
	Icon        *string        `json:"icon"`
	Category    *string        `json:"category"`
	Description *string        `json:"description"`
	Requirement *string        `json:"requirement"`
	XPReward    *int           `json:"xp_reward"`
	Rarity      *string        `json:"rarity"`
	Criteria    map[string]any `json:"criteria"`
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
	Mine(ctx context.Context, chapterID, memberID string) ([]*MemberBadge, error)
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
		SELECT id, chapter_id, name, icon, category, description, requirement,
		       xp_reward, rarity, is_active, criteria, created_at
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
		if err := rows.Scan(&b.ID, &b.ChapterID, &b.Name, &b.Icon, &b.Category, &b.Description,
			&b.Requirement, &b.XPReward, &b.Rarity, &b.IsActive, &b.Criteria, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan badge: %w", err)
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*Badge, error) {
	b := &Badge{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, name, icon, category, description, requirement,
		       xp_reward, rarity, is_active, criteria, created_at
		FROM badges
		WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&b.ID, &b.ChapterID, &b.Name, &b.Icon, &b.Category, &b.Description,
		&b.Requirement, &b.XPReward, &b.Rarity, &b.IsActive, &b.Criteria, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get badge: %w", err)
	}
	return b, nil
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Badge, error) {
	if input.Rarity == "" {
		input.Rarity = "common"
	}
	if input.Criteria == nil {
		input.Criteria = map[string]any{}
	}
	b := &Badge{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO badges (chapter_id, name, icon, category, description, requirement, xp_reward, rarity, criteria)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, chapter_id, name, icon, category, description, requirement,
		          xp_reward, rarity, is_active, criteria, created_at
	`, chapterID, input.Name, input.Icon, input.Category, input.Description,
		input.Requirement, input.XPReward, input.Rarity, input.Criteria).
		Scan(&b.ID, &b.ChapterID, &b.Name, &b.Icon, &b.Category, &b.Description,
			&b.Requirement, &b.XPReward, &b.Rarity, &b.IsActive, &b.Criteria, &b.CreatedAt)
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
			icon        = COALESCE($4, icon),
			category    = COALESCE($5, category),
			description = COALESCE($6, description),
			requirement = COALESCE($7, requirement),
			xp_reward   = COALESCE($8, xp_reward),
			rarity      = COALESCE($9, rarity)
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, name, icon, category, description, requirement,
		          xp_reward, rarity, is_active, criteria, created_at
	`, id, chapterID, input.Name, input.Icon, input.Category, input.Description,
		input.Requirement, input.XPReward, input.Rarity).
		Scan(&b.ID, &b.ChapterID, &b.Name, &b.Icon, &b.Category, &b.Description,
			&b.Requirement, &b.XPReward, &b.Rarity, &b.IsActive, &b.Criteria, &b.CreatedAt)
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

func (s *service) Mine(ctx context.Context, chapterID, memberID string) ([]*MemberBadge, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, member_id, badge_id, awarded_at, awarded_by
		FROM member_badges
		WHERE chapter_id = $1 AND member_id = $2
		ORDER BY awarded_at DESC
	`, chapterID, memberID)
	if err != nil {
		return nil, fmt.Errorf("list my badges: %w", err)
	}
	defer rows.Close()

	var result []*MemberBadge
	for rows.Next() {
		mb := &MemberBadge{}
		if err := rows.Scan(&mb.ID, &mb.ChapterID, &mb.MemberID, &mb.BadgeID, &mb.AwardedAt, &mb.AwardedBy); err != nil {
			return nil, fmt.Errorf("scan member badge: %w", err)
		}
		result = append(result, mb)
	}
	if result == nil {
		result = []*MemberBadge{}
	}
	return result, rows.Err()
}

func (s *service) Award(ctx context.Context, chapterID, badgeID, memberID, awardedBy string) (*MemberBadge, error) {
	mb := &MemberBadge{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO member_badges (chapter_id, member_id, badge_id, awarded_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, chapter_id, member_id, badge_id, awarded_at, awarded_by
	`, chapterID, memberID, badgeID, awardedBy).
		Scan(&mb.ID, &mb.ChapterID, &mb.MemberID, &mb.BadgeID, &mb.AwardedAt, &mb.AwardedBy)
	if err != nil {
		return nil, fmt.Errorf("award badge: %w", err)
	}
	return mb, nil
}
