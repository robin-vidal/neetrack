package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleUpsertLink(t *testing.T) {
	srv, _ := newTestServer(t)

	var gotUserID int64
	var gotSlug, gotNeetcode string
	srv.store = fakeStore{
		upsertLink: func(ctx context.Context, userID int64, problemSlug, neetcodeSlug string) error {
			gotUserID = userID
			gotSlug = problemSlug
			gotNeetcode = neetcodeSlug
			return nil
		},
	}

	body, err := json.Marshal(upsertLinkRequest{NeetcodeSlug: "custom-two-sum"})
	if err != nil {
		t.Fatalf("marshaling request: %v", err)
	}

	req := authenticatedRequest(t, srv, http.MethodPut, "/api/links/two-sum", body)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if gotUserID != 1 {
		t.Errorf("userID = %d, want 1", gotUserID)
	}
	if gotSlug != "two-sum" {
		t.Errorf("slug = %q, want %q", gotSlug, "two-sum")
	}
	if gotNeetcode != "custom-two-sum" {
		t.Errorf("neetcodeSlug = %q, want %q", gotNeetcode, "custom-two-sum")
	}
}

func TestHandleUpsertLinkUnknownProblem(t *testing.T) {
	srv, _ := newTestServer(t)

	body, err := json.Marshal(upsertLinkRequest{NeetcodeSlug: "whatever"})
	if err != nil {
		t.Fatalf("marshaling request: %v", err)
	}

	req := authenticatedRequest(t, srv, http.MethodPut, "/api/links/does-not-exist", body)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandleUpsertLinkEmptySlug(t *testing.T) {
	srv, _ := newTestServer(t)

	body, err := json.Marshal(upsertLinkRequest{NeetcodeSlug: "   "})
	if err != nil {
		t.Fatalf("marshaling request: %v", err)
	}

	req := authenticatedRequest(t, srv, http.MethodPut, "/api/links/two-sum", body)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleUpsertLinkUnauthenticated(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPut, "/api/links/two-sum", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
