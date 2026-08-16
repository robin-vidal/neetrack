package config

import (
	"log/slog"
	"testing"
)

func setEnv(t *testing.T, values map[string]string) {
	t.Helper()
	for key, value := range values {
		t.Setenv(key, value)
	}
}

func TestLoad(t *testing.T) {
	valid := map[string]string{
		"DATABASE_URL":  "postgres://user:pass@localhost:5432/neetrack",
		"LISTEN_ADDR":   ":8080",
		"COOKIE_SECURE": "true",
		"LOG_LEVEL":     "INFO",
	}

	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
	}{
		{
			name: "all set",
			env:  valid,
		},
		{
			name: "missing database url",
			env: map[string]string{
				"LISTEN_ADDR":   valid["LISTEN_ADDR"],
				"COOKIE_SECURE": valid["COOKIE_SECURE"],
				"LOG_LEVEL":     valid["LOG_LEVEL"],
			},
			wantErr: true,
		},
		{
			name: "missing listen addr",
			env: map[string]string{
				"DATABASE_URL":  valid["DATABASE_URL"],
				"COOKIE_SECURE": valid["COOKIE_SECURE"],
				"LOG_LEVEL":     valid["LOG_LEVEL"],
			},
			wantErr: true,
		},
		{
			name: "invalid cookie secure",
			env: map[string]string{
				"DATABASE_URL":  valid["DATABASE_URL"],
				"LISTEN_ADDR":   valid["LISTEN_ADDR"],
				"COOKIE_SECURE": "not-a-bool",
				"LOG_LEVEL":     valid["LOG_LEVEL"],
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			env: map[string]string{
				"DATABASE_URL":  valid["DATABASE_URL"],
				"LISTEN_ADDR":   valid["LISTEN_ADDR"],
				"COOKIE_SECURE": valid["COOKIE_SECURE"],
				"LOG_LEVEL":     "not-a-level",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.env)

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.DatabaseURL != valid["DATABASE_URL"] {
				t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, valid["DATABASE_URL"])
			}
			if cfg.ListenAddr != valid["LISTEN_ADDR"] {
				t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, valid["LISTEN_ADDR"])
			}
			if !cfg.CookieSecure {
				t.Errorf("CookieSecure = false, want true")
			}
			if cfg.LogLevel != slog.LevelInfo {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
			}
		})
	}
}
