package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Notification represents a row from the notifications table.
type Notification struct {
	ID        string    `json:"id"`
	ChapterID string    `json:"chapter_id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Link      *string   `json:"link,omitempty"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationsService defines the notifications domain interface.
type NotificationsService interface {
	ListForUser(ctx context.Context, userID string, page, perPage int) ([]*Notification, int, error)
	MarkRead(ctx context.Context, notificationID, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, notificationID, userID string) error
	Create(ctx context.Context, chapterID, userID, notifType, title, body string, link *string) (*Notification, error)
}

type notificationsService struct {
	db *pgxpool.Pool
}

// NewNotificationsService creates a new notifications service.
func NewNotificationsService(db *pgxpool.Pool) NotificationsService {
	return &notificationsService{db: db}
}

func (s *notificationsService) ListForUser(ctx context.Context, userID string, page, perPage int) ([]*Notification, int, error) {
	offset := (page - 1) * perPage

	var total int
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID).Scan(&total)

	rows, err := s.db.Query(ctx, `
		SELECT id, chapter_id, user_id, type, title, body, link, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY is_read ASC, created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.ChapterID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Link, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, &n)
	}

	return notifications, total, nil
}

func (s *notificationsService) MarkRead(ctx context.Context, notificationID, userID string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`,
		notificationID, userID)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

func (s *notificationsService) MarkAllRead(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`,
		userID)
	if err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

func (s *notificationsService) Delete(ctx context.Context, notificationID, userID string) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM notifications WHERE id = $1 AND user_id = $2`,
		notificationID, userID)
	return err
}

func (s *notificationsService) Create(ctx context.Context, chapterID, userID, notifType, title, body string, link *string) (*Notification, error) {
	var id string
	err := s.db.QueryRow(ctx, `
		INSERT INTO notifications (chapter_id, user_id, type, title, body, link)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		chapterID, userID, notifType, title, body, link,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	return &Notification{
		ID:        id,
		ChapterID: chapterID,
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Body:      body,
		Link:      link,
		IsRead:    false,
		CreatedAt: time.Now(),
	}, nil
}
