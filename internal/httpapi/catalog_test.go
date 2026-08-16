package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/scheduler"
)

func TestHandleCatalog(t *testing.T) {
	srv, _ := newTestServer(t)

	now := time.Now()
	dueCard := scheduler.Card{Pattern: "Arrays & Hashing"}
	dueCard.Stability = 5
	dueCard.State = fsrs.Review
	dueCard.LastReview = now.AddDate(0, 0, -30)
	dueCard.Due = now.AddDate(0, 0, -1)

	srv.store = fakeStore{
		cards:      []scheduler.Card{dueCard},
		doneCounts: map[string]int{"Arrays & Hashing": 3},
	}

	loginRec := doLogin(t, srv, "robin", "s3cret")
	cookies := loginRec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("login cookies = %v, want exactly one", cookies)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/catalog", nil)
	req.AddCookie(cookies[0])
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body catalogResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}

	if len(body.Patterns) != 18 {
		t.Fatalf("len(Patterns) = %d, want 18", len(body.Patterns))
	}

	var arraysHashing, twoPointers *patternProgress
	for i := range body.Patterns {
		switch body.Patterns[i].Name {
		case "Arrays & Hashing":
			arraysHashing = &body.Patterns[i]
		case "Two Pointers":
			twoPointers = &body.Patterns[i]
		}
	}
	if arraysHashing == nil || twoPointers == nil {
		t.Fatalf("missing expected patterns in response: %+v", body.Patterns)
	}

	if arraysHashing.Total != 9 {
		t.Errorf("Arrays & Hashing Total = %d, want 9", arraysHashing.Total)
	}
	if arraysHashing.Done != 3 {
		t.Errorf("Arrays & Hashing Done = %d, want 3", arraysHashing.Done)
	}
	if !arraysHashing.Due {
		t.Errorf("Arrays & Hashing Due = false, want true (card is overdue)")
	}
	if arraysHashing.Retrievability == nil {
		t.Errorf("Arrays & Hashing Retrievability = nil, want a value (card exists)")
	}

	if twoPointers.Done != 0 {
		t.Errorf("Two Pointers Done = %d, want 0", twoPointers.Done)
	}
	if twoPointers.Due {
		t.Errorf("Two Pointers Due = true, want false (no card)")
	}
	if twoPointers.Retrievability != nil {
		t.Errorf("Two Pointers Retrievability = %v, want nil (no card)", *twoPointers.Retrievability)
	}
}

func TestHandleCatalogUnauthenticated(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/catalog", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
