package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

// CreateUser inserts a new user with an already-hashed password and returns
// its id.
func (s *Store) CreateUser(ctx context.Context, name, passwordHash string) (int64, error) {
	u, err := s.queries.CreateUser(ctx, storedb.CreateUserParams{
		Name:         name,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return 0, fmt.Errorf("creating user: %w", err)
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
