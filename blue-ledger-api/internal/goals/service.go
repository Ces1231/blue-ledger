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
	ID          string     `json:"id"`
	ChapterID   string     `json:"chapter_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Category    *string    `json:"category,omitempty"`
	Target      float64    `json:"target"`
	Current     float64    `json:"current"`
	Unit        *string    `json:"unit,omitempty"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	OwnerID     *string    `json:"owner_id,omitempty"`
	Status      string     `json:"status"`
	XPReward    int        `json:"xp_reward"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateInput holds data for creating a goal.
type CreateInput struct {
	Title       string     `json:"title" validate:"required,min=2,max=200"`
	Description *string    `json:"description"`
	Category    *string    `json:"category"`
	Target      float64    `json:"target" validate:"required,min=0.01"`
	Unit        *string    `json:"unit"`
	Deadline    *time.Time `json:"deadline"`
	OwnerID     *string    `json:"owner_id"`
	XPReward    int        `json:"xp_reward" validate:"min=0"`
}

// UpdateInput holds data for updating a goal.
type UpdateInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Target      *float64   `json:"target"`
	Unit        *string    `json:"unit"`
	Deadline    *time.Time `json:"deadline"`
	Status      *string    `json:"status"`
	XPReward    *int       `json:"xp_reward"`
}

// ProgressInput holds an increment for goal progress.
type ProgressInput struct {
	Increment float64 `json:"increment" validate:"required,min=0.01"`
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
		SELECT id, chapter_id, title, description, category, target, current,
			unit, deadline, owner_id, status, xp_reward, created_at, updated_at
		FROM chapter_goals
		WHERE chapter_id = $1
		ORDER BY status ASC, created_at DESC
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
			&g.Target, &g.Current, &g.Unit, &g.Deadline, &g.OwnerID,
			&g.Status, &g.XPReward, &g.CreatedAt, &g.UpdatedAt,
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
		INSERT INTO chapter_goals (chapter_id, title, description, category, target, current, unit, deadline, owner_id, status, xp_reward)
		VALUES ($1, $2, $3, $4, $5, 0, $6, $7, $8, 'in_progress', $9)
		RETURNING id, chapter_id, title, description, category, target, current,
			unit, deadline, owner_id, status, xp_reward, created_at, updated_at
	`, chapterID, input.Title, input.Description, input.Category, input.Target,
		input.Unit, input.Deadline, input.OwnerID, input.XPReward).
		Scan(&g.ID, &g.ChapterID, &g.Title, &g.Description, &g.Category,
			&g.Target, &g.Current, &g.Unit, &g.Deadline, &g.OwnerID,
			&g.Status, &g.XPReward, &g.CreatedAt, &g.UpdatedAt)
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
			title       = COALESCE($3, title),
			description = COALESCE($4, description),
			target      = COALESCE($5, target),
			unit        = COALESCE($6, unit),
			deadline    = COALESCE($7, deadline),
			status      = COALESCE($8, status),
			xp_reward   = COALESCE($9, xp_reward),
			updated_at  = NOW()
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, title, description, category, target, current,
			unit, deadline, owner_id, status, xp_reward, created_at, updated_at
	`, goalID, chapterID, input.Title, input.Description, input.Target,
		input.Unit, input.Deadline, input.Status, input.XPReward).
		Scan(&g.ID, &g.ChapterID, &g.Title, &g.Description, &g.Category,
			&g.Target, &g.Current, &g.Unit, &g.Deadline, &g.OwnerID,
			&g.Status, &g.XPReward, &g.CreatedAt, &g.UpdatedAt)
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
			current    = LEAST(current + $3, target),
			updated_at = NOW()
		WHERE id = $1 AND chapter_id = $2 AND status = 'in_progress'
		RETURNING id, chapter_id, title, description, category, target, current,
			unit, deadline, owner_id, status, xp_reward, created_at, updated_at
	`, goalID, chapterID, input.Increment).
		Scan(&g.ID, &g.ChapterID, &g.Title, &g.Description, &g.Category,
			&g.Target, &g.Current, &g.Unit, &g.Deadline, &g.OwnerID,
			&g.Status, &g.XPReward, &g.CreatedAt, &g.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update goal progress: %w", err)
	}
	return g, nil
}
