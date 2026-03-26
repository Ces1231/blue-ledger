package goals

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("goal not found")

// Goal is the domain model for a chapter OKR goal.
type Goal struct {
	ID          string    `json:"id"`
	ChapterID   string    `json:"chapter_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Category    string    `json:"category"`
	TargetValue int       `json:"target_value"`
	CurrentValue int      `json:"current_value"`
	XPReward    int       `json:"xp_reward"`
	IsActive    bool      `json:"is_active"`
	DueDate     *string   `json:"due_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateInput holds data for creating a goal.
type CreateInput struct {
	Title       string  `json:"title" validate:"required,min=2,max=200"`
	Description *string `json:"description"`
	Category    string  `json:"category" validate:"required"`
	TargetValue int     `json:"target_value" validate:"required,min=1"`
	XPReward    int     `json:"xp_reward" validate:"min=0"`
	DueDate     *string `json:"due_date"`
}

// UpdateInput holds data for updating a goal.
type UpdateInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	TargetValue *int    `json:"target_value"`
	XPReward    *int    `json:"xp_reward"`
	DueDate     *string `json:"due_date"`
	IsActive    *bool   `json:"is_active"`
}

// ProgressInput holds an increment for goal progress.
type ProgressInput struct {
	Increment int `json:"increment" validate:"required,min=1"`
}

// Service defines the goals business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Goal, error)
	Create(ctx context.Context, chapterID string, input CreateInput) (*Goal, error)
	Update(ctx context.Context, chapterID, goalID string, input UpdateInput) (*Goal, error)
	UpdateProgress(ctx context.Context, chapterID, goalID string, input ProgressInput) (*Goal, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new goals service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string) ([]*Goal, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, title, description, category, target_value, current_value,
			xp_reward, is_active, due_date, created_at, updated_at
		FROM chapter_goals
		WHERE chapter_id = $1
		ORDER BY is_active DESC, created_at DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list goals: %w", err)
	}
	defer rows.Close()

	var result []*Goal
	for rows.Next() {
		g := &Goal{}
		if err := rows.Scan(
			&g.ID, &g.ChapterID, &g.Title, &g.Description, &g.Category,
			&g.TargetValue, &g.CurrentValue, &g.XPReward, &g.IsActive, &g.DueDate,
			&g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan goal: %w", err)
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Goal, error) {
	g := &Goal{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO chapter_goals (chapter_id, title, description, category, target_value, current_value, xp_reward, is_active, due_date)
		VALUES ($1, $2, $3, $4, $5, 0, $6, TRUE, $7)
		RETURNING id, chapter_id, title, description, category, target_value, current_value,
			xp_reward, is_active, due_date, created_at, updated_at
	`, chapterID, input.Title, input.Description, input.Category, input.TargetValue, input.XPReward, input.DueDate).
		Scan(&g.ID, &g.ChapterID, &g.Title, &g.Description, &g.Category,
			&g.TargetValue, &g.CurrentValue, &g.XPReward, &g.IsActive, &g.DueDate,
			&g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create goal: %w", err)
	}
	return g, nil
}

func (s *service) Update(ctx context.Context, chapterID, goalID string, input UpdateInput) (*Goal, error) {
	g := &Goal{}
	err := s.pool.QueryRow(ctx, `
		UPDATE chapter_goals
		SET
			title        = COALESCE($3, title),
			description  = COALESCE($4, description),
			target_value = COALESCE($5, target_value),
			xp_reward    = COALESCE($6, xp_reward),
			due_date     = COALESCE($7, due_date),
			is_active    = COALESCE($8, is_active),
			updated_at   = NOW()
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, title, description, category, target_value, current_value,
			xp_reward, is_active, due_date, created_at, updated_at
	`, goalID, chapterID, input.Title, input.Description, input.TargetValue, input.XPReward, input.DueDate, input.IsActive).
		Scan(&g.ID, &g.ChapterID, &g.Title, &g.Description, &g.Category,
			&g.TargetValue, &g.CurrentValue, &g.XPReward, &g.IsActive, &g.DueDate,
			&g.CreatedAt, &g.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update goal: %w", err)
	}
	return g, nil
}

func (s *service) UpdateProgress(ctx context.Context, chapterID, goalID string, input ProgressInput) (*Goal, error) {
	g := &Goal{}
	err := s.pool.QueryRow(ctx, `
		UPDATE chapter_goals
		SET
			current_value = LEAST(current_value + $3, target_value),
			updated_at    = NOW()
		WHERE id = $1 AND chapter_id = $2 AND is_active = TRUE
		RETURNING id, chapter_id, title, description, category, target_value, current_value,
			xp_reward, is_active, due_date, created_at, updated_at
	`, goalID, chapterID, input.Increment).
		Scan(&g.ID, &g.ChapterID, &g.Title, &g.Description, &g.Category,
			&g.TargetValue, &g.CurrentValue, &g.XPReward, &g.IsActive, &g.DueDate,
			&g.CreatedAt, &g.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update goal progress: %w", err)
	}
	return g, nil
}
