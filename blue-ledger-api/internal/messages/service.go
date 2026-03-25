package messages

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("message thread not found")

// Thread is the domain model for a message thread.
type Thread struct {
	ID          string    `json:"id"`
	ChapterID   string    `json:"chapter_id"`
	Subject     string    `json:"subject"`
	CreatedBy   string    `json:"created_by"`
	IsGroupChat bool      `json:"is_group_chat"`
	CreatedAt   time.Time `json:"created_at"`
	// Aggregated
	LastMessageAt  *time.Time `json:"last_message_at,omitempty"`
	LastMessageBody *string   `json:"last_message_body,omitempty"`
	MessageCount   int        `json:"message_count"`
	// Viewer join
	AuthorFirstName *string `json:"author_first_name,omitempty"`
	AuthorLastName  *string `json:"author_last_name,omitempty"`
}

// Message is the domain model for a single message in a thread.
type Message struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	ChapterID string    `json:"chapter_id"`
	SenderID  string    `json:"sender_id"`
	Body      string    `json:"body"`
	SentAt    time.Time `json:"sent_at"`
	// Joined
	SenderFirstName *string `json:"sender_first_name,omitempty"`
	SenderLastName  *string `json:"sender_last_name,omitempty"`
	SenderAvatarURL *string `json:"sender_avatar_url,omitempty"`
}

// CreateThreadInput holds data for creating a new thread.
type CreateThreadInput struct {
	Subject     string   `json:"subject" validate:"required,min=1,max=200"`
	IsGroupChat bool     `json:"is_group_chat"`
	ParticipantIDs []string `json:"participant_ids"`
	InitialMessage string `json:"initial_message" validate:"required,min=1"`
}

// SendMessageInput holds data for sending a message.
type SendMessageInput struct {
	Body string `json:"body" validate:"required,min=1"`
}

// Service defines the messages business logic interface.
type Service interface {
	ListThreads(ctx context.Context, chapterID, memberID string) ([]*Thread, error)
	CreateThread(ctx context.Context, chapterID, memberID string, input CreateThreadInput) (*Thread, error)
	GetThread(ctx context.Context, chapterID, threadID, memberID string) ([]*Message, error)
	SendMessage(ctx context.Context, chapterID, threadID, senderID string, input SendMessageInput) (*Message, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new messages service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) ListThreads(ctx context.Context, chapterID, memberID string) ([]*Thread, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id, t.chapter_id, t.subject, t.created_by, t.is_group_chat, t.created_at,
			MAX(m.sent_at) AS last_message_at,
			(SELECT body FROM messages WHERE thread_id = t.id ORDER BY sent_at DESC LIMIT 1) AS last_message_body,
			COUNT(m.id) AS message_count,
			cr.first_name, cr.last_name
		FROM message_threads t
		LEFT JOIN messages m ON m.thread_id = t.id
		LEFT JOIN members cr ON cr.id = t.created_by
		JOIN thread_participants tp ON tp.thread_id = t.id AND tp.member_id = $2
		WHERE t.chapter_id = $1
		GROUP BY t.id, cr.first_name, cr.last_name
		ORDER BY MAX(m.sent_at) DESC NULLS LAST, t.created_at DESC
	`, chapterID, memberID)
	if err != nil {
		return nil, fmt.Errorf("list threads: %w", err)
	}
	defer rows.Close()

	var result []*Thread
	for rows.Next() {
		th := &Thread{}
		if err := rows.Scan(
			&th.ID, &th.ChapterID, &th.Subject, &th.CreatedBy, &th.IsGroupChat, &th.CreatedAt,
			&th.LastMessageAt, &th.LastMessageBody, &th.MessageCount,
			&th.AuthorFirstName, &th.AuthorLastName,
		); err != nil {
			return nil, fmt.Errorf("scan thread: %w", err)
		}
		result = append(result, th)
	}
	return result, rows.Err()
}

func (s *service) CreateThread(ctx context.Context, chapterID, memberID string, input CreateThreadInput) (*Thread, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	th := &Thread{}
	if err := tx.QueryRow(ctx, `
		INSERT INTO message_threads (chapter_id, subject, created_by, is_group_chat)
		VALUES ($1, $2, $3, $4)
		RETURNING id, chapter_id, subject, created_by, is_group_chat, created_at
	`, chapterID, input.Subject, memberID, input.IsGroupChat).
		Scan(&th.ID, &th.ChapterID, &th.Subject, &th.CreatedBy, &th.IsGroupChat, &th.CreatedAt); err != nil {
		return nil, fmt.Errorf("create thread: %w", err)
	}

	// Add creator as participant
	if _, err := tx.Exec(ctx,
		`INSERT INTO thread_participants (thread_id, member_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		th.ID, memberID,
	); err != nil {
		return nil, fmt.Errorf("add creator to thread: %w", err)
	}

	// Add other participants
	for _, pid := range input.ParticipantIDs {
		if pid == memberID {
			continue
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO thread_participants (thread_id, member_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			th.ID, pid,
		); err != nil {
			return nil, fmt.Errorf("add participant to thread: %w", err)
		}
	}

	// Send initial message
	if _, err := tx.Exec(ctx,
		`INSERT INTO messages (thread_id, chapter_id, sender_id, body) VALUES ($1, $2, $3, $4)`,
		th.ID, chapterID, memberID, input.InitialMessage,
	); err != nil {
		return nil, fmt.Errorf("send initial message: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit thread creation: %w", err)
	}

	return th, nil
}

func (s *service) GetThread(ctx context.Context, chapterID, threadID, memberID string) ([]*Message, error) {
	// Verify member is a participant
	var participantExists bool
	if err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM thread_participants WHERE thread_id = $1 AND member_id = $2)`,
		threadID, memberID,
	).Scan(&participantExists); err != nil {
		return nil, fmt.Errorf("check participant: %w", err)
	}
	if !participantExists {
		return nil, ErrNotFound
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			m.id, m.thread_id, m.chapter_id, m.sender_id, m.body, m.sent_at,
			mb.first_name, mb.last_name, mb.avatar_url
		FROM messages m
		LEFT JOIN members mb ON mb.id = m.sender_id
		WHERE m.thread_id = $1 AND m.chapter_id = $2
		ORDER BY m.sent_at ASC
	`, threadID, chapterID)
	if err != nil {
		return nil, fmt.Errorf("get thread messages: %w", err)
	}
	defer rows.Close()

	var result []*Message
	for rows.Next() {
		msg := &Message{}
		if err := rows.Scan(
			&msg.ID, &msg.ThreadID, &msg.ChapterID, &msg.SenderID, &msg.Body, &msg.SentAt,
			&msg.SenderFirstName, &msg.SenderLastName, &msg.SenderAvatarURL,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		result = append(result, msg)
	}
	return result, rows.Err()
}

func (s *service) SendMessage(ctx context.Context, chapterID, threadID, senderID string, input SendMessageInput) (*Message, error) {
	// Verify sender is a participant
	var participantExists bool
	if err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM thread_participants WHERE thread_id = $1 AND member_id = $2)`,
		threadID, senderID,
	).Scan(&participantExists); err != nil {
		return nil, fmt.Errorf("check participant: %w", err)
	}
	if !participantExists {
		return nil, ErrNotFound
	}

	msg := &Message{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO messages (thread_id, chapter_id, sender_id, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, thread_id, chapter_id, sender_id, body, sent_at
	`, threadID, chapterID, senderID, input.Body).
		Scan(&msg.ID, &msg.ThreadID, &msg.ChapterID, &msg.SenderID, &msg.Body, &msg.SentAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	return msg, nil
}
