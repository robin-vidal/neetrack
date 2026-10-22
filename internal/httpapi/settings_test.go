package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robin-vidal/neetrack/internal/store"
)

func TestHandleGetSettings(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.store = fakeStore{
		settings: store.UserSettings{
			TargetRetention: 0.9,
			BlockSize:       4,
			OpenMax:         1,
			NewPct:          40,
			SessionSize:     5,
			IncludeHard:     true,
		},
	}

	req := authenticatedRequest(t, srv, http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body settingsResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
	if body.TargetRetention != 0.9 || body.BlockSize != 4 || body.OpenMax != 1 {
		t.Errorf("body = %+v, want defaults", body)
	}
	if body.TargetDate != nil {
		t.Errorf("TargetDate = %v, want nil", *body.TargetDate)
	}
	if !body.IncludeHard {
		t.Errorf("IncludeHard = false, want true")
	}
}

func TestHandlePatchSettings(t *testing.T) {
	tests := []struct {
		name       string
		initial    store.UserSettings
		patch      map[string]any
		wantStatus int
		check      func(t *testing.T, got store.UserSettings)
	}{
		{
			name:       "updates target retention",
			initial:    store.UserSettings{TargetRetention: 0.9, BlockSize: 4, OpenMax: 1, NewPct: 40, SessionSize: 5},
			patch:      map[string]any{"target_retention": 0.95},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, got store.UserSettings) {
				if got.TargetRetention != 0.95 {
					t.Errorf("TargetRetention = %v, want 0.95", got.TargetRetention)
				}
				if got.BlockSize != 4 {
					t.Errorf("BlockSize = %v, want unchanged 4", got.BlockSize)
				}
			},
		},
		{
			name:       "sets target date",
			initial:    store.UserSettings{TargetRetention: 0.9, BlockSize: 4, OpenMax: 1, NewPct: 40, SessionSize: 5},
			patch:      map[string]any{"target_date": "2026-12-01"},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, got store.UserSettings) {
				if got.TargetDate == nil || !got.TargetDate.Equal(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)) {
					t.Errorf("TargetDate = %v, want 2026-12-01", got.TargetDate)
				}
			},
		},
		{
			name: "clears target date with empty string",
			initial: store.UserSettings{
				TargetRetention: 0.9, TargetDate: timePtr(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)),
				BlockSize: 4, OpenMax: 1, NewPct: 40, SessionSize: 5,
			},
			patch:      map[string]any{"target_date": ""},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, got store.UserSettings) {
				if got.TargetDate != nil {
					t.Errorf("TargetDate = %v, want nil", got.TargetDate)
				}
			},
		},
		{
			name:       "disables hard problems",
			initial:    store.UserSettings{TargetRetention: 0.9, BlockSize: 4, OpenMax: 1, NewPct: 40, SessionSize: 5, IncludeHard: true},
			patch:      map[string]any{"include_hard": false},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, got store.UserSettings) {
				if got.IncludeHard {
					t.Errorf("IncludeHard = true, want false")
				}
			},
		},
		{
			name:       "leaves include hard untouched when absent",
			initial:    store.UserSettings{TargetRetention: 0.9, BlockSize: 4, OpenMax: 1, NewPct: 40, SessionSize: 5, IncludeHard: true},
			patch:      map[string]any{"block_size": 3},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, got store.UserSettings) {
				if !got.IncludeHard {
					t.Errorf("IncludeHard = false, want unchanged true")
				}
			},
		},
		{
			name:       "rejects out of range target retention",
			initial:    store.UserSettings{TargetRetention: 0.9, BlockSize: 4, OpenMax: 1, NewPct: 40, SessionSize: 5},
			patch:      map[string]any{"target_retention": 0.5},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects out of range block size",
			initial:    store.UserSettings{TargetRetention: 0.9, BlockSize: 4, OpenMax: 1, NewPct: 40, SessionSize: 5},
			patch:      map[string]any{"block_size": 10},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects invalid target date format",
			initial:    store.UserSettings{TargetRetention: 0.9, BlockSize: 4, OpenMax: 1, NewPct: 40, SessionSize: 5},
			patch:      map[string]any{"target_date": "not-a-date"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			saved := tt.initial
			srv.store = fakeStore{
				settings: tt.initial,
				updateSettings: func(_ context.Context, _ int64, s store.UserSettings) error {
					saved = s
					return nil
				},
			}

			body, err := json.Marshal(tt.patch)
			if err != nil {
				t.Fatalf("marshaling patch: %v", err)
			}

			req := authenticatedRequest(t, srv, http.MethodPatch, "/api/settings", body)
			rec := httptest.NewRecorder()
			srv.Routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.check != nil {
				tt.check(t, saved)
			}
		})
	}
}

func TestHandleSettingsUnauthenticated(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
