package quests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("quest not found")

// Quest is the domain model for a chapter quest.
type Quest struct {
	ID            string           `json:"id"`
	ChapterID     string           `json:"chapter_id"`
	Title         string           `json:"title"`
	Description   *string          `json:"description,omitempty"`
	XPReward      int              `json:"xp_reward"`
	BadgeRewardID *string          `json:"badge_reward_id,omitempty"`
	Steps         []map[string]any `json:"steps"`
	IsActive      bool             `json:"is_active"`
	CreatedAt     time.Time        `json:"created_at"`
}

// QuestProgress tracks a member's progress on a quest.
type QuestProgress struct {
	ID          string         `json:"id,omitempty"`
	ChapterID   string         `json:"chapter_id,omitempty"`
	MemberID    string         `json:"member_id"`
	QuestID     string         `json:"quest_id"`
	Progress    map[string]any `json:"progress"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// CreateInput holds data for creating a quest.
type CreateInput struct {
	Title         string           `json:"title" validate:"required,min=2,max=100"`
	Description   *string          `json:"description"`
	XPReward      int              `json:"xp_reward"`
	BadgeRewardID *string          `json:"badge_reward_id"`
	Steps         []map[string]any `json:"steps"`
	IsActive      bool             `json:"is_active"`
}

// UpdateInput holds data for updating a quest.
type UpdateInput struct {
	Title       *string          `json:"title"`
	Description *string          `json:"description"`
	XPReward    *int             `json:"xp_reward"`
	Steps       []map[string]any `json:"steps"`
	IsActive    *bool            `json:"is_active"`
}

// Service defines quest business logic.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Quest, error)
	GetByID(ctx context.Context, chapterID, id string) (*Quest, error)
	GetProgress(ctx context.Context, memberID, questID string) (*QuestProgress, error)
	Create(ctx context.Context, chapterID string, input CreateInput) (*Quest, error)
	Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Quest, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new quests service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string) ([]*Quest, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, title, description, xp_reward, badge_reward_id, steps, is_active, created_at
		FROM quests WHERE chapter_id = $1 ORDER BY title ASC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list quests: %w", err)
	}
	defer rows.Close()
	var result []*Quest
	for rows.Next() {
		q := &Quest{}
		if err := rows.Scan(&q.ID, &q.ChapterID, &q.Title, &q.Description, &q.XPReward,
			&q.BadgeRewardID, &q.Steps, &q.IsActive, &q.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan quest: %w", err)
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*Quest, error) {
	q := &Quest{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, title, description, xp_reward, badge_reward_id, steps, is_active, created_at
		FROM quests WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&q.ID, &q.ChapterID, &q.Title, &q.Description, &q.XPReward,
		&q.BadgeRewardID, &q.Steps, &q.IsActive, &q.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return q, err
}

func (s *service) GetProgress(ctx context.Context, memberID, questID string) (*QuestProgress, error) {
	p := &QuestProgress{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, member_id, quest_id, progress, completed_at
		FROM member_quest_progress WHERE member_id = $1 AND quest_id = $2
	`, memberID, questID).Scan(&p.ID, &p.ChapterID, &p.MemberID, &p.QuestID, &p.Progress, &p.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &QuestProgress{MemberID: memberID, QuestID: questID, Progress: map[string]any{}}, nil
	}
	return p, err
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Quest, error) {
	if input.Steps == nil {
		input.Steps = []map[string]any{}
	}
	q := &Quest{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO quests (chapter_id, title, description, xp_reward, badge_reward_id, steps, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, chapter_id, title, description, xp_reward, badge_reward_id, steps, is_active, created_at
	`, chapterID, input.Title, input.Description, input.XPReward, input.BadgeRewardID,
		input.Steps, input.IsActive).
		Scan(&q.ID, &q.ChapterID, &q.Title, &q.Description, &q.XPReward,
			&q.BadgeRewardID, &q.Steps, &q.IsActive, &q.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create quest: %w", err)
	}
	return q, nil
}

func (s *service) Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Quest, error) {
	q := &Quest{}
	err := s.pool.QueryRow(ctx, `
		UPDATE quests
		SET title       = COALESCE($3, title),
		    description = COALESCE($4, description),
		    xp_reward   = COALESCE($5, xp_reward),
		    is_active   = COALESCE($6, is_active)
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, title, description, xp_reward, badge_reward_id, steps, is_active, created_at
	`, id, chapterID, input.Title, input.Description, input.XPReward, input.IsActive).
		Scan(&q.ID, &q.ChapterID, &q.Title, &q.Description, &q.XPReward,
			&q.BadgeRewardID, &q.Steps, &q.IsActive, &q.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update quest: %w", err)
	}
	return q, nil
}
