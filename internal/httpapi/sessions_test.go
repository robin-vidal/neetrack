package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robin-vidal/neetrack/internal/catalog"
	"github.com/robin-vidal/neetrack/internal/store"
)

func TestHandleComposeSession(t *testing.T) {
	srv, _ := newTestServer(t)

	loginRec := doLogin(t, srv, "robin", "s3cret")
	cookies := loginRec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("login cookies = %v, want exactly one", cookies)
	}

	body, err := json.Marshal(composeSessionRequest{Size: 3, NewPct: 40})
	if err != nil {
		t.Fatalf("marshaling request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
	req.AddCookie(cookies[0])
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var items []sessionItem
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}

	if len(items) == 0 {
		t.Fatalf("expected at least one session item for a blank user")
	}
	for _, it := range items {
		if it.Reason != "new" {
			t.Errorf("item %q has reason %q, want %q (no cards exist yet)", it.Slug, it.Reason, "new")
		}
		if it.Retrievability != nil {
			t.Errorf("item %q has retrievability %v, want nil for a new problem", it.Slug, *it.Retrievability)
		}
	}
}

func TestHandleComposeSessionIncludeHard(t *testing.T) {
	tests := []struct {
		name        string
		includeHard bool
		wantHard    bool
	}{
		{name: "hard problems scheduled when enabled", includeHard: true, wantHard: true},
		{name: "hard problems skipped when disabled", includeHard: false, wantHard: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			srv.store = fakeStore{
				settings: store.UserSettings{
					TargetRetention: 0.9,
					BlockSize:       6,
					OpenMax:         3,
					NewPct:          100,
					SessionSize:     20,
					IncludeHard:     tt.includeHard,
				},
			}

			body, err := json.Marshal(composeSessionRequest{Size: 20, NewPct: 100})
			if err != nil {
				t.Fatalf("marshaling request: %v", err)
			}

			req := authenticatedRequest(t, srv, http.MethodPost, "/api/sessions", body)
			rec := httptest.NewRecorder()
			srv.Routes().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			var items []sessionItem
			if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
				t.Fatalf("decoding response body: %v", err)
			}

			gotHard := false
			for _, it := range items {
				if it.Difficulty == string(catalog.Hard) {
					gotHard = true
				}
			}
			if gotHard != tt.wantHard {
				t.Errorf("session contains hard = %v, want %v", gotHard, tt.wantHard)
			}
		})
	}
}

func TestHandleComposeSessionUnauthenticated(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandleComposeSessionInvalidBody(t *testing.T) {
	srv, _ := newTestServer(t)

	loginRec := doLogin(t, srv, "robin", "s3cret")
	cookies := loginRec.Result().Cookies()

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader([]byte("not json")))
	req.AddCookie(cookies[0])
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
