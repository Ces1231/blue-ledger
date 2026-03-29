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
	ErrNotFound      = errors.New("mentorship record not found")
	ErrAlreadyMatched = errors.New("member already has an active mentorship match")
)

// Mentor represents a member registered as a mentor.
type Mentor struct {
	ID         string    `json:"id"`
	ChapterID  string    `json:"chapter_id"`
	MemberID   string    `json:"member_id"`
	FocusAreas []string  `json:"specialties"`
	Bio        *string   `json:"bio,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	// Joined
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Role      *string `json:"role,omitempty"`
}

// Match represents a mentorship pairing.
type Match struct {
	ID        string     `json:"id"`
	ChapterID string     `json:"chapter_id"`
	MentorID  string     `json:"mentor_id"`
	MenteeID  *string    `json:"mentee_id,omitempty"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
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
		SELECT DISTINCT ON (mt.mentor_id)
			mt.id, mt.chapter_id, mt.mentor_id, mt.focus_areas, mt.bio, mt.is_active, mt.created_at,
				u.first_name, u.last_name, u.avatar_url, m.role
			FROM mentorships mt
			JOIN members m ON m.id = mt.mentor_id
			LEFT JOIN users u ON u.id = m.user_id
		WHERE mt.chapter_id = $1 AND mt.is_active = TRUE
		ORDER BY mt.mentor_id, mt.created_at DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list mentors: %w", err)
	}
	defer rows.Close()

	var result []*Mentor
	for rows.Next() {
		mt := &Mentor{}
		if err := rows.Scan(
			&mt.ID, &mt.ChapterID, &mt.MemberID, &mt.FocusAreas, &mt.Bio, &mt.IsActive, &mt.CreatedAt,
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
		INSERT INTO mentorships (chapter_id, mentor_id, focus_areas, bio, is_active)
		VALUES ($1, $2, $3, $4, TRUE)
		RETURNING id, chapter_id, mentor_id, focus_areas, bio, is_active, created_at
	`, chapterID, memberID, input.Specialties, input.Bio).
		Scan(&mt.ID, &mt.ChapterID, &mt.MemberID, &mt.FocusAreas, &mt.Bio, &mt.IsActive, &mt.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("become mentor: %w", err)
	}
	return mt, nil
}

func (s *service) RequestMatch(ctx context.Context, chapterID, menteeID string, input RequestMatchInput) (*Match, error) {
	// Check for existing active match for this mentee
	var existingID string
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM mentorships WHERE chapter_id = $1 AND mentee_id = $2 AND is_active = TRUE`,
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
		UPDATE mentorships
		SET mentee_id = $3, started_at = NOW(), updated_at = NOW()
		WHERE id = (
			SELECT id FROM mentorships
			WHERE chapter_id = $1 AND mentor_id = $2 AND mentee_id IS NULL AND is_active = TRUE
			ORDER BY created_at DESC
			LIMIT 1
		)
		RETURNING id, chapter_id, mentor_id, mentee_id, started_at, created_at
	`, chapterID, input.MentorID, menteeID).
		Scan(&match.ID, &match.ChapterID, &match.MentorID, &match.MenteeID, &match.StartedAt, &match.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("create mentorship match: %w", err)
	}
	match.Status = "active"
	return match, nil
}

func (s *service) GetMyMatch(ctx context.Context, chapterID, memberID string) (*Match, error) {
	match := &Match{}
	err := s.pool.QueryRow(ctx, `
		SELECT mt.id, mt.chapter_id, mt.mentor_id, mt.mentee_id, mt.started_at, mt.created_at,
				utor.first_name, utor.last_name, utee.first_name, utee.last_name
			FROM mentorships mt
			LEFT JOIN members tor ON tor.id = mt.mentor_id
			LEFT JOIN members tee ON tee.id = mt.mentee_id
			LEFT JOIN users utor ON utor.id = tor.user_id
			LEFT JOIN users utee ON utee.id = tee.user_id
		WHERE mt.chapter_id = $1 AND (mt.mentee_id = $2 OR mt.mentor_id = $2) AND mt.is_active = TRUE
		ORDER BY mt.created_at DESC
		LIMIT 1
	`, chapterID, memberID).
		Scan(&match.ID, &match.ChapterID, &match.MentorID, &match.MenteeID,
			&match.StartedAt, &match.CreatedAt,
			&match.MentorFirstName, &match.MentorLastName,
			&match.MenteeFirstName, &match.MenteeLastName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get my match: %w", err)
	}
	match.Status = "active"
	return match, nil
}
