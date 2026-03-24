package members

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Member represents a chapter membership record joined with user profile data.
type Member struct {
	ID              string     `json:"id"`
	ChapterID       string     `json:"chapter_id"`
	UserID          string     `json:"user_id"`
	MemberDisplayID string     `json:"member_display_id"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	DisplayName     *string    `json:"display_name,omitempty"`
	Email           string     `json:"email"`
	AvatarURL       *string    `json:"avatar_url,omitempty"`
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	InductedYear    *int       `json:"inducted_year,omitempty"`
	Employer        *string    `json:"employer,omitempty"`
	JobTitle        *string    `json:"job_title,omitempty"`
	City            *string    `json:"city,omitempty"`
	LinkedinURL     *string    `json:"linkedin_url,omitempty"`
	XPTotal         int        `json:"xp_total"`
	XPSemester      int        `json:"xp_semester"`
	Level           string     `json:"level"`
	LevelKey        string     `json:"level_key"`
	DuesStatus      string     `json:"dues_status"`
	AvatarBg        *string    `json:"avatar_bg,omitempty"`
	AvatarFg        *string    `json:"avatar_fg,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// XPHistoryEntry is a single record from the engagement_log.
type XPHistoryEntry struct {
	ID            string     `json:"id"`
	Activity      string     `json:"activity"`
	XPAwarded     int        `json:"xp_awarded"`
	Source        string     `json:"source"`
	ReferenceID   *string    `json:"reference_id,omitempty"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	Semester      *string    `json:"semester,omitempty"`
	Note          *string    `json:"note,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Repository provides database access for the members domain.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new members repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const memberSelectCols = `
	m.id, m.chapter_id, m.user_id, m.member_display_id,
	u.first_name, u.last_name, u.display_name, u.email, u.avatar_url,
	m.role, m.status, m.inducted_year, m.employer, m.job_title,
	m.city, m.linkedin_url, m.xp_total, m.xp_semester,
	m.level, m.level_key, m.dues_status, m.avatar_bg, m.avatar_fg,
	m.created_at, m.updated_at
`

func scanMember(row pgx.Row) (*Member, error) {
	var m Member
	err := row.Scan(
		&m.ID, &m.ChapterID, &m.UserID, &m.MemberDisplayID,
		&m.FirstName, &m.LastName, &m.DisplayName, &m.Email, &m.AvatarURL,
		&m.Role, &m.Status, &m.InductedYear, &m.Employer, &m.JobTitle,
		&m.City, &m.LinkedinURL, &m.XPTotal, &m.XPSemester,
		&m.Level, &m.LevelKey, &m.DuesStatus, &m.AvatarBg, &m.AvatarFg,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetByID fetches a single member by their members.id (UUID).
func (r *Repository) GetByID(ctx context.Context, id string) (*Member, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+memberSelectCols+`
		FROM members m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = $1 AND m.deleted_at IS NULL`, id)

	m, err := scanMember(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get member by id: %w", err)
	}
	return m, nil
}

// GetByDisplayID fetches a member by their chapter-scoped display ID (e.g. "ΤΤΣ-001").
func (r *Repository) GetByDisplayID(ctx context.Context, chapterID, displayID string) (*Member, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+memberSelectCols+`
		FROM members m
		JOIN users u ON u.id = m.user_id
		WHERE m.chapter_id = $1 AND m.member_display_id = $2 AND m.deleted_at IS NULL`,
		chapterID, displayID)

	m, err := scanMember(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get member by display id: %w", err)
	}
	return m, nil
}

// List returns all active members for a chapter with pagination.
func (r *Repository) List(ctx context.Context, chapterID string, page, perPage int) ([]*Member, int, error) {
	offset := (page - 1) * perPage

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM members WHERE chapter_id = $1 AND deleted_at IS NULL`, chapterID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count members: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+memberSelectCols+`
		FROM members m
		JOIN users u ON u.id = m.user_id
		WHERE m.chapter_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.xp_total DESC
		LIMIT $2 OFFSET $3`,
		chapterID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var members []*Member
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}

	return members, total, nil
}

// UpdateInput holds the fields that can be updated on a member profile.
type UpdateInput struct {
	InductedYear *int
	Employer     *string
	JobTitle     *string
	City         *string
	LinkedinURL  *string
	AvatarURL    *string
	AvatarBg     *string
	AvatarFg     *string
	// Admin-only fields
	Role       *string
	Status     *string
	DuesStatus *string
}

// Update applies partial updates to a member record.
func (r *Repository) Update(ctx context.Context, memberID string, input UpdateInput) (*Member, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE members SET
			inducted_year  = COALESCE($2, inducted_year),
			employer       = COALESCE($3, employer),
			job_title      = COALESCE($4, job_title),
			city           = COALESCE($5, city),
			linkedin_url   = COALESCE($6, linkedin_url),
			avatar_bg      = COALESCE($7, avatar_bg),
			avatar_fg      = COALESCE($8, avatar_fg),
			role           = COALESCE($9, role),
			status         = COALESCE($10, status),
			dues_status    = COALESCE($11, dues_status),
			updated_at     = NOW()
		WHERE id = $1 AND deleted_at IS NULL`,
		memberID, input.InductedYear, input.Employer, input.JobTitle, input.City,
		input.LinkedinURL, input.AvatarBg, input.AvatarFg, input.Role, input.Status, input.DuesStatus,
	)
	if err != nil {
		return nil, fmt.Errorf("update member: %w", err)
	}

	return r.GetByID(ctx, memberID)
}

// SoftDelete marks a member as deleted.
func (r *Repository) SoftDelete(ctx context.Context, memberID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE members SET deleted_at = NOW(), status = 'inactive' WHERE id = $1`, memberID)
	if err != nil {
		return fmt.Errorf("soft delete member: %w", err)
	}
	return nil
}

// GetXPHistory returns the engagement_log entries for a member.
func (r *Repository) GetXPHistory(ctx context.Context, memberID string, page, perPage int) ([]*XPHistoryEntry, int, error) {
	offset := (page - 1) * perPage

	var total int
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM engagement_log WHERE member_id = $1`, memberID).Scan(&total)

	rows, err := r.db.Query(ctx, `
		SELECT id, activity, xp_awarded, source, reference_id, reference_type, semester, note, created_at
		FROM engagement_log
		WHERE member_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		memberID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get xp history: %w", err)
	}
	defer rows.Close()

	var entries []*XPHistoryEntry
	for rows.Next() {
		var e XPHistoryEntry
		if err := rows.Scan(&e.ID, &e.Activity, &e.XPAwarded, &e.Source,
			&e.ReferenceID, &e.ReferenceType, &e.Semester, &e.Note, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan xp entry: %w", err)
		}
		entries = append(entries, &e)
	}

	return entries, total, nil
}
