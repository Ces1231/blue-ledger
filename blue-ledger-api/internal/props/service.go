package props

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ces1231/blue-ledger-api/internal/xp"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSelfProps = errors.New("cannot give props to yourself")

// Prop is the domain model for a props record.
type Prop struct {
	ID         string    `json:"id"`
	ChapterID  string    `json:"chapter_id"`
	FromID     string    `json:"from_id"`
	ToID       string    `json:"to_id"`
	Category   string    `json:"category"`
	Message    *string   `json:"message,omitempty"`
	XPAwarded  int       `json:"xp_awarded"`
	CreatedAt  time.Time `json:"created_at"`
	// Joined
	FromFirstName *string `json:"from_first_name,omitempty"`
	FromLastName  *string `json:"from_last_name,omitempty"`
	ToFirstName   *string `json:"to_first_name,omitempty"`
	ToLastName    *string `json:"to_last_name,omitempty"`
}

// GiveInput holds the data for giving props to a member.
type GiveInput struct {
	ToMemberID string  `json:"to_member_id" validate:"required"`
	Category   string  `json:"category" validate:"required,oneof=Leadership Brotherhood Service Academic Professionalism Other"`
	Message    *string `json:"message"`
}

// Service defines the props business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string, page, perPage int) ([]*Prop, int, error)
	Give(ctx context.Context, chapterID, fromMemberID string, input GiveInput) (*Prop, error)
	GetReceived(ctx context.Context, chapterID, memberID string, page, perPage int) ([]*Prop, int, error)
}

type service struct {
	pool  *pgxpool.Pool
	xpSvc xp.XPService
}

// NewService creates a new props service.
func NewService(pool *pgxpool.Pool, xpSvc xp.XPService) Service {
	return &service{pool: pool, xpSvc: xpSvc}
}

func (s *service) List(ctx context.Context, chapterID string, page, perPage int) ([]*Prop, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM props WHERE chapter_id = $1`, chapterID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count props: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			p.id, p.chapter_id, p.from_id, p.to_id, p.category, p.message,
			p.xp_awarded, p.created_at,
				uf.first_name, uf.last_name,
				ut.first_name, ut.last_name
			FROM props p
			LEFT JOIN members f ON f.id = p.from_id
			LEFT JOIN members t ON t.id = p.to_id
			LEFT JOIN users uf ON uf.id = f.user_id
			LEFT JOIN users ut ON ut.id = t.user_id
		WHERE p.chapter_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, chapterID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list props: %w", err)
	}
	defer rows.Close()

	var result []*Prop
	for rows.Next() {
		p := &Prop{}
		if err := rows.Scan(
			&p.ID, &p.ChapterID, &p.FromID, &p.ToID, &p.Category, &p.Message,
			&p.XPAwarded, &p.CreatedAt,
			&p.FromFirstName, &p.FromLastName,
			&p.ToFirstName, &p.ToLastName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan prop: %w", err)
		}
		result = append(result, p)
	}
	return result, total, rows.Err()
}

func (s *service) Give(ctx context.Context, chapterID, fromMemberID string, input GiveInput) (*Prop, error) {
	if fromMemberID == input.ToMemberID {
		return nil, ErrSelfProps
	}

	const xpAmount = 10
	p := &Prop{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO props (chapter_id, from_id, to_id, category, message, xp_awarded)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, chapter_id, from_id, to_id, category, message, xp_awarded, created_at
	`, chapterID, fromMemberID, input.ToMemberID, input.Category, input.Message, xpAmount).
		Scan(&p.ID, &p.ChapterID, &p.FromID, &p.ToID, &p.Category, &p.Message, &p.XPAwarded, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert prop: %w", err)
	}

	// Award XP to the recipient
	activity := fmt.Sprintf("Props received (%s)", input.Category)
	refType := "prop"
	_ = s.xpSvc.AwardXP(ctx, chapterID, fromMemberID, xp.AwardXPInput{
		MemberID:  input.ToMemberID,
		XPAmount:  xpAmount,
		Activity:  activity,
		Note:      input.Message,
		Semester:  nil,
	})
	_ = refType

	return p, nil
}

func (s *service) GetReceived(ctx context.Context, chapterID, memberID string, page, perPage int) ([]*Prop, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM props WHERE chapter_id = $1 AND to_id = $2`, chapterID, memberID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count received props: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			p.id, p.chapter_id, p.from_id, p.to_id, p.category, p.message,
			p.xp_awarded, p.created_at,
				uf.first_name, uf.last_name,
				ut.first_name, ut.last_name
			FROM props p
			LEFT JOIN members f ON f.id = p.from_id
			LEFT JOIN members t ON t.id = p.to_id
			LEFT JOIN users uf ON uf.id = f.user_id
			LEFT JOIN users ut ON ut.id = t.user_id
		WHERE p.chapter_id = $1 AND p.to_id = $2
		ORDER BY p.created_at DESC
		LIMIT $3 OFFSET $4
	`, chapterID, memberID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get received props: %w", err)
	}
	defer rows.Close()

	var result []*Prop
	for rows.Next() {
		p := &Prop{}
		if err := rows.Scan(
			&p.ID, &p.ChapterID, &p.FromID, &p.ToID, &p.Category, &p.Message,
			&p.XPAwarded, &p.CreatedAt,
			&p.FromFirstName, &p.FromLastName,
			&p.ToFirstName, &p.ToLastName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan received prop: %w", err)
		}
		result = append(result, p)
	}
	return result, total, rows.Err()
}
