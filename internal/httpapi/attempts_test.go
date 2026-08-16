package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/robin-vidal/neetrack/internal/scheduler"
	"github.com/robin-vidal/neetrack/internal/store"
)

func authenticatedRequest(t *testing.T, srv *Server, method, path string, body []byte) *http.Request {
	t.Helper()
	loginRec := doLogin(t, srv, "robin", "s3cret")
	cookies := loginRec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("login cookies = %v, want exactly one", cookies)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.AddCookie(cookies[0])
	return req
}

func TestHandleCreateAttempt(t *testing.T) {
	srv, _ := newTestServer(t)

	wantDue := time.Now().AddDate(0, 0, 5)
	srv.store = fakeStore{
		recordAttempt: func(ctx context.Context, p store.AttemptParams) (store.AttemptResult, error) {
			if p.Problem.Slug != "two-sum" {
				t.Errorf("Problem.Slug = %q, want %q", p.Problem.Slug, "two-sum")
			}
			return store.AttemptResult{
				Rating:       fsrs.Easy,
				Pattern:      "Arrays & Hashing",
				NextDueAt:    wantDue,
				IntervalDays: 5,
			}, nil
		},
	}

	body, err := json.Marshal(createAttemptRequest{ProblemSlug: "two-sum", DurationSeconds: intPtr(300)})
	if err != nil {
		t.Fatalf("marshaling request: %v", err)
	}

	req := authenticatedRequest(t, srv, http.MethodPost, "/api/attempts", body)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp createAttemptResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
	if resp.Rating != int(fsrs.Easy) {
		t.Errorf("Rating = %d, want %d", resp.Rating, fsrs.Easy)
	}
	if resp.Pattern != "Arrays & Hashing" {
		t.Errorf("Pattern = %q, want %q", resp.Pattern, "Arrays & Hashing")
	}
	if resp.IntervalDays != 5 {
		t.Errorf("IntervalDays = %d, want 5", resp.IntervalDays)
	}
}

func TestHandleCreateAttemptUnknownProblem(t *testing.T) {
	srv, _ := newTestServer(t)

	body, err := json.Marshal(createAttemptRequest{ProblemSlug: "does-not-exist"})
	if err != nil {
		t.Fatalf("marshaling request: %v", err)
	}

	req := authenticatedRequest(t, srv, http.MethodPost, "/api/attempts", body)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandleCreateAttemptMissingDurationAndRating(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.store = fakeStore{
		recordAttempt: func(ctx context.Context, p store.AttemptParams) (store.AttemptResult, error) {
			return store.AttemptResult{}, scheduler.ErrRatingRequired
		},
	}

	body, err := json.Marshal(createAttemptRequest{ProblemSlug: "two-sum"})
	if err != nil {
		t.Fatalf("marshaling request: %v", err)
	}

	req := authenticatedRequest(t, srv, http.MethodPost, "/api/attempts", body)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleCreateAttemptUnauthenticated(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/attempts", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func intPtr(v int) *int {
	return &v
}
