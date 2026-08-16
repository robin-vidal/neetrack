package auth

import "errors"

var (
	// ErrInvalidCredentials is returned when a login fails, whether the user
	// name is unknown or the password is wrong. The two cases are never
	// distinguished, to avoid leaking which user names exist.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserNotFound is returned by a UserStore when no user matches the
	// requested name.
	ErrUserNotFound = errors.New("user not found")

	// ErrSessionNotFound is returned by a SessionStore when the session token
	// is unknown or expired.
	ErrSessionNotFound = errors.New("session not found")
)
