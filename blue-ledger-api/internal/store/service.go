package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound         = errors.New("store item not found")
	ErrInsufficientXP   = errors.New("insufficient XP balance for this purchase")
	ErrItemUnavailable  = errors.New("item is not available")
)

// Item is the domain model for a store item.
type Item struct {
	ID          string     `json:"id"`
	ChapterID   string     `json:"chapter_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ImageURL    *string    `json:"image_url,omitempty"`
	XPCost      int        `json:"xp_cost"`
	Quantity    *int       `json:"quantity,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Order is the domain model for a store purchase order.
type Order struct {
	ID        string    `json:"id"`
	ChapterID string    `json:"chapter_id"`
	MemberID  string    `json:"member_id"`
	ItemID    string    `json:"item_id"`
	XPSpent   int       `json:"xp_spent"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	// Joined
	ItemName        *string `json:"item_name,omitempty"`
	MemberFirstName *string `json:"member_first_name,omitempty"`
	MemberLastName  *string `json:"member_last_name,omitempty"`
}

// CreateItemInput holds data for creating a store item.
type CreateItemInput struct {
	Name        string  `json:"name" validate:"required,min=2,max=200"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url"`
	XPCost      int     `json:"xp_cost" validate:"required,min=1"`
	Quantity    *int    `json:"quantity"`
}

// Service defines the store business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string) ([]*Item, error)
	Create(ctx context.Context, chapterID string, input CreateItemInput) (*Item, error)
	Purchase(ctx context.Context, chapterID, memberID, itemID string) (*Order, error)
	ListOrders(ctx context.Context, chapterID string, page, perPage int) ([]*Order, int, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new store service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) List(ctx context.Context, chapterID string) ([]*Item, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, name, description, image_url, xp_cost, quantity, is_active, created_at, updated_at
		FROM store_items
		WHERE chapter_id = $1 AND is_active = TRUE
		ORDER BY xp_cost ASC, name
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("list store items: %w", err)
	}
	defer rows.Close()

	var result []*Item
	for rows.Next() {
		item := &Item{}
		if err := rows.Scan(
			&item.ID, &item.ChapterID, &item.Name, &item.Description, &item.ImageURL,
			&item.XPCost, &item.Quantity, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan store item: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateItemInput) (*Item, error) {
	item := &Item{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO store_items (chapter_id, name, description, image_url, xp_cost, quantity, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE)
		RETURNING id, chapter_id, name, description, image_url, xp_cost, quantity, is_active, created_at, updated_at
	`, chapterID, input.Name, input.Description, input.ImageURL, input.XPCost, input.Quantity).
		Scan(&item.ID, &item.ChapterID, &item.Name, &item.Description, &item.ImageURL,
			&item.XPCost, &item.Quantity, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create store item: %w", err)
	}
	return item, nil
}

func (s *service) Purchase(ctx context.Context, chapterID, memberID, itemID string) (*Order, error) {
	// Get item details
	item := &Item{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, xp_cost, quantity, is_active FROM store_items WHERE id = $1 AND chapter_id = $2`,
		itemID, chapterID,
	).Scan(&item.ID, &item.XPCost, &item.Quantity, &item.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get store item: %w", err)
	}
	if !item.IsActive {
		return nil, ErrItemUnavailable
	}

	// Check member XP balance
	var xpTotal int
	if err := s.pool.QueryRow(ctx,
		`SELECT xp_total FROM members WHERE id = $1 AND chapter_id = $2`, memberID, chapterID,
	).Scan(&xpTotal); err != nil {
		return nil, fmt.Errorf("get member XP: %w", err)
	}
	if xpTotal < item.XPCost {
		return nil, ErrInsufficientXP
	}

	// Use transaction: deduct XP and create order atomically
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Deduct XP
	if _, err := tx.Exec(ctx,
		`UPDATE members SET xp_total = xp_total - $1 WHERE id = $2 AND chapter_id = $3`,
		item.XPCost, memberID, chapterID,
	); err != nil {
		return nil, fmt.Errorf("deduct XP: %w", err)
	}

	// Decrement quantity if tracked
	if item.Quantity != nil {
		tag, err := tx.Exec(ctx,
			`UPDATE store_items SET quantity = quantity - 1 WHERE id = $1 AND quantity > 0`,
			itemID,
		)
		if err != nil {
			return nil, fmt.Errorf("decrement item quantity: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil, ErrItemUnavailable
		}
	}

	// Create order
	order := &Order{}
	if err := tx.QueryRow(ctx, `
		INSERT INTO store_orders (chapter_id, member_id, item_id, xp_spent, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, chapter_id, member_id, item_id, xp_spent, status, created_at
	`, chapterID, memberID, itemID, item.XPCost).
		Scan(&order.ID, &order.ChapterID, &order.MemberID, &order.ItemID,
			&order.XPSpent, &order.Status, &order.CreatedAt); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit purchase transaction: %w", err)
	}

	return order, nil
}

func (s *service) ListOrders(ctx context.Context, chapterID string, page, perPage int) ([]*Order, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	offset := (page - 1) * perPage

	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM store_orders WHERE chapter_id = $1`, chapterID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			so.id, so.chapter_id, so.member_id, so.item_id, so.xp_spent, so.status, so.created_at,
			si.name AS item_name,
			m.first_name, m.last_name
		FROM store_orders so
		LEFT JOIN store_items si ON si.id = so.item_id
		LEFT JOIN members m ON m.id = so.member_id
		WHERE so.chapter_id = $1
		ORDER BY so.created_at DESC
		LIMIT $2 OFFSET $3
	`, chapterID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var result []*Order
	for rows.Next() {
		o := &Order{}
		if err := rows.Scan(
			&o.ID, &o.ChapterID, &o.MemberID, &o.ItemID, &o.XPSpent, &o.Status, &o.CreatedAt,
			&o.ItemName, &o.MemberFirstName, &o.MemberLastName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		result = append(result, o)
	}
	return result, total, rows.Err()
}
