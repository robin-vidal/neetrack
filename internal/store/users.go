package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

var (
	// ErrManagedUserNotFound reports an account that no longer exists.
	ErrManagedUserNotFound = errors.New("managed user not found")
	// ErrLastAdmin reports an attempt to remove or demote the final administrator.
	ErrLastAdmin = errors.New("cannot remove the last administrator")
	// ErrManagedUserNameTaken reports a duplicate account name.
	ErrManagedUserNameTaken = errors.New("managed user name already exists")
)

// ManagedUser is the safe account summary used by the admin panel.
type ManagedUser struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUser inserts a new user with an already-hashed password and its
// default settings row, and returns the user's id.
func (s *Store) CreateUser(ctx context.Context, name, passwordHash string) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(914837215)"); err != nil {
		return 0, fmt.Errorf("locking initial account creation: %w", err)
	}

	q := s.queries.WithTx(tx)

	var u storedb.User
	err = tx.QueryRow(ctx, `
		INSERT INTO users (name, password_hash, is_admin)
		VALUES ($1, $2, NOT EXISTS (SELECT 1 FROM users))
		RETURNING id, name, password_hash, created_at`, name, passwordHash).Scan(&u.ID, &u.Name, &u.PasswordHash, &u.CreatedAt)
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

// AdminUsers lists account summaries without credential material.
func (s *Store) AdminUsers(ctx context.Context) ([]ManagedUser, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, is_admin, created_at FROM users ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()
	users := make([]ManagedUser, 0)
	for rows.Next() {
		var user ManagedUser
		if err := rows.Scan(&user.ID, &user.Name, &user.IsAdmin, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading users: %w", err)
	}
	return users, nil
}

// CreateManagedUser creates an account with default settings.
func (s *Store) CreateManagedUser(ctx context.Context, name, passwordHash string, isAdmin bool) (ManagedUser, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("beginning user creation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(914837215)"); err != nil {
		return ManagedUser{}, fmt.Errorf("locking user creation: %w", err)
	}
	var user ManagedUser
	err = tx.QueryRow(ctx, `INSERT INTO users (name, password_hash, is_admin) VALUES ($1, $2, $3) RETURNING id, name, is_admin, created_at`, name, passwordHash, isAdmin).Scan(&user.ID, &user.Name, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ManagedUser{}, ErrManagedUserNameTaken
		}
		return ManagedUser{}, fmt.Errorf("inserting managed user: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO user_settings (user_id) VALUES ($1)`, user.ID); err != nil {
		return ManagedUser{}, fmt.Errorf("creating managed user settings: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ManagedUser{}, fmt.Errorf("committing managed user: %w", err)
	}
	return user, nil
}

// UpdateManagedUser applies provided account fields and revokes sessions after a password change.
func (s *Store) UpdateManagedUser(ctx context.Context, id int64, name *string, passwordHash *string, isAdmin *bool) (ManagedUser, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("beginning user update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(914837215)"); err != nil {
		return ManagedUser{}, fmt.Errorf("locking user update: %w", err)
	}
	var user ManagedUser
	err = tx.QueryRow(ctx, `SELECT id, name, is_admin, created_at FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&user.ID, &user.Name, &user.IsAdmin, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ManagedUser{}, ErrManagedUserNotFound
	}
	if err != nil {
		return ManagedUser{}, fmt.Errorf("loading managed user: %w", err)
	}
	if user.IsAdmin && isAdmin != nil && !*isAdmin {
		var admins int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM users WHERE is_admin`).Scan(&admins); err != nil {
			return ManagedUser{}, fmt.Errorf("counting administrators: %w", err)
		}
		if admins <= 1 {
			return ManagedUser{}, ErrLastAdmin
		}
	}
	newName := user.Name
	newAdmin := user.IsAdmin
	if name != nil {
		newName = *name
	}
	if isAdmin != nil {
		newAdmin = *isAdmin
	}
	_, err = tx.Exec(ctx, `UPDATE users SET name=$2, is_admin=$3, password_hash=COALESCE($4, password_hash) WHERE id=$1`, id, newName, newAdmin, nullableString(passwordHash))
	if err != nil {
		if isUniqueViolation(err) {
			return ManagedUser{}, ErrManagedUserNameTaken
		}
		return ManagedUser{}, fmt.Errorf("updating managed user: %w", err)
	}
	if passwordHash != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, id); err != nil {
			return ManagedUser{}, fmt.Errorf("revoking managed user sessions: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return ManagedUser{}, fmt.Errorf("committing managed user update: %w", err)
	}
	user.Name = newName
	user.IsAdmin = newAdmin
	return user, nil
}

// DeleteManagedUser removes an account and its dependent data.
func (s *Store) DeleteManagedUser(ctx context.Context, id int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning user deletion: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(914837215)"); err != nil {
		return fmt.Errorf("locking user deletion: %w", err)
	}
	var isAdmin bool
	err = tx.QueryRow(ctx, `SELECT is_admin FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&isAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrManagedUserNotFound
	}
	if err != nil {
		return fmt.Errorf("loading user for deletion: %w", err)
	}
	if isAdmin {
		var admins int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM users WHERE is_admin`).Scan(&admins); err != nil {
			return fmt.Errorf("counting administrators: %w", err)
		}
		if admins <= 1 {
			return ErrLastAdmin
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, id); err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing user deletion: %w", err)
	}
	return nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
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
