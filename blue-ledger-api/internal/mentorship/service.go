package mentorship

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ces1231/blue-ledger-api/internal/xp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound        = errors.New("mentorship record not found")
	ErrAlreadyMentor   = errors.New("member is already registered as a mentor")
	ErrAlreadyMatched  = errors.New("member already has an active mentorship match")
)

// Mentor represents a member registered as a mentor.
type Mentor struct {
	ID          string    `json:"id"`
	ChapterID   string    `json:"chapter_id"`
	MemberID    string    `json:"member_id"`
	Bio         *string   `json:"bio,omitempty"`
	Specialties []string  `json:"specialties"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	// Joined
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Role      *string `json:"role,omitempty"`
}

// Match represents a mentorship pairing.
type Match struct {
	ID         string     `json:"id"`
	ChapterID  string     `json:"chapter_id"`
	MentorID   string     `json:"mentor_id"`
	MenteeID   string     `json:"mentee_id"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	// Joined
	MentorFirstName *string `json:"mentor_first_name,omitempty"`
	MentorLastName  *string `json:"mentor_last_name,omitempty"`
	MenteeFirstName *string `json:"mentee_first_name,omitempty"`
	MenteeLastName  *string `json:"mentee_last_name,omitempty"`
}

// BecomeMentorInput holds data for registering as a mentor.
type BecomeMentorInput struct {
	Bio         *string  `json:"bio"`
	Specialties []string `json:"specialties"`
}

// RequestMatchInput holds data for requesting a mentor.
type RequestMatchInput struct {
	MentorID string `json:"mentor_id" validate:"required"`
}

// Service defines the mentorship business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Mentor, error)
	BecomeMentor(ctx context.Context, chapterID, memberID string, input BecomeMentorInput) (*Mentor, error)
	RequestMatch(ctx context.Context, chapterID, menteeID string, input RequestMatchInput) (*Match, error)
	GetMyMatch(ctx context.Context, chapterID, memberID string) (*Match, error)
}

type service struct {
	pool  *pgxpool.Pool
	xpSvc xp.XPService
}

// NewService creates a new mentorship service.
func NewService(pool *pgxpool.Pool, xpSvc xp.XPService) Service {
	return &service{pool: pool, xpSvc: xpSvc}
}

func (s *service) List(ctx context.Context, chapterID string) ([]*Mentor, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			mt.id, mt.chapter_id, mt.member_id, mt.bio, mt.specialties, mt.is_active, mt.created_at,
			m.first_name, m.last_name, m.avatar_url, m.role
		FROM mentors mt
		JOIN members m ON m.id = mt.member_id
		WHERE mt.chapter_id = $1 AND mt.is_active = TRUE
		ORDER BY m.last_name, m.first_name
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list mentors: %w", err)
	}
	defer rows.Close()

	var result []*Mentor
	for rows.Next() {
		mt := &Mentor{}
		if err := rows.Scan(
			&mt.ID, &mt.ChapterID, &mt.MemberID, &mt.Bio, &mt.Specialties, &mt.IsActive, &mt.CreatedAt,
			&mt.FirstName, &mt.LastName, &mt.AvatarURL, &mt.Role,
		); err != nil {
			return nil, fmt.Errorf("scan mentor: %w", err)
		}
		result = append(result, mt)
	}
	return result, rows.Err()
}

func (s *service) BecomeMentor(ctx context.Context, chapterID, memberID string, input BecomeMentorInput) (*Mentor, error) {
	mt := &Mentor{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO mentors (chapter_id, member_id, bio, specialties, is_active)
		VALUES ($1, $2, $3, $4, TRUE)
		ON CONFLICT (chapter_id, member_id) DO UPDATE
			SET bio = EXCLUDED.bio, specialties = EXCLUDED.specialties, is_active = TRUE
		RETURNING id, chapter_id, member_id, bio, specialties, is_active, created_at
	`, chapterID, memberID, input.Bio, input.Specialties).
		Scan(&mt.ID, &mt.ChapterID, &mt.MemberID, &mt.Bio, &mt.Specialties, &mt.IsActive, &mt.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("become mentor: %w", err)
	}
	return mt, nil
}

func (s *service) RequestMatch(ctx context.Context, chapterID, menteeID string, input RequestMatchInput) (*Match, error) {
	// Check for existing active match
	var existingID string
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM mentorship_matches WHERE chapter_id = $1 AND mentee_id = $2 AND status IN ('pending','active')`,
		chapterID, menteeID,
	).Scan(&existingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check existing match: %w", err)
	}
	if existingID != "" {
		return nil, ErrAlreadyMatched
	}

	match := &Match{}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO mentorship_matches (chapter_id, mentor_id, mentee_id, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id, chapter_id, mentor_id, mentee_id, status, started_at, ended_at, created_at
	`, chapterID, input.MentorID, menteeID).
		Scan(&match.ID, &match.ChapterID, &match.MentorID, &match.MenteeID,
			&match.Status, &match.StartedAt, &match.EndedAt, &match.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create mentorship match: %w", err)
	}

	// Award XP to the mentee for finding a mentor
	_ = s.xpSvc.AwardXP(ctx, chapterID, menteeID, xp.AwardXPInput{
		MemberID: menteeID,
		XPAmount: 50,
		Activity: "Mentorship match requested",
	})

	return match, nil
}

func (s *service) GetMyMatch(ctx context.Context, chapterID, memberID string) (*Match, error) {
	match := &Match{}
	err := s.pool.QueryRow(ctx, `
		SELECT
			mm.id, mm.chapter_id, mm.mentor_id, mm.mentee_id,
			mm.status, mm.started_at, mm.ended_at, mm.created_at,
			tor.first_name AS mentor_first_name, tor.last_name AS mentor_last_name,
			tee.first_name AS mentee_first_name, tee.last_name AS mentee_last_name
		FROM mentorship_matches mm
		LEFT JOIN members tor ON tor.id = mm.mentor_id
		LEFT JOIN members tee ON tee.id = mm.mentee_id
		WHERE mm.chapter_id = $1 AND (mm.mentee_id = $2 OR mm.mentor_id = $2)
			AND mm.status IN ('pending', 'active')
		ORDER BY mm.created_at DESC
		LIMIT 1
	`, chapterID, memberID).
		Scan(&match.ID, &match.ChapterID, &match.MentorID, &match.MenteeID,
			&match.Status, &match.StartedAt, &match.EndedAt, &match.CreatedAt,
			&match.MentorFirstName, &match.MentorLastName,
			&match.MenteeFirstName, &match.MenteeLastName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get my match: %w", err)
	}
	return match, nil
}
