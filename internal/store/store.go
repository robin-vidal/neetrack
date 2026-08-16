// Package store owns the database connection pool and migration lifecycle.
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

// Store wraps the connection pool used by the rest of the application.
type Store struct {
	pool    *pgxpool.Pool
	queries *storedb.Queries
}

// New builds a Store backed by a pgx connection pool.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &Store{pool: pool, queries: storedb.New(pool)}, nil
}

// Ping reports whether the database is reachable.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Close releases every connection held by the pool.
func (s *Store) Close() {
	s.pool.Close()
}
