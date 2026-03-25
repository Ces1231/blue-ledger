package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ChapterSummary is a lightweight chapter row for the sysadmin overview.
type ChapterSummary struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	GreekLetters       string     `json:"greek_letters"`
	City               string     `json:"city"`
	StateCode          string     `json:"state_code"`
	University         *string    `json:"university,omitempty"`
	MemberCount        int        `json:"member_count"`
	SubscriptionStatus string     `json:"subscription_status"`
	PlanTier           string     `json:"plan_tier"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// UserSummary is a row in the sysadmin user management view.
type UserSummary struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	IsSysadmin   bool      `json:"is_sysadmin"`
	ChapterCount int       `json:"chapter_count"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// AuditEntry is a single entry in the audit log.
type AuditEntry struct {
	ID         string    `json:"id"`
	ActorID    *string   `json:"actor_id,omitempty"`
	ActorEmail *string   `json:"actor_email,omitempty"`
	Action     string    `json:"action"`
	EntityType *string   `json:"entity_type,omitempty"`
	EntityID   *string   `json:"entity_id,omitempty"`
	IPAddress  *string   `json:"ip_address,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateChapterInput holds data for creating a new chapter via sysadmin.
type CreateChapterInput struct {
	Name         string `json:"name" validate:"required,min=3"`
	GreekLetters string `json:"greek_letters" validate:"required"`
	City         string `json:"city" validate:"required"`
	StateCode    string `json:"state_code" validate:"required,len=2"`
	University   string `json:"university"`
}

// UpdateSubscriptionInput holds subscription update data.
type UpdateSubscriptionInput struct {
	SubscriptionStatus string  `json:"subscription_status" validate:"required,oneof=trialing active past_due canceled paused"`
	PlanTier           string  `json:"plan_tier" validate:"required,oneof=starter growth chapter_pro"`
}

// SysService defines the sysadmin platform operations interface.
type SysService interface {
	ListChapters(ctx context.Context, page, perPage int) ([]*ChapterSummary, int, error)
	CreateChapter(ctx context.Context, input CreateChapterInput) (*ChapterSummary, error)
	GetChapter(ctx context.Context, chapterID string) (*ChapterSummary, error)
	UpdateSubscription(ctx context.Context, chapterID string, input UpdateSubscriptionInput) (*ChapterSummary, error)
	DeleteChapter(ctx context.Context, chapterID string) error
	ListUsers(ctx context.Context, page, perPage int) ([]*UserSummary, int, error)
	GetAuditLog(ctx context.Context, page, perPage int) ([]*AuditEntry, int, error)
	RecalculateAllXP(ctx context.Context) (int, error)
}

type sysService struct {
	pool *pgxpool.Pool
}

// NewSysService creates a new platform sysadmin service.
func NewSysService(pool *pgxpool.Pool) SysService {
	return &sysService{pool: pool}
}

func (s *sysService) ListChapters(ctx context.Context, page, perPage int) ([]*ChapterSummary, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM chapters WHERE deleted_at IS NULL`,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count chapters: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			c.id, c.name, c.greek_letters, c.city, c.state_code, c.university,
			COUNT(m.id) AS member_count,
			c.subscription_status, c.plan_tier, c.trial_ends_at, c.created_at
		FROM chapters c
		LEFT JOIN members m ON m.chapter_id = c.id AND m.status = 'active'
		WHERE c.deleted_at IS NULL
		GROUP BY c.id
		ORDER BY c.created_at DESC
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()

	var result []*ChapterSummary
	for rows.Next() {
		ch := &ChapterSummary{}
		if err := rows.Scan(
			&ch.ID, &ch.Name, &ch.GreekLetters, &ch.City, &ch.StateCode, &ch.University,
			&ch.MemberCount, &ch.SubscriptionStatus, &ch.PlanTier, &ch.TrialEndsAt, &ch.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan chapter: %w", err)
		}
		result = append(result, ch)
	}
	return result, total, rows.Err()
}

func (s *sysService) CreateChapter(ctx context.Context, input CreateChapterInput) (*ChapterSummary, error) {
	ch := &ChapterSummary{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO chapters (name, greek_letters, city, state_code, university, subscription_status, plan_tier)
		VALUES ($1, $2, $3, $4, $5, 'trialing', 'starter')
		RETURNING id, name, greek_letters, city, state_code, university,
			0, subscription_status, plan_tier, trial_ends_at, created_at
	`, input.Name, input.GreekLetters, input.City, input.StateCode, input.University).
		Scan(&ch.ID, &ch.Name, &ch.GreekLetters, &ch.City, &ch.StateCode, &ch.University,
			&ch.MemberCount, &ch.SubscriptionStatus, &ch.PlanTier, &ch.TrialEndsAt, &ch.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create chapter: %w", err)
	}
	return ch, nil
}

func (s *sysService) GetChapter(ctx context.Context, chapterID string) (*ChapterSummary, error) {
	ch := &ChapterSummary{}
	err := s.pool.QueryRow(ctx, `
		SELECT
			c.id, c.name, c.greek_letters, c.city, c.state_code, c.university,
			COUNT(m.id) AS member_count,
			c.subscription_status, c.plan_tier, c.trial_ends_at, c.created_at
		FROM chapters c
		LEFT JOIN members m ON m.chapter_id = c.id AND m.status = 'active'
		WHERE c.id = $1 AND c.deleted_at IS NULL
		GROUP BY c.id
	`, chapterID).
		Scan(&ch.ID, &ch.Name, &ch.GreekLetters, &ch.City, &ch.StateCode, &ch.University,
			&ch.MemberCount, &ch.SubscriptionStatus, &ch.PlanTier, &ch.TrialEndsAt, &ch.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get chapter: %w", err)
	}
	return ch, nil
}

func (s *sysService) UpdateSubscription(ctx context.Context, chapterID string, input UpdateSubscriptionInput) (*ChapterSummary, error) {
	ch := &ChapterSummary{}
	err := s.pool.QueryRow(ctx, `
		UPDATE chapters
		SET subscription_status = $2, plan_tier = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, greek_letters, city, state_code, university,
			0, subscription_status, plan_tier, trial_ends_at, created_at
	`, chapterID, input.SubscriptionStatus, input.PlanTier).
		Scan(&ch.ID, &ch.Name, &ch.GreekLetters, &ch.City, &ch.StateCode, &ch.University,
			&ch.MemberCount, &ch.SubscriptionStatus, &ch.PlanTier, &ch.TrialEndsAt, &ch.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("update chapter subscription: %w", err)
	}
	return ch, nil
}

func (s *sysService) DeleteChapter(ctx context.Context, chapterID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE chapters SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		chapterID,
	)
	if err != nil {
		return fmt.Errorf("delete chapter: %w", err)
	}
	return nil
}

func (s *sysService) ListUsers(ctx context.Context, page, perPage int) ([]*UserSummary, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			u.id, u.email, u.first_name, u.last_name, u.is_sysadmin,
			COUNT(DISTINCT m.chapter_id) AS chapter_count,
			MAX(s.created_at) AS last_login_at,
			u.created_at
		FROM users u
		LEFT JOIN members m ON m.user_id = u.id
		LEFT JOIN auth_sessions s ON s.user_id = u.id
		GROUP BY u.id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var result []*UserSummary
	for rows.Next() {
		u := &UserSummary{}
		if err := rows.Scan(
			&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.IsSysadmin,
			&u.ChapterCount, &u.LastLoginAt, &u.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, u)
	}
	return result, total, rows.Err()
}

func (s *sysService) GetAuditLog(ctx context.Context, page, perPage int) ([]*AuditEntry, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_log`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit log: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT al.id, al.actor_id, u.email, al.action, al.entity_type, al.entity_id, al.ip_address, al.created_at
		FROM audit_log al
		LEFT JOIN users u ON u.id = al.actor_id
		ORDER BY al.created_at DESC
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get audit log: %w", err)
	}
	defer rows.Close()

	var result []*AuditEntry
	for rows.Next() {
		e := &AuditEntry{}
		if err := rows.Scan(
			&e.ID, &e.ActorID, &e.ActorEmail, &e.Action, &e.EntityType, &e.EntityID, &e.IPAddress, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit entry: %w", err)
		}
		result = append(result, e)
	}
	return result, total, rows.Err()
}

func (s *sysService) RecalculateAllXP(ctx context.Context) (int, error) {
	// Recalculate xp_total for every member from the engagement_log source of truth
	tag, err := s.pool.Exec(ctx, `
		UPDATE members m
		SET xp_total = COALESCE((
			SELECT SUM(xp_awarded) FROM engagement_log WHERE member_id = m.id
		), 0)
	`)
	if err != nil {
		return 0, fmt.Errorf("recalculate all XP: %w", err)
	}
	return int(tag.RowsAffected()), nil
}
