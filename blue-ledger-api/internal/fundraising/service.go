package fundraising

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("campaign not found")

// Campaign is a fundraising campaign.
type Campaign struct {
	ID          string     `json:"id"`
	ChapterID   string     `json:"chapter_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	GoalCents   int        `json:"goal_cents"`
	RaisedCents int        `json:"raised_cents"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Contribution is a member's donation to a campaign.
type Contribution struct {
	ID          string    `json:"id"`
	CampaignID  string    `json:"campaign_id"`
	MemberID    string    `json:"member_id"`
	AmountCents int       `json:"amount_cents"`
	Note        *string   `json:"note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateInput for creating a campaign.
type CreateInput struct {
	Title       string     `json:"title" validate:"required,min=2,max=200"`
	Description string     `json:"description"`
	GoalCents   int        `json:"goal_cents"`
	Deadline    *time.Time `json:"deadline"`
	Active      bool       `json:"active"`
}

// UpdateInput for updating a campaign.
type UpdateInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	GoalCents   *int       `json:"goal_cents"`
	Deadline    *time.Time `json:"deadline"`
	Active      *bool      `json:"active"`
}

// ContributeInput for posting a contribution.
type ContributeInput struct {
	AmountCents int     `json:"amount_cents" validate:"required,min=1"`
	Note        *string `json:"note"`
}

// Service defines fundraising business logic.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Campaign, error)
	GetByID(ctx context.Context, chapterID, id string) (*Campaign, error)
	Create(ctx context.Context, chapterID string, input CreateInput) (*Campaign, error)
	Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Campaign, error)
	Delete(ctx context.Context, chapterID, id string) error
	Contribute(ctx context.Context, campaignID, memberID string, input ContributeInput) (*Contribution, error)
}

type service struct{ pool *pgxpool.Pool }

// NewService creates a new fundraising service.
func NewService(pool *pgxpool.Pool) Service { return &service{pool: pool} }

func (s *service) List(ctx context.Context, chapterID string) ([]*Campaign, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, title, description, goal_cents, raised_cents, deadline, active, created_at
		FROM fundraising_campaigns WHERE chapter_id = $1 ORDER BY created_at DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}
	defer rows.Close()
	var result []*Campaign
	for rows.Next() {
		c := &Campaign{}
		if err := rows.Scan(&c.ID, &c.ChapterID, &c.Title, &c.Description, &c.GoalCents, &c.RaisedCents, &c.Deadline, &c.Active, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan campaign: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (s *service) GetByID(ctx context.Context, chapterID, id string) (*Campaign, error) {
	c := &Campaign{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, title, description, goal_cents, raised_cents, deadline, active, created_at
		FROM fundraising_campaigns WHERE id = $1 AND chapter_id = $2
	`, id, chapterID).Scan(&c.ID, &c.ChapterID, &c.Title, &c.Description, &c.GoalCents, &c.RaisedCents, &c.Deadline, &c.Active, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Campaign, error) {
	c := &Campaign{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO fundraising_campaigns (chapter_id, title, description, goal_cents, deadline, active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, chapter_id, title, description, goal_cents, raised_cents, deadline, active, created_at
	`, chapterID, input.Title, input.Description, input.GoalCents, input.Deadline, input.Active).
		Scan(&c.ID, &c.ChapterID, &c.Title, &c.Description, &c.GoalCents, &c.RaisedCents, &c.Deadline, &c.Active, &c.CreatedAt)
	return c, err
}

func (s *service) Update(ctx context.Context, chapterID, id string, input UpdateInput) (*Campaign, error) {
	c := &Campaign{}
	err := s.pool.QueryRow(ctx, `
		UPDATE fundraising_campaigns
		SET title       = COALESCE($3, title),
		    description = COALESCE($4, description),
		    goal_cents  = COALESCE($5, goal_cents),
		    active      = COALESCE($6, active)
		WHERE id = $1 AND chapter_id = $2
		RETURNING id, chapter_id, title, description, goal_cents, raised_cents, deadline, active, created_at
	`, id, chapterID, input.Title, input.Description, input.GoalCents, input.Active).
		Scan(&c.ID, &c.ChapterID, &c.Title, &c.Description, &c.GoalCents, &c.RaisedCents, &c.Deadline, &c.Active, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *service) Delete(ctx context.Context, chapterID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM fundraising_campaigns WHERE id = $1 AND chapter_id = $2`, id, chapterID)
	if err != nil {
		return fmt.Errorf("delete campaign: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *service) Contribute(ctx context.Context, campaignID, memberID string, input ContributeInput) (*Contribution, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	contrib := &Contribution{}
	if err := tx.QueryRow(ctx, `
		INSERT INTO fundraising_contributions (campaign_id, member_id, amount_cents, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, campaign_id, member_id, amount_cents, note, created_at
	`, campaignID, memberID, input.AmountCents, input.Note).
		Scan(&contrib.ID, &contrib.CampaignID, &contrib.MemberID, &contrib.AmountCents, &contrib.Note, &contrib.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert contribution: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE fundraising_campaigns SET raised_cents = raised_cents + $2 WHERE id = $1
	`, campaignID, input.AmountCents); err != nil {
		return nil, fmt.Errorf("update raised amount: %w", err)
	}

	return contrib, tx.Commit(ctx)
}
