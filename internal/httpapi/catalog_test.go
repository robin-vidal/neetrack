package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robin-vidal/neetrack/internal/catalog"
)

func TestHandleCatalog(t *testing.T) {
	srv, _ := newTestServer(t)
	srv.catalog = fakeCatalog{
		patterns: []string{"Arrays & Hashing", "Two Pointers"},
		byPattern: map[string][]catalog.Problem{
			"Arrays & Hashing": {
				{Slug: "two-sum", Title: "Two Sum", Pattern: "Arrays & Hashing", Difficulty: catalog.Easy, Order: 0},
				{Slug: "group-anagrams", Title: "Group Anagrams", Pattern: "Arrays & Hashing", Difficulty: catalog.Medium, Order: 1},
			},
			"Two Pointers": {
				{Slug: "3sum", Title: "3Sum", Pattern: "Two Pointers", Difficulty: catalog.Medium, Order: 100},
			},
		},
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

	if len(body.Patterns) != 2 {
		t.Fatalf("len(Patterns) = %d, want 2", len(body.Patterns))
	}

	want := []patternProgress{
		{Name: "Arrays & Hashing", Total: 2, Done: 0, Due: false, Retrievability: nil},
		{Name: "Two Pointers", Total: 1, Done: 0, Due: false, Retrievability: nil},
	}
	for i, w := range want {
		got := body.Patterns[i]
		if got.Name != w.Name || got.Total != w.Total || got.Done != w.Done || got.Due != w.Due || got.Retrievability != nil {
			t.Errorf("Patterns[%d] = %+v, want %+v", i, got, w)
		}
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
