package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSPAHandlerServesIndexForRoot(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	spaHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "<html") {
		t.Errorf("body does not look like the placeholder index.html: %s", rec.Body.String())
	}
}

func TestSPAHandlerFallsBackToIndexForUnknownRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/session/some-client-side-route", nil)
	rec := httptest.NewRecorder()

	spaHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "<html") {
		t.Errorf("unknown route did not fall back to index.html: %s", rec.Body.String())
	}
}

func TestSPAHandlerDoesNotSwallowAPIRoutes(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (SPA catch-all must not shadow /api/*)", rec.Code, http.StatusOK)
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Errorf("/api/health returned HTML instead of JSON: %s", rec.Body.String())
	}
}
