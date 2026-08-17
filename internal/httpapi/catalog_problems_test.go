package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlePatternProblems(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.store = fakeStore{
		problemMastery: map[string]int{"two-sum": 3},
		problemLinks:   map[string]string{"contains-duplicate": "custom-contains-duplicate"},
	}

	req := authenticatedRequest(t, srv, http.MethodGet, "/api/catalog/problems?pattern=Arrays+%26+Hashing", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body []patternProblem
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}

	if len(body) != 9 {
		t.Fatalf("len(body) = %d, want 9", len(body))
	}

	var twoSum, containsDuplicate *patternProblem
	for i := range body {
		switch body[i].Slug {
		case "two-sum":
			twoSum = &body[i]
		case "contains-duplicate":
			containsDuplicate = &body[i]
		}
	}
	if twoSum == nil || containsDuplicate == nil {
		t.Fatalf("missing expected problems in response: %+v", body)
	}

	if twoSum.Mastery != 3 {
		t.Errorf("two-sum Mastery = %d, want 3", twoSum.Mastery)
	}
	if containsDuplicate.NeetcodeSlug != "custom-contains-duplicate" {
		t.Errorf("contains-duplicate NeetcodeSlug = %q, want %q", containsDuplicate.NeetcodeSlug, "custom-contains-duplicate")
	}
	if containsDuplicate.Mastery != 0 {
		t.Errorf("contains-duplicate Mastery = %d, want 0 (never attempted)", containsDuplicate.Mastery)
	}
}

func TestHandlePatternProblemsMissingParam(t *testing.T) {
	srv, _ := newTestServer(t)

	req := authenticatedRequest(t, srv, http.MethodGet, "/api/catalog/problems", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlePatternProblemsUnknownPattern(t *testing.T) {
	srv, _ := newTestServer(t)

	req := authenticatedRequest(t, srv, http.MethodGet, "/api/catalog/problems?pattern=does-not-exist", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlePatternProblemsUnauthenticated(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/catalog/problems?pattern=Arrays+%26+Hashing", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
