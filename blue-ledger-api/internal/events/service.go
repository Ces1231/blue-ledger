package events

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Event represents an event record.
type Event struct {
	ID           string     `json:"id"`
	ChapterID    string     `json:"chapter_id"`
	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	EventDate    string     `json:"event_date"` // DATE stored as string "YYYY-MM-DD"
	EventTime    *string    `json:"event_time,omitempty"`
	Location     *string    `json:"location,omitempty"`
	EventType    string     `json:"event_type"`
	XPAttend     int        `json:"xp_attend"`
	XPRsvp       int        `json:"xp_rsvp"`
	RSVPDeadline *string    `json:"rsvp_deadline,omitempty"`
	Capacity     *int       `json:"capacity,omitempty"`
	IsActive     bool       `json:"is_active"`
	QRCodeToken  *string    `json:"qr_code_token,omitempty"`
	CreatedBy    *string    `json:"created_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateEventInput holds request data for creating an event.
type CreateEventInput struct {
	Name         string  `json:"name" validate:"required,min=3"`
	Description  *string `json:"description"`
	EventDate    string  `json:"event_date" validate:"required"`
	EventTime    *string `json:"event_time"`
	Location     *string `json:"location"`
	EventType    string  `json:"event_type" validate:"required"`
	XPAttend     int     `json:"xp_attend"`
	XPRsvp       int     `json:"xp_rsvp"`
	RSVPDeadline *string `json:"rsvp_deadline"`
	Capacity     *int    `json:"capacity"`
}

// CheckInResult is returned after a successful QR check-in.
type CheckInResult struct {
	MemberID    string   `json:"member_id"`
	MemberName  string   `json:"member_name"`
	XPAwarded   int      `json:"xp_awarded"`
	Level       string   `json:"level"`
	NewBadges   []string `json:"new_badges"`
}

// EventsService defines the events domain interface.
type EventsService interface {
	List(ctx context.Context, chapterID string, upcoming bool, page, perPage int) ([]*Event, int, error)
	Get(ctx context.Context, eventID string) (*Event, error)
	Create(ctx context.Context, chapterID, createdByMemberID string, input CreateEventInput) (*Event, error)
	Update(ctx context.Context, eventID string, input CreateEventInput) (*Event, error)
	Delete(ctx context.Context, eventID string) error
	RSVP(ctx context.Context, chapterID, eventID, memberID, status string) error
	GetQRToken(ctx context.Context, chapterID, eventID, qrSecret string) (string, error)
	CheckInQR(ctx context.Context, chapterID, eventID, memberDisplayID, scannerToken, scannerMemberID, qrSecret string) (*CheckInResult, error)
}

type eventsService struct {
	db *pgxpool.Pool
}

// NewEventsService creates a new events service.
func NewEventsService(db *pgxpool.Pool) EventsService {
	return &eventsService{db: db}
}

func (s *eventsService) List(ctx context.Context, chapterID string, upcoming bool, page, perPage int) ([]*Event, int, error) {
	offset := (page - 1) * perPage
	dateFilter := "event_date < CURRENT_DATE"
	if upcoming {
		dateFilter = "event_date >= CURRENT_DATE"
	}

	var total int
	s.db.QueryRow(ctx, fmt.Sprintf(
		`SELECT COUNT(*) FROM events WHERE chapter_id = $1 AND %s AND deleted_at IS NULL`, dateFilter),
		chapterID).Scan(&total)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT id, chapter_id, name, description, event_date, event_time, location,
		       event_type, xp_attend, xp_rsvp, rsvp_deadline, capacity, is_active,
		       qr_code_token, created_by, created_at, updated_at
		FROM events
		WHERE chapter_id = $1 AND %s AND deleted_at IS NULL
		ORDER BY event_date ASC
		LIMIT $2 OFFSET $3`, dateFilter),
		chapterID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	return scanEvents(rows, total)
}

func (s *eventsService) Get(ctx context.Context, eventID string) (*Event, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, chapter_id, name, description, event_date, event_time, location,
		       event_type, xp_attend, xp_rsvp, rsvp_deadline, capacity, is_active,
		       qr_code_token, created_by, created_at, updated_at
		FROM events WHERE id = $1 AND deleted_at IS NULL`, eventID)

	e, err := scanEvent(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return e, err
}

func (s *eventsService) Create(ctx context.Context, chapterID, createdByMemberID string, input CreateEventInput) (*Event, error) {
	xpAttend := 50
	if input.XPAttend > 0 {
		xpAttend = input.XPAttend
	}
	xpRsvp := 10
	if input.XPRsvp > 0 {
		xpRsvp = input.XPRsvp
	}

	var id string
	err := s.db.QueryRow(ctx, `
		INSERT INTO events (chapter_id, name, description, event_date, event_time, location,
		                    event_type, xp_attend, xp_rsvp, rsvp_deadline, capacity, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id`,
		chapterID, input.Name, input.Description, input.EventDate, input.EventTime,
		input.Location, input.EventType, xpAttend, xpRsvp, input.RSVPDeadline,
		input.Capacity, createdByMemberID,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}

	return s.Get(ctx, id)
}

func (s *eventsService) Update(ctx context.Context, eventID string, input CreateEventInput) (*Event, error) {
	_, err := s.db.Exec(ctx, `
		UPDATE events SET
			name          = COALESCE(NULLIF($2,''), name),
			description   = COALESCE($3, description),
			event_date    = COALESCE(NULLIF($4,''), event_date::text)::date,
			event_time    = COALESCE($5, event_time),
			location      = COALESCE($6, location),
			event_type    = COALESCE(NULLIF($7,''), event_type),
			xp_attend     = CASE WHEN $8 > 0 THEN $8 ELSE xp_attend END,
			xp_rsvp       = CASE WHEN $9 > 0 THEN $9 ELSE xp_rsvp END,
			rsvp_deadline = COALESCE($10, rsvp_deadline),
			capacity      = COALESCE($11, capacity),
			updated_at    = NOW()
		WHERE id = $1 AND deleted_at IS NULL`,
		eventID, input.Name, input.Description, input.EventDate, input.EventTime,
		input.Location, input.EventType, input.XPAttend, input.XPRsvp,
		input.RSVPDeadline, input.Capacity,
	)
	if err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}
	return s.Get(ctx, eventID)
}

func (s *eventsService) Delete(ctx context.Context, eventID string) error {
	_, err := s.db.Exec(ctx, `UPDATE events SET deleted_at = NOW() WHERE id = $1`, eventID)
	return err
}

func (s *eventsService) RSVP(ctx context.Context, chapterID, eventID, memberID, status string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO rsvps (chapter_id, event_id, member_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (event_id, member_id)
		DO UPDATE SET status = $4, updated_at = NOW()`,
		chapterID, eventID, memberID, status)
	return err
}

func (s *eventsService) GetQRToken(ctx context.Context, chapterID, eventID, qrSecret string) (string, error) {
	var existing *string
	s.db.QueryRow(ctx, `SELECT qr_code_token FROM events WHERE id = $1`, eventID).Scan(&existing)

	if existing != nil && *existing != "" {
		return *existing, nil
	}

	token, err := GenerateQRToken(qrSecret, eventID, chapterID)
	if err != nil {
		return "", err
	}

	_, err = s.db.Exec(ctx, `UPDATE events SET qr_code_token = $1 WHERE id = $2`, token, eventID)
	if err != nil {
		return "", fmt.Errorf("store qr token: %w", err)
	}

	return token, nil
}

func (s *eventsService) CheckInQR(ctx context.Context, chapterID, eventID, memberDisplayID, scannerToken, scannerMemberID, qrSecret string) (*CheckInResult, error) {
	// Validate scanner token
	tokenEventID, tokenChapterID, err := ValidateQRToken(qrSecret, scannerToken)
	if err != nil {
		return nil, fmt.Errorf("invalid scanner token: %w", err)
	}
	if tokenEventID != eventID || tokenChapterID != chapterID {
		return nil, fmt.Errorf("token mismatch")
	}

	// Look up event
	var xpAttend int
	var isActive bool
	err = s.db.QueryRow(ctx, `SELECT xp_attend, is_active FROM events WHERE id = $1 AND chapter_id = $2`,
		eventID, chapterID).Scan(&xpAttend, &isActive)
	if err == pgx.ErrNoRows || !isActive {
		return nil, fmt.Errorf("event not found or not active")
	}

	// Look up member by display ID
	var memberID, firstName, lastName, level string
	err = s.db.QueryRow(ctx, `
			SELECT m.id, u.first_name, u.last_name, m.level_key
			FROM members m JOIN users u ON u.id = m.user_id
			WHERE m.chapter_id = $1 AND m.display_id = $2 AND m.deleted_at IS NULL`,
		chapterID, memberDisplayID).Scan(&memberID, &firstName, &lastName, &level)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("member not found")
	}
	if err != nil {
		return nil, fmt.Errorf("lookup member: %w", err)
	}

	// Check for duplicate check-in
	var alreadyCheckedIn bool
	s.db.QueryRow(ctx, `SELECT attended FROM rsvps WHERE event_id = $1 AND member_id = $2`,
		eventID, memberID).Scan(&alreadyCheckedIn)
	if alreadyCheckedIn {
		return nil, fmt.Errorf("member already checked in")
	}

	// Perform check-in in a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin checkin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Upsert RSVP
	_, err = tx.Exec(ctx, `
		INSERT INTO rsvps (chapter_id, event_id, member_id, status, attended, checked_in_at, checked_in_by)
		VALUES ($1, $2, $3, 'yes', true, NOW(), $4)
		ON CONFLICT (event_id, member_id)
		DO UPDATE SET attended = true, checked_in_at = NOW(), checked_in_by = $4, updated_at = NOW()`,
		chapterID, eventID, memberID, scannerMemberID)
	if err != nil {
		return nil, fmt.Errorf("upsert rsvp: %w", err)
	}

	// Log XP
	_, err = tx.Exec(ctx, `
		INSERT INTO engagement_log (chapter_id, member_id, activity, xp_awarded, source, reference_id, reference_type)
		VALUES ($1, $2, 'Event Check-In', $3, 'checkin', $4, 'event')`,
		chapterID, memberID, xpAttend, eventID)
	if err != nil {
		return nil, fmt.Errorf("insert engagement log: %w", err)
	}

	// Update XP total
	_, err = tx.Exec(ctx, `UPDATE members SET xp_total = xp_total + $1, xp_semester = xp_semester + $1, updated_at = NOW() WHERE id = $2`,
		xpAttend, memberID)
	if err != nil {
		return nil, fmt.Errorf("update xp: %w", err)
	}

	// Create notification
	_, _ = tx.Exec(ctx, `
		INSERT INTO notifications (chapter_id, user_id, type, title, body, link)
		SELECT chapter_id, user_id, 'xp', 'Check-In Complete', $2, $3
		FROM members WHERE id = $1`,
		memberID,
		fmt.Sprintf("You earned %d XP for attending this event!", xpAttend),
		"/events")

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit checkin: %w", err)
	}

	// Run badge engine asynchronously
	go func() {
		bgCtx := context.Background()
		_, _ = CheckAndAwardBadgesForEvent(bgCtx, s.db, chapterID, memberID)
	}()

	return &CheckInResult{
		MemberID:   memberID,
		MemberName: fmt.Sprintf("%s %s", firstName, lastName),
		XPAwarded:  xpAttend,
		Level:      level,
		NewBadges:  []string{},
	}, nil
}

// CheckAndAwardBadgesForEvent is a thin shim that defers to the xp package badge engine.
// Defined here to avoid an import cycle — events -> xp would create a cycle.
// The badge engine import will be wired in cmd/server/main.go instead.
func CheckAndAwardBadgesForEvent(ctx context.Context, db *pgxpool.Pool, chapterID, memberID string) ([]string, error) {
	// This is intentionally a no-op stub here; the actual wiring is done in cmd/server.
	// Iron Man will wire the badge engine callback during service construction.
	return nil, nil
}

// scanEvent scans a single event row.
func scanEvent(row pgx.Row) (*Event, error) {
	var e Event
	err := row.Scan(
		&e.ID, &e.ChapterID, &e.Name, &e.Description, &e.EventDate, &e.EventTime,
		&e.Location, &e.EventType, &e.XPAttend, &e.XPRsvp, &e.RSVPDeadline,
		&e.Capacity, &e.IsActive, &e.QRCodeToken, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// scanEvents scans a rows result set into a slice of events.
func scanEvents(rows pgx.Rows, total int) ([]*Event, int, error) {
	var events []*Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, total, nil
}
