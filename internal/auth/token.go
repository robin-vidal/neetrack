package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

const (
	// CookieName is the name of the session cookie.
	CookieName = "neetrack_session"

	// SessionDuration is how long a session stays valid without activity.
	SessionDuration = 30 * 24 * time.Hour

	tokenLen = 32
)

// Token is a freshly minted session token: Raw is the cookie value sent to
// the client, Hash is what gets stored in the database.
type Token struct {
	Raw  string
	Hash []byte
}

// NewToken generates a random session token.
func NewToken() (Token, error) {
	raw := make([]byte, tokenLen)
	if _, err := rand.Read(raw); err != nil {
		return Token{}, fmt.Errorf("generating session token: %w", err)
	}

	sum := sha256.Sum256(raw)

	return Token{
		Raw:  base64.RawURLEncoding.EncodeToString(raw),
		Hash: sum[:],
	}, nil
}

// HashToken derives the storage hash of a raw cookie value.
func HashToken(raw string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decoding session token: %w", err)
	}

	sum := sha256.Sum256(decoded)

	return sum[:], nil
}
