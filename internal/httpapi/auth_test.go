package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/store"
)

type memoryAuthStore struct {
	mu       sync.Mutex
	users    map[string]auth.Credentials
	names    map[int64]string
	sessions map[string]auth.SessionRecord
}

func newMemoryAuthStore() *memoryAuthStore {
	return &memoryAuthStore{
		users:    map[string]auth.Credentials{},
		names:    map[int64]string{},
		sessions: map[string]auth.SessionRecord{},
	}
}

func (m *memoryAuthStore) addUser(id int64, name, passwordHash string) {
	m.users[name] = auth.Credentials{UserID: id, PasswordHash: passwordHash}
	m.names[id] = name
}

func (m *memoryAuthStore) UserByName(_ context.Context, name string) (auth.Credentials, error) {
	c, ok := m.users[name]
	if !ok {
		return auth.Credentials{}, auth.ErrUserNotFound
	}
	return c, nil
}

func (m *memoryAuthStore) CreateSession(_ context.Context, hash []byte, userID int64, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[string(hash)] = auth.SessionRecord{UserID: userID, UserName: m.names[userID], ExpiresAt: expiresAt}
	return nil
}

func (m *memoryAuthStore) SessionUser(_ context.Context, hash []byte) (auth.SessionRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.sessions[string(hash)]
	if !ok {
		return auth.SessionRecord{}, auth.ErrSessionNotFound
	}
	return rec, nil
}

func (m *memoryAuthStore) ExtendSession(_ context.Context, hash []byte, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.sessions[string(hash)]
	if !ok {
		return auth.ErrSessionNotFound
	}
	rec.ExpiresAt = expiresAt
	m.sessions[string(hash)] = rec
	return nil
}

func (m *memoryAuthStore) DeleteSession(_ context.Context, hash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, string(hash))
	return nil
}

func newTestServer(t *testing.T) (*Server, *memoryAuthStore) {
	t.Helper()
	authStore := newMemoryAuthStore()
	hash, err := auth.HashPassword("s3cret")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	authStore.addUser(1, "robin", hash)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	authService := auth.NewService(authStore, authStore, false)

	dataStore := fakeStore{
		settings: store.UserSettings{
			TargetRetention: 0.90,
			BlockSize:       4,
			OpenMax:         1,
			NewPct:          40,
			SessionSize:     5,
		},
	}
	return New(logger, dataStore, authService, testCatalog(t)), authStore
}

func doLogin(t *testing.T, srv *Server, name, password string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(loginRequest{Name: name, Password: password})
	if err != nil {
		t.Fatalf("marshaling login body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	return rec
}

func TestHandleLogin(t *testing.T) {
	tests := []struct {
		name       string
		user       string
		password   string
		wantStatus int
		wantCode   string
	}{
		{name: "correct credentials", user: "robin", password: "s3cret", wantStatus: http.StatusNoContent},
		{name: "wrong password", user: "robin", password: "wrong", wantStatus: http.StatusUnauthorized, wantCode: "invalid_credentials"},
		{name: "unknown user", user: "ghost", password: "whatever", wantStatus: http.StatusUnauthorized, wantCode: "invalid_credentials"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			rec := doLogin(t, srv, tt.user, tt.password)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusNoContent {
				cookies := rec.Result().Cookies()
				if len(cookies) != 1 || cookies[0].Name != auth.CookieName {
					t.Fatalf("cookies = %v, want a single %q cookie", cookies, auth.CookieName)
				}
				return
			}

			var body errorResponse
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decoding response body: %v", err)
			}
			if body.Error != tt.wantCode {
				t.Errorf("error code = %q, want %q", body.Error, tt.wantCode)
			}
		})
	}
}

func TestHandleLoginRateLimit(t *testing.T) {
	srv, _ := newTestServer(t)

	var last *httptest.ResponseRecorder
	for range loginRateLimit + 1 {
		last = doLogin(t, srv, "robin", "wrong")
	}

	if last.Code != http.StatusTooManyRequests {
		t.Errorf("status after exceeding quota = %d, want %d", last.Code, http.StatusTooManyRequests)
	}
}

func TestHandleMe(t *testing.T) {
	srv, _ := newTestServer(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		rec := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("authenticated", func(t *testing.T) {
		loginRec := doLogin(t, srv, "robin", "s3cret")
		cookies := loginRec.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("login cookies = %v, want exactly one", cookies)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		req.AddCookie(cookies[0])
		rec := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var body meResponse
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decoding response body: %v", err)
		}
		if body.Name != "robin" {
			t.Errorf("name = %q, want %q", body.Name, "robin")
		}
	})
}

func TestHandleLogout(t *testing.T) {
	srv, _ := newTestServer(t)

	loginRec := doLogin(t, srv, "robin", "s3cret")
	cookies := loginRec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("login cookies = %v, want exactly one", cookies)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	req.AddCookie(cookies[0])
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.AddCookie(cookies[0])
	meRec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(meRec, meReq)

	if meRec.Code != http.StatusUnauthorized {
		t.Errorf("status after logout = %d, want %d", meRec.Code, http.StatusUnauthorized)
	}
}
