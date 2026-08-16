package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robin-vidal/neetrack/internal/auth"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(ctx context.Context) error {
	return f.err
}

type emptyAuthStore struct{}

func (emptyAuthStore) UserByName(ctx context.Context, name string) (auth.Credentials, error) {
	return auth.Credentials{}, auth.ErrUserNotFound
}

func (emptyAuthStore) CreateSession(ctx context.Context, hash []byte, userID int64, expiresAt time.Time) error {
	return nil
}

func (emptyAuthStore) SessionUser(ctx context.Context, hash []byte) (auth.SessionRecord, error) {
	return auth.SessionRecord{}, auth.ErrSessionNotFound
}

func (emptyAuthStore) ExtendSession(ctx context.Context, hash []byte, expiresAt time.Time) error {
	return nil
}

func (emptyAuthStore) DeleteSession(ctx context.Context, hash []byte) error {
	return nil
}

func newTestAuthService() *auth.Service {
	return auth.NewService(emptyAuthStore{}, emptyAuthStore{}, false)
}

func TestHandleHealth(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "database reachable",
			pingErr:    nil,
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "database unreachable",
			pingErr:    errors.New("connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "database_unreachable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			srv := New(logger, fakePinger{err: tt.pingErr}, newTestAuthService())

			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			rec := httptest.NewRecorder()

			srv.Routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var body map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decoding response body: %v", err)
			}

			found := false
			for _, v := range body {
				if s, ok := v.(string); ok && s == tt.wantBody {
					found = true
				}
			}
			if !found {
				t.Errorf("response body %v does not contain %q", body, tt.wantBody)
			}
		})
	}
}
