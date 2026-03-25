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
	ID           string         `json:"id"`
	ChapterID    string         `json:"chapter_id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Requirements map[string]any `json:"requirements"`
	XPReward     int            `json:"xp_reward"`
	BadgeID      *string        `json:"badge_id,omitempty"`
	Active       bool           `json:"active"`
	CreatedAt    time.Time      `json:"created_at"`
}

// QuestProgress tracks a member's progress on a quest.
type QuestProgress struct {
	ID          string     `json:"id"`
	MemberID    string     `json:"member_id"`
	QuestID     string     `json:"quest_id"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// CreateInput holds data for creating a quest.
type CreateInput struct {
	Name         string         `json:"name" validate:"required,min=2,max=100"`
	Description  string         `json:"description" validate:"required"`
	Requirements map[string]any `json:"requirements"`
	XPReward     int            `json:"xp_reward"`
	BadgeID      *string        `json:"badge_id"`
	Active       bool           `json:"active"`
}

// UpdateInput holds data for updating a quest.
type UpdateInput struct {
	Name         *string        `json:"name"`
	Description  *string        `json:"description"`
	Requirements map[string]any `json:"requirements"`
	XPReward     *int           `json:"xp_reward"`
	Active       *bool          `json:"active"`
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
		SELECT id, chapter_id, name, description, requirements, xp_reward, badge_id, active, created_at
		FROM quests WHERE chapter_id = $1 ORDER BY name ASC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list quests: %w", err)
	}
	defer rows.Close()
	var result []*Quest
	for rows.Next() {
		q := &Quest{}
		if err := rows.Scan(&q.ID, &q.ChapterID, &q.Name, &q.Description, &q.Requirements, &q.XPReward, &q.BadgeID, &q.Active, &q.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan quest: %w", err)
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*Quest, error) {
	q := &Quest{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, name, description, requirements, xp_reward, badge_id, active, created_at
		FROM quests WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&q.ID, &q.ChapterID, &q.Name, &q.Description, &q.Requirements, &q.XPReward, &q.BadgeID, &q.Active, &q.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return q, err
}

func (s *service) GetProgress(ctx context.Context, memberID, questID string) (*QuestProgress, error) {
	p := &QuestProgress{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, member_id, quest_id, status, completed_at
		FROM quest_progress WHERE member_id = $1 AND quest_id = $2
	`, memberID, questID).Scan(&p.ID, &p.MemberID, &p.QuestID, &p.Status, &p.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &QuestProgress{MemberID: memberID, QuestID: questID, Status: "not_started"}, nil
	}
	return p, err
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Quest, error) {
	q := &Quest{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO quests (chapter_id, name, description, requirements, xp_reward, badge_id, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, chapter_id, name, description, requirements, xp_reward, badge_id, active, created_at
	`, chapterID, input.Name, input.Description, input.Requirements, input.XPReward, input.BadgeID, input.Active).
		Scan(&q.ID, &q.ChapterID, &q.Name, &q.Description, &q.Requirements, &q.XPReward, &q.BadgeID, &q.Active, &q.CreatedAt)
	return q, err
}

func (s *service) Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Quest, error) {
	q := &Quest{}
	err := s.pool.QueryRow(ctx, `
		UPDATE quests
		SET name        = COALESCE($3, name),
		    description = COALESCE($4, description),
		    xp_reward   = COALESCE($5, xp_reward),
		    active      = COALESCE($6, active)
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, name, description, requirements, xp_reward, badge_id, active, created_at
	`, id, chapterID, input.Name, input.Description, input.XPReward, input.Active).
		Scan(&q.ID, &q.ChapterID, &q.Name, &q.Description, &q.Requirements, &q.XPReward, &q.BadgeID, &q.Active, &q.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return q, err
}
