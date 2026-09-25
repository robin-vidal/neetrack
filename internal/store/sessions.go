package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

// CreateSession implements auth.SessionStore.
func (s *Store) CreateSession(ctx context.Context, hash []byte, userID int64, expiresAt time.Time) error {
	err := s.queries.CreateSession(ctx, storedb.CreateSessionParams{
		ID:        hash,
		UserID:    userID,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	return nil
}

// SessionUser implements auth.SessionStore.
func (s *Store) SessionUser(ctx context.Context, hash []byte) (auth.SessionRecord, error) {
	var row auth.SessionRecord
	err := s.pool.QueryRow(ctx, `
		SELECT sessions.user_id, users.name, users.is_admin, sessions.expires_at
		FROM sessions JOIN users ON users.id = sessions.user_id
		WHERE sessions.id = $1`, hash).Scan(&row.UserID, &row.UserName, &row.IsAdmin, &row.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.SessionRecord{}, auth.ErrSessionNotFound
	}
	if err != nil {
		return auth.SessionRecord{}, fmt.Errorf("querying session: %w", err)
	}

	return row, nil
}

// ExtendSession implements auth.SessionStore.
func (s *Store) ExtendSession(ctx context.Context, hash []byte, expiresAt time.Time) error {
	err := s.queries.ExtendSession(ctx, storedb.ExtendSessionParams{
		ID:        hash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("extending session: %w", err)
	}

	return nil
}

// DeleteSession implements auth.SessionStore.
func (s *Store) DeleteSession(ctx context.Context, hash []byte) error {
	if err := s.queries.DeleteSession(ctx, hash); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	return nil
}
