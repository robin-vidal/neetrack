package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/stats"
)

func TestHandleStats(t *testing.T) {
	srv, _ := newTestServer(t)
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	srv.store = fakeStore{
		attemptLog: []stats.Attempt{
			{ProblemSlug: "contains-duplicate", Pattern: "Arrays & Hashing", Rating: fsrs.Easy, Duration: 4 * time.Minute, CreatedAt: now},
			{ProblemSlug: "group-anagrams", Pattern: "Arrays & Hashing", Rating: fsrs.Again, Duration: 30 * time.Minute, SawSolution: true, CreatedAt: now},
		},
		problemLinks: map[string]string{"contains-duplicate": "duplicate-integer"},
	}

	req := authenticatedRequest(t, srv, http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body statsResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}

	if body.Overview.Attempts != 2 || body.Overview.TotalSeconds != 34*60 || body.Overview.AverageSeconds != 17*60 {
		t.Errorf("Overview = %+v, want 2 attempts, 2040 s total, 1020 s average", body.Overview)
	}
	if body.Overview.SolutionRate != 0.5 {
		t.Errorf("SolutionRate = %v, want 0.5", body.Overview.SolutionRate)
	}
	if body.Ratings != (statsRatings{Again: 1, Easy: 1}) {
		t.Errorf("Ratings = %+v, want one Again and one Easy", body.Ratings)
	}
	if len(body.ByDifficulty) != 3 || body.ByDifficulty[0].TargetSeconds != 12*60 {
		t.Errorf("ByDifficulty = %+v, want three tiers starting with easy at 720 s", body.ByDifficulty)
	}
	if len(body.Best) != 1 || body.Best[0].Slug != "contains-duplicate" {
		t.Fatalf("Best = %+v, want contains-duplicate only", body.Best)
	}
	if body.Best[0].NeetcodeSlug != "duplicate-integer" {
		t.Errorf("Best[0].NeetcodeSlug = %q, want the user's link override", body.Best[0].NeetcodeSlug)
	}
	if len(body.Worst) != 1 || body.Worst[0].Slug != "group-anagrams" || body.Worst[0].LastRating != 1 {
		t.Errorf("Worst = %+v, want group-anagrams rated 1", body.Worst)
	}
}

func TestHandleStatsEmptyListsAreArrays(t *testing.T) {
	srv, _ := newTestServer(t)

	req := authenticatedRequest(t, srv, http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	for _, key := range []string{`"by_pattern":[]`, `"best":[]`, `"worst":[]`} {
		if !strings.Contains(rec.Body.String(), key) {
			t.Errorf("body = %s, want it to contain %s", rec.Body.String(), key)
		}
	}
}

func TestHandleStatsUnauthenticated(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
