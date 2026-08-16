package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireAuth(t *testing.T) {
	store := newFakeStore()
	store.addUser(1, "robin", "unused")
	svc := NewService(store, store, false)

	token, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken() error: %v", err)
	}
	if err := store.CreateSession(context.Background(), token.Hash, 1, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	protected := RequireAuth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Errorf("expected user in context")
		}
		if user.Name != "robin" {
			t.Errorf("user.Name = %q, want %q", user.Name, "robin")
		}
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("no cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		req.AddCookie(&http.Cookie{Name: CookieName, Value: "garbage"})
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("valid session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		req.AddCookie(&http.Cookie{Name: CookieName, Value: token.Raw})
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}
