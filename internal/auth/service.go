package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Service implements login, logout and session verification.
type Service struct {
	users        UserStore
	sessions     SessionStore
	cookieSecure bool
}

// NewService builds a Service backed by the given stores.
func NewService(users UserStore, sessions SessionStore, cookieSecure bool) *Service {
	return &Service{users: users, sessions: sessions, cookieSecure: cookieSecure}
}

// Login verifies name and password, and on success issues a new session
// cookie. It returns ErrInvalidCredentials for both an unknown user and a
// wrong password, never distinguishing the two.
func (s *Service) Login(ctx context.Context, name, password string) (*http.Cookie, error) {
	creds, err := s.users.UserByName(ctx, name)
	if errors.Is(err, ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	ok, err := VerifyPassword(creds.PasswordHash, password)
	if err != nil {
		return nil, fmt.Errorf("verifying password: %w", err)
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}

	token, err := NewToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(SessionDuration)
	if err := s.sessions.CreateSession(ctx, token.Hash, creds.UserID, expiresAt); err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	return s.cookie(token.Raw, expiresAt), nil
}

// Logout deletes the session identified by rawToken. An unknown or
// malformed token is treated as already logged out.
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	hash, err := HashToken(rawToken)
	if err != nil {
		return nil
	}

	if err := s.sessions.DeleteSession(ctx, hash); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	return nil
}

// Authenticate resolves rawToken to its user and extends the session's
// expiry. It returns ErrSessionNotFound for an unknown, malformed or
// expired token.
func (s *Service) Authenticate(ctx context.Context, rawToken string) (User, error) {
	hash, err := HashToken(rawToken)
	if err != nil {
		return User{}, ErrSessionNotFound
	}

	record, err := s.sessions.SessionUser(ctx, hash)
	if errors.Is(err, ErrSessionNotFound) {
		return User{}, ErrSessionNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("resolving session: %w", err)
	}

	if time.Now().After(record.ExpiresAt) {
		_ = s.sessions.DeleteSession(ctx, hash)
		return User{}, ErrSessionNotFound
	}

	if err := s.sessions.ExtendSession(ctx, hash, time.Now().Add(SessionDuration)); err != nil {
		return User{}, fmt.Errorf("extending session: %w", err)
	}

	return User{ID: record.UserID, Name: record.UserName}, nil
}

// ClearCookie returns a cookie that immediately expires the session cookie
// on the client.
func (s *Service) ClearCookie() *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
}

func (s *Service) cookie(rawToken string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	}
}
