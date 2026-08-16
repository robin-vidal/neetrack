// Package config parses the process environment into a validated Config.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

// Config holds every runtime setting. It is built once in main and passed
// explicitly, never read from the environment again elsewhere.
type Config struct {
	DatabaseURL  string
	ListenAddr   string
	CookieSecure bool
	LogLevel     slog.Level
}

// Load reads DATABASE_URL, LISTEN_ADDR, COOKIE_SECURE and LOG_LEVEL from the
// environment and fails with an explicit error if any is missing or malformed.
func Load() (Config, error) {
	databaseURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}

	listenAddr, err := requireEnv("LISTEN_ADDR")
	if err != nil {
		return Config{}, err
	}

	cookieSecureRaw, err := requireEnv("COOKIE_SECURE")
	if err != nil {
		return Config{}, err
	}
	cookieSecure, err := strconv.ParseBool(cookieSecureRaw)
	if err != nil {
		return Config{}, fmt.Errorf("parsing COOKIE_SECURE=%q: %w", cookieSecureRaw, err)
	}

	logLevelRaw, err := requireEnv("LOG_LEVEL")
	if err != nil {
		return Config{}, err
	}
	logLevel, err := parseLogLevel(logLevelRaw)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DatabaseURL:  databaseURL,
		ListenAddr:   listenAddr,
		CookieSecure: cookieSecure,
		LogLevel:     logLevel,
	}, nil
}

func requireEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return "", fmt.Errorf("missing required environment variable %s", name)
	}
	return value, nil
}

func parseLogLevel(raw string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return 0, fmt.Errorf("parsing LOG_LEVEL=%q: %w", raw, err)
	}
	return level, nil
}
