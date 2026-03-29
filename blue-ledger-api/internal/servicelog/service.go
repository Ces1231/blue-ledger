package servicelog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ces1231/blue-ledger-api/internal/xp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("service log entry not found")

// Entry is the domain model for a service log record.
type Entry struct {
	ID           string     `json:"id"`
	ChapterID    string     `json:"chapter_id"`
	MemberID     string     `json:"member_id"`
	EventName    string     `json:"event_name"`
	Organization *string    `json:"organization,omitempty"`
	ServiceDate  string     `json:"service_date"`
	Hours        float64    `json:"hours"`
	XPAwarded    int        `json:"xp_awarded"`
	Verified     bool       `json:"verified"`
	VerifiedBy   *string    `json:"verified_by,omitempty"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	// Joined
	MemberFirstName *string `json:"member_first_name,omitempty"`
	MemberLastName  *string `json:"member_last_name,omitempty"`
}

// CreateInput holds data for logging service hours.
type CreateInput struct {
	EventName    string  `json:"event_name" validate:"required,min=2,max=200"`
	Organization *string `json:"organization"`
	ServiceDate  string  `json:"service_date" validate:"required"`
	Hours        float64 `json:"hours" validate:"required,gt=0,lte=24"`
	Notes        *string `json:"notes"`
}

// Service defines the service log business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string, page, perPage int) ([]*Entry, int, error)
	Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*Entry, error)
	Verify(ctx context.Context, chapterID, entryID, verifierMemberID string) (*Entry, error)
	GetByMember(ctx context.Context, chapterID, memberID string, page, perPage int) ([]*Entry, int, error)
}

type service struct {
	pool  *pgxpool.Pool
	xpSvc xp.XPService
}

// NewService creates a new service log service.
func NewService(pool *pgxpool.Pool, xpSvc xp.XPService) Service {
	return &service{pool: pool, xpSvc: xpSvc}
}

func (s *service) List(ctx context.Context, chapterID string, page, perPage int) ([]*Entry, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM service_log WHERE chapter_id = $1`, chapterID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count service log: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			sl.id, sl.chapter_id, sl.member_id, sl.event_name, sl.organization,
			sl.service_date::text, sl.hours, sl.xp_awarded, sl.verified,
			sl.verified_by, sl.verified_at, sl.notes, sl.created_at,
				u.first_name, u.last_name
			FROM service_log sl
			LEFT JOIN members m ON m.id = sl.member_id
			LEFT JOIN users u ON u.id = m.user_id
		WHERE sl.chapter_id = $1
		ORDER BY sl.service_date DESC, sl.created_at DESC
		LIMIT $2 OFFSET $3
	`, chapterID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list service log: %w", err)
	}
	defer rows.Close()

	return scanEntries(rows, total)
}

func (s *service) Create(ctx context.Context, chapterID, memberID string, input CreateInput) (*Entry, error) {
	e := &Entry{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO service_log
			(chapter_id, member_id, event_name, organization, service_date, hours, xp_awarded, verified)
		VALUES ($1, $2, $3, $4, $5, $6, 0, FALSE)
		RETURNING id, chapter_id, member_id, event_name, organization,
			service_date::text, hours, xp_awarded, verified, verified_by, verified_at, notes, created_at
	`, chapterID, memberID, input.EventName, input.Organization, input.ServiceDate, input.Hours).
		Scan(&e.ID, &e.ChapterID, &e.MemberID, &e.EventName, &e.Organization,
			&e.ServiceDate, &e.Hours, &e.XPAwarded, &e.Verified,
			&e.VerifiedBy, &e.VerifiedAt, &e.Notes, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create service log entry: %w", err)
	}
	return e, nil
}

func (s *service) Verify(ctx context.Context, chapterID, entryID, verifierMemberID string) (*Entry, error) {
	e := &Entry{}
	err := s.pool.QueryRow(ctx, `
		UPDATE service_log
		SET verified = TRUE, verified_by = $3, verified_at = NOW()
		WHERE id = $1 AND chapter_id = $2 AND verified = FALSE
		RETURNING id, chapter_id, member_id, event_name, organization,
			service_date::text, hours, xp_awarded, verified, verified_by, verified_at, notes, created_at
	`, entryID, chapterID, verifierMemberID).
		Scan(&e.ID, &e.ChapterID, &e.MemberID, &e.EventName, &e.Organization,
			&e.ServiceDate, &e.Hours, &e.XPAwarded, &e.Verified,
			&e.VerifiedBy, &e.VerifiedAt, &e.Notes, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("verify service log entry: %w", err)
	}

	// Award XP: 10 XP per hour verified
	xpToAward := int(e.Hours * 10)
	if xpToAward > 0 {
		activity := fmt.Sprintf("Service hours verified: %s (%.1f hrs)", e.EventName, e.Hours)
		_ = s.xpSvc.AwardXP(ctx, chapterID, verifierMemberID, xp.AwardXPInput{
			MemberID: e.MemberID,
			XPAmount: xpToAward,
			Activity: activity,
		})

		// Update xp_awarded on the row
		_, _ = s.pool.Exec(ctx,
			`UPDATE service_log SET xp_awarded = $2 WHERE id = $1`, entryID, xpToAward)
		e.XPAwarded = xpToAward
	}

	return e, nil
}

func (s *service) GetByMember(ctx context.Context, chapterID, memberID string, page, perPage int) ([]*Entry, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM service_log WHERE chapter_id = $1 AND member_id = $2`, chapterID, memberID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count member service log: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			sl.id, sl.chapter_id, sl.member_id, sl.event_name, sl.organization,
			sl.service_date::text, sl.hours, sl.xp_awarded, sl.verified,
			sl.verified_by, sl.verified_at, sl.notes, sl.created_at,
				u.first_name, u.last_name
			FROM service_log sl
			LEFT JOIN members m ON m.id = sl.member_id
			LEFT JOIN users u ON u.id = m.user_id
		WHERE sl.chapter_id = $1 AND sl.member_id = $2
		ORDER BY sl.service_date DESC
		LIMIT $3 OFFSET $4
	`, chapterID, memberID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get member service log: %w", err)
	}
	defer rows.Close()

	return scanEntries(rows, total)
}

func scanEntries(rows pgx.Rows, total int) ([]*Entry, int, error) {
	var result []*Entry
	for rows.Next() {
		e := &Entry{}
		if err := rows.Scan(
			&e.ID, &e.ChapterID, &e.MemberID, &e.EventName, &e.Organization,
			&e.ServiceDate, &e.Hours, &e.XPAwarded, &e.Verified,
			&e.VerifiedBy, &e.VerifiedAt, &e.Notes, &e.CreatedAt,
			&e.MemberFirstName, &e.MemberLastName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan service log entry: %w", err)
		}
		result = append(result, e)
	}
	return result, total, rows.Err()
}
