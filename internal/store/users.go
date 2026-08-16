package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

// CreateUser inserts a new user with an already-hashed password and its
// default settings row, and returns the user's id.
func (s *Store) CreateUser(ctx context.Context, name, passwordHash string) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := s.queries.WithTx(tx)

	u, err := q.CreateUser(ctx, storedb.CreateUserParams{
		Name:         name,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return 0, fmt.Errorf("creating user: %w", err)
	}

	if err := q.CreateDefaultSettings(ctx, u.ID); err != nil {
		return 0, fmt.Errorf("creating default settings: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("committing transaction: %w", err)
	}

	return u.ID, nil
}

// UserByName implements auth.UserStore.
func (s *Store) UserByName(ctx context.Context, name string) (auth.Credentials, error) {
	u, err := s.queries.UserByName(ctx, name)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Credentials{}, auth.ErrUserNotFound
	}
	if err != nil {
		return auth.Credentials{}, fmt.Errorf("querying user: %w", err)
	}

	return auth.Credentials{UserID: u.ID, PasswordHash: u.PasswordHash}, nil
}
