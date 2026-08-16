package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

// HashPassword derives an argon2id hash of password, PHC-encoded with a
// random salt.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword reports whether password matches a PHC-encoded argon2id
// hash produced by HashPassword, comparing in constant time.
func VerifyPassword(encoded, password string) (bool, error) {
	version, memory, iterations, threads, salt, hash, err := decodeHash(encoded)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, fmt.Errorf("unsupported argon2 version %d", version)
	}

	candidate := argon2.IDKey([]byte(password), salt, iterations, memory, threads, uint32(len(hash)))

	return subtle.ConstantTimeCompare(hash, candidate) == 1, nil
}

func decodeHash(encoded string) (version int, memory, iterations uint32, threads uint8, salt, hash []byte, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return 0, 0, 0, 0, nil, nil, fmt.Errorf("invalid argon2id hash format")
	}

	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return 0, 0, 0, 0, nil, nil, fmt.Errorf("parsing argon2id version: %w", err)
	}

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads); err != nil {
		return 0, 0, 0, 0, nil, nil, fmt.Errorf("parsing argon2id params: %w", err)
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return 0, 0, 0, 0, nil, nil, fmt.Errorf("decoding argon2id salt: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, 0, nil, nil, fmt.Errorf("decoding argon2id hash: %w", err)
	}

	return version, memory, iterations, threads, salt, hash, nil
}
