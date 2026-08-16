package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robin-vidal/neetrack/internal/store"
)

func TestStreak(t *testing.T) {
	now := time.Date(2026, 3, 10, 15, 0, 0, 0, time.UTC) // a Tuesday

	tests := []struct {
		name   string
		active []string // dates with at least one attempt
		want   int
	}{
		{
			name:   "no activity",
			active: nil,
			want:   0,
		},
		{
			name:   "active today only",
			active: []string{"2026-03-10"},
			want:   1,
		},
		{
			name:   "active today and yesterday",
			active: []string{"2026-03-10", "2026-03-09"},
			want:   2,
		},
		{
			name:   "nothing today, active yesterday keeps the streak alive",
			active: []string{"2026-03-09", "2026-03-08"},
			want:   2,
		},
		{
			name:   "nothing today or yesterday breaks the streak",
			active: []string{"2026-03-08"},
			want:   0,
		},
		{
			name:   "gap in the middle stops the count",
			active: []string{"2026-03-10", "2026-03-09", "2026-03-07"},
			want:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byDate := make(map[string]int)
			for _, d := range tt.active {
				byDate[d] = 1
			}

			if got := streak(byDate, now); got != tt.want {
				t.Errorf("streak() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHandleActivity(t *testing.T) {
	srv, _ := newTestServer(t)

	now := time.Now().UTC()
	today := now.Format(dateLayout)
	yesterday := now.AddDate(0, 0, -1).Format(dateLayout)

	srv.store = fakeStore{
		activityCounts: []store.DayCount{
			{Date: now.AddDate(0, 0, -1), Count: 2},
			{Date: now, Count: 3},
		},
	}

	req := authenticatedRequest(t, srv, http.MethodGet, "/api/activity", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body activityResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}

	if len(body.Days) != 2 {
		t.Fatalf("len(Days) = %d, want 2", len(body.Days))
	}
	if body.Days[0].Date != yesterday || body.Days[0].Count != 2 {
		t.Errorf("Days[0] = %+v, want {%s 2}", body.Days[0], yesterday)
	}
	if body.Days[1].Date != today || body.Days[1].Count != 3 {
		t.Errorf("Days[1] = %+v, want {%s 3}", body.Days[1], today)
	}
	if body.Streak != 2 {
		t.Errorf("Streak = %d, want 2", body.Streak)
	}
}

func TestHandleActivityUnauthenticated(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/activity", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
