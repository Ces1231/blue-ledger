package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates and validates a new pgx connection pool from the given database URL.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	// Tuning: adjust pool size for a small SaaS deployment.
	config.MaxConns = 25
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// SetChapterContext sets the PostgreSQL session variable used by RLS policies.
// Call this at the start of every tenant-scoped request handler (via TenantSetter middleware).
// It sets: SET LOCAL app.chapter_id = '<uuid>'
// "LOCAL" means it resets at the end of the current transaction, providing per-request isolation.
func SetChapterContext(ctx context.Context, pool *pgxpool.Pool, chapterID string) error {
	_, err := pool.Exec(ctx,
		fmt.Sprintf("SET LOCAL app.chapter_id = '%s'", chapterID),
	)
	if err != nil {
		return fmt.Errorf("set chapter context: %w", err)
	}
	return nil
}

// SetChapterContextOnConn sets the session variable on an acquired connection.
// Use this when you already hold a pgxpool.Conn (e.g., inside a transaction).
func SetChapterContextOnConn(ctx context.Context, conn interface {
	Exec(ctx context.Context, sql string, arguments ...any) (interface{}, error)
}, chapterID string) error {
	_, err := conn.Exec(ctx, fmt.Sprintf("SET LOCAL app.chapter_id = '%s'", chapterID))
	if err != nil {
		return fmt.Errorf("set chapter context on conn: %w", err)
	}
	return nil
}

// Ping checks that the database is reachable. Used by the /healthz endpoint.
func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
