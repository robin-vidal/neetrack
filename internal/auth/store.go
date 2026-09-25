package auth

import (
	"context"
	"time"
)

// Credentials is what a UserStore returns to verify a login attempt.
type Credentials struct {
	UserID       int64
	PasswordHash string
}

// SessionRecord is what a SessionStore returns to resolve a session token.
type SessionRecord struct {
	UserID    int64
	UserName  string
	IsAdmin   bool
	ExpiresAt time.Time
}

// UserStore looks up the credentials needed to verify a login attempt.
type UserStore interface {
	UserByName(ctx context.Context, name string) (Credentials, error)
}

// SessionStore persists and resolves session tokens.
type SessionStore interface {
	CreateSession(ctx context.Context, hash []byte, userID int64, expiresAt time.Time) error
	SessionUser(ctx context.Context, hash []byte) (SessionRecord, error)
	ExtendSession(ctx context.Context, hash []byte, expiresAt time.Time) error
	DeleteSession(ctx context.Context, hash []byte) error
}
